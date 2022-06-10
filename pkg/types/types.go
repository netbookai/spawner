package types

type ResourceRecordValue struct {
	Value string
}

// Record is a representation of a DNS record.
type Route53ResourceRecordSet struct {
	// provider-specific metadata
	Type         string
	Name         string
	TTLInSeconds int64
	// type-dependent record fields
	ResourceRecords []ResourceRecordValue
}
