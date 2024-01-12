package httputil

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type userTestSuite struct {
	suite.Suite
	unm string
	ugr []string
	ugg string
}

func (s *userTestSuite) TestGetUsername() {
	usr := &User{Username: s.unm}

	s.Assert().Equal(s.unm, usr.GetUsername())
}

func (s *userTestSuite) TestGetUsernameNil() {
	var usr *User

	s.Assert().Equal("", usr.GetUsername())
}

func (s *userTestSuite) TestSetUsername() {
	usr := new(User)
	usr.SetUsername(s.unm)

	s.Assert().Equal(s.unm, usr.Username)
}

func (s *userTestSuite) TestGetGroups() {
	usr := &User{Groups: s.ugr}

	s.Assert().Equal(s.ugr, usr.GetGroups())
}

func (s *userTestSuite) TestGetGroupsNil() {
	var usr *User

	s.Assert().Equal([]string{}, usr.GetGroups())
}

func (s *userTestSuite) TestSetGroups() {
	usr := new(User)
	usr.SetGroups(s.ugr)

	s.Assert().Equal(s.ugr, usr.Groups)
}

func (s *userTestSuite) TestIsInGroup() {
	usr := new(User)
	usr.SetGroups(s.ugr)

	for _, grp := range s.ugr {
		s.Assert().True(usr.IsInGroup(grp))
	}
}

func (s *userTestSuite) TestIsInGroupNil() {
	var usr *User

	for _, grp := range s.ugr {
		s.Assert().False(usr.IsInGroup(grp))
	}
}

func (s *userTestSuite) TestGetGroupNil() {
	var usr *User

	s.Assert().Empty(usr.GetGroup())
}

func (s *userTestSuite) TestGetGroup() {
	usr := new(User)
	usr.SetGroups(s.ugr)

	s.Assert().Equal(s.ugg, usr.GetGroup())
}

func TestUser(t *testing.T) {
	for _, testCase := range []*userTestSuite{
		{
			unm: "john.doe",
			ugr: []string{"group_2", "group_1"},
			ugg: "group_2",
		},
	} {
		suite.Run(t, testCase)
	}
}
