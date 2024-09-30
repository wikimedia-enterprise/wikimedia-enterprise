package config

import (
	_ "embed"
	"encoding/json"

	"github.com/gocarina/gocsv"
)

//go:embed projects.json
var projectsConfig []byte

//go:embed languages.json
var languagesConfig []byte

//go:embed namespaces.json
var namespacesConfig []byte

//go:embed partitions_v2.csv
var partitionsConfig []byte

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

// PartitionsGetter interface to expose GetPartitions method.
type PartitionsGetter interface {
	GetPartitions(string, int) []int
}

// StructuredProjectsGetter interface to expose GetStructuredProjects method.
type StructuredProjectsGetter interface {
	GetStructuredProjects() []string
}

// API exposes full config API under a single interface.
type API interface {
	ProjectsGetter
	LanguageGetter
	NamespacesGetter
	PartitionsGetter
	StructuredProjectsGetter
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
		projects:   []string{},
		languages:  map[string]string{},
		namespaces: []int{},
		partitions: map[string]map[int][]int{},
		structured: []string{},
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

// Config struct to do config lookups.
type Config struct {
	projects   []string
	languages  map[string]string
	namespaces []int
	partitions map[string]map[int][]int
	structured []string
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

// GetPartitions returns a list of partitions for a given project.
func (c *Config) GetPartitions(dtb string, nid int) []int {
	return c.partitions[dtb][nid]
}

// GetStructuredProjects returns a lost of structured projects.
func (c *Config) GetStructuredProjects() []string {
	return c.structured
}
