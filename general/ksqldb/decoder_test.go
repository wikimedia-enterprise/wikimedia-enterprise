package ksqldb

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type decoderTestVersion struct {
	Identifier       int      `avro:"identifier"`
	ParentIdentifier int      `avro:"parent_identifier"`
	Comment          string   `avro:"comment"`
	Tags             []string `avro:"tags"`
}

type decoderTestNamespace struct {
	Identifier    int    `avro:"identifier"`
	AlternateName string `avro:"alternate_name"`
	Name          string `avro:"name"`
}

type decoderTestArticle struct {
	TableName    struct{}              `ksql:"articles"`
	Name         string                `avro:"name"`
	Identifier   int                   `avro:"identifier"`
	DateModified *time.Time            `avro:"date_modified"`
	Namespace    *decoderTestNamespace `avro:"namespace"`
	Version      *decoderTestVersion   `avro:"version" ksql:"versions"`
}

type decoderTestSuite struct {
	suite.Suite
	joins   bool
	val     *decoderTestArticle
	result  *decoderTestArticle
	row     Row
	columns []string
	dec     *Decoder
}

func (s *decoderTestSuite) SetupSuite() {
	s.val = new(decoderTestArticle)
	s.dec = NewDecoder(s.row, s.columns, func(d *Decoder) {
		d.HasJoins = s.joins
	})
}

func (s *decoderTestSuite) TestDecode() {
	s.Assert().NoError(s.dec.Decode(s.val))
	s.Assert().Equal(s.result, s.val)
}

func TestDecoder(t *testing.T) {
	dateModifiedMicro := 1.637448825222886e+15
	dateModified := time.UnixMicro(int64(dateModifiedMicro))

	main := new(decoderTestArticle)
	main.Name = "Earth"
	main.Identifier = 100
	main.DateModified = &dateModified
	main.Namespace = new(decoderTestNamespace)
	main.Namespace.Identifier = 99
	main.Namespace.Name = "Articles"
	main.Namespace.AlternateName = "Стаття"
	main.Version = new(decoderTestVersion)
	main.Version.Identifier = 12
	main.Version.ParentIdentifier = 11
	main.Version.Comment = "Awesome comment"
	main.Version.Tags = []string{"hello", "world"}

	fail := new(decoderTestArticle)
	fail.Name = "Earth"
	fail.Identifier = 100
	fail.DateModified = &dateModified
	fail.Namespace = new(decoderTestNamespace)
	fail.Namespace.Identifier = 99
	fail.Namespace.Name = "Articles"
	fail.Namespace.AlternateName = "Стаття"

	for _, testCase := range []*decoderTestSuite{
		{
			result: main,
			row: Row{
				main.Name,
				main.Identifier,
				dateModifiedMicro,
				map[string]interface{}{
					"ALTERNATE_NAME": main.Namespace.AlternateName,
					"IDENTIFIER":     main.Namespace.Identifier,
					"NAME":           main.Namespace.Name,
				},
				map[string]interface{}{
					"IDENTIFIER":        main.Version.Identifier,
					"PARENT_IDENTIFIER": main.Version.ParentIdentifier,
					"COMMENT":           main.Version.Comment,
					"TAGS":              main.Version.Tags,
				},
			},
			columns: []string{"NAME", "IDENTIFIER", "DATE_MODIFIED", "NAMESPACE", "VERSION"},
		},
		{
			result:  main,
			row:     Row{main.Name, main.Identifier, dateModifiedMicro, map[string]interface{}{"ALTERNATE_NAME": main.Namespace.AlternateName, "IDENTIFIER": main.Namespace.Identifier, "NAME": main.Namespace.Name}, main.Version.Identifier, main.Version.ParentIdentifier, main.Version.Comment, []interface{}{main.Version.Tags[0], main.Version.Tags[1]}},
			columns: []string{"ARTICLES_NAME", "ARTICLES_IDENTIFIER", "ARTICLES_DATE_MODIFIED", "ARTICLES_NAMESPACE", "VERSIONS_IDENTIFIER", "VERSIONS_PARENT_IDENTIFIER", "VERSIONS_COMMENT", "VERSIONS_TAGS"},
			joins:   true,
		},
		{
			result:  fail,
			row:     Row{main.Name, main.Identifier, dateModifiedMicro, map[string]interface{}{"ALTERNATE_NAME": main.Namespace.AlternateName, "IDENTIFIER": main.Namespace.Identifier, "NAME": main.Namespace.Name}, main.Version.Identifier, main.Version.ParentIdentifier, main.Version.Comment, []interface{}{main.Version.Tags[0], main.Version.Tags[1]}},
			columns: []string{"ARTICLES_NAME", "ARTICLES_IDENTIFIER", "ARTICLES_DATE_MODIFIED", "ARTICLES_NAMESPACE"},
			joins:   true,
		},
	} {
		suite.Run(t, testCase)
	}
}
