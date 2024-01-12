// Package wmf handlers all of the API request for WMF APIs.
// Including Actions API, REST API and dumps.
package wmf

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"dario.cat/mergo"
)

// ErrProjectNotFound appears when project was not found in the site matrix.
var ErrProjectNotFound = errors.New("project not found")

// ErrLanguageNotFound appears when language was not found in the site matrix.
var ErrLanguageNotFound = errors.New("language not found")

// ErrPageNotFound appears when page was not found on the website.
var ErrPageNotFound = errors.New("page not found")

// ErrUserNotFound appears when user was not found in the response payload.
var ErrUserNotFound = errors.New("user not found")

// ErrHTTPClientNotFound appears http client is set to null when new API client was created.
var ErrHTTPClientNotFound = errors.New("http client not found")

// ErrLiftWingModelNotFound appears when LiftWing model was not found in the configuration.
var ErrLiftWingModelNotFound = errors.New("liftwing model not found")

// AllPagesGetter interface for all pages enumeration in specific namespace.
type AllPagesGetter interface {
	GetAllPages(ctx context.Context, dtb string, cbk func([]*Page), ops ...func(*url.Values)) error
}

// PagesGetter interface to expose method that gets list of pages for specific project.
type PagesGetter interface {
	GetPages(ctx context.Context, dtb string, tls []string, ops ...func(*url.Values)) (map[string]*Page, error)
}

// PageGetter interface to expose method that gets single page for specific database name.
type PageGetter interface {
	GetPage(ctx context.Context, dtb string, ttl string, ops ...func(*url.Values)) (*Page, error)
}

// PageHTMLGetter interface to expose method that gets single page HTML content.
type PageHTMLGetter interface {
	GetPageHTML(ctx context.Context, dtb string, ttl string, ops ...func(*url.Values)) (string, error)
}

// PagesHTMLGetter interface to expose method that gets a list page HTML content concurrently.
type PagesHTMLGetter interface {
	GetPagesHTML(ctx context.Context, dtb string, tls []string, mxc int, ops ...func(*url.Values)) map[string]*PageHTML
}

// LanguagesGetter interface to expose method that gets list of available languages using database name.
type LanguagesGetter interface {
	GetLanguages(ctx context.Context, dtb string, ops ...func(*url.Values)) ([]*Language, error)
}

// LanguageGetter interface to expose method that gets single language using the database name.
type LanguageGetter interface {
	GetLanguage(ctx context.Context, dtb string) (*Language, error)
}

// ProjectsGetter interface to expose method that gets an array of projects using the database name.
type ProjectsGetter interface {
	GetProjects(ctx context.Context, dtb string) ([]*Project, error)
}

// ProjectGetter interface to expose method that gets single project using the database name.
type ProjectGetter interface {
	GetProject(ctx context.Context, dtb string) (*Project, error)
}

// NamespaceGetter interface to expose method that gets a list of namespaces using the database name.
type NamespacesGetter interface {
	GetNamespaces(ctx context.Context, dtb string, ops ...func(*url.Values)) ([]*Namespace, error)
}

// RandomPagesGetter interface to expose method that gets a list of random articles from a project.
type RandomPagesGetter interface {
	GetRandomPages(ctx context.Context, dtb string, ops ...func(*url.Values)) ([]*Page, error)
}

// UsersGetter interface to expose method that finds a list of users by identifiers and database name.
type UsersGetter interface {
	GetUsers(ctx context.Context, dtb string, ids []int, ops ...func(*url.Values)) (map[int]*User, error)
}

// UserGetter interface to expose method that finds a single user by identifier and database name.
type UserGetter interface {
	GetUser(ctx context.Context, dtb string, id int, ops ...func(*url.Values)) (*User, error)
}

// ScoreGetter interface to expose method that gets a score for a revision.
type ScoreGetter interface {
	GetScore(ctx context.Context, rev int, lng string, prj string, mdl string) (*Score, error)
}

// PageSummaryGetter interface to expose method that gets page summary for specific page title.
type PageSummaryGetter interface {
	GetPageSummary(ctx context.Context, dtb string, ttl string, ops ...func(*url.Values)) (*PageSummary, error)
}

