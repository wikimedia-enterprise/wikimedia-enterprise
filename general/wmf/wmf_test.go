package wmf

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const errorResponse = `{
	"error": {
		"code": "badvalue",
		"info": "Unrecognized value for parameter \"apfilterredir\": asdas.",
		"docref": "See https://en.wikipedia.org/w/api.php for API usage. Subscribe to the mediawiki-api-announce mailing list at &lt;https://lists.wikimedia.org/postorius/lists/mediawiki-api-announce.lists.wikimedia.org/&gt; for notice of API deprecations and breaking changes."
	},
	"servedby": "mw1426"
}`

const getPagesResponse = `{
	"batchcomplete":true,
	"query":{
		"pages":[
			{"pageid":9228,"ns":0,"title":"Earth","contentmodel":"wikitext","pagelanguage":"en","pagelanguagehtmlcode":"en","pagelanguagedir":"ltr","touched":"2022-11-07T09:53:09Z","lastrevid":1119965032,"length":190421,"protection":[{"type":"edit","level":"autoconfirmed","expiry":"infinity"},{"type":"move","level":"sysop","expiry":"infinity"}],"restrictiontypes":["edit","move"],"fullurl":"https://en.wikipedia.org/wiki/Earth","editurl":"https://en.wikipedia.org/w/index.php?title=Earth&action=edit","canonicalurl":"https://en.wikipedia.org/wiki/Earth","displaytitle":"Earth","revisions":[{"revid":1119965032,"parentid":1119594067,"minor":true,"user":"Finnusertop","userid":19089174,"timestamp":"2022-11-04T10:47:24Z","slots":{"main":{"contentmodel":"wikitext","contentformat":"text/x-wiki","content":"...wikitext..."}},"comment":"/*References*/cs1","tags":["wikieditor"],"oresscores":{"damaging":{"true":0.011,"false":0.989},"goodfaith":{"true":0.995,"false":0.0050000000000000044},"articlequality":{"Stub":0.752}}}],"wbentityusage":{"Q2":{"aspects":["C","D.en","O","S","T"]}},"pageprops":{"wikibase_item":"Q2"},"redirects":[{"pageid":9215,"ns":0,"title":"EartH"},{"pageid":307601,"ns":0,"title":"Sol3"},{"pageid":603544,"ns":0,"title":"TheEarth"},{"pageid":896072,"ns":0,"title":"Earth(Planet)"},{"pageid":1096589,"ns":0,"title":"Surfaceareaofearth"},{"pageid":1191327,"ns":0,"title":"Terra(planet)"},{"pageid":1324754,"ns":0,"title":"Theplanetearth"},{"pageid":1415438,"ns":0,"title":"Terra(namefortheearth)"},{"pageid":1788541,"ns":0,"title":"LocalPlanet"},{"pageid":2237401,"ns":0,"title":"ThirdPlanet"},{"pageid":2742548,"ns":0,"title":"Globe(Earth)"},{"pageid":3520701,"ns":0,"title":"ThirdplanetfromtheSun"},{"pageid":3601947,"ns":0,"title":"Tellus(Planet)"},{"pageid":4476832,"ns":0,"title":"SolIII"},{"pageid":5222588,"ns":0,"title":"Planetearth"},{"pageid":5423363,"ns":0,"title":"World(geography)"},{"pageid":8678510,"ns":0,"title":"Earth(planet)"},{"pageid":9090641,"ns":0,"title":"ThePlanetEarth"},{"pageid":9455987,"ns":0,"title":"HomePlanet"},{"pageid":9898684,"ns":0,"title":"Earth(word)"},{"pageid":13181153,"ns":0,"title":"MeandensityoftheEarth"},{"pageid":13935837,"ns":0,"title":"Eareth"},{"pageid":15203095,"ns":0,"title":"Blueandgreenplanet"},{"pageid":15203100,"ns":0,"title":"Greenandblueplanet"},{"pageid":16430764,"ns":0,"title":"Earth’ssurface"},{"pageid":16972296,"ns":0,"title":"Earth'ssurface"},{"pageid":18755374,"ns":0,"title":"PlanetofWater"},{"pageid":19790623,"ns":0,"title":"Sol-3"},{"pageid":20384608,"ns":0,"title":"EARTH"},{"pageid":22759962,"ns":0,"title":"Thirdplanet"},{"pageid":23775266,"ns":0,"title":"Earth'smeandensity"},{"pageid":26366190,"ns":0,"title":"CompositionoftheEarth"},{"pageid":27384837,"ns":0,"title":"Telluris"},{"pageid":27706257,"ns":0,"title":"SolPrime"},{"pageid":28257717,"ns":0,"title":"LexicographyofEarth"},{"pageid":31193038,"ns":0,"title":"Earth,Sol"},{"pageid":33364470,"ns":0,"title":"FormationoftheEarth"},{"pageid":33810062,"ns":0,"title":"Etymologyoftheword\"Earth\""},{"pageid":35531228,"ns":0,"title":"SurfaceoftheEarth"},{"pageid":43507855,"ns":0,"title":"Tierra(planet)"},{"pageid":43822591,"ns":0,"title":"3rdplanet"},{"pageid":47103485,"ns":0,"title":"806.4616.0110"},{"pageid":48120239,"ns":0,"title":"PlanetTerra"},{"pageid":56078981,"ns":0,"title":"TheplanetEarth"},{"pageid":56079851,"ns":0,"title":"PlanetEarth"},{"pageid":57857323,"ns":0,"title":"Sizeoftheearth"},{"pageid":58783959,"ns":0,"title":"Theearth"},{"pageid":63738768,"ns":0,"title":"Earthsurface"},{"pageid":64715694,"ns":0,"title":"Earth'sdensity"},{"pageid":64715695,"ns":0,"title":"DensityoftheEarth"},{"pageid":66075911,"ns":0,"title":"PlanetThree"},{"pageid":67560020,"ns":0,"title":"ClimateofEarth"},{"pageid":67713162,"ns":118,"title":"Draft:Earth"},{"pageid":68559411,"ns":0,"title":"FormationofEarth"}]},
			{"pageid":46396,"ns":0,"title":"Ninja","contentmodel":"wikitext","pagelanguage":"en","pagelanguagehtmlcode":"en","pagelanguagedir":"ltr","touched":"2022-11-07T21:42:13Z","lastrevid":1117662369,"length":79921,"protection":[{"type":"edit","level":"autoconfirmed","expiry":"infinity"},{"type":"move","level":"autoconfirmed","expiry":"infinity"}],"restrictiontypes":["edit","move"],"fullurl":"https://en.wikipedia.org/wiki/Ninja","editurl":"https://en.wikipedia.org/w/index.php?title=Ninja&action=edit","canonicalurl":"https://en.wikipedia.org/wiki/Ninja","displaytitle":"Ninja","revisions":[{"revid":1117662369,"parentid":1117662080,"minor":false,"user":"ToastforTeddy","userid":44098681,"timestamp":"2022-10-22T23:00:58Z","slots":{"main":{"contentmodel":"wikitext","contentformat":"text/x-wiki","content":"...wikitext..."}},"comment":"Contentsofsectionmovedto==Seealso==","tags":["visualeditor-wikitext"],"oresscores":{"damaging":{"true":0.268,"false":0.732},"goodfaith":{"true":0.923,"false":0.07699999999999996},"articlequality":{"Stub":0.747}}}],"wbentityusage":{"Q7430520":{"aspects":["S"]},"Q9402":{"aspects":["C","D.en","O","S","T"]}},"pageprops":{"wikibase_item":"Q9402"},"redirects":[{"pageid":340047,"ns":0,"title":"Ninjas"},{"pageid":597512,"ns":0,"title":"Ninzya"},{"pageid":617095,"ns":0,"title":"Sinobi"},{"pageid":1799908,"ns":0,"title":"Shinobishōzoku"},{"pageid":3088559,"ns":0,"title":"Nukenin"},{"pageid":3562652,"ns":0,"title":"Historyoftheninja"},{"pageid":3563229,"ns":0,"title":"HistoryoftheNinja"},{"pageid":5032477,"ns":0,"title":"Shinobishozoku"},{"pageid":5110667,"ns":0,"title":"Jonin"},{"pageid":5355512,"ns":0,"title":"ShinobiShozoku"},{"pageid":6796660,"ns":0,"title":"Chunin"},{"pageid":10690516,"ns":0,"title":"Shinobi"},{"pageid":15203846,"ns":0,"title":"忍者"},{"pageid":17425138,"ns":0,"title":"Suppa"},{"pageid":22173937,"ns":0,"title":"NINJA"},{"pageid":22248985,"ns":0,"title":"Shinobishozoku"},{"pageid":24163001,"ns":0,"title":"Shinobi-no-mono"},{"pageid":56504885,"ns":0,"title":"Chūnin"},{"pageid":59779622,"ns":0,"title":"忍び"},{"pageid":62995468,"ns":0,"title":"🥷"},{"pageid":63676735,"ns":0,"title":"🥷🏻"},{"pageid":63676737,"ns":0,"title":"🥷🏼"},{"pageid":63676738,"ns":0,"title":"🥷🏽"},{"pageid":63676740,"ns":0,"title":"🥷🏾"},{"pageid":63676741,"ns":0,"title":"🥷🏿"},{"pageid":70535257,"ns":0,"title":"Goshiki-mai"},{"pageid":70535258,"ns":0,"title":"Goshiki-Mai"}]}
		]
	}
}`

