package schema

import (
	"encoding/json"
	"errors"
	"fmt"
)

// ErrNamespaceNotSupported appears when when namespace does not have corresponding topic.
var ErrNamespaceNotSupported = errors.New("namespaces is not supported")

// Topics allows to locate topic by namespace.
type Topics struct {
	Version     string `json:"version"`
	ServiceName string `json:"service_name"`
	Location    string `json:"location"`
}

// UnmarshalEnvironmentValue will unmarshal struct if it's being used as environment variable.
func (t *Topics) UnmarshalEnvironmentValue(dta string) error {
	_ = json.Unmarshal([]byte(dta), t)

	if len(t.Version) == 0 {
		t.Version = "v1"
	}

	if len(t.ServiceName) == 0 {
		t.ServiceName = "structured-data"
	}

	if len(t.Location) == 0 {
		t.Location = "aws"
	}

	return nil
}

// GetName returns topic name according to project and namespace.
func (t *Topics) GetName(dtb string, nid int) (string, error) {
	switch nid {
	case 0:
		return fmt.Sprintf("%s.%s.%s-articles-compacted.%s", t.Location, t.ServiceName, dtb, t.Version), nil
	case 6:
		return fmt.Sprintf("%s.%s.%s-files-compacted.%s", t.Location, t.ServiceName, dtb, t.Version), nil
	case 10:
		return fmt.Sprintf("%s.%s.%s-templates-compacted.%s", t.Location, t.ServiceName, dtb, t.Version), nil
	case 14:
		return fmt.Sprintf("%s.%s.%s-categories-compacted.%s", t.Location, t.ServiceName, dtb, t.Version), nil
	default:
		return "", ErrNamespaceNotSupported
	}
}
