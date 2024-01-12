package schema

// Available config types.
const (
	ConfigTypeValue = "value"
	ConfigTypeKey   = "key"
)

// Config contains AVRO schema, type, name and
// all schema references for particular entity.
type Config struct {
	Type       string
	Name       string
	Schema     string
	Reflection interface{}
	References []*Config
}

// ResolveReferences recursively resolves all dependencies and
// creates list of all references that are used by configuration.
func (c *Config) ResolveReferences() []*Config {
	refs := []*Config{}

	for _, ref := range c.References {
		refs = append(refs, ref.ResolveReferences()...)
	}

	refs = append(refs, c)

	return refs
}