//go:embed testdata/get_languages.json
var getLanguagesResponse string

const getUsersResponse = `{"batchcomplete":true,"query":{"users":[{"userid":2,"name":"AxelBoldt","editcount":43615,"registration":"2001-07-26T14:50:09Z","groups":["sysop","*","user","autoconfirmed"],"groupmemberships":[{"group":"sysop","expiry":"infinity"}],"emailable":true},{"userid":3,"name":"Tobias Hoevekamp","editcount":2321,"registration":"2001-03-26T20:21:05Z","groups":["extendedconfirmed","*","user","autoconfirmed"],"groupmemberships":[{"group":"extendedconfirmed","expiry":"infinity"}],"emailable":false},{"userid":5,"name":"Hoevekam~enwiki","editcount":14,"registration":"2003-07-09T19:21:37Z","groups":["*","user","autoconfirmed"],"groupmemberships":[],"emailable":false}]}}`

const liftWingPayload = `
{
    "output": {
        "prediction": false,
        "probabilities": {
            "true": 0.05083514377474785,
            "false": 0.9491648562252522
        }
    }
}
`

const referenceRiskError = `{"error":"Could not make prediction for revision 41642294 (pt). Reason: parent_revision_missing"}`

const referenceRiskResponse = `{
    "reference_count": 1,
    "survival_ratio": {
        "min": 0.5816893664255616,
        "mean": 0.5816893664255616,
        "median": 0.5816893664255616
    },
    "reference_risk_score": 0.1,
    "references": [
        {
            "url": "http://www.american-pictures.com/genealogy/persons/per00884.htm",
            "domain_name": "american-pictures.com",
            "domain_metadata": {
                "survival_ratio": 0.5816893664255616,
                "page_count": 19,
                "editors_count": 16
            }
        }
    ]
}
`

const referenceNeedResponse = `{"reference_need_score":0.87}`

const referenceNeedError = `{"error": "Could not make prediction for revision 143077740 (pt). Reason: revision_missing"}`

//go:embed testdata/get_contributors_10025_page1.json
var getContributorsResponse10025Page1 string

//go:embed testdata/get_contributors_10025_page2.json
var getContributorsResponse10025Page2 string

//go:embed testdata/get_contributors_100709.json
var getContributorsResponse100709 string

func createActionsAPIServer(status int, payloads ...string) *httptest.Server {
	rtr := http.NewServeMux()

	nextResponse := 0
	rtr.HandleFunc("/w/api.php", func(w http.ResponseWriter, r *http.Request) {
		if nextResponse > len(payloads) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(status)
		_, _ = w.Write([]byte(payloads[nextResponse]))
		nextResponse++
	})

	return httptest.NewServer(rtr)
}

type allPagesTestSuite struct {
	suite.Suite
	ctx context.Context
	srv *httptest.Server
	clt *Client
	hse bool
	dtb string
	pld string
	sts int
	pgs map[string]*Page
}

func mockTracer(context.Context, map[string]string) (func(error, string), context.Context) {
	return func(error, string) {}, context.Background()
}

func (s *allPagesTestSuite) SetupSuite() {
	s.srv = createActionsAPIServer(s.sts, s.pld)
	s.dtb = "enwiki"
	s.ctx = context.Background()
	s.clt = &Client{
		DefaultURL:      s.srv.URL,
		DefaultDatabase: s.dtb,
		HTTPClient:      &http.Client{},
		Tracer:          mockTracer,
	}

	rsp := new(Response)
	_ = json.Unmarshal([]byte(s.pld), rsp)

	s.pgs = map[string]*Page{}

	if rsp.Query != nil {
		for _, pge := range rsp.Query.AllPages {
			s.pgs[pge.Title] = pge
		}
	}
}

func (s *allPagesTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *allPagesTestSuite) TestGetAllPages() {
	cbk := func(pgs []*Page) {
		for _, pge := range pgs {
			s.Assert().Equal(s.pgs[pge.Title], pge)
		}
	}
	err := s.clt.GetAllPages(s.ctx, s.dtb, cbk)

	if s.hse {
		s.Assert().Error(err)
	} else {
		s.Assert().NoError(err)
	}
}

func TestAllPages(t *testing.T) {
	for _, testCase := range []*allPagesTestSuite{
		{
			hse: false,
			pld: `{
				"batchcomplete": true,
				"query": {
					"allpages": [
						{ "pageid": 341265, "ns": 0, "title": "Jungle" },
						{ "pageid": 56461312, "ns": 0, "title": "Jungle-runner" },
						{ "pageid": 18698572, "ns": 0, "title": "Jungle/Drum n bass" },
						{ "pageid": 1487899, "ns": 0, "title": "Jungle2jungle" },
						{ "pageid": 30391179, "ns": 0, "title": "JunglePup" },
						{ "pageid": 18470226, "ns": 0, "title": "Jungle (2000 film)" },
						{ "pageid": 52695829, "ns": 0, "title": "Jungle (2017 film)" },
						{ "pageid": 54000049, "ns": 0, "title": "Jungle (A Boogie wit da Hoodie song)" },
						{ "pageid": 39013520, "ns": 0, "title": "Jungle (Andre Nickatina song)" },
						{ "pageid": 43710655, "ns": 0, "title": "Jungle (Bakufu Slump album)" }
					]
				}
			}`,
			sts: http.StatusOK,
		},
		{
			hse: true,
			pld: errorResponse,
			sts: http.StatusOK,
		},
		{
			hse: true,
			sts: http.StatusInternalServerError,
		},
	} {
		suite.Run(t, testCase)
	}
}

type getPagesTestSuite struct {
	suite.Suite
	ctx context.Context
	srv *httptest.Server
	clt *Client
	hse bool
	dtb string
	pld string
	sts int
	rsp *Response
	tts []string
}

func (s *getPagesTestSuite) SetupSuite() {
	s.srv = createActionsAPIServer(s.sts, s.pld)
	s.dtb = "enwiki"
	s.ctx = context.Background()
	s.clt = &Client{
		DefaultURL:      s.srv.URL,
		DefaultDatabase: s.dtb,
		HTTPClient:      &http.Client{},
		Tracer:          mockTracer,
	}

	if len(s.pld) > 0 {
		s.rsp = new(Response)
		_ = json.Unmarshal([]byte(s.pld), s.rsp)
	}
}

func (s *getPagesTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *getPagesTestSuite) TestGetPages() {
	pgs, err := s.clt.GetPages(s.ctx, s.dtb, s.tts)

	if s.rsp != nil && s.rsp.Query != nil {
		for _, pge := range s.rsp.Query.Pages {
			s.Assert().Equal(pgs[pge.Title], pge)
		}
	}

	if s.hse {
		s.Assert().Error(err)
	} else {
		s.Assert().NoError(err)
	}
}

func TestGetPages(t *testing.T) {
	for _, testCase := range []*getPagesTestSuite{
		{
			tts: []string{"Earth", "Ninja"},
			sts: http.StatusOK,
			pld: getPagesResponse,
		},
		{
			hse: true,
			pld: errorResponse,
			sts: http.StatusOK,
		},
		{
			hse: true,
			sts: http.StatusInternalServerError,
		},
	} {
		suite.Run(t, testCase)
	}
}

type getPageTestSuite struct {
	suite.Suite
	ctx context.Context
	srv *httptest.Server
	clt *Client
	hse bool
	err error
	dtb string
	pld string
	sts int
	rsp *Response
	ttl string
}

func (s *getPageTestSuite) SetupSuite() {
	s.srv = createActionsAPIServer(s.sts, s.pld)
	s.dtb = "enwiki"
	s.ctx = context.Background()
	s.clt = &Client{
		DefaultURL:      s.srv.URL,
		DefaultDatabase: s.dtb,
		HTTPClient:      &http.Client{},
		Tracer:          mockTracer,
	}

	if len(s.pld) > 0 {
		s.rsp = new(Response)
		_ = json.Unmarshal([]byte(s.pld), s.rsp)
	}
}

func (s *getPageTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *getPageTestSuite) TestGetPage() {
	pge, err := s.clt.GetPage(s.ctx, s.dtb, s.ttl)

	if s.rsp != nil && s.rsp.Query != nil && err == nil {
		s.Assert().Equal(*pge, *s.rsp.Query.Pages[0])
	} else {
		s.Assert().Nil(pge)
	}

	if s.hse {
		s.Assert().Error(err)
	} else {
		s.Assert().NoError(err)
	}

	if s.err != nil {
		s.Assert().Equal(err, s.err)
	}
}

func TestGetPage(t *testing.T) {
	for _, testCase := range []*getPageTestSuite{
		{
			sts: http.StatusOK,
			ttl: "Earth",
			pld: getPagesResponse,
		},
		{
			sts: http.StatusOK,
			ttl: "Not Found",
			pld: getPagesResponse,
			err: ErrPageNotFound,
			hse: true,
		},
		{
			ttl: "Earth",
			sts: http.StatusInternalServerError,
			hse: true,
		},
	} {
		suite.Run(t, testCase)
	}
}

