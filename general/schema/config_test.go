package schema

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type configTestSuite struct {
	suite.Suite
	cfg  *Config
	refs []*Config
}

func (s *configTestSuite) TestResolveReferences() {
	s.Assert().Equal(s.refs, s.cfg.ResolveReferences())
}

func TestConfig(t *testing.T) {
	for _, testCase := range []*configTestSuite{
		{
			refs: []*Config{
				ConfigEvent,
				ConfigLanguage,
			},
			cfg: ConfigLanguage,
		},
		{
			refs: []*Config{
				ConfigEvent,
				ConfigEditor,
				ConfigProbabilityScore,
				ConfigScores,
				ConfigSize,
				ConfigDelta,
				ConfigSize,
				ConfigDiff,
				ConfigMaintenanceTags,
				ConfigVersion,
			},
			cfg: ConfigVersion,
		},
	} {
		suite.Run(t, testCase)
	}
}
