package config

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/gocarina/gocsv"
)

// TopicArticleConfig represents the configuration for topic articles.
type TopicArticleConfig struct {
	Location  string   `json:"location"`
	Service   string   `json:"service_name"`
	Namespace string   `json:"namespace,omitempty"`
	Version   []string `json:"version"`
}

// UnmarshalEnvironmentValue deserializes the TopicArticleConfig from a JSON string.
func (c *TopicArticleConfig) UnmarshalEnvironmentValue(data string) error {
	return json.Unmarshal([]byte(data), c)
}

//go:embed projects.json
var projectsConfig []byte

//go:embed languages.json
var languagesConfig []byte

//go:embed namespaces.json
var namespacesConfig []byte

//go:embed namespaces_metadata.json
var namespacesMetadataConfig []byte

//go:embed partitions_v2.csv
var partitionsConfig []byte

//go:embed structured_projects.json
var structuredConfig []byte

var (
	namespaceMap = map[int]string{
		0:  "articles",
		6:  "files",
		10: "templates",
		14: "categories",
	}
)

// TopicTemplate is the template used to generate topics.
const TopicTemplate = "{{.Location}}.{{.Service}}.{{.Project}}{{.Namespace}}-compacted.{{.Version}}"

// ProjectsGetter interface to expose GetProjects method.
type ProjectsGetter interface {
	GetProjects() []string
}

// LanguageGetter interface to expose GetLanguage method.
type LanguageGetter interface {
	GetLanguage(string) string
}

// NamespacesGetter interface to expose GetNamespaces method.
type NamespacesGetter interface {
	GetNamespaces() []int
}

// NamespacesGetter interface to expose GetNamespaces method.
type NamespacesMetadataGetter interface {
	GetNamespacesMetadata() map[int]NamespaceMetadata
}

// PartitionsGetter interface to expose GetPartitions method.
type PartitionsGetter interface {
	GetPartitions(string, int) []int
}

// StructuredProjectsGetter interface to expose GetStructuredProjects method.
type StructuredProjectsGetter interface {
	GetStructuredProjects() []string
}

// ArticleTopicsGetter interface to expose GetArticleTopics method.
type ArticleTopicsGetter interface {
	GetArticleTopics(*TopicArticleConfig) ([]string, error)
}

// API exposes full config API under a single interface.
type API interface {
	ProjectsGetter
	LanguageGetter
	NamespacesGetter
	NamespacesMetadataGetter
	PartitionsGetter
	StructuredProjectsGetter
	ArticleTopicsGetter
}

// Partition struct to represent project partitions configuration.
type Partition struct {
	Identifier string `csv:"identifier"`
	Articles   int    `csv:"articles"`
	Files      int    `csv:"files"`
	Templates  int    `csv:"templates"`
	Categories int    `csv:"categories"`
}

// New creates new instance of configuration.
func New() (API, error) {
	cfg := &Config{
		projects:           []string{},
		languages:          map[string]string{},
		namespaces:         []int{},
		namespacesMetadata: map[int]NamespaceMetadata{},
		partitions:         map[string]map[int][]int{},
		structured:         []string{},
	}

	if err := json.Unmarshal(projectsConfig, &cfg.projects); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(languagesConfig, &cfg.languages); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(namespacesConfig, &cfg.namespaces); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(namespacesMetadataConfig, &cfg.namespacesMetadata); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(structuredConfig, &cfg.structured); err != nil {
		return nil, err
	}

	pts := []*Partition{}

	if err := gocsv.UnmarshalBytes(partitionsConfig, &pts); err != nil {
		return nil, err
	}

	for _, ptn := range pts {
		cfg.partitions[ptn.Identifier] = map[int][]int{}

		pmn := map[int]int{
			0:  ptn.Articles,
			6:  ptn.Files,
			10: ptn.Templates,
			14: ptn.Categories,
		}

		for nid, npt := range pmn {
			for i := 0; i < npt; i++ {
				cfg.partitions[ptn.Identifier][nid] = append(cfg.partitions[ptn.Identifier][nid], i)
			}
		}
	}

	return cfg, nil
}

type NamespaceMetadata struct {
	Description string `json:"description"`
}

// Config struct to do config lookups.
type Config struct {
	projects           []string
	languages          map[string]string
	namespaces         []int
	namespacesMetadata map[int]NamespaceMetadata
	partitions         map[string]map[int][]int
	structured         []string
}

// GetProjects get list of projects form configuration.
func (c *Config) GetProjects() []string {
	return c.projects
}

// GetLanguage lookups language by database name.
func (c *Config) GetLanguage(dbn string) string {
	return c.languages[dbn]
}

// GetNamespaces returns a list of supported namespaces.
func (c *Config) GetNamespaces() []int {
	return c.namespaces
}

// GetNamespaces returns metadata for all supported namespaces.
func (c *Config) GetNamespacesMetadata() map[int]NamespaceMetadata {
	return c.namespacesMetadata
}

// GetPartitions returns a list of partitions for a given project.
func (c *Config) GetPartitions(dtb string, nid int) []int {
	return c.partitions[dtb][nid]
}

// GetStructuredProjects returns a list of structured projects.
func (c *Config) GetStructuredProjects() []string {
	return c.structured
}

// GetArticleTopics dynamically generates topics based on the provided pattern and configuration.
func (c *Config) GetArticleTopics(cfg *TopicArticleConfig) ([]string, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	// Determine namespaces
	var nss []string
	if len(cfg.Namespace) > 0 {
		nss = []string{cfg.Namespace}
	} else {
		// If no namespace is provided, use all available namespaces from the configuration
		for _, ns := range namespaceMap {
			nss = append(nss, ns)
		}
	}

	var topics []string
	tmpl, err := template.New("topic").Parse(TopicTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template pattern: %w", err)
	}

	for _, prj := range c.GetProjects() {
		for _, ns := range nss {
			// Add a hyphen before the namespace to separate it from the project name
			nsWithHyphen := fmt.Sprintf("-%s", ns)
			for _, vrn := range cfg.Version {
				data := map[string]interface{}{
					"Location":  cfg.Location,
					"Service":   cfg.Service,
					"Project":   prj,
					"Namespace": nsWithHyphen,
					"Version":   vrn,
				}

				var buf bytes.Buffer
				if err := tmpl.Execute(&buf, data); err != nil {
					return nil, fmt.Errorf("failed to execute template: %w", err)
				}

				topics = append(topics, buf.String())
			}
		}
	}

	return topics, nil
}

// validateConfig ensures the provided TopicArticleConfig is valid.
func validateConfig(cfg *TopicArticleConfig) error {
	if cfg.Location == "" {
		return fmt.Errorf("location is required")
	}

	if cfg.Service == "" {
		return fmt.Errorf("service is required")
	}

	if len(cfg.Version) == 0 {
		return fmt.Errorf("versions cannot be empty")
	}

	return nil
}