type getPagesHTMLTestSuite struct {
	suite.Suite
	ctx context.Context
	srv *httptest.Server
	clt *Client
	dtb string
	pls []string
	tls []string
	ers []bool
	sts []int
}

func (s *getPagesHTMLTestSuite) createServer() {
	rtr := http.NewServeMux()

	for i, ttl := range s.tls {
		func(ttl string, i int) {
			url := fmt.Sprintf("/w/rest.php/v1/page/%s/html", strings.ReplaceAll(ttl, " ", "_"))

			rtr.HandleFunc(url, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(s.sts[i])
				_, _ = w.Write([]byte(s.pls[i]))
			})
		}(ttl, i)
	}

	s.srv = httptest.NewServer(rtr)
}

func (s *getPagesHTMLTestSuite) SetupSuite() {
	s.createServer()

	s.dtb = "enwiki"
	s.ctx = context.Background()
	s.clt = &Client{
		DefaultURL:      s.srv.URL,
		DefaultDatabase: s.dtb,
		HTTPClient:      &http.Client{},
		Tracer:          mockTracer,
	}
}

func (s *getPagesHTMLTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *getPagesHTMLTestSuite) TestGetPagesHTML() {
	phl := s.clt.GetPagesHTML(s.ctx, s.dtb, s.tls, 10)

	for i, ttl := range s.tls {
		phm := phl[ttl]

		if s.ers[i] {
			s.Assert().Error(phm.Error)
			s.Assert().Empty(phm.Content)
		} else {
			s.Assert().NoError(phm.Error)
			s.Assert().Equal(s.pls[i], phm.Content)
		}

		s.Assert().Equal(ttl, phm.Title)
	}

	s.Assert().Equal(len(s.tls), len(phl))
}

func TestGetPagesHTML(t *testing.T) {
	for _, testCase := range []*getPagesHTMLTestSuite{
		{
			sts: []int{http.StatusOK},
			tls: []string{"Earth"},
			pls: []string{"<p>...html goes here...</p>"},
			ers: []bool{false},
		},
		{
			sts: []int{http.StatusNotFound},
			tls: []string{"Not Found"},
			pls: []string{
				`{
					"type":"https://mediawiki.org/wiki/HyperSwitch/errors/not_found",
					"title":"Not found.",
					"method":"get",
					"detail":"Page or revision not found.",
					"uri":"/en.wikipedia.org/v1/page/html/Sadsd"
				}`,
			},
			ers: []bool{true},
		},
		{
			sts: []int{http.StatusOK, http.StatusNotFound},
			tls: []string{"Earth", "Not Found"},
			pls: []string{
				"<p>...html goes here...</p>",
				`{
					"type":"https://mediawiki.org/wiki/HyperSwitch/errors/not_found",
					"title":"Not found.",
					"method":"get",
					"detail":"Page or revision not found.",
					"uri":"/en.wikipedia.org/v1/page/html/Sadsd"
				}`,
			},
			ers: []bool{false, true},
		},
	} {
		suite.Run(t, testCase)
	}
}

type getPageHTMLTestSuite struct {
	suite.Suite
	ctx context.Context
	srv *httptest.Server
	clt *Client
	hse bool
	dtb string
	pld string
	sts int
	ttl string
}

func (s *getPageHTMLTestSuite) createServer() {
	rtr := http.NewServeMux()

	rtr.HandleFunc(fmt.Sprintf("/w/rest.php/v1/page/%s/html", strings.ReplaceAll(s.ttl, " ", "_")), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(s.sts)
		_, _ = w.Write([]byte(s.pld))
	})

	s.srv = httptest.NewServer(rtr)
}

func (s *getPageHTMLTestSuite) SetupSuite() {
	s.createServer()

	s.dtb = "enwiki"
	s.ctx = context.Background()
	s.clt = &Client{
		DefaultURL:      s.srv.URL,
		DefaultDatabase: s.dtb,
		HTTPClient:      &http.Client{},
		Tracer:          mockTracer,
	}
}

func (s *getPageHTMLTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *getPageHTMLTestSuite) TestGetPageHTML() {
	htm, err := s.clt.GetPageHTML(s.ctx, s.dtb, s.ttl)

	if err == nil {
		s.Assert().Equal(s.pld, htm)
	} else {
		s.Assert().Empty(htm)
	}

	if s.hse {
		s.Assert().Error(err)
	} else {
		s.Assert().NoError(err)
	}
}

func TestGetPageHTML(t *testing.T) {
	for _, testCase := range []*getPageHTMLTestSuite{
		{
			sts: http.StatusOK,
			ttl: "Earth",
			pld: "<p>...html goes here...</p>",
		},
		{
			sts: http.StatusFound,
			ttl: "Earth",
			pld: "<p>...html goes here...</p>",
		},
		{
			sts: http.StatusNotFound,
			ttl: "Not Found",
			pld: `{
				"type":"https://mediawiki.org/wiki/HyperSwitch/errors/not_found",
				"title":"Not found.",
				"method":"get",
				"detail":"Page or revision not found.",
				"uri":"/en.wikipedia.org/v1/page/html/Sadsd"
			}`,
			hse: true,
		},
	} {
		suite.Run(t, testCase)
	}
}

type getRevisionHTMLTestSuite struct {
	suite.Suite
	ctx context.Context
	srv *httptest.Server
	clt *Client
	hse bool
	dtb string
	pld string
	sts int
	rid string
}

func (s *getRevisionHTMLTestSuite) createServer() {
	rtr := http.NewServeMux()

	rtr.HandleFunc(fmt.Sprintf("/w/rest.php/v1/page/%s/html", strings.ReplaceAll(s.rid, " ", "_")), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(s.sts)
		_, _ = w.Write([]byte(s.pld))
	})

	s.srv = httptest.NewServer(rtr)
}

func (s *getRevisionHTMLTestSuite) SetupSuite() {
	s.createServer()

	s.dtb = "enwiki"
	s.ctx = context.Background()
	s.clt = &Client{
		DefaultURL:      s.srv.URL,
		DefaultDatabase: s.dtb,
		HTTPClient:      &http.Client{},
		Tracer:          mockTracer,
	}
}

func (s *getRevisionHTMLTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *getRevisionHTMLTestSuite) TestGetRevisionHTML() {
	htm, err := s.clt.GetPageHTML(s.ctx, s.dtb, s.rid)

	if err == nil {
		s.Assert().Equal(s.pld, htm)
	} else {
		s.Assert().Empty(htm)
	}

	if s.hse {
		s.Assert().Error(err)
	} else {
		s.Assert().NoError(err)
	}
}

func TestGetRevisionHTML(t *testing.T) {
	for _, testCase := range []*getRevisionHTMLTestSuite{
		{
			sts: http.StatusOK,
			rid: "1234567",
			pld: "<p>...html goes here...</p>",
		},
		{
			sts: http.StatusFound,
			rid: "2345678",
			pld: "<p>...html goes here...</p>",
		},
		{
			sts: http.StatusNotFound,
			rid: "Not Found",
			pld: `{
				"type":"https://mediawiki.org/wiki/HyperSwitch/errors/not_found",
				"title":"Not found.",
				"method":"get",
				"detail":"Page or revision not found.",
				"uri":"/en.wikipedia.org/v1/page/html/Sadsd"
			}`,
			hse: true,
		},
	} {
		suite.Run(t, testCase)
	}
}

type getPageSummaryTestSuite struct {
	suite.Suite
	ctx context.Context
	srv *httptest.Server
	clt *Client
	hse bool
	dtb string
	sum string
	sts int
	ttl string
}

func (s *getPageSummaryTestSuite) createServer() {
	rtr := http.NewServeMux()

	rtr.HandleFunc(fmt.Sprintf("/api/rest_v1/page/summary/%s", strings.ReplaceAll(s.ttl, " ", "_")), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(s.sts)
		_, _ = w.Write([]byte(s.sum))
	})

	s.srv = httptest.NewServer(rtr)
}

func (s *getPageSummaryTestSuite) SetupSuite() {
	s.createServer()

	s.dtb = "enwiki"
	s.ctx = context.Background()
	s.clt = &Client{
		DefaultURL:      s.srv.URL,
		DefaultDatabase: s.dtb,
		HTTPClient:      &http.Client{},
		Tracer:          mockTracer,
	}
}

func (s *getPageSummaryTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *getPageSummaryTestSuite) TestGetPageSummary() {
	smr, err := s.clt.GetPageSummary(s.ctx, s.dtb, s.ttl)

	if s.hse {
		s.Assert().Error(err)
		s.Assert().Nil(smr)
	} else {
		s.Assert().NoError(err)
		s.Assert().NotNil(smr)
	}
}

