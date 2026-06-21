package repository

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
