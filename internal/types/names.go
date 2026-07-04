package types

import (
	"github.com/oh-tarnished/runtime-go/ulid"
	"github.com/the-protobuf-project/resourcename"
)

// Resource-name helpers (Google AIP-122). The freebusy database stores bare ids
// as primary keys, while the protobuf API addresses records by hierarchical
// resource names such as "promoCodes/{promo_code}". These helpers convert between
// the two forms using github.com/the-protobuf-project/resourcename, so callers
// never hand-assemble or hand-parse name strings.
//
// Each helper struct carries its template on a blank field (detected by the
// braces in its tag) and exposes the id segment(s) as tagged value fields.

type promoCodeName struct {
	_  struct{} `resource:"promoCodes/{promo_code}"`
	ID string   `resource:"promo_code"`
}

type resourceName struct {
	_  struct{} `resource:"resources/{resource}"`
	ID string   `resource:"resource"`
}

type offeringName struct {
	_        struct{} `resource:"resources/{resource}/offerings/{offering}"`
	Resource string   `resource:"resource"`
	Offering string   `resource:"offering"`
}

type propertyName struct {
	_  struct{} `resource:"properties/{property}"`
	ID string   `resource:"property"`
}

type unitName struct {
	_        struct{} `resource:"properties/{property}/units/{unit}"`
	Property string   `resource:"property"`
	Unit     string   `resource:"unit"`
}

type organisationName struct {
	_  struct{} `resource:"organisations/{organisation}"`
	ID string   `resource:"organisation"`
}

type memberName struct {
	_            struct{} `resource:"organisations/{organisation}/members/{member}"`
	Organisation string   `resource:"organisation"`
	Member       string   `resource:"member"`
}

// PromoCodeName builds the resource name "promoCodes/{id}" from a bare id.
func PromoCodeName(id string) (string, error) {
	return resourcename.MarshalResource(&promoCodeName{ID: id})
}

// PromoCodeID extracts the bare id from a "promoCodes/{id}" resource name.
func PromoCodeID(name string) (string, error) {
	var n promoCodeName
	if err := resourcename.UnmarshalResource(name, &n); err != nil {
		return "", err
	}
	return n.ID, nil
}

// ResolvePromoCodeName returns the bare id and full resource name for a write.
// When name is empty a fresh ULID id is generated and formatted into a name;
// otherwise the id is parsed out of the supplied name. Both adapters use it so
// id generation and name parsing stay identical across providers.
func ResolvePromoCodeName(name string) (id, full string, err error) {
	if name == "" {
		id = ulid.GenerateString()
		full, err = PromoCodeName(id)
		return id, full, err
	}
	if id, err = PromoCodeID(name); err != nil {
		return "", "", err
	}
	return id, name, nil
}

// ResourceName builds the resource name "resources/{id}" from a bare id.
func ResourceName(id string) (string, error) {
	return resourcename.MarshalResource(&resourceName{ID: id})
}

// ResourceID extracts the bare id from a "resources/{id}" resource name.
func ResourceID(name string) (string, error) {
	var n resourceName
	if err := resourcename.UnmarshalResource(name, &n); err != nil {
		return "", err
	}
	return n.ID, nil
}

// OfferingName builds "resources/{resource}/offerings/{offering}" from bare ids.
func OfferingName(resourceID, offeringID string) (string, error) {
	return resourcename.MarshalResource(&offeringName{Resource: resourceID, Offering: offeringID})
}

// OfferingID extracts the offering id segment from a
// "resources/{resource}/offerings/{offering}" resource name.
func OfferingID(name string) (string, error) {
	var n offeringName
	if err := resourcename.UnmarshalResource(name, &n); err != nil {
		return "", err
	}
	return n.Offering, nil
}

// PropertyName builds the resource name "properties/{id}" from a bare id.
func PropertyName(id string) (string, error) {
	return resourcename.MarshalResource(&propertyName{ID: id})
}

// PropertyID extracts the bare id from a "properties/{id}" resource name.
func PropertyID(name string) (string, error) {
	var n propertyName
	if err := resourcename.UnmarshalResource(name, &n); err != nil {
		return "", err
	}
	return n.ID, nil
}

