package schema

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type topicsTestSuite struct {
	suite.Suite
	cfg string
	nid int
	dtb string
	ets []string
	tps *Topics
	evt string
	err error
}

func (s *topicsTestSuite) SetupSuite() {
	s.evt = "Create"
	s.tps = new(Topics)
	s.Assert().NoError(
		s.tps.UnmarshalEnvironmentValue(s.cfg),
	)
}

func (s *topicsTestSuite) TestGetName() {
	tpc, err := s.tps.GetName(s.dtb, s.nid)

	s.Assert().Equal(s.err, err)

	if len(s.ets) > 0 {
		s.Assert().Equal(s.ets[len(s.ets)-1], tpc)
	} else {
		s.Assert().Empty(tpc)
	}
}

func (s *topicsTestSuite) TestGetNameByVersion() {
	tpc, err := s.tps.GetNameByVersion(s.dtb, s.nid, s.tps.Versions[0])

	s.Assert().Equal(s.err, err)

	if len(s.ets) > 0 {
		s.Assert().Equal(s.ets[0], tpc)
	} else {
		s.Assert().Empty(tpc)
	}
}

func (s *topicsTestSuite) TestGetNames() {
	tps, err := s.tps.GetNames(s.dtb, s.nid)

	s.Assert().Equal(s.err, err)
	s.Assert().Equal(s.ets, tps)
}

func (s *topicsTestSuite) TestGetNameByEventType() {
	tpc := s.tps.GetNameByEventType(s.dtb, s.evt, s.tps.Versions[0])

	if s.dtb == "commonswiki" {
		s.Assert().Equal(fmt.Sprintf("%s.%s.commons-%s.%s", s.tps.Location, s.tps.ServiceName, s.evt, s.tps.Versions[0]), tpc)
	} else {
		s.Assert().Equal(fmt.Sprintf("%s.%s.article-%s.%s", s.tps.Location, s.tps.ServiceName, s.evt, s.tps.Versions[0]), tpc)
	}
}

func (s *topicsTestSuite) TestGetNamesByEventType() {
	tpcs := s.tps.GetNamesByEventType(s.dtb, s.evt)

	for _, vrs := range s.tps.Versions {
		if s.dtb == "commonswiki" {
			s.Assert().Contains(tpcs, fmt.Sprintf("%s.%s.commons-%s.%s", s.tps.Location, s.tps.ServiceName, s.evt, vrs))
		} else {
			s.Assert().Contains(tpcs, fmt.Sprintf("%s.%s.article-%s.%s", s.tps.Location, s.tps.ServiceName, s.evt, vrs))
		}
	}
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
			ets: []string{"aws.structured-data.enwiki-articles-compacted.v1"},
		},
		{
			cfg: `{}`,
			nid: 6,
			dtb: "enwiki",

			ets: []string{"aws.structured-data.enwiki-files-compacted.v1"},
		},
		{
			cfg: `{}`,
			nid: 10,
			dtb: "enwiki",
			ets: []string{"aws.structured-data.enwiki-templates-compacted.v1"},
		},
		{
			cfg: `{}`,
			nid: 14,
			dtb: "enwiki",
			ets: []string{"aws.structured-data.enwiki-categories-compacted.v1"},
		},
		{
			cfg: `{
				"version": ["v2"],
				"service_name": "new",
				"location": "local"
			}`,
			dtb: "enwiki",
			nid: 0,
			ets: []string{"local.new.enwiki-articles-compacted.v2"},
		},
		{
			cfg: `{
				"version": ["v2"],
				"service_name": "new",
				"location": "local"
			}`,
			nid: 6,
			dtb: "enwiki",
			ets: []string{"local.new.enwiki-files-compacted.v2"},
		},
		{
			cfg: `{
				"version": ["v2"],
				"service_name": "new",
				"location": "local"
			}`,
			nid: 10,
			dtb: "enwiki",
			ets: []string{"local.new.enwiki-templates-compacted.v2"},
		},
		{
			cfg: `{
				"version": ["v2"],
				"service_name": "new",
				"location": "local"
			}`,
			nid: 14,
			dtb: "enwiki",
			ets: []string{"local.new.enwiki-categories-compacted.v2"},
		},
		{
			cfg: `{
				"version": ["v2"]
			}`,
			nid: 14,
			dtb: "enwiki",
			ets: []string{"aws.structured-data.enwiki-categories-compacted.v2"},
		},
		{
			cfg: `{
				"version": ["v2", "v1"]
			}`,
			nid: 14,
			dtb: "enwiki",
			ets: []string{"aws.structured-data.enwiki-categories-compacted.v2", "aws.structured-data.enwiki-categories-compacted.v1"},
		},
		{
			cfg: `{
				"version": ["v2", "v1"],
				"service_name": "structured-contents"
			}`,
			nid: 0,
			dtb: "enwiki",
			ets: []string{"aws.structured-contents.enwiki-articles-compacted.v2", "aws.structured-contents.enwiki-articles-compacted.v1"},
		},
	} {
		suite.Run(t, testCase)
	}
}
