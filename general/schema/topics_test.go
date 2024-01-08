package schema

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type topicsTestSuite struct {
	suite.Suite
	cfg string
	nid int
	dtb string
	tpc string
	tps *Topics
	err error
}

func (s *topicsTestSuite) SetupSuite() {
	s.tps = new(Topics)
	s.Assert().NoError(
		s.tps.UnmarshalEnvironmentValue(s.cfg),
	)
}

func (s *topicsTestSuite) TestGetByNamespace() {
	tpc, err := s.tps.GetName(s.dtb, s.nid)

	s.Assert().Equal(s.err, err)
	s.Assert().Equal(s.tpc, tpc)
}

func TestTopics(t *testing.T) {
	for _, testCase := range []*topicsTestSuite{
		{
			cfg: `{}`,
			nid: -1,
			err: ErrNamespaceNotSupported,
		},
		{
			cfg: `{}`,
			dtb: "enwiki",
			nid: 0,
			tpc: "aws.structured-data.enwiki-articles-compacted.v1",
		},
		{
			cfg: `{}`,
			nid: 6,
			dtb: "enwiki",
			tpc: "aws.structured-data.enwiki-files-compacted.v1",
		},
		{
			cfg: `{}`,
			nid: 10,
			dtb: "enwiki",
			tpc: "aws.structured-data.enwiki-templates-compacted.v1",
		},
		{
			cfg: `{}`,
			nid: 14,
			dtb: "enwiki",
			tpc: "aws.structured-data.enwiki-categories-compacted.v1",
		},
		{
			cfg: `{
				"version": "v2",
				"service_name": "new",
				"location": "local"
			}`,
			dtb: "enwiki",
			nid: 0,
			tpc: "local.new.enwiki-articles-compacted.v2",
		},
		{
			cfg: `{
				"version": "v2",
				"service_name": "new",
				"location": "local"
			}`,
			nid: 6,
			dtb: "enwiki",
			tpc: "local.new.enwiki-files-compacted.v2",
		},
		{
			cfg: `{
				"version": "v2",
				"service_name": "new",
				"location": "local"
			}`,
			nid: 10,
			dtb: "enwiki",
			tpc: "local.new.enwiki-templates-compacted.v2",
		},
		{
			cfg: `{
				"version": "v2",
				"service_name": "new",
				"location": "local"
			}`,
			nid: 14,
			dtb: "enwiki",
			tpc: "local.new.enwiki-categories-compacted.v2",
		},
	} {
		suite.Run(t, testCase)
	}
}
