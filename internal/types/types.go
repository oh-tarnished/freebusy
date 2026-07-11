// Package types holds the shared, provider-agnostic vocabulary for the freebusy
// API conventions (Google AIP): sentinel errors, list pagination/ordering,
// field-mask handling, and resource-name parsing. Its error sentinels alias the
// generated repox ones and its filter shapes bridge to the generated filterx
// engines, so hand-written and generated repositories are indistinguishable to
// the gRPC layer; beyond those two bridges it stays free of ORM and GraphQL
// concerns.
package types