// ResolvePropertyName returns the bare id and full resource name for a write.
// When name is empty a fresh ULID id is generated; otherwise the id is parsed
// out of the supplied name.
func ResolvePropertyName(name string) (id, full string, err error) {
	if name == "" {
		id = ulid.GenerateString()
		full, err = PropertyName(id)
		return id, full, err
	}
	if id, err = PropertyID(name); err != nil {
		return "", "", err
	}
	return id, name, nil
}

// UnitName builds "properties/{property}/units/{unit}" from bare ids.
func UnitName(propertyID, unitID string) (string, error) {
	return resourcename.MarshalResource(&unitName{Property: propertyID, Unit: unitID})
}

// UnitID extracts the unit id segment from a
// "properties/{property}/units/{unit}" resource name.
func UnitID(name string) (string, error) {
	var n unitName
	if err := resourcename.UnmarshalResource(name, &n); err != nil {
		return "", err
	}
	return n.Unit, nil
}

// UnitParentID extracts the {property} segment (the parent property id) from a
// "properties/{property}/units/{unit}" resource name.
func UnitParentID(name string) (string, error) {
	var n unitName
	if err := resourcename.UnmarshalResource(name, &n); err != nil {
		return "", err
	}
	return n.Property, nil
}

// ResolveUnitName returns the parent property id, unit id, and full unit
// resource name for a write. When name is set it is parsed; otherwise a fresh
// ULID unit id is minted under the property parsed from parent
// ("properties/{property}"). Both adapters use it so id generation and name
// parsing stay identical across providers.
func ResolveUnitName(parent, name string) (propertyID, unitID, full string, err error) {
	if name != "" {
		var n unitName
		if err = resourcename.UnmarshalResource(name, &n); err != nil {
			return "", "", "", err
		}
		return n.Property, n.Unit, name, nil
	}
	if propertyID, err = PropertyID(parent); err != nil {
		return "", "", "", err
	}
	unitID = ulid.GenerateString()
	full, err = UnitName(propertyID, unitID)
	return propertyID, unitID, full, err
}

// OrganisationName builds the resource name "organisations/{id}" from a bare id.
func OrganisationName(id string) (string, error) {
	return resourcename.MarshalResource(&organisationName{ID: id})
}

// OrganisationID extracts the bare id from an "organisations/{id}" resource name.
func OrganisationID(name string) (string, error) {
	var n organisationName
	if err := resourcename.UnmarshalResource(name, &n); err != nil {
		return "", err
	}
	return n.ID, nil
}

// ResolveOrganisationName returns the bare id and full resource name for a write.
func ResolveOrganisationName(name string) (id, full string, err error) {
	if name == "" {
		id = ulid.GenerateString()
		full, err = OrganisationName(id)
		return id, full, err
	}
	if id, err = OrganisationID(name); err != nil {
		return "", "", err
	}
	return id, name, nil
}

// MemberName builds "organisations/{organisation}/members/{member}" from bare ids.
func MemberName(organisationID, memberID string) (string, error) {
	return resourcename.MarshalResource(&memberName{Organisation: organisationID, Member: memberID})
}

// MemberID extracts the member id segment from a member resource name.
func MemberID(name string) (string, error) {
	var n memberName
	if err := resourcename.UnmarshalResource(name, &n); err != nil {
		return "", err
	}
	return n.Member, nil
}

// ResolveMemberName returns the parent organisation id, member id, and full member
// resource name for a write. When name is set it is parsed; otherwise a fresh ULID
// member id is minted under the organisation parsed from parent.
func ResolveMemberName(parent, name string) (organisationID, memberID, full string, err error) {
	if name != "" {
		var n memberName
		if err = resourcename.UnmarshalResource(name, &n); err != nil {
			return "", "", "", err
		}
		return n.Organisation, n.Member, name, nil
	}
	if organisationID, err = OrganisationID(parent); err != nil {
		return "", "", "", err
	}
	memberID = ulid.GenerateString()
	full, err = MemberName(organisationID, memberID)
	return organisationID, memberID, full, err
}