// API interface fot the whole API client.
type API interface {
	AllPagesGetter
	PagesGetter
	PageGetter
	PagesHTMLGetter
	PageHTMLGetter
	LanguagesGetter
	LanguageGetter
	ProjectGetter
	NamespacesGetter
	UsersGetter
	UserGetter
	ScoreGetter
	PageSummaryGetter
}

// Response generic structure for Actions API response.
type Response struct {
	BatchComplete bool                 `json:"batchcomplete,omitempty"`
	Continue      map[string]string    `json:"continue,omitempty"`
	Query         *Query               `json:"query,omitempty"`
	Error         *Error               `json:"error,omitempty"`
	ServedBy      string               `json:"servedby,omitempty"`
	SiteMatrix    map[string]*Language `json:"sitematrix,omitempty"`
}

// Error generic structure for Actions API error response.
type Error struct {
	Type      string `json:"type,omitempty"`
	Message   string `json:"message,omitempty"`
	Code      string `json:"code,omitempty"`
	Info      string `json:"info,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	LowLimit  int    `json:"lowlimit,omitempty"`
	HighLimit int    `json:"highlimit,omitempty"`
	DocRef    string `json:"docref,omitempty"`
}

// Query response type for the Actions API if you run a query request.
type Query struct {
	Redirects  []*Redirect        `json:"redirects,omitempty"`
	Normalized []*Normalization   `json:"normalized,omitempty"`
	Pages      []*Page            `json:"pages,omitempty"`
	Namespaces map[int]*Namespace `json:"namespaces,omitempty"`
	Users      []*User            `json:"users,omitempty"`
	AllPages   []*Page            `json:"allpages,omitempty"`
	Random     []*Page            `json:"random,omitempty"`
}

// Redirect shows title redirection in Actions API.
type Redirect struct {
	From   string `json:"from,omitempty"`
	To     string `json:"to,omitempty"`
	PageID int    `json:"pageid"`
	Ns     int    `json:"ns"`
	Title  string `json:"title"`
}

// Category representation of category in Actions API.
type Category struct {
	Ns     int    `json:"ns"`
	Title  string `json:"title"`
	Hidden bool   `json:"hidden"`
}

// Template representation of template in Actions API.
type Template struct {
	Ns    int    `json:"ns"`
	Title string `json:"title"`
}

// Flagged shows is revision was flagged by community in Actions API.
type Flagged struct {
	StableRevID  int        `json:"stable_revid"`
	Level        int        `json:"level"`
	LevelText    string     `json:"level_text"`
	PendingSince *time.Time `json:"pending_since"`
}

// Projection levels of protection and expiration for Actions API.
type Protection struct {
	Type   string `json:"type"`
	Level  string `json:"level"`
	Expiry string `json:"expiry"`
}

// Revision representation of revision data in Actions API.
type Revision struct {
	RevID     int        `json:"revid"`
	ParentID  int        `json:"parentid"`
	User      string     `json:"user"`
	UserID    int        `json:"userid"`
	Minor     bool       `json:"minor"`
	Timestamp *time.Time `json:"timestamp"`
	Slots     *Slots     `json:"slots"`
	Comment   string     `json:"comment"`
	Tags      []string   `json:"tags"`
}

// Slots revision slots data structure for Actions API.
type Slots struct {
	Main *Main `json:"main"`
}

// Main revision main slot data structure for Actions API.
type Main struct {
	ContentModel  string `json:"contentmodel"`
	ContentFormat string `json:"contentformat"`
	Content       string `json:"content"`
}

// Normalization shows title normalization in Actions API.
type Normalization struct {
	FromEncoded bool   `json:"fromencoded,omitempty"`
	From        string `json:"from,omitempty"`
	To          string `json:"to,omitempty"`
}

// Page represent page data response in Actions API.
type Page struct {
	PageID               int                       `json:"pageid,omitempty"`
	Title                string                    `json:"title,omitempty"`
	Ns                   int                       `json:"ns,omitempty"`
	WbEntityUsage        map[string]*WbEntityUsage `json:"wbentityusage,omitempty"`
	PageProps            *PageProps                `json:"pageprops,omitempty"`
	Watchers             int                       `json:"watchers"`
	ContentModel         string                    `json:"contentmodel"`
	PageLanguage         string                    `json:"pagelanguage"`
	PageLanguageHTMLCode string                    `json:"pagelanguagehtmlcode"`
	PageLanguageDir      string                    `json:"pagelanguagedir"`
	Touched              *time.Time                `json:"touched"`
	LastRevID            int                       `json:"lastrevid"`
	Length               int                       `json:"length"`
	Missing              bool                      `json:"missing"`
	Protection           []*Protection             `json:"protection"`
	RestrictionTypes     []string                  `json:"restrictiontypes"`
	FullURL              string                    `json:"fullurl"`
	EditURL              string                    `json:"editurl"`
	CanonicalURL         string                    `json:"canonicalurl"`
	DisplayTitle         string                    `json:"displaytitle"`
	Revisions            []*Revision               `json:"revisions"`
	Redirects            []*Redirect               `json:"redirects"`
	Categories           []*Category               `json:"categories"`
	Templates            []*Template               `json:"templates"`
	Flagged              *Flagged                  `json:"flagged"`
	Original             *Image                    `json:"original,omitempty"`
	Thumbnail            *Image                    `json:"thumbnail,omitempty"`
}

// WbEntityUsage represents wikibase entity usage for the page.
type WbEntityUsage struct {
	Aspects []string `json:"aspects,omitempty"`
}

// PageProps represents page properties response in Actions API.
type PageProps struct {
	WikiBaseItem string `json:"wikibase_item,omitempty"`
}

// Project represents a project form Site Matrix Actions API call.
type Project struct {
	URL      string `json:"url"`
	DBName   string `json:"dbname"`
	Code     string `json:"code"`
	SiteName string `json:"sitename"`
	Closed   bool   `json:"closed,omitempty"`
}

// Language represents a language form Site Matrix Actions API call.
type Language struct {
	Code      string     `json:"code"`
	Name      string     `json:"name"`
	Projects  []*Project `json:"site"`
	Dir       string     `json:"dir"`
	LocalName string     `json:"localname"`
}

// Namespace representation of namespace properties in Actions API.
type Namespace struct {
	ID            int    `json:"id"`
	Case          string `json:"case"`
	Name          string `json:"name"`
	SubPages      bool   `json:"subpages"`
	Canonical     string `json:"canonical"`
	Content       bool   `json:"content"`
	NonIncludable bool   `json:"nonincludable"`
}

// User actions API user representation.
type User struct {
	UserID           int           `json:"userid,omitempty"`
	Name             string        `json:"name"`
	EditCount        int           `json:"editcount,omitempty"`
	Registration     *time.Time    `json:"registration,omitempty"`
	Groups           []string      `json:"groups,omitempty"`
	GroupMemberships []interface{} `json:"groupmemberships,omitempty"`
	Emailable        bool          `json:"emailable,omitempty"`
	Missing          bool          `json:"missing,omitempty"`
}

// PageHTML representation of concurrent HTML API response.
type PageHTML struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Error   error  `json:"error"`
}

// Titles represents title properties for page summary.
type Titles struct {
	Canonical  string `json:"canonical"`
	Normalized string `json:"normalized"`
	Display    string `json:"display"`
}

// Titles represents image properties for page summary.
type Image struct {
	Source string `json:"source"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// ContentUrls represents page URLs for page summary.
type ContentUrls struct {
	Page      string `json:"page"`
	Revisions string `json:"revisions"`
	Edit      string `json:"edit"`
	Talk      string `json:"talk"`
}

// Coordinates represents place coordinates for page summary.
type Coordinates struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// PageSummary the content for page summary returned by the REST API.
type PageSummary struct {
	Type              string                  `json:"type"`
	Namespace         *Namespace              `json:"namespace"`
	WikiBaseItem      string                  `json:"wikibase_item"`
	Titles            *Titles                 `json:"titles"`
	PageID            int                     `json:"pageid"`
	Thumbnail         *Image                  `json:"thumbnail"`
	OriginalImage     *Image                  `json:"originalimage"`
	Lang              string                  `json:"lang"`
	Dir               string                  `json:"dir"`
	Revision          string                  `json:"revision"`
	TID               string                  `json:"tid"`
	Timestamp         *time.Time              `json:"timestamp"`
	Description       string                  `json:"description"`
	DescriptionSource string                  `json:"description_source"`
	Coordinates       *Coordinates            `json:"coordinates"`
	ContentUrls       map[string]*ContentUrls `json:"content_urls"`
	Extract           string                  `json:"extract"`
	ExtractHTML       string                  `json:"extract_html"`
}

// LiftWingScore represents model scores for a single revision.
type LiftWingScore struct {
	Prediction  bool                `json:"prediction"`
	Probability *BooleanProbability `json:"probabilities"`
}

// Score is the output response for the LiftWing API.
type Score struct {
	Output *LiftWingScore `json:"output,omitempty"`
	Error  *Error         `json:"error,omitempty"`
}

// BooleanProbability represents probability for boolean values.
type BooleanProbability struct {
	True  float64 `json:"true"`
	False float64 `json:"false"`
}

// NewAPI creates WFM API(s) client under the interface.
func NewAPI(ops ...ClientOption) API {
	return NewClient(ops...)
}

// ClientOption enables optional configuration for the client.
type ClientOption func(*Client)

// NewClient creates WMF API(s) client.
func NewClient(ops ...ClientOption) *Client {
	cl := &Client{
		HTTPClient:         &http.Client{},
		HTTPClientLiftWing: &http.Client{},
		DefaultRetryAfter:  time.Second * 1,
		EnableRetryAfter:   true,
		DefaultURL:         "https://en.wikipedia.org",
		LiftWingBaseURL:    "https://api.wikimedia.org/service/lw/inference/v1/models/",
		DefaultDatabase:    "enwiki",
		UserAgent:          "WME/2.0 (https://enterprise.wikimedia.com/; wme_mgmt@wikimedia.org)",
	}

	// Apply optional configurations
	for _, opt := range ops {
		opt(cl)
	}

	return cl
}

// Client all encompassing client for WFM API(s).
type Client struct {
	HTTPClient         *http.Client
	HTTPClientLiftWing *http.Client
	DefaultURL         string
	LiftWingBaseURL    string
	OAuthToken         string
	DefaultDatabase    string
	UserAgent          string
	DefaultRetryAfter  time.Duration
	EnableRetryAfter   bool
	projects           map[string]*Project
	languages          map[string]*Language
	mutex              sync.Mutex
}

func (c *Client) init(ctx context.Context) error {
	if c.projects == nil || c.languages == nil {
		lns, err := c.GetLanguages(ctx, c.DefaultDatabase)

		if err != nil {
			return err
		}

		c.mutex.Lock()
		defer c.mutex.Unlock()

		c.projects = map[string]*Project{}
		c.languages = map[string]*Language{}

		for _, lng := range lns {
			for _, prj := range lng.Projects {
				c.languages[prj.DBName] = lng
				c.projects[prj.DBName] = prj
			}
		}
	}

	return nil
}

func (c *Client) getProjectURL(ctx context.Context, dtb string) (string, error) {
	if dtb == c.DefaultDatabase {
		return c.DefaultURL, nil
	}

	prj, err := c.GetProject(ctx, dtb)

	if err != nil {
		return "", err
	}

	return prj.URL, nil
}

func (c *Client) newRESTRequest(ctx context.Context, dtb string, path string, qry url.Values) (*http.Request, error) {
	url, err := c.getProjectURL(ctx, dtb)

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/%s?%s", url, path, qry.Encode()), nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", c.UserAgent)

	return req, nil
}

