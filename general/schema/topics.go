package schema

import (
	"encoding/json"
	"errors"
	"fmt"
)

// ErrNamespaceNotSupported appears when when namespace does not have corresponding topic.
var ErrNamespaceNotSupported = errors.New("namespaces is not supported")

// DefaultTopicVersion is a default version for the topics.
const DefaultTopicVersion = "v1"

// Topics allows to locate topic by namespace.
type Topics struct {
	// Versions is a list of versions for the topic.
	// For example: ["v1", "v2", "v3"]
	Versions []string `json:"version"`

	// ServiceName is a name of the service.
	// For example: "structured-data"
	ServiceName string `json:"service_name"`

	// Location is a location of the topic.
	// For example: "aws", "gcp", "azure"
	Location string `json:"location"`
}

// UnmarshalEnvironmentValue will unmarshal struct if it's being used as environment variable.
func (t *Topics) UnmarshalEnvironmentValue(dta string) error {
	_ = json.Unmarshal([]byte(dta), t)

	if len(t.Versions) == 0 {
		t.Versions = []string{DefaultTopicVersion}
	}

	if len(t.ServiceName) == 0 {
		t.ServiceName = "structured-data"
	}

	if len(t.Location) == 0 {
		t.Location = "aws"
	}

	return nil
}

// GetNameByVersion returns topic name according to project, namespace and version.
func (t *Topics) GetNameByVersion(dtb string, nid int, vrs string) (string, error) {
	switch nid {
	case 0:
		return fmt.Sprintf("%s.%s.%s-articles-compacted.%s", t.Location, t.ServiceName, dtb, vrs), nil
	case 6:
		return fmt.Sprintf("%s.%s.%s-files-compacted.%s", t.Location, t.ServiceName, dtb, vrs), nil
	case 10:
		return fmt.Sprintf("%s.%s.%s-templates-compacted.%s", t.Location, t.ServiceName, dtb, vrs), nil
	case 14:
		return fmt.Sprintf("%s.%s.%s-categories-compacted.%s", t.Location, t.ServiceName, dtb, vrs), nil
	default:
		return "", ErrNamespaceNotSupported
	}
}

// GetName returns topic name according to project and namespace for primary version.
func (t *Topics) GetName(dtb string, nid int) (string, error) {
	vrs := DefaultTopicVersion

	if len(t.Versions) > 0 {
		vrs = t.Versions[len(t.Versions)-1]
	}

	return t.GetNameByVersion(dtb, nid, vrs)
}

// GetNames returns topic names according to project and namespace for all versions.
func (t *Topics) GetNames(dtb string, nid int) ([]string, error) {
	nms := []string{}

	for _, vrs := range t.Versions {
		nme, err := t.GetNameByVersion(dtb, nid, vrs)

		if err != nil {
			return nil, err
		}

		nms = append(nms, nme)
	}

	return nms, nil
}
