package text_test

import (
	"context"
	"errors"
	"testing"
	"wikimedia-enterprise/services/structured-data/libraries/text"
	"wikimedia-enterprise/services/structured-data/packages/textprocessor"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
)

type textProcessorMock struct {
	mock.Mock
	textprocessor.TextProcessorClient
}

func (m *textProcessorMock) GetDictionaryWords(_ context.Context, in *textprocessor.DictionaryRequest, _ ...grpc.CallOption) (*textprocessor.DictionaryResponse, error) {
	ags := m.Called(in.Tokens)

	wrs := ags.Get(0).([]string)
	res := &textprocessor.DictionaryResponse{
		DictWords:    wrs,
		NonDictWords: wrs,
	}

	return res, ags.Error(1)
}

func (m *textProcessorMock) GetNonSafeWords(_ context.Context, in *textprocessor.NonSafeRequest, _ ...grpc.CallOption) (*textprocessor.NonSafeResponse, error) {
	ags := m.Called(in.Tokens)
	res := &textprocessor.NonSafeResponse{
		NonSafeWords: ags.Get(0).([]string),
	}

	return res, ags.Error(1)
}

func (m *textProcessorMock) GetInformalWords(_ context.Context, in *textprocessor.InformalRequest, _ ...grpc.CallOption) (*textprocessor.InformalResponse, error) {
	ags := m.Called(in.Tokens)
	res := &textprocessor.InformalResponse{
		InformalWords: ags.Get(0).([]string),
	}

	return res, ags.Error(1)
}

func (m *textProcessorMock) GetUpperCaseWords(_ context.Context, in *textprocessor.UppercaseRequest, _ ...grpc.CallOption) (*textprocessor.UppercaseResponse, error) {
	ags := m.Called(in.Tokens)
	res := &textprocessor.UppercaseResponse{
		UppercaseWords: ags.Get(0).([]string),
	}

	return res, ags.Error(1)
}

type textTestSuite struct {
	suite.Suite
	ctx context.Context
	tpm *textProcessorMock
	txt *text.Text
	tks []string
	dws []string
	nsw []string
	ifw []string
	ucw []string
	dwe error
	nse error
	iwe error
	uce error
}

func (s *textTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.tpm = new(textProcessorMock)

	s.tpm.On("GetDictionaryWords", s.tks).Return(s.dws, s.dwe)
	s.tpm.On("GetNonSafeWords", s.tks).Return(s.nsw, s.nse)
	s.tpm.On("GetInformalWords", s.tks).Return(s.ifw, s.iwe)
	s.tpm.On("GetUpperCaseWords", s.tks).Return(s.ucw, s.uce)

	s.txt = &text.Text{
		Processor: s.tpm,
	}
}

func (s *textTestSuite) TestGetWords() {
	wds, err := s.txt.GetWords(s.ctx, s.tks)

	if s.dwe != nil {
		s.Assert().Equal(s.dwe, err)
	} else if s.nse != nil {
		s.Assert().Equal(s.nse, err)
	} else if s.iwe != nil {
		s.Assert().Equal(s.iwe, err)
	} else if s.uce != nil {
		s.Assert().Equal(s.uce, err)
	} else {
		s.Assert().Equal(s.dws, wds.Dictionary)
		s.Assert().Equal(s.dws, wds.NonDictionary)
		s.Assert().Equal(s.ifw, wds.Informal)
		s.Assert().Equal(s.nsw, wds.NonSafe)
		s.Assert().Equal(s.ucw, wds.UpperCase)
	}
}

func (s *textTestSuite) TestGetWordsPair() {
	wdf, wds, err := s.txt.GetWordsPair(s.ctx, s.tks, s.tks)

	if s.dwe != nil {
		s.Assert().Contains(err.Error(), s.dwe.Error())
	} else if s.nse != nil {
		s.Assert().Contains(err.Error(), s.nse.Error())
	} else if s.iwe != nil {
		s.Assert().Contains(err.Error(), s.iwe.Error())
	} else if s.uce != nil {
		s.Assert().Contains(err.Error(), s.uce.Error())
	} else {
		s.Assert().Equal(s.dws, wds.Dictionary)
		s.Assert().Equal(s.dws, wds.NonDictionary)
		s.Assert().Equal(s.ifw, wds.Informal)
		s.Assert().Equal(s.nsw, wds.NonSafe)
		s.Assert().Equal(s.ucw, wds.UpperCase)
		s.Assert().Equal(s.dws, wdf.Dictionary)
		s.Assert().Equal(s.dws, wdf.NonDictionary)
		s.Assert().Equal(s.ifw, wdf.Informal)
		s.Assert().Equal(s.nsw, wdf.NonSafe)
		s.Assert().Equal(s.ucw, wdf.UpperCase)
	}
}

func TestText(t *testing.T) {
	for _, testCase := range []*textTestSuite{
		{
			tks: []string{"this", "is", "a", "test", "String"},
			dws: []string{"test", "string"},
			ifw: []string{"is", "a"},
			ucw: []string{"String"},
		},
		{
			tks: []string{"this", "is", "a", "test", "String"},
			dws: []string{"test", "string"},
			ifw: []string{"is", "a"},
			ucw: []string{"String"},
			dwe: errors.New("this is an error"),
		},
		{
			tks: []string{"this", "is", "a", "test", "String"},
			dws: []string{"test", "string"},
			ifw: []string{"is", "a"},
			ucw: []string{"String"},
			nse: errors.New("this is an error"),
		},
		{
			tks: []string{"this", "is", "a", "test", "String"},
			dws: []string{"test", "string"},
			ifw: []string{"is", "a"},
			ucw: []string{"String"},
			iwe: errors.New("this is an error"),
		},
		{
			tks: []string{"this", "is", "a", "test", "String"},
			dws: []string{"test", "string"},
			ifw: []string{"is", "a"},
			ucw: []string{"String"},
			uce: errors.New("this is an error"),
		},
	} {
		suite.Run(t, testCase)
	}
}