func (c *Client) newActionsRequest(ctx context.Context, dtb string, bdy url.Values) (*http.Request, error) {
	url, err := c.getProjectURL(ctx, dtb)

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/w/api.php", url), strings.NewReader(bdy.Encode()))

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", c.UserAgent)

	return req, nil
}

func (c *Client) do(clt *http.Client, req *http.Request) (*http.Response, error) {
	if clt == nil {
		return nil, ErrHTTPClientNotFound
	}

	res, err := clt.Do(req)

	if err != nil {
		return nil, fmt.Errorf("wmf api call failed for url %s with error %v", req.URL.String(), err)
	}

	if c.EnableRetryAfter && res.StatusCode == http.StatusTooManyRequests {
		if est, _ := getErrorString(res); len(est) > 0 {
			log.Println(est)
		}

		rtv, err := getRetryAfterValue(res, c.DefaultRetryAfter)

		if err != nil {
			return nil, err
		}

		time.Sleep(rtv)

		return c.do(clt, req)
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusFound {
		dta, err := getErrorString(res)

		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("wmf api call returned status not 200 or 302 for url %s with error %s", req.URL.String(), dta)
	}

	return res, nil
}

// GetAllPAges lists all pages in alphabetical order.
// Not that default response includes only limited amount of properties.
// Such as page ID, title and namespace.
// In order to use other namespace than 0 use `ops ...func(*url.Values)` and update `apnamespace` property.
func (c *Client) GetAllPages(ctx context.Context, dtb string, cbk func([]*Page), ops ...func(*url.Values)) error {
	var rsp *Response
	swg := new(sync.WaitGroup)
	pgs := make(chan []*Page, 400000)

	swg.Add(1)
	go func() {
		defer swg.Done()

		for pgs := range pgs {
			cbk(pgs)
		}
	}()

	for {
		bdy := url.Values{}
		bdy.Set("action", "query")
		bdy.Set("list", "allpages")
		bdy.Set("apnamespace", "0")
		bdy.Set("aplimit", "500")
		bdy.Set("apfilterredir", "nonredirects")
		bdy.Set("format", "json")
		bdy.Set("formatversion", "2")

		for _, opt := range ops {
			opt(&bdy)
		}

		if rsp != nil {
			for name, val := range rsp.Continue {
				bdy.Set(name, val)
			}
		}

		req, err := c.newActionsRequest(ctx, dtb, bdy)

		if err != nil {
			return err
		}

		res, err := c.do(c.HTTPClient, req)

		if err != nil {
			return err
		}

		defer res.Body.Close()
		rsp = new(Response)

		if err := json.NewDecoder(res.Body).Decode(rsp); err != nil {
			return err
		}

		if err := getResponseError(rsp); err != nil {
			return fmt.Errorf("%s:%v", http.StatusText(res.StatusCode), err)
		}

		pgs <- rsp.Query.AllPages

		if len(rsp.Continue) == 0 {
			break
		}
	}

	close(pgs)
	swg.Wait()

	return nil
}

// GetPages gets a list of pages from actions API by titles (max batch limit is 50) and merges the request from continuation prop.
// If you need to pass a specific property or update the API request use `ops ...func(*url.Values)` property.
func (c *Client) GetPages(ctx context.Context, dtb string, tls []string, ops ...func(*url.Values)) (map[string]*Page, error) {
	var rsp *Response
	pgs := map[string]*Page{}
	nls := map[string]string{}
	cfg := func(c *mergo.Config) {
		c.AppendSlice = true
	}

	for {
		bdy := url.Values{}
		bdy.Set("action", "query")
		bdy.Set("prop", "info|revisions|wbentityusage|pageprops|redirects|flagged|pageimages")
		bdy.Set("rvprop", "comment|content|ids|timestamp|tags|user|userid|flags")
		bdy.Set("piprop", "thumbnail|original")
		bdy.Set("rvslots", "main")
		bdy.Set("inprop", "displaytitle|protection|url|watchers")
		bdy.Set("ppprop", "wikibase_item")
		bdy.Set("redirects", "1")
		bdy.Set("titles", strings.Join(tls, "|"))
		bdy.Set("format", "json")
		bdy.Set("formatversion", "2")
		bdy.Set("rdlimit", "500")
		bdy.Set("wbeulimit", "500")

		if rsp != nil {
			for name, val := range rsp.Continue {
				bdy.Set(name, val)
			}
		}

		for _, opt := range ops {
			opt(&bdy)
		}

		req, err := c.newActionsRequest(ctx, dtb, bdy)

		if err != nil {
			return nil, err
		}

		res, err := c.do(c.HTTPClient, req)

		if err != nil {
			return nil, err
		}

		defer res.Body.Close()
		rsp = new(Response)

		if err := json.NewDecoder(res.Body).Decode(rsp); err != nil {
			return nil, err
		}

		if err := getResponseError(rsp); err != nil {
			return nil, fmt.Errorf("%s:%v", http.StatusText(res.StatusCode), err)
		}

		if rsp.Query != nil {
			for _, nlz := range rsp.Query.Normalized {
				nls[nlz.To] = nlz.From
			}

			for _, rdr := range rsp.Query.Redirects {
				ttl, ok := nls[rdr.From]

				if ok {
					nls[rdr.To] = ttl
				} else {
					nls[rdr.To] = rdr.From
				}
			}

			for _, page := range rsp.Query.Pages {
				ttl, ok := nls[page.Title]

				if !ok {
					ttl = page.Title
				}

				if ppg := pgs[ttl]; ppg != nil {
					if err := mergo.Merge(ppg, page, cfg); err != nil {
						return nil, err
					}
				} else {
					pgs[ttl] = page
				}
			}
		}

		if len(rsp.Continue) > 0 {
			// this is handling the infinite loops case when we are using
			// `rvlimit` property, cuz it will use `rvcontinue` to go through all revisions
			delete(rsp.Continue, "rvcontinue")

			if _, ok := rsp.Continue["continue"]; ok && len(rsp.Continue) == 1 {
				break
			}
		}

		if len(rsp.Continue) == 0 {
			break
		}
	}

	return pgs, nil
}

// GetPage gets a single page from Actions API.
// Request body can be updated using `ops ...func(*url.Values)` property.
func (c *Client) GetPage(ctx context.Context, dtb string, ttl string, ops ...func(*url.Values)) (*Page, error) {
	pgs, err := c.GetPages(ctx, dtb, []string{ttl}, ops...)

	if err != nil {
		return nil, err
	}

	if pge, ok := pgs[ttl]; ok {
		return pge, nil
	}

	return nil, ErrPageNotFound
}

// GetPageHTML gets HTML of the page using page title.
// Request query can be updated using `ops ...func(*url.Values)` property.
func (c *Client) GetPageHTML(ctx context.Context, dtb string, ttl string, ops ...func(*url.Values)) (string, error) {
	qry := url.Values{}

	for _, opt := range ops {
		opt(&qry)
	}

	// Switch to the new endpoint after the Content transform team fixes all the bugs
	// req, err := c.newRESTRequest(ctx, dtb, fmt.Sprintf("w/rest.php/v1/page/%s/html", url.QueryEscape(strings.ReplaceAll(ttl, " ", "_"))), qry)
	req, err := c.newRESTRequest(ctx, dtb, fmt.Sprintf("api/rest_v1/page/html/%s", url.QueryEscape(strings.ReplaceAll(ttl, " ", "_"))), qry)

	if err != nil {
		return "", err
	}

	res, err := c.do(c.HTTPClient, req)

	if err != nil {
		return "", err
	}

	defer res.Body.Close()
	dta, err := io.ReadAll(res.Body)

	if err != nil {
		return "", err
	}

	return string(dta), nil
}

// GetPagesHTML makes concurrent requests to get HTML of the pages.
// Parameter `mxc int` - is responsible for max amount concurrent requests.
func (c *Client) GetPagesHTML(ctx context.Context, dtb string, tls []string, mxc int, ops ...func(*url.Values)) map[string]*PageHTML {
	tln := len(tls)
	que := make(chan string, tln)
	out := make(chan *PageHTML, tln)

	for i := 0; i < mxc; i++ {
		go func() {
			for ttl := range que {
				cnt, err := c.GetPageHTML(ctx, dtb, ttl, ops...)

				out <- &PageHTML{
					Title:   ttl,
					Content: cnt,
					Error:   err,
				}
			}
		}()
	}

	for _, ttl := range tls {
		que <- ttl
	}

	close(que)
	rsp := map[string]*PageHTML{}

	for i := 0; i < tln; i++ {
		phm := <-out
		rsp[phm.Title] = phm
	}

	return rsp
}

// GetPageSummary gets summary of the page using page title.
// Request query can be updated using `ops ...func(*url.Values)` property.
func (c *Client) GetPageSummary(ctx context.Context, dtb string, ttl string, ops ...func(*url.Values)) (*PageSummary, error) {
	qry := url.Values{}

	for _, opt := range ops {
		opt(&qry)
	}

	req, err := c.newRESTRequest(ctx, dtb, fmt.Sprintf("api/rest_v1/page/summary/%s", url.QueryEscape(strings.ReplaceAll(ttl, " ", "_"))), qry)

	if err != nil {
		return nil, err
	}

	res, err := c.do(c.HTTPClient, req)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	psm := new(PageSummary)

	if err := json.NewDecoder(res.Body).Decode(psm); err != nil {
		return nil, err
	}

	return psm, nil
}

// GetLanguages gets a list of languages from the Actions API using Site Matrix api.
// If you need to pass a specific property or update the API request use `ops ...func(*url.Values)` property.
func (c *Client) GetLanguages(ctx context.Context, dtb string, ops ...func(*url.Values)) ([]*Language, error) {
	bdy := url.Values{}
	bdy.Set("action", "sitematrix")
	bdy.Set("format", "json")
	bdy.Set("formatversion", "2")

	for _, opt := range ops {
		opt(&bdy)
	}

	req, err := c.newActionsRequest(ctx, dtb, bdy)

	if err != nil {
		return nil, err
	}

	res, err := c.do(c.HTTPClient, req)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	rsp := new(Response)

	_ = json.NewDecoder(res.Body).Decode(rsp)

	if err := getResponseError(rsp); err != nil {
		return nil, fmt.Errorf("%s:%v", http.StatusText(res.StatusCode), err)
	}

	lns := []*Language{}

	for num, lng := range rsp.SiteMatrix {
		if num != "count" && num != "specials" {
			lns = append(lns, lng)
		}
	}

	return lns, nil
}

// GetLanguage gets a single language using a database name.
func (c *Client) GetLanguage(ctx context.Context, dtb string) (*Language, error) {
	if err := c.init(ctx); err != nil {
		return nil, err
	}

	if lng, ok := c.languages[dtb]; ok {
		return lng, nil
	}

	return nil, ErrLanguageNotFound
}

// GetProjects gets an array of all projects using database name.
func (c *Client) GetProjects(ctx context.Context, dtb string) ([]*Project, error) {
	if err := c.init(ctx); err != nil {
		return nil, err
	}

	prl := []*Project{}

	if dtb == c.DefaultDatabase {
		for _, prj := range c.projects {
			prl = append(prl, prj)
		}

		return prl, nil
	}

	lgs, err := c.GetLanguages(ctx, dtb)

	if err != nil {
		return nil, err
	}

	for _, lng := range lgs {
		prl = append(prl, lng.Projects...)
	}

	return prl, nil
}

// GetProject gets a single project using a database name.
func (c *Client) GetProject(ctx context.Context, dtb string) (*Project, error) {
	if err := c.init(ctx); err != nil {
		return nil, err
	}

	if prj, ok := c.projects[dtb]; ok {
		return prj, nil
	}

	return nil, ErrProjectNotFound
}

// GetNamespaces returns a list of namespaces supported by the projects.
// If you need to pass a specific property or update the API request use `ops ...func(*url.Values)` property.
func (c *Client) GetNamespaces(ctx context.Context, dtb string, ops ...func(*url.Values)) ([]*Namespace, error) {
	bdy := url.Values{}
	bdy.Set("action", "query")
	bdy.Set("meta", "siteinfo")
	bdy.Set("siprop", "namespaces")
	bdy.Set("format", "json")
	bdy.Set("formatversion", "2")

	for _, opt := range ops {
		opt(&bdy)
	}

	req, err := c.newActionsRequest(ctx, dtb, bdy)

	if err != nil {
		return nil, err
	}

	res, err := c.do(c.HTTPClient, req)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	rsp := new(Response)

	if err := json.NewDecoder(res.Body).Decode(rsp); err != nil {
		return nil, err
	}

	if err := getResponseError(rsp); err != nil {
		return nil, fmt.Errorf("%s:%v", http.StatusText(res.StatusCode), err)
	}

	nss := []*Namespace{}

	for _, nsp := range rsp.Query.Namespaces {
		nss = append(nss, nsp)
	}

	return nss, nil
}

// GetRandomPages returns a list of random article titles from a project. rnlimit should be between 1 to 500.
// If you need to pass a specific property or update the API request use `ops ...func(*url.Values)` property.
func (c *Client) GetRandomPages(ctx context.Context, dtb string, ops ...func(*url.Values)) ([]*Page, error) {
	bdy := url.Values{}
	bdy.Set("action", "query")
	bdy.Set("format", "json")
	bdy.Set("formatversion", "2")
	bdy.Set("list", "random")
	bdy.Set("rnfilterredir", "nonredirects")
	bdy.Set("rnnamespace", "0")
	bdy.Set("rnlimit", "1")

	for _, opt := range ops {
		opt(&bdy)
	}

	req, err := c.newActionsRequest(ctx, dtb, bdy)

	if err != nil {
		return nil, err
	}

	res, err := c.do(c.HTTPClient, req)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	rsp := new(Response)

	if err := json.NewDecoder(res.Body).Decode(rsp); err != nil {
		return nil, err
	}

	if rsp.Error != nil {
		return nil, fmt.Errorf("%s:%v", http.StatusText(res.StatusCode), errors.New(rsp.Error.Info))
	}

	rns := append([]*Page{}, rsp.Query.Random...)

	return rns, nil
}

// GetUsers gets a list of users using identifiers and database name.
// If you need to pass a specific property or update the API request use `ops ...func(*url.Values)` property.
func (c *Client) GetUsers(ctx context.Context, dtb string, ids []int, ops ...func(*url.Values)) (map[int]*User, error) {
	uds := []string{}

	for _, id := range ids {
		uds = append(uds, strconv.Itoa(id))
	}

	bdy := url.Values{}
	bdy.Add("action", "query")
	bdy.Add("list", "users")
	bdy.Add("usprop", "groups|editcount|groupmemberships|registration|emailable")
	bdy.Add("ususerids", strings.Join(uds, "|"))
	bdy.Add("format", "json")
	bdy.Add("formatversion", "2")

	for _, opt := range ops {
		opt(&bdy)
	}

	req, err := c.newActionsRequest(ctx, dtb, bdy)

	if err != nil {
		return nil, err
	}

	res, err := c.do(c.HTTPClient, req)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	rsp := new(Response)

	if err := json.NewDecoder(res.Body).Decode(rsp); err != nil {
		return nil, err
	}

	if err := getResponseError(rsp); err != nil {
		return nil, fmt.Errorf("%s:%v", http.StatusText(res.StatusCode), err)
	}

	uss := map[int]*User{}

	for _, usr := range rsp.Query.Users {
		uss[usr.UserID] = usr
	}

	return uss, nil
}

// GetUser gets a single user using identifiers and database name.
// If you need to pass a specific property or update the API request use `ops ...func(*url.Values)` property.
func (c *Client) GetUser(ctx context.Context, dtb string, id int, ops ...func(*url.Values)) (*User, error) {
	uss, err := c.GetUsers(ctx, dtb, []int{id}, ops...)

	if err != nil {
		return nil, err
	}

	if usr, ok := uss[id]; ok {
		return usr, nil
	}

	return nil, ErrUserNotFound
}

// GetScore gets a single score using revision ID, language, model and project.
func (c *Client) GetScore(ctx context.Context, rev int, lng string, prj string, mdl string) (*Score, error) {
	req, err := c.newLiftWingRequest(ctx, rev, lng, prj, mdl)

	if err != nil {
		return nil, err
	}

	res, err := c.do(c.HTTPClientLiftWing, req)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	scr := &Score{}

	if err := json.NewDecoder(res.Body).Decode(scr); err != nil {
		return nil, err
	}

	return scr, nil
}

func (c *Client) newLiftWingRequest(ctx context.Context, rev int, lng string, prj string, mdl string) (*http.Request, error) {
	bdy := map[string]interface{}{
		"rev_id": rev,
		"lang":   lng,
	}

	url := fmt.Sprintf("%s%s-%s:predict", c.LiftWingBaseURL, prj, mdl)

	if mdl == "revertrisk" {
		url = fmt.Sprintf("%s%s-language-agnostic:predict", c.LiftWingBaseURL, mdl)
	}

	jby, err := json.Marshal(bdy)

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jby))

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.OAuthToken))
	return req, nil
}
