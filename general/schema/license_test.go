package schema

import (
	"testing"

	"github.com/hamba/avro/v2"
	"github.com/stretchr/testify/suite"
)

type licenseTestSuite struct {
	suite.Suite
	license *License
}

func (s *licenseTestSuite) SetupTest() {
	s.license = &License{
		Name:       "Creative Commons Attribution Share Alike 3.0 Unported",
		Identifier: "CC-BY-SA-3.0",
		URL:        "https://creativecommons.org/licenses/by-sa/3.0/",
	}
}

func (s *licenseTestSuite) TestNewLicenseSchema() {
	sch, err := NewLicenseSchema()
	s.Assert().NoError(err)

	data, err := avro.Marshal(sch, s.license)
	s.Assert().NoError(err)

	license := new(License)
	s.Assert().NoError(avro.Unmarshal(sch, data, license))
	s.Assert().Equal(s.license.Identifier, license.Identifier)
	s.Assert().Equal(s.license.Name, license.Name)
	s.Assert().Equal(s.license.URL, license.URL)
}

func TestLicense(t *testing.T) {
	suite.Run(t, new(licenseTestSuite))
}