func TestGetPageSummary(t *testing.T) {
	for _, testCase := range []*getPageSummaryTestSuite{
		{
			sts: http.StatusOK,
			ttl: "Test",
			sum: `{"type":"disambiguation","title":"Test","displaytitle":"<span class=\"mw-page-title-main\">Test</span>","namespace":{"id":0,"text":""},"wikibase_item":"Q224615","titles":{"canonical":"Test","normalized":"Test","display":"<span class=\"mw-page-title-main\">Test</span>"},"pageid":11089416,"lang":"en","dir":"ltr","revision":"1114329112","tid":"bf9b9c00-4506-11ed-84c5-0daf85508ff9","timestamp":"2022-10-05T23:37:58Z","description":"Topics referred to by the same term","description_source":"local","content_urls":{"desktop":{"page":"https://en.wikipedia.org/wiki/Test","revisions":"https://en.wikipedia.org/wiki/Test?action=history","edit":"https://en.wikipedia.org/wiki/Test?action=edit","talk":"https://en.wikipedia.org/wiki/Talk:Test"},"mobile":{"page":"https://en.m.wikipedia.org/wiki/Test","revisions":"https://en.m.wikipedia.org/wiki/Special:History/Test","edit":"https://en.m.wikipedia.org/wiki/Test?action=edit","talk":"https://en.m.wikipedia.org/wiki/Talk:Test"}},"extract":"Test(s), testing, or TEST may refer to:Test (assessment), an educational assessment intended to measure the respondents' knowledge or other abilities","extract_html":"<p><b>Test(s)</b>, <b>testing</b>, or <b>TEST</b> may refer to:</p><ul><li>Test (assessment), an educational assessment intended to measure the respondents' knowledge or other abilities</li></ul>"}`,
		},
		{
			sts: http.StatusNotFound,
			ttl: "Not Found",
			sum: `{
				"type":"https://mediawiki.org/wiki/HyperSwitch/errors/not_found",
				"title":"Not found.",
				"method":"get",
				"detail":"Page or revision not found.",
				"uri":"/en.wikipedia.org/v1/page/summary/Sadsd"
			}`,
			hse: true,
		},
	} {
		suite.Run(t, testCase)
	}
}

type getLanguagesTestSuite struct {
	suite.Suite
	ctx context.Context
	srv *httptest.Server
	clt *Client
	hse bool
	dtb string
	pld string
	sts int
	lns []*Language
}

func (s *getLanguagesTestSuite) SetupSuite() {
	s.srv = createActionsAPIServer(s.sts, s.pld)
	s.dtb = "enwiki"
	s.ctx = context.Background()
	s.clt = &Client{
		DefaultURL:      s.srv.URL,
		DefaultDatabase: s.dtb,
		HTTPClient:      &http.Client{},
		Tracer:          mockTracer,
	}

	s.lns = []*Language{}
	rsp := new(Response)
	_ = json.Unmarshal([]byte(s.pld), rsp)

	if rsp.SiteMatrix != nil {
		for _, lng := range rsp.SiteMatrix.Languages {
			s.lns = append(s.lns, lng)
		}

		for _, prj := range rsp.SiteMatrix.Specials {
			if prj.DBName == "commonswiki" || prj.DBName == "wikidata" {
				s.lns = append(s.lns, &Language{Projects: []*Project{prj}})
			}
		}
	}
}

func (s *getLanguagesTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *getLanguagesTestSuite) TestGetLanguages() {
	lns, err := s.clt.GetLanguages(s.ctx, s.dtb)

	if len(s.lns) > 0 {
		s.Assert().Len(lns, len(s.lns))
	}

	if s.hse {
		s.Assert().Error(err)
	} else {
		s.Assert().NoError(err)
	}
}

func TestGetLanguages(t *testing.T) {
	for _, testCase := range []*getLanguagesTestSuite{
		{
			sts: http.StatusOK,
			pld: getLanguagesResponse,
		},
		{
			sts: http.StatusOK,
			pld: errorResponse,
			hse: true,
		},
		{
			sts: http.StatusInternalServerError,
			hse: true,
		},
	} {
		suite.Run(t, testCase)
	}
}

type getLanguageTestSuite struct {
	suite.Suite
	ctx context.Context
	srv *httptest.Server
	clt *Client
	hse bool
	dtb string
	pld string
	sts int
	lns map[string]*Language
	err error
}

func (s *getLanguageTestSuite) SetupSuite() {
	s.srv = createActionsAPIServer(s.sts, s.pld)
	s.ctx = context.Background()
	s.clt = &Client{
		DefaultURL:      s.srv.URL,
		DefaultDatabase: "enwiki",
		HTTPClient:      &http.Client{},
		Tracer:          mockTracer,
	}

	s.lns = map[string]*Language{}
	rsp := new(Response)
	_ = json.Unmarshal([]byte(s.pld), rsp)

	if rsp.SiteMatrix != nil {
		for _, lng := range rsp.SiteMatrix.Languages {
			s.lns[lng.Code] = lng
		}
	}
}

func (s *getLanguageTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *getLanguageTestSuite) TestGetLanguage() {
	lng, err := s.clt.GetLanguage(s.ctx, s.dtb)

	if s.hse {
		s.Assert().Nil(lng)
		s.Assert().Error(err)
	} else {
		s.Assert().Equal(s.lns[lng.Code], lng)
		s.Assert().NoError(err)
	}

	if s.err != nil {
		s.Assert().Equal(s.err, err)
	}
}

func TestGetLanguage(t *testing.T) {
	for _, testCase := range []*getLanguageTestSuite{
		{
			sts: http.StatusOK,
			pld: getLanguagesResponse,
			dtb: "afwikibooks",
		},
		{
			sts: http.StatusOK,
			pld: getLanguagesResponse,
			hse: true,
			dtb: "not found",
			err: ErrLanguageNotFound,
		},
		{
			sts: http.StatusOK,
			pld: errorResponse,
			hse: true,
		},
		{
			sts: http.StatusInternalServerError,
			hse: true,
		},
	} {
		suite.Run(t, testCase)
	}
}

type getProjectTestSuite struct {
	suite.Suite
	ctx context.Context
	srv *httptest.Server
	clt *Client
	hse bool
	dtb string
	pld string
	sts int
	prs map[string]*Project
	err error
}

func (s *getProjectTestSuite) SetupSuite() {
	s.srv = createActionsAPIServer(s.sts, s.pld)
	s.ctx = context.Background()
	s.clt = &Client{
		DefaultURL:      s.srv.URL,
		DefaultDatabase: "enwiki",
		HTTPClient:      &http.Client{},
		Tracer:          mockTracer,
	}

	s.prs = map[string]*Project{}
	rsp := new(Response)
	_ = json.Unmarshal([]byte(s.pld), rsp)

	if rsp.SiteMatrix != nil {
		for _, lng := range rsp.SiteMatrix.Languages {
			for _, prj := range lng.Projects {
				s.prs[prj.DBName] = prj
			}
		}
	}
}

func (s *getProjectTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *getProjectTestSuite) TestGetProject() {
	prj, err := s.clt.GetProject(s.ctx, s.dtb)

	if s.hse {
		s.Assert().Nil(prj)
		s.Assert().Error(err)
	} else {
		s.Assert().Equal(s.prs[prj.DBName], prj)
		s.Assert().NoError(err)
	}

	if s.err != nil {
		s.Assert().Equal(s.err, err)
	}
}

func TestGetProject(t *testing.T) {
	for _, testCase := range []*getProjectTestSuite{
		{
			sts: http.StatusOK,
			pld: getLanguagesResponse,
			dtb: "afwikibooks",
		},
		{
			sts: http.StatusOK,
			pld: getLanguagesResponse,
			hse: true,
			dtb: "not found",
			err: ErrProjectNotFound,
		},
		{
			sts: http.StatusOK,
			pld: errorResponse,
			hse: true,
		},
		{
			sts: http.StatusInternalServerError,
			hse: true,
		},
	} {
		suite.Run(t, testCase)
	}
}

type getProjectsTestSuite struct {
	suite.Suite
	ctx context.Context
	srv *httptest.Server
	clt *Client
	hse bool
	dtb string
	pld string
	sts int
	prs []*Project
	err error
}

func (s *getProjectsTestSuite) SetupSuite() {
	s.srv = createActionsAPIServer(s.sts, s.pld)
	s.ctx = context.Background()
	s.clt = &Client{
		DefaultURL:      s.srv.URL,
		DefaultDatabase: "enwiki",
		HTTPClient:      &http.Client{},
		Tracer:          mockTracer,
	}

	s.prs = []*Project{}
	rsp := new(Response)
	_ = json.Unmarshal([]byte(s.pld), rsp)

	if rsp.SiteMatrix != nil {
		for _, lng := range rsp.SiteMatrix.Languages {
			s.prs = append(s.prs, lng.Projects...)
		}
	}
}

func (s *getProjectsTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *getProjectsTestSuite) TestGetProjects() {
	prl, err := s.clt.GetProjects(s.ctx, s.dtb)

	if s.hse {
		s.Assert().Nil(prl)
		s.Assert().Error(err)
	} else {
		s.Assert().NoError(err)
	}

	if s.err != nil {
		s.Assert().Equal(s.err, err)
	}
}

func TestGetProjects(t *testing.T) {
	for _, testCase := range []*getProjectsTestSuite{
		{
			sts: http.StatusOK,
			pld: getLanguagesResponse,
			dtb: "enwiki",
		},
		{
			sts: http.StatusOK,
			pld: getLanguagesResponse,
			hse: true,
			dtb: "not found",
			err: ErrProjectNotFound,
		},
		{
			sts: http.StatusOK,
			pld: errorResponse,
			hse: true,
		},
		{
			sts: http.StatusInternalServerError,
			hse: true,
		},
	} {
		suite.Run(t, testCase)
	}
}

type getNamespacesTestSuite struct {
	suite.Suite
	ctx context.Context
	srv *httptest.Server
	clt *Client
	rsp *Response
	hse bool
	dtb string
	pld string
	sts int
}

