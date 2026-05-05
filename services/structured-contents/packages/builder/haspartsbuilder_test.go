package builder_test

import (
	"testing"
	"wikimedia-enterprise/services/structured-contents/packages/builder"
	"wikimedia-enterprise/services/structured-contents/submodules/schema"
	parser "wikimedia-enterprise/services/structured-contents/submodules/structured-content-parser"

	"github.com/stretchr/testify/suite"
)

type haspartsBuilderTestSuite struct {
	suite.Suite
	p       *parser.Part
	builder *builder.HasPartsBuilder
}

// compareImages compares all fields of parser.Image and schema.Image
func (s *haspartsBuilderTestSuite) compareImages(pimgs []*parser.Image, himgs []*schema.Image) {
	for i := range pimgs {
		s.Assert().Equal(pimgs[i].ContentUrl, himgs[i].ContentUrl)
		s.Assert().Equal(pimgs[i].AlternativeText, himgs[i].AlternativeText)
		s.Assert().Equal(pimgs[i].Caption, himgs[i].Caption)
		s.Assert().Equal(pimgs[i].Height, himgs[i].Height)
		s.Assert().Equal(pimgs[i].Width, himgs[i].Width)
	}
}

// compareHasParts recursively compares all fields of parser.Part and schema.HasParts
func (s *haspartsBuilderTestSuite) compareHasParts(p *parser.Part, hpts *schema.HasParts) {
	s.Assert().Equal(p.Name, hpts.Name)
	s.Assert().Equal(p.Type, hpts.Type)
	s.Assert().Equal(p.Value, hpts.Value)
	s.Assert().Equal(p.Values, hpts.Values)

	// Compare Images
	s.Assert().Equal(len(p.Images), len(hpts.Images))
	s.compareImages(p.Images, hpts.Images)

	// Compare Links
	s.Assert().Equal(len(p.Links), len(hpts.Links))
	for i := range p.Links {
		s.Assert().Equal(p.Links[i].URL, hpts.Links[i].URL)
		s.Assert().Equal(p.Links[i].Text, hpts.Links[i].Text)

		// Compare Link Images
		s.Assert().Equal(len(p.Links[i].Images), len(hpts.Links[i].Images))
		s.compareImages(p.Links[i].Images, hpts.Links[i].Images)
	}

	// Recursively compare HasParts
	s.Assert().Equal(len(p.HasParts), len(hpts.HasParts))
	for i := range p.HasParts {
		s.compareHasParts(p.HasParts[i], hpts.HasParts[i])
	}
}

func (s *haspartsBuilderTestSuite) SetupTest() {
	s.builder = builder.NewHasPartsBuilder()
}

func (s *haspartsBuilderTestSuite) TestNewHasPartsBuilder() {
	s.Assert().NotNil(s.builder)

	hpts := s.builder.Build()
	s.Assert().NotNil(hpts)
}

func (s *haspartsBuilderTestSuite) TestFromPart() {
	hpts := s.builder.FromPart(s.p).Build()
	s.compareHasParts(s.p, hpts)
}

func TestHasPartsBuilder(t *testing.T) {
	for _, testCase := range []*haspartsBuilderTestSuite{
		{
			p: &parser.Part{
				Name:   "Part1",
				Type:   "Type1",
				Value:  "Value1",
				Values: []string{"Value1.1", "Value1.2"},
				HasParts: []*parser.Part{
					{
						Name:  "SubPart1",
						Type:  "Type1",
						Value: "Value1",
						HasParts: []*parser.Part{
							{
								Name:  "SubPart1.1",
								Type:  "Type1.1",
								Value: "Value1.1",
							},
							{
								Name:  "SubPart1.2",
								Type:  "Type1.2",
								Value: "Value1.2",
								HasParts: []*parser.Part{
									{
										Name:  "SubPart1.2.1",
										Type:  "Type1.2.1",
										Value: "Value1.2.1",
									},
									{
										Name:  "SubPart1.2.2",
										Type:  "Type1.2.2",
										Value: "Value1.2.2",
									},
								},
							},
						},
					},
				},
				Images: []*parser.Image{{ContentUrl: "image1.jpg"}},
				Links: []*parser.Link{{URL: "link1.com", Text: "Link 1",
					Images: []*parser.Image{
						{ContentUrl: "image1"},
						{Caption: "cation1"},
						{AlternativeText: "text1"},
					}}},
			},
		},
		{
			p: &parser.Part{
				Name:     "",
				Type:     "",
				Value:    "",
				Values:   []string{},
				HasParts: []*parser.Part{},
				Images:   []*parser.Image{},
				Links:    []*parser.Link{},
			},
		},
		{
			p: &parser.Part{
				Name:   "Part1",
				Type:   "Type1",
				Value:  "Value1",
				Values: []string{"Value1.1", "Value1.2"},
				HasParts: []*parser.Part{
					{
						Name:  "SubPart1",
						Type:  "Type1",
						Value: "Value1",
						HasParts: []*parser.Part{
							{
								Name:  "SubPart1.1",
								Type:  "Type1.1",
								Value: "Value1.1",
							},
							{
								Name:  "SubPart1.2",
								Type:  "Type1.2",
								Value: "Value1.2",
								HasParts: []*parser.Part{
									{
										Name:  "SubPart1.2.1",
										Type:  "Type1.2.1",
										Value: "Value1.2.1",
									},
									{
										Name:  "SubPart1.2.2",
										Type:  "Type1.2.2",
										Value: "Value1.2.2",
										HasParts: []*parser.Part{
											{
												Links: []*parser.Link{{URL: "link1.com", Text: "Link 1",
													Images: []*parser.Image{
														{ContentUrl: "image1"},
														{Caption: "cation1"},
														{AlternativeText: "text1"},
													}}},
											},
											{HasParts: []*parser.Part{
												{Values: []string{"a", "b"}},
												{HasParts: []*parser.Part{
													{Images: []*parser.Image{
														{AlternativeText: "imagey"},
													}},
													{Name: "part_z"},
												}},
											}},
										},
									},
								},
							},
						},
					},
				},
				Images: []*parser.Image{{ContentUrl: "image1.jpg"}},
				Links: []*parser.Link{{URL: "link1.com", Text: "Link 1",
					Images: []*parser.Image{
						{ContentUrl: "image1"},
						{Caption: "cation1"},
						{AlternativeText: "text1"},
					}}},
			},
		},
	} {
		suite.Run(t, testCase)
	}
}
