package resolver

// NewSlice creates new slice instance for the resolver.
func NewSlice(str *Struct) *Slice {
	return &Slice{
		Struct: str,
	}
}

// Slice slice representation for the resolver.
// Used for mapping golang slice representation to ksqldb query.
type Slice struct {
	Struct *Struct
}
