package hasura

import (
	"github.com/oh-tarnished/freebusy/internal/database/repository/repox"
	"github.com/oh-tarnished/freebusy/internal/service/dbutil"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/commonql/postaladdressql"
	commonschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/commonql/schemaql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/mediasql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/policiesql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/propertiesql"
	pschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/schemaql"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/property/v1/propertypbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const stateActive = "ACTIVE"

// --- Property graph ----------------------------------------------------------

type propertyGraph struct {
	property propertiesql.CreateInput
	address  *postaladdressql.CreateInput
	policy   *policiesql.CreateInput
	medias   []mediasql.CreateInput
}

func buildPropertyGraph(p *propertypbv1.Property, now time.Time) *propertyGraph {
	g := &propertyGraph{}
	nowStr := dbutil.TsToStr(timestamppb.New(now))
	g.property = propertiesql.CreateInput{
		Organisation: repox.LastSegment(p.GetOrganisation()),
		DisplayName:  p.GetDisplayName(),
		Description:  p.GetDescription(),
		TimeZone:     p.GetTimeZone(),
		Tags:         strSliceToPtrs(p.GetTags()),
		Attributes:   structToJSON(p.GetAttributes()),
		State:        stateActive,
		CreateTime:   nowStr,
		UpdateTime:   nowStr,
	}
	if a := p.GetAddress(); a != nil {
		aID := ulid.GenerateString()
		ci := addressInput(aID, a)
		g.address = &ci
		g.property.AddressId = aID
	}
	if pol := p.GetPolicy(); pol != nil {
		pID := ulid.GenerateString()
		ci := policiesql.CreateInput{
			Id:           pID,
			CheckinTime:  todToStr(pol.GetCheckinTime()),
			CheckoutTime: todToStr(pol.GetCheckoutTime()),
			HouseRules:   strSliceToPtrs(pol.GetHouseRules()),
			Notes:        pol.GetNotes(),
		}
		g.policy = &ci
		g.property.PolicyId = pID
	}
	for _, m := range p.GetMedia() {
		g.medias = append(g.medias, mediasql.CreateInput{
			Id:          ulid.GenerateString(),
			Uri:         m.GetUri(),
			Type:        mediaTypeToStr(m.GetType()),
			Title:       m.GetTitle(),
			Description: m.GetDescription(),
			MimeType:    m.GetMimeType(),
			SortOrder:   m.GetSortOrder(),
			Primary:     m.GetPrimary(),
		})
	}
	return g
}

type propertyParts struct {
	res          *pschema.PropertyProperties
	address      *commonschema.CommonPostalAddress
	policy       *pschema.PropertyPolicies
	medias       []pschema.PropertyMedias
	unitNames    []string
	licenceNames []string
}

func propertyFromParts(p propertyParts) *propertypbv1.Property {
	res := p.res
	out := &propertypbv1.Property{
		Name:         res.Name,
		Organisation: orgName(res.Organisation),
		DisplayName:  res.DisplayName,
		Description:  repox.Deref(res.Description),
		Address:      addressFromModel(p.address),
		TimeZone:     res.TimeZone,
		Policy:       policyFromModel(p.policy),
		Tags:         strPtrsToSlice(res.Tags),
		Attributes:   jsonToStruct(jsonBytes(res.Attributes)),
		State:        propertyStateFromStr(res.State),
		CreateTime:   strToTS(res.CreateTime),
		UpdateTime:   strToTS(res.UpdateTime),
		Etag:         repox.Deref(res.Etag),
		Units:        p.unitNames,
		Licences:     p.licenceNames,
	}
	for i := range p.medias {
		out.Media = append(out.Media, mediaFromModel(&p.medias[i]))
	}
	return out
}

func policyFromModel(p *pschema.PropertyPolicies) *propertypbv1.Policy {
	if p == nil {
		return nil
	}
	return &propertypbv1.Policy{
		CheckinTime:  strToTOD(p.CheckinTime),
		CheckoutTime: strToTOD(p.CheckoutTime),
		HouseRules:   strPtrsToSlice(p.HouseRules),
		Notes:        repox.Deref(p.Notes),
	}
}

func mediaFromModel(m *pschema.PropertyMedias) *propertypbv1.Media {
	return &propertypbv1.Media{
		Uri:         m.Uri,
		Type:        mediaTypeFromStr(m.Type),
		Title:       repox.Deref(m.Title),
		Description: repox.Deref(m.Description),
		MimeType:    repox.Deref(m.MimeType),
		SortOrder:   repox.Deref(m.SortOrder),
		Primary:     repox.Deref(m.Primary),
	}
}

func propertyStateFromStr(s *string) propertypbv1.PropertyState {
	if s == nil || *s == "" {
		return propertypbv1.PropertyState_PROPERTY_STATE_UNSPECIFIED
	}
	return propertypbv1.PropertyState(propertypbv1.PropertyState_value["PROPERTY_STATE_"+*s])
}

// --- Unit graph --------------------------------------------------------------
