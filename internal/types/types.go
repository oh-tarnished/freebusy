// Package types holds the shared, provider-agnostic vocabulary for the freebusy
// API conventions (Google AIP): sentinel errors, list pagination/ordering,
// field-mask handling, and resource-name parsing. It depends on nothing in the
// project, so every layer — services, the database adapters, future packages —
// can share these without coupling to a concrete ORM or GraphQL client.
package types
