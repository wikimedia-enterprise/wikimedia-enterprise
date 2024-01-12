package aggregate

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"time"
	"wikimedia-enterprise/general/wmf"
)

// ErrScorePrefix is the prefix for the ScoresGetter errors.
var ErrScorePrefix = errors.New("score: ")

// DefaultGetter implements the basic behaviour tha will inject API client.
type DefaultGetter struct {
	API wmf.API
}

// SetAPI sets API client as a dependency for the data getters.
func (d *DefaultGetter) SetAPI(api wmf.API) {
	d.API = api
}

// WithPages initializes Page getter with required arguments.
func WithPages(rvl int, mxc int) *PagesGetter {
	return &PagesGetter{
		RevisionsLimit: rvl,
		MaxConcurrency: mxc,
	}
}

// PagesGetter allows to get page data in the aggregator.
type PagesGetter struct {
	DefaultGetter
	RevisionsLimit int
	MaxConcurrency int
}

// GetData builds up the actions API query to get page metadata.
func (p *PagesGetter) GetData(ctx context.Context, dtb string, tls []string) (interface{}, error) {
	if p.RevisionsLimit < 2 {
		return p.API.GetPages(ctx, dtb, tls)
	}

	qln := len(tls)
	pgs := make(chan *wmf.Page, qln)
	que := make(chan string, qln)
	ers := make(chan error, qln)
	opt := func(v *url.Values) {
		v.Set("rvlimit", strconv.Itoa(p.RevisionsLimit))
	}

	for i := 0; i < p.MaxConcurrency; i++ {
		go func() {
			for ttl := range que {
				res, err := p.API.GetPage(ctx, dtb, ttl, opt)
				pgs <- res
				ers <- err
			}
		}()
	}

	for _, ttl := range tls {
		que <- ttl
	}

	close(que)

	erm := ""

	for i := 0; i < qln; i++ {
		if err := <-ers; err != nil {
			erm += err.Error()
		}
	}

	if len(erm) > 0 {
		return nil, errors.New(erm)
	}

	rvs := map[string]*wmf.Page{}

	for i := 0; i < qln; i++ {
		pge := <-pgs
		rvs[pge.Title] = pge
	}

	return rvs, nil
}

// WithPagesHTML initializes Pages HTML getter with required parameters.
func WithPagesHTML(mxc int) *PagesHTMLGetter {
	return &PagesHTMLGetter{
		MaxConcurrency: mxc,
	}
}

// PagesHTMLGetter integrates Page HTML API requests with concurrency into aggregation.
type PagesHTMLGetter struct {
	DefaultGetter
	MaxConcurrency int
}

// GetData does concurrent Page HTML lookups using actions API.
func (p *PagesHTMLGetter) GetData(ctx context.Context, dtb string, tls []string) (interface{}, error) {
	hts := p.API.GetPagesHTML(ctx, dtb, tls, p.MaxConcurrency)
	msg := ""
	ner := 0

	for _, res := range hts {
		if res.Error != nil {
			msg += res.Error.Error()
			ner++
		}
	}

	if len(msg) > 0 && ner == len(tls) {
		return nil, errors.New(msg)
	}

	return hts, nil
}

// WithRevisions initializes Revisions getter with required parameters.
func WithRevisions(mxc int) *RevisionsGetter {
	return &RevisionsGetter{
		MaxConcurrency: mxc,
	}
}

// RevisionsGetter integrates Revisions API lookups into aggregation.
type RevisionsGetter struct {
	DefaultGetter
	MaxConcurrency int
}

// GetData does concurrent revisions lookup in the Actions API.
func (r *RevisionsGetter) GetData(ctx context.Context, dtb string, tls []string) (interface{}, error) {
	qln := len(tls)
	pgs := make(chan *wmf.Page, qln)
	que := make(chan string, qln)
	ers := make(chan error, qln)

	for i := 0; i < r.MaxConcurrency; i++ {
		go func() {
			for ttl := range que {
				res, err := r.API.GetPage(ctx, dtb, ttl, func(v *url.Values) {
					v.Set("rvdir", "newer")
					v.Set("prop", "revisions")
					v.Set("rvprop", "ids|timestamp")
					v.Set("rvlimit", "1")
					v.Del("ppprop")
					v.Del("inprop")
					v.Del("rdlimit")
					v.Del("rvslots")
					v.Del("wbeulimit")
				})
				pgs <- res
				ers <- err
			}
		}()
	}

	for _, ttl := range tls {
		que <- ttl
	}

	close(que)

	erm := ""

	for i := 0; i < qln; i++ {
		if err := <-ers; err != nil {
			erm += err.Error()
		}
	}

	if len(erm) > 0 {
		return nil, errors.New(erm)
	}

	rvs := map[string]*wmf.Revision{}

	for i := 0; i < qln; i++ {
		pge := <-pgs

		if len(pge.Revisions) > 0 {
			rvs[pge.Title] = pge.Revisions[0]
		} else {
			rvs[pge.Title] = nil
		}
	}

	return rvs, nil
}

// WithUser creates new instance of the Users getter with required params (for single user).
func WithUser(ttl string, uid int) *UsersGetter {
	return &UsersGetter{
		Users: map[string]int{
			ttl: uid,
		},
	}
}

// WithUser creates new instance of the Users getter with required params (for multiple users).
func WithUsers(urs map[string]int) *UsersGetter {
	return &UsersGetter{
		Users: urs,
	}
}

// UsersGetter integrates Actions API users lookup into the aggregation.
type UsersGetter struct {
	DefaultGetter
	Users map[string]int
}

// GetData does concurrent users lookup in Actions API.
func (u *UsersGetter) GetData(ctx context.Context, dtb string, tls []string) (interface{}, error) {
	uis := []int{}

	for _, ttl := range tls {
		uid := u.Users[ttl]

		if uid != 0 {
			uis = append(uis, uid)
		}
	}

	sort.Ints(uis)

	urs, err := u.API.GetUsers(ctx, dtb, uis)

	if err != nil {
		return nil, err
	}

	rur := map[string]*wmf.User{}

	for _, ttl := range tls {
		uid := u.Users[ttl]

		if uid != 0 {
			rur[ttl] = urs[uid]
		}
	}

	return rur, nil
}

// WithScore creates new instance of the Score getter with required params.
func WithScore(rid int, lng string, ttl string, mdl string, prj string) *ScoreGetter {
	return &ScoreGetter{
		Revision: map[string]int{
			ttl: rid,
		},
		Language:       lng,
		Model:          mdl,
		RequestTimeout: time.Millisecond * 500,
		Project:        prj,
	}
}

// ScoreGetter stores required params for liftwing score lookup.
type ScoreGetter struct {
	DefaultGetter
	Language       string
	Revision       map[string]int
	Model          string
	Project        string
	RequestTimeout time.Duration
}

// GetData builds up the LiftWing API query to get score metadata.
func (s *ScoreGetter) GetData(ctx context.Context, dtb string, _ []string) (interface{}, error) {
	scs := map[string]*wmf.Score{}

	ctx, cancel := context.WithTimeout(ctx, s.RequestTimeout)
	defer cancel()

	for ttl, rid := range s.Revision {
		scr, err := s.API.GetScore(ctx, rid, s.Language, s.Project, s.Model)

		if err != nil {
			return nil, fmt.Errorf("%s%s", ErrScorePrefix, err)
		}

		scs[ttl] = scr
	}

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("%s%s", ErrScorePrefix, ctx.Err())
	default:
		return scs, nil
	}
}