func (s *getNamespacesTestSuite) SetupSuite() {
	s.srv = createActionsAPIServer(s.sts, s.pld)
	s.dtb = "enwiki"
	s.ctx = context.Background()
	s.clt = &Client{
		DefaultURL:      s.srv.URL,
		DefaultDatabase: s.dtb,
		HTTPClient:      &http.Client{},
		Tracer:          mockTracer,
	}

	s.rsp = new(Response)
	_ = json.Unmarshal([]byte(s.pld), s.rsp)
}

func (s *getNamespacesTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *getNamespacesTestSuite) TestGetNamespaces() {
	nms, err := s.clt.GetNamespaces(s.ctx, s.dtb)

	if s.rsp.Query != nil && len(s.rsp.Query.Namespaces) > 0 {
		s.Assert().Len(nms, len(s.rsp.Query.Namespaces))

		for _, nsp := range s.rsp.Query.Namespaces {
			s.Assert().Equal(s.rsp.Query.Namespaces[nsp.ID], nsp)
		}
	}

	if s.hse {
		s.Assert().Error(err)
	} else {
		s.Assert().NoError(err)
	}
}

func TestGetNamespaces(t *testing.T) {
	for _, testCase := range []*getNamespacesTestSuite{
		{
			sts: http.StatusOK,
			pld: `{"batchcomplete":true,"query":{"namespaces":{"-2":{"id":-2,"case":"first-letter","name":"Media","subpages":false,"canonical":"Media","content":false,"nonincludable":false},"-1":{"id":-1,"case":"first-letter","name":"Special","subpages":false,"canonical":"Special","content":false,"nonincludable":false},"0":{"id":0,"case":"first-letter","name":"","subpages":false,"content":true,"nonincludable":false},"1":{"id":1,"case":"first-letter","name":"Talk","subpages":true,"canonical":"Talk","content":false,"nonincludable":false},"2":{"id":2,"case":"first-letter","name":"User","subpages":true,"canonical":"User","content":false,"nonincludable":false},"3":{"id":3,"case":"first-letter","name":"User talk","subpages":true,"canonical":"User talk","content":false,"nonincludable":false},"4":{"id":4,"case":"first-letter","name":"Wikipedia","subpages":true,"canonical":"Project","content":false,"nonincludable":false},"5":{"id":5,"case":"first-letter","name":"Wikipedia talk","subpages":true,"canonical":"Project talk","content":false,"nonincludable":false},"6":{"id":6,"case":"first-letter","name":"File","subpages":false,"canonical":"File","content":false,"nonincludable":false},"7":{"id":7,"case":"first-letter","name":"File talk","subpages":true,"canonical":"File talk","content":false,"nonincludable":false},"8":{"id":8,"case":"first-letter","name":"MediaWiki","subpages":false,"canonical":"MediaWiki","content":false,"nonincludable":false,"namespaceprotection":"editinterface"},"9":{"id":9,"case":"first-letter","name":"MediaWiki talk","subpages":true,"canonical":"MediaWiki talk","content":false,"nonincludable":false},"10":{"id":10,"case":"first-letter","name":"Template","subpages":true,"canonical":"Template","content":false,"nonincludable":false},"11":{"id":11,"case":"first-letter","name":"Template talk","subpages":true,"canonical":"Template talk","content":false,"nonincludable":false},"12":{"id":12,"case":"first-letter","name":"Help","subpages":true,"canonical":"Help","content":false,"nonincludable":false},"13":{"id":13,"case":"first-letter","name":"Help talk","subpages":true,"canonical":"Help talk","content":false,"nonincludable":false},"14":{"id":14,"case":"first-letter","name":"Category","subpages":true,"canonical":"Category","content":false,"nonincludable":false},"15":{"id":15,"case":"first-letter","name":"Category talk","subpages":true,"canonical":"Category talk","content":false,"nonincludable":false},"100":{"id":100,"case":"first-letter","name":"Portal","subpages":true,"canonical":"Portal","content":false,"nonincludable":false},"101":{"id":101,"case":"first-letter","name":"Portal talk","subpages":true,"canonical":"Portal talk","content":false,"nonincludable":false},"118":{"id":118,"case":"first-letter","name":"Draft","subpages":true,"canonical":"Draft","content":false,"nonincludable":false},"119":{"id":119,"case":"first-letter","name":"Draft talk","subpages":true,"canonical":"Draft talk","content":false,"nonincludable":false},"710":{"id":710,"case":"first-letter","name":"TimedText","subpages":false,"canonical":"TimedText","content":false,"nonincludable":false},"711":{"id":711,"case":"first-letter","name":"TimedText talk","subpages":false,"canonical":"TimedText talk","content":false,"nonincludable":false},"828":{"id":828,"case":"first-letter","name":"Module","subpages":true,"canonical":"Module","content":false,"nonincludable":false},"829":{"id":829,"case":"first-letter","name":"Module talk","subpages":true,"canonical":"Module talk","content":false,"nonincludable":false},"2300":{"id":2300,"case":"case-sensitive","name":"Gadget","subpages":false,"canonical":"Gadget","content":false,"nonincludable":false,"namespaceprotection":"gadgets-edit"},"2301":{"id":2301,"case":"case-sensitive","name":"Gadget talk","subpages":false,"canonical":"Gadget talk","content":false,"nonincludable":false},"2302":{"id":2302,"case":"case-sensitive","name":"Gadget definition","subpages":false,"canonical":"Gadget definition","content":false,"nonincludable":false,"namespaceprotection":"gadgets-definition-edit","defaultcontentmodel":"GadgetDefinition"},"2303":{"id":2303,"case":"case-sensitive","name":"Gadget definition talk","subpages":false,"canonical":"Gadget definition talk","content":false,"nonincludable":false}}}}`,
		},
		{
			sts: http.StatusOK,
			pld: errorResponse,
			hse: true,
		},
		{
			sts: http.StatusInternalServerError,
			hse: true,
		},
	} {
		suite.Run(t, testCase)
	}
}

type getRandomPagesTestSuite struct {
	suite.Suite
	ctx context.Context
	srv *httptest.Server
	clt *Client
	rsp *Response
	hse bool
	dtb string
	pld string
	sts int
}

func (s *getRandomPagesTestSuite) SetupSuite() {
	s.srv = createActionsAPIServer(s.sts, s.pld)
	s.dtb = "enwiki"
	s.ctx = context.Background()
	s.clt = &Client{
		DefaultURL:      s.srv.URL,
		DefaultDatabase: s.dtb,
		HTTPClient:      &http.Client{},
		Tracer:          mockTracer,
	}

	s.rsp = new(Response)
	_ = json.Unmarshal([]byte(s.pld), s.rsp)
}

func (s *getRandomPagesTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *getRandomPagesTestSuite) TestGetRandomPages() {
	pgs, err := s.clt.GetRandomPages(s.ctx, s.dtb)

	if s.rsp.Query != nil && len(s.rsp.Query.Random) > 0 {
		s.Assert().Len(pgs, len(s.rsp.Query.Random))

		itl := []string{}
		rtl := []string{}

		for _, rnd := range s.rsp.Query.Random {
			itl = append(itl, rnd.Title)
		}

		for _, pge := range pgs {
			rtl = append(rtl, pge.Title)
		}

		s.Assert().ElementsMatch(itl, rtl)
	}

	if s.hse {
		s.Assert().Error(err)
	} else {
		s.Assert().NoError(err)
	}
}

func TestGetRandomPages(t *testing.T) {
	for _, testCase := range []*getRandomPagesTestSuite{
		{
			sts: http.StatusOK,
			pld: `{
				"batchcomplete": true,
				"continue": {
				  "rncontinue": "0.626892872077|0.626893505289|40964313|0",
				  "continue": "-||"
				},
				"query": {
				  "random": [
					{
					  "id": 823840,
					  "ns": 0,
					  "title": "Nintendo Power"
					},
					{
					  "id": 60232671,
					  "ns": 0,
					  "title": "Umenomoto Station"
					},
					{
					  "id": 17902603,
					  "ns": 0,
					  "title": "2008 ANAPROF Clausura"
					},
					{
					  "id": 5101698,
					  "ns": 0,
					  "title": "Siege of Otate"
					},
					{
					  "id": 27747856,
					  "ns": 0,
					  "title": "William J. Hughes Technical Center"
					}
				  ]
				}
			  }`,
		},
		{
			sts: http.StatusOK,
			pld: errorResponse,
			hse: true,
		},
		{
			sts: http.StatusInternalServerError,
			hse: true,
		},
	} {
		suite.Run(t, testCase)
	}
}

type getUsersTestSuite struct {
	suite.Suite
	ctx context.Context
	srv *httptest.Server
	clt *Client
	urs map[int]*User
	hse bool
	dtb string
	ids []int
	pld string
	sts int
}

func (s *getUsersTestSuite) SetupSuite() {
	s.srv = createActionsAPIServer(s.sts, s.pld)
	s.dtb = "enwiki"
	s.ctx = context.Background()
	s.clt = &Client{
		DefaultURL:      s.srv.URL,
		DefaultDatabase: s.dtb,
		HTTPClient:      &http.Client{},
		Tracer:          mockTracer,
	}

	s.urs = map[int]*User{}
	rsp := new(Response)
	_ = json.Unmarshal([]byte(s.pld), rsp)

	if rsp.Query != nil {
		for _, usr := range rsp.Query.Users {
			s.urs[usr.UserID] = usr
		}
	}
}

