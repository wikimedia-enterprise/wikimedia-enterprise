package langid_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"runtime"
	"testing"
	"wikimedia-enterprise/services/event-bridge/libraries/langid"

	"github.com/stretchr/testify/suite"
)

var errDictTest = errors.New("no language identifier for corresponding dbname")
var testId = "aa"
var sitematrixTestUrl = "/w/api.php"

func createLangIdServer() http.Handler {
	router := http.NewServeMux()

	router.HandleFunc(sitematrixTestUrl, func(w http.ResponseWriter, r *http.Request) {
		_, fnm, _, _ := runtime.Caller(0)
		data, _ := os.ReadFile(fmt.Sprintf("%s/sitematrix.json", path.Dir(fnm)))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	})

	return router
}

type dictionaryTestSuite struct {
	suite.Suite
	dict langid.Dictionarer
	ctx  context.Context
	srv  *httptest.Server
}

func (s *dictionaryTestSuite) SetupSuite() {
	s.srv = httptest.NewServer(createLangIdServer())

	dict := &langid.Dictionary{
		URL: "https://en.wikipedia.org",
	}
	dict.URL = s.srv.URL

	s.dict = dict
	s.ctx = context.Background()
}

func (s *dictionaryTestSuite) TearDownSuite() {
	defer s.srv.Close()
}

func (s *dictionaryTestSuite) TestDictionarySuccess() {
	langid, err := s.dict.GetLanguage(s.ctx, "aawiki")
	s.Assert().NoError(err)
	s.Assert().Equal(testId, langid)
}

func (s *dictionaryTestSuite) TestDictionaryFail() {
	langid, err := s.dict.GetLanguage(s.ctx, "some_db")
	s.Assert().Empty(langid)
	s.Assert().Equal(errDictTest, err)
}

func TestDictionary(t *testing.T) {
	suite.Run(t, new(dictionaryTestSuite))
}
