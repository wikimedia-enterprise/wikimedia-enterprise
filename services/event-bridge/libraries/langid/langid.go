package langid

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/protsack-stephan/mediawiki-api-client"
)

// ErrNoCorrespondingLanguage if there is no identifier for corresponding dbname
var ErrNoCorrespondingLanguage = errors.New("no language identifier for corresponding dbname")

// Dictionarer is an interface that wraps GetLanguage method for tests
type Dictionarer interface {
	GetLanguage(ctx context.Context, dbname string) (string, error)
}

// NewDictionary creates new language dictionary instance
func NewDictionary() Dictionarer {
	return &Dictionary{
		URL: "https://en.wikipedia.org",
	}
}

// Represents dictionary structure
type Dictionary struct {
	URL    string
	lookup map[string]string
}

func (d *Dictionary) getSiteMatrix(ctx context.Context) error {
	smx, err := mediawiki.
		NewClient(d.URL).
		Sitematrix(ctx)

	if err != nil {
		return err
	}

	for _, prj := range smx.Projects {
		for _, ste := range prj.Site {
			d.lookup[ste.DBName] = prj.Code
		}
	}

	return nil
}

func (d *Dictionary) getFallback(ctx context.Context) error {
	res := struct {
		Sitematrix map[string]mediawiki.Project `json:"sitematrix"`
	}{}

	_, fnm, _, ok := runtime.Caller(0)

	if !ok {
		return errors.New("caller not found")
	}

	data, err := os.ReadFile(fmt.Sprintf("%s/sitematrix.json", path.Dir(fnm)))

	if err != nil {
		return err
	}

	_ = json.Unmarshal(data, &res)

	for num, prj := range res.Sitematrix {
		if num != "count" && num != "specials" {
			for _, ste := range prj.Site {
				d.lookup[ste.DBName] = prj.Code
			}
		}
	}
	return nil
}

// GetLanguage returns language identifier for corresponding database
func (d *Dictionary) GetLanguage(ctx context.Context, dbname string) (string, error) {
	if len(d.lookup) == 0 {
		d.lookup = map[string]string{}

		if err := d.getSiteMatrix(ctx); err != nil {
			if err := d.getFallback(ctx); err != nil {
				return "", err
			}
		}
	}

	if val, ok := d.lookup[dbname]; ok {
		return val, nil
	}

	return "", ErrNoCorrespondingLanguage
}