func (s *getUsersTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *getUsersTestSuite) TestGetUsers() {
	urs, err := s.clt.GetUsers(s.ctx, s.dtb, s.ids)

	if len(s.urs) > 0 {
		s.Assert().Len(urs, len(s.urs))

		for _, usr := range urs {
			s.Assert().Equal(s.urs[usr.UserID], usr)
		}
	}

	if s.hse {
		s.Assert().Error(err)
	} else {
		s.Assert().NoError(err)
	}
}

func TestGetUsers(t *testing.T) {
	for _, testCase := range []*getUsersTestSuite{
		{
			ids: []int{2, 3, 5},
			sts: http.StatusOK,
			pld: getUsersResponse,
		},
		{
			sts: http.StatusOK,
			pld: errorResponse,
			hse: true,
		},
		{
			sts: http.StatusInternalServerError,
			hse: true,
		},
	} {
		suite.Run(t, testCase)
	}
}

type getUserTestSuite struct {
	suite.Suite
	ctx context.Context
	srv *httptest.Server
	clt *Client
	err error
	urs map[int]*User
	hse bool
	dtb string
	uid int
	pld string
	sts int
}

func (s *getUserTestSuite) SetupSuite() {
	s.srv = createActionsAPIServer(s.sts, s.pld)
	s.dtb = "enwiki"
	s.ctx = context.Background()
	s.clt = &Client{
		DefaultURL:      s.srv.URL,
		DefaultDatabase: s.dtb,
		HTTPClient:      &http.Client{},
		Tracer:          mockTracer,
	}

	s.urs = map[int]*User{}
	rsp := new(Response)
	_ = json.Unmarshal([]byte(s.pld), rsp)

	if rsp.Query != nil {
		for _, usr := range rsp.Query.Users {
			s.urs[usr.UserID] = usr
		}
	}
}

func (s *getUserTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *getUserTestSuite) TestGetUser() {
	usr, err := s.clt.GetUser(s.ctx, s.dtb, s.uid)

	if len(s.urs) > 0 {
		s.Assert().Equal(s.urs[usr.UserID], usr)
	}

	if s.hse {
		s.Assert().Error(err)
	} else {
		s.Assert().NoError(err)
	}

	if s.err != nil {
		s.Assert().Equal(s.err, err)
	}
}

func TestGetUser(t *testing.T) {
	for _, testCase := range []*getUserTestSuite{
		{
			uid: 5,
			sts: http.StatusOK,
			pld: getUsersResponse,
		},
		{
			uid: 10,
			sts: http.StatusOK,
			pld: getPagesResponse,
			err: ErrUserNotFound,
			hse: true,
		},
		{
			sts: http.StatusOK,
			pld: errorResponse,
			hse: true,
		},
		{
			sts: http.StatusInternalServerError,
			hse: true,
		},
	} {
		suite.Run(t, testCase)
	}
}

type retryAfterTestSuite struct {
	suite.Suite
	ctx context.Context
	srv *httptest.Server
	clt *Client
	rtv int
	era bool
	hre bool
}

