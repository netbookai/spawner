package types

// Record is a representation of a DNS record.
type Route53Record struct {
	// provider-specific metadata
	ID           string
	Type         string
	Name         string
	Value        string
	TTLInSeconds int64
	// type-dependent record fields
	Priority int64 // used by MX, SRV, and URI records
}