func (s *retryAfterTestSuite) SetupSuite() {
	rtr := http.NewServeMux()
	cnt := 2

	rtr.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if cnt > 0 {
			w.Header().Add("Retry-After", "Wed, 21 Oct 2015 07:28:00 GMT")
			w.Header().Add("Retry-After", strconv.Itoa(s.rtv))
			w.WriteHeader(http.StatusTooManyRequests)
		} else {
			if strings.Contains(r.URL.Path, "/w/rest.php/v1/page/") {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("<html><body>Test Page</body></html>"))
			} else {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"batchcomplete": true, "sitematrix": {"count": 0, "languages": {}}}`))
			}
		}

		cnt--
	})

	s.srv = httptest.NewServer(rtr)
	s.ctx = context.Background()
	s.clt = &Client{
		HTTPClient:       &http.Client{},
		DefaultURL:       s.srv.URL,
		DefaultDatabase:  "enwiki",
		EnableRetryAfter: s.era,
		Tracer:           mockTracer,
	}
}

func (s *retryAfterTestSuite) TestRetryAfrer() {
	_, err := s.clt.GetLanguages(s.ctx, s.clt.DefaultDatabase)

	if s.hre {
		s.Assert().Error(err)
	} else {
		s.Assert().NoError(err)
	}

	_, err = s.clt.GetPageHTML(s.ctx, s.clt.DefaultDatabase, "Earth")

	if s.hre {
		s.Assert().Error(err)
	} else {
		s.Assert().NoError(err)
	}
}

func TestRetryAfter(t *testing.T) {
	for _, testCase := range []*retryAfterTestSuite{
		{
			era: true,
			hre: false,
			rtv: 1,
		},
		{
			era: false,
			hre: true,
			rtv: 1,
		},
	} {
		suite.Run(t, testCase)
	}
}

func createLiftWingAPIServer(sts int, pld string) *httptest.Server {
	rtr := http.NewServeMux()

	rtr.HandleFunc("/service/lw/inference/v1/models/revertrisk-language-agnostic:predict", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(sts)
		if sts == http.StatusOK {
			_, _ = w.Write([]byte(pld))
		} else {
			_, _ = w.Write([]byte("invalid status code"))
		}
	})

	rtr.HandleFunc("/service/lw/inference/v1/models/reference-need:predict", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(sts)
		if sts == http.StatusOK {
			_, _ = w.Write([]byte(pld))
		} else {
			_, _ = w.Write([]byte("invalid status code"))
		}
	})

	rtr.HandleFunc("/service/lw/inference/v1/models/reference-risk:predict", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(sts)
		if sts == http.StatusOK {
			_, _ = w.Write([]byte(pld))
		} else {
			_, _ = w.Write([]byte("invalid status code"))
		}
	})

	return httptest.NewServer(rtr)
}

type getScoreTestSuite struct {
	suite.Suite
	ctx context.Context
	clt *Client
	srv *httptest.Server
	sts int    // status codes
	pld string // payload
	rid int
	lng string
	prj string
	mdl string
	err error
}

func (s *getScoreTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.pld = strings.ReplaceAll(s.pld, "\n", "")
	s.pld = strings.ReplaceAll(s.pld, " ", "")
	s.srv = createLiftWingAPIServer(s.sts, s.pld)
	s.prj = ""
	s.clt = &Client{
		HTTPClientLiftWing: &http.Client{},
		LiftWingBaseURL:    s.srv.URL + "/service/lw/inference/v1/models/",
		Tracer:             mockTracer,
	}
}

func (s *getScoreTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *getScoreTestSuite) TestGetScore() {
	rsp, err := s.clt.GetScore(s.ctx, s.rid, s.lng, s.prj, s.mdl)

	if s.err != nil {
		s.Assert().Error(err)
		s.Assert().Contains(err.Error(), s.err.Error())
	} else {
		s.Assert().NoError(err)
		rrk, _ := json.Marshal(rsp)
		s.Assert().Equal(s.pld, string(rrk))
	}
}

type getReferenceNeedScoreTestSuite struct {
	suite.Suite
	ctx context.Context
	clt *Client
	srv *httptest.Server
	sts int
	pld string
	rid int
	lng string
	prj string
	err error
	mdl string
}

func (s *getReferenceNeedScoreTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.pld = strings.ReplaceAll(s.pld, "\n", "")
	s.pld = strings.ReplaceAll(s.pld, " ", "")
	s.srv = createLiftWingAPIServer(s.sts, s.pld)
	s.clt = &Client{
		HTTPClientLiftWing: &http.Client{},
		LiftWingBaseURL:    s.srv.URL + "/service/lw/inference/v1/models/",
		Tracer:             mockTracer,
	}
}

type getReferenceRiskScoreTestSuite struct {
	suite.Suite
	ctx context.Context
	clt *Client
	srv *httptest.Server
	sts int
	pld string
	rid int
	lng string
	prj string
	err error
	mdl string
}

func (s *getReferenceRiskScoreTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.pld = strings.ReplaceAll(s.pld, "\n", "")
	s.pld = strings.ReplaceAll(s.pld, " ", "")
	s.srv = createLiftWingAPIServer(s.sts, s.pld)
	s.clt = &Client{
		HTTPClientLiftWing: &http.Client{},
		LiftWingBaseURL:    s.srv.URL + "/service/lw/inference/v1/models/",
		Tracer:             mockTracer,
	}
}

func (s *getReferenceNeedScoreTestSuite) TearDownTest() {
	s.srv.Close()
}

func (s *getReferenceNeedScoreTestSuite) TestGetReferenceNeedScore() {
	rsp, err := s.clt.GetReferenceNeedScore(s.ctx, s.rid, s.lng, s.prj)

	if s.err != nil {
		s.Assert().Error(err)
		s.Assert().Contains(err.Error(), s.err.Error())
	} else {
		s.Assert().NoError(err)
		encoded, _ := json.Marshal(rsp)
		s.Assert().Equal(s.pld, string(encoded))
	}
}

func (s *getReferenceRiskScoreTestSuite) TestGetReferenceRiskScore() {
	rsp, err := s.clt.GetReferenceRiskScore(s.ctx, s.rid, s.lng, s.prj)

	if s.err != nil {
		s.Assert().Error(err)
		s.Assert().Contains(err.Error(), s.err.Error())
	} else {
		s.Assert().NoError(err)
		encoded, _ := json.Marshal(rsp)
		s.Assert().Equal(s.pld, string(encoded))
	}
}

func TestGetScore(t *testing.T) {
	for _, testCase := range []*getScoreTestSuite{
		{
			sts: http.StatusOK,
			pld: liftWingPayload,
			mdl: "revertrisk",
			lng: "en",
			err: nil,
		},
		{
			sts: http.StatusBadRequest,
			pld: liftWingPayload,
			mdl: "revertrisk",
			lng: "en",
			err: errors.New("400 Bad Request:invalidstatuscode"),
		},
	} {
		suite.Run(t, testCase)
	}
}

func TestGetReferenceNeedScore(t *testing.T) {
	for _, testCase := range []*getReferenceNeedScoreTestSuite{
		{
			sts: http.StatusOK,
			pld: referenceNeedResponse,
			mdl: "reference-need",
			lng: "en",
			err: nil,
		},
		{
			sts: http.StatusBadRequest,
			pld: referenceNeedError,
			lng: "en",
			err: errors.New("400 Bad Request:invalidstatuscode"),
		},
	} {
		suite.Run(t, testCase)
	}
}

func TestGetReferenceRiskScore(t *testing.T) {
	for _, testCase := range []*getReferenceRiskScoreTestSuite{
		{
			sts: http.StatusOK,
			pld: referenceRiskResponse,
			mdl: "reference-risk",
			lng: "en",
			err: nil,
		},
		{
			sts: http.StatusBadRequest,
			pld: referenceRiskError,
			lng: "en",
			err: errors.New("400 Bad Request:invalidstatuscode"),
		},
	} {
		suite.Run(t, testCase)
	}
}

func createCommonsServer(sts int, pld []byte) *httptest.Server {
	rtr := http.NewServeMux()

	rtr.HandleFunc("/commons", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(sts)
		if sts == http.StatusOK || sts == http.StatusPartialContent {
			_, _ = w.Write(pld)
		} else {
			_, _ = w.Write([]byte("invalid status code"))
		}
	})

	return httptest.NewServer(rtr)
}

type downloadFileTestSuite struct {
	suite.Suite
	ctx context.Context
	clt *Client
	srv *httptest.Server
	sts int
	pld []byte
	ops func(*http.Request)
	url string
	err error
}

func (s *downloadFileTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.srv = createCommonsServer(s.sts, s.pld)
	s.clt = &Client{
		DefaultURL: s.srv.URL,
		HTTPClient: &http.Client{},
		Tracer:     mockTracer,
	}
}

func (s *downloadFileTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *downloadFileTestSuite) TestDownload() {
	s.url = s.srv.URL + "/commons"

	rsp, err := s.clt.DownloadFile(s.ctx, s.url, s.ops)

	if s.err != nil {
		s.Assert().Error(err)
		s.Assert().Contains(err.Error(), s.err.Error())
	} else {
		s.Assert().NoError(err)
		s.Assert().Equal(s.pld, rsp)
	}
}

func TestDownloadFile(t *testing.T) {
	for _, testCase := range []*downloadFileTestSuite{
		{
			sts: http.StatusOK,
			pld: []byte("test payload"),
			err: nil,
			ops: func(r *http.Request) {},
		},
		{
			sts: http.StatusPartialContent,
			pld: []byte("test payload"),
			ops: func(r *http.Request) { r.Header.Set("Range", "bytes=0-99") },
			err: nil,
		},
		{
			sts: http.StatusInternalServerError,
			pld: nil,
			err: errors.New("500 Internal Server Error"),
			ops: func(r *http.Request) {},
		},
	} {
		suite.Run(t, testCase)
	}
}

type headFileTestSuite struct {
	suite.Suite
	ctx context.Context
	clt *Client
	srv *httptest.Server
	sts int
	pld []byte
	ops func(*http.Request)
	url string
	err error
}

func (s *headFileTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.srv = createCommonsServer(s.sts, s.pld)
	s.clt = &Client{
		DefaultURL: s.srv.URL,
		HTTPClient: &http.Client{},
		Tracer:     mockTracer,
	}
}

func (s *headFileTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *headFileTestSuite) TestHead() {
	s.url = s.srv.URL + "/commons"

	_, err := s.clt.HeadFile(s.ctx, s.url, s.ops)

	if s.err != nil {
		s.Assert().Error(err)
		s.Assert().Contains(err.Error(), s.err.Error())
	} else {
		s.Assert().NoError(err)
	}
}

func TestHeadFile(t *testing.T) {
	for _, testCase := range []*headFileTestSuite{
		{
			sts: http.StatusOK,
			pld: []byte("test payload"),
			err: nil,
			ops: func(r *http.Request) {},
		},
		{
			sts: http.StatusInternalServerError,
			pld: nil,
			err: errors.New("500 Internal Server Error"),
			ops: func(r *http.Request) {},
		},
	} {
		suite.Run(t, testCase)
	}
}

type siteMatrixUnmarshalTestSuite struct {
	suite.Suite
	jsonData string
	count    int
	hasError bool
}

func (s *siteMatrixUnmarshalTestSuite) TestUnmarshal() {
	var sm SiteMatrix
	err := json.Unmarshal([]byte(s.jsonData), &sm)

	if s.hasError {
		s.Assert().Error(err)
	} else {
		s.Assert().NotNil(sm)
	}
}

func TestSiteMatrixUnmarshal(t *testing.T) {
	for _, testCase := range []*siteMatrixUnmarshalTestSuite{
		{
			jsonData: `
            {
                "count": 1044,
                "357": {
                    "code": "zh",
                    "name": "中文",
                    "site": [
                        {
                            "url": "https://zh.wikipedia.org",
                            "dbname": "zhwiki",
                            "code": "wiki",
                            "sitename": "Wikipedia"
                        }
                    ],
                    "dir": "ltr",
                    "localname": "Chinese"
                },
                "specials": [
                    {
                        "url": "https://commons.wikimedia.org",
                        "dbname": "commonswiki",
                        "code": "commons",
                        "lang": "commons",
                        "sitename": "Wikimedia Commons"
                    }
                ]
            }`,
			count:    1044,
			hasError: false,
		},
		{
			jsonData: `{"count": "invalid"}`,
			count:    0,
			hasError: true,
		},
	} {
		suite.Run(t, testCase)
	}
}

type getPageByRevisionSuite struct {
	suite.Suite
	ctx               context.Context
	srv               *httptest.Server
	clt               *Client
	revid             int
	errorExpected     bool
	database          string
	apiResponseJson   string
	apiResponseParsed *Response
	apiStatus         int
	expectedPageId    int
}

func (s *getPageByRevisionSuite) SetupSuite() {
	s.srv = createActionsAPIServer(s.apiStatus, s.apiResponseJson)
	s.database = "enwiki"
	s.ctx = context.Background()
	s.clt = &Client{
		DefaultURL:      s.srv.URL,
		DefaultDatabase: s.database,
		HTTPClient:      &http.Client{},
		Tracer:          mockTracer,
	}

	s.apiResponseParsed = new(Response)
	_ = json.Unmarshal([]byte(s.apiResponseJson), s.apiResponseParsed)
}

func (s *getPageByRevisionSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *getPageByRevisionSuite) TestGetPageByRevision() {
	page, err := s.clt.GetPageByRevision(s.ctx, s.database, s.revid)

	if s.errorExpected {
		s.Assert().Error(err)
		s.Assert().Nil(page)
		return
	}

	s.Assert().NoError(err)
	s.Assert().Equal(page.PageID, s.expectedPageId)
}

func TestGetPageByRevision(t *testing.T) {
	for _, testCase := range []*getPageByRevisionSuite{
		// Example: https://en.wikipedia.org/w/api.php?action=query&revids=143507357&format=json&formatversion=2
		{
			apiStatus: http.StatusOK,
			apiResponseJson: `
{
  "batchcomplete": true,
  "query": {
    "pages": [
      {
        "pageid": 4187252,
        "ns": 0,
        "title": "Salvadoran Civil War"
      }
    ]
  }
}`,
			expectedPageId: 4187252,
		},
		// Example: https://en.wikipedia.org/w/api.php?action=query&revids=abc&format=json&formatversion=2
		{
			apiStatus: http.StatusOK,
			apiResponseJson: `
{
  "error": {
    "code": "badinteger",
    "info": "Invalid value \"abc\" for integer parameter \"revids\".",
    "docref": "See https://en.wikipedia.org/w/api.php for API usage. Subscribe to the mediawiki-api-announce mailing list at &lt;https://lists.wikimedia.org/postorius/lists/mediawiki-api-announce.lists.wikimedia.org/&gt; for notice of API deprecations and breaking changes."
  },
  "servedby": "mw-api-ext.eqiad.main-846b9fd56-vlcnp"
}`,
			errorExpected: true,
		},
		{
			apiStatus:     http.StatusInternalServerError,
			errorExpected: true,
		},
	} {
		suite.Run(t, testCase)
	}
}

type getAllRevisionsSuite struct {
	suite.Suite
	ctx                   context.Context
	srv                   *httptest.Server
	clt                   *Client
	pageid                int
	errorExpected         bool
	database              string
	apiFirstResponse      string
	apiFollowUpResponse   string
	apiStatus             int
	expectedPageId        int
	expectedFollowUpParam string
	expectedRevisions     []*Revision

	// Test state
	requestsSent int
}

func (s *getAllRevisionsSuite) setUpApiServer() *httptest.Server {
	rtr := http.NewServeMux()

	rtr.HandleFunc("/w/api.php", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(s.apiStatus)

		bdy, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			log.Fatalln("Failed to read request")
		}

		params, err := url.ParseQuery(string(bdy))
		if err != nil {
			log.Fatalln("Failed to parse request body")
		}
		s.Assert().Equal(params.Get("prop"), "revisions")

		if s.expectedFollowUpParam != "" && s.requestsSent == 1 {
			s.Assert().Equal(params.Get("rvcontinue"), s.expectedFollowUpParam)

			_, _ = w.Write([]byte(s.apiFollowUpResponse))
		} else {
			s.Assert().Empty(params.Get("rvcontinue"))

			_, _ = w.Write([]byte(s.apiFirstResponse))
		}

		s.requestsSent++
	})

	return httptest.NewServer(rtr)
}

func (s *getAllRevisionsSuite) SetupSuite() {
	s.srv = s.setUpApiServer()
	s.database = "enwiki"
	s.ctx = context.Background()
	s.clt = &Client{
		DefaultURL:      s.srv.URL,
		DefaultDatabase: s.database,
		HTTPClient:      &http.Client{},
		Tracer:          mockTracer,
	}
}

func (s *getAllRevisionsSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *getAllRevisionsSuite) TestGetAllRevisions() {
	revisions, err := s.clt.GetAllRevisions(s.ctx, s.database, s.pageid)

	if s.errorExpected {
		s.Assert().Error(err)
		s.Assert().Nil(revisions)
		return
	}

	s.Assert().NoError(err)
	s.Assert().ElementsMatch(revisions, s.expectedRevisions)
}

func parseTimestamp(ts string) *time.Time {
	out, err := time.Parse("2006-01-02T15:04:05Z", ts)
	if err != nil {
		log.Fatalln("Failed to parse timestamp", ts, err.Error())
	}
	return &out
}

func TestGetAllRevisions(t *testing.T) {
	for _, testCase := range []*getAllRevisionsSuite{
		// Example: https://en.wikipedia.org/w/api.php?action=query&prop=revisions&pageids=4187252&rvlimit=1&format=json&formatversion=2
		{
			apiStatus: http.StatusOK,
			apiFirstResponse: `
{
  "continue": {
    "rvcontinue": "20250219173559|1276579474",
    "continue": "||"
  },
  "query": {
    "pages": [
      {
        "pageid": 4187252,
        "ns": 0,
        "title": "Salvadoran Civil War",
        "revisions": [
          {
            "revid": 1276633532,
            "user": "Sir Sputnik",
            "timestamp": "2025-02-19T23:48:01Z"
          }
        ]
      }
    ]
  }
}`,
			expectedFollowUpParam: "20250219173559|1276579474",
			apiFollowUpResponse: `
{
  "batchcomplete": true,
  "query": {
    "pages": [
      {
        "pageid": 4187252,
        "ns": 0,
        "title": "Salvadoran Civil War",
        "revisions": [
          {
            "revid": 1276579474,
            "user": "Andrew Pendelton",
            "timestamp": "2025-02-19T17:35:59Z"
          }
        ]
      }
    ]
  }
}
`,
			expectedPageId: 4187252,
			// All revisions should be returned:
			expectedRevisions: []*Revision{
				{RevID: 1276633532, User: "Sir Sputnik", Timestamp: parseTimestamp("2025-02-19T23:48:01Z")},
				{RevID: 1276579474, User: "Andrew Pendelton", Timestamp: parseTimestamp("2025-02-19T17:35:59Z")},
			},
		},
		// Example: https://en.wikipedia.org/w/api.php?action=query&prop=revisions&pageids=abc&format=json&formatversion=2
		{
			apiStatus: http.StatusOK,
			apiFirstResponse: `
{
  "error": {
    "code": "badinteger",
    "info": "Invalid value \"abc\" for integer parameter \"pageids\".",
    "docref": "See https://en.wikipedia.org/w/api.php for API usage. Subscribe to the mediawiki-api-announce mailing list at &lt;https://lists.wikimedia.org/postorius/lists/mediawiki-api-announce.lists.wikimedia.org/&gt; for notice of API deprecations and breaking changes."
  },
  "servedby": "mw-api-ext.eqiad.main-674d4cc8c7-q42qt"
}`,
			errorExpected: true,
		},
		{
			apiStatus:     http.StatusInternalServerError,
			errorExpected: true,
		},
	} {
		suite.Run(t, testCase)
	}
}

func TestGetContributorsComplete(t *testing.T) {
	assert := assert.New(t)

	server := createActionsAPIServer(200, getContributorsResponse100709)
	client := &Client{
		DefaultURL:      server.URL,
		DefaultDatabase: "enwiki",
		HTTPClient:      &http.Client{},
		Tracer:          mockTracer,
	}

	maxContrib := 500
	pages, fetchedAll, err := client.GetContributors(context.Background(), "enwiki", 100709, &maxContrib)
	assert.NoError(err)
	assert.True(fetchedAll)
	assert.Equal(213, pages[0].AnonymousContributors)
	assert.Equal(383, len(pages[0].Contributors))
}

func TestGetContributorsCountComplete(t *testing.T) {
	assert := assert.New(t)

	server := createActionsAPIServer(200, getContributorsResponse100709)
	client := &Client{
		DefaultURL:      server.URL,
		DefaultDatabase: "enwiki",
		HTTPClient:      &http.Client{},
		Tracer:          mockTracer,
	}

	maxContrib := 500
	editors, err := client.GetContributorsCount(context.Background(), "enwiki", 100709, &maxContrib)
	assert.NoError(err)
	assert.Equal(ContributorsCount{
		AnonymousCount:        213,
		RegisteredCount:       383,
		AllRegisteredIncluded: true,
	}, *editors)
}

func TestGetContributorsIncomplete(t *testing.T) {
	assert := assert.New(t)

	server := createActionsAPIServer(200, getContributorsResponse10025Page1)
	client := &Client{
		DefaultURL:      server.URL,
		DefaultDatabase: "enwiki",
		HTTPClient:      &http.Client{},
		Tracer:          mockTracer,
	}

	maxContrib := 500
	pages, fetchedAll, err := client.GetContributors(context.Background(), "enwiki", 10025, &maxContrib)
	assert.NoError(err)
	assert.False(fetchedAll)
	assert.Equal(1, len(pages))
	assert.Equal(374, pages[0].AnonymousContributors)
	assert.Equal(500, len(pages[0].Contributors))
}

func TestGetContributorsCompleteWithPagination(t *testing.T) {
	assert := assert.New(t)

	server := createActionsAPIServer(200, getContributorsResponse10025Page1, getContributorsResponse10025Page2)
	client := &Client{
		DefaultURL:      server.URL,
		DefaultDatabase: "enwiki",
		HTTPClient:      &http.Client{},
		Tracer:          mockTracer,
	}

	maxContrib := 1000
	pages, fetchedAll, err := client.GetContributors(context.Background(), "enwiki", 10025, &maxContrib)
	assert.NoError(err)
	assert.True(fetchedAll)
	assert.Equal(374, pages[0].AnonymousContributors)
	assert.Equal(500, len(pages[0].Contributors))
	assert.Equal(69, len(pages[1].Contributors))
}

func TestGetContributorsCountIncomplete(t *testing.T) {
	assert := assert.New(t)

	server := createActionsAPIServer(200, getContributorsResponse10025Page1)
	client := &Client{
		DefaultURL:      server.URL,
		DefaultDatabase: "enwiki",
		HTTPClient:      &http.Client{},
		Tracer:          mockTracer,
	}

	maxContrib := 500
	editors, err := client.GetContributorsCount(context.Background(), "enwiki", 10025, &maxContrib)
	assert.NoError(err)
	assert.Equal(ContributorsCount{
		AnonymousCount:        374,
		RegisteredCount:       500,
		AllRegisteredIncluded: false,
	}, *editors)
}

func TestGetContributorsCountCompleteWithPagination(t *testing.T) {
	assert := assert.New(t)

	server := createActionsAPIServer(200, getContributorsResponse10025Page1, getContributorsResponse10025Page2)
	client := &Client{
		DefaultURL:      server.URL,
		DefaultDatabase: "enwiki",
		HTTPClient:      &http.Client{},
		Tracer:          mockTracer,
	}

	maxContrib := 1000
	editors, err := client.GetContributorsCount(context.Background(), "enwiki", 10025, &maxContrib)
	assert.NoError(err)
	assert.Equal(ContributorsCount{
		AnonymousCount:        374,
		RegisteredCount:       569,
		AllRegisteredIncluded: true,
	}, *editors)
}

func TestFakeCompatibility(t *testing.T) {
	// If this test fails to compile, update APIFake to support the interface changes.
	var fake API = NewAPIFake()
	_ = fake
}
