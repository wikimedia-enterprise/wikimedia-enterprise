package wmf

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type FakePage struct {
	Title                  string
	HTML                   string
	Description            string
	RevID                  int
	PageID                 int
	AnonymousContributors  int
	RegisteredContributors int
}

func (p *FakePage) asPage() *Page {
	contrib := []*User{}
	for range p.RegisteredContributors {
		contrib = append(contrib, &User{})
	}
	return &Page{
		Title:                 p.Title,
		PageID:                p.PageID,
		AnonymousContributors: p.AnonymousContributors,
		Contributors:          contrib,
	}
}

func (p *FakePage) asPageHTML() *PageHTML {
	return &PageHTML{
		Title:   p.Title,
		Content: p.HTML,
	}
}

func (p *FakePage) asRevision() *Revision {
	return &Revision{
		RevID: p.RevID,
	}
}

func (p *FakePage) asPageSummary() *PageSummary {
	return &PageSummary{
		Titles:      &Titles{Canonical: p.Title, Normalized: p.Title, Display: p.Title},
		PageID:      p.PageID,
		Revision:    strconv.FormatInt(int64(p.RevID), 10),
		Description: p.Description,
	}
}

type APIFake struct {
	pages      []*FakePage
	users      map[int]*User
	languages  map[string]*Language
	projects   map[string]*Project
	namespaces []*Namespace
	files      map[string][]byte
}

func NewAPIFake() *APIFake {
	return &APIFake{
		pages:     make([]*FakePage, 0),
		users:     make(map[int]*User),
		languages: make(map[string]*Language),
		projects:  make(map[string]*Project),
		files:     make(map[string][]byte),
	}
}

// Convenience function for dependency injection.
func (f *APIFake) AsAPI() API {
	return f
}

func (f *APIFake) InsertPage(page *FakePage) {
	f.pages = append(f.pages, page)
}
func (f *APIFake) InsertUser(id int, user *User) {
	f.users[id] = user
}
func (f *APIFake) InsertLanguage(code string, lang *Language) {
	f.languages[code] = lang
}
func (f *APIFake) InsertProject(name string, prj *Project) {
	f.projects[name] = prj
}
func (f *APIFake) InsertNamespace(ns *Namespace) {
	f.namespaces = append(f.namespaces, ns)
}
func (f *APIFake) InsertFile(url string, data []byte) {
	f.files[url] = data
}

func (f *APIFake) GetPage(ctx context.Context, dtb, ttl string, ops ...func(*url.Values)) (*Page, error) {
	for _, p := range f.pages {
		if p.Title == ttl {
			return p.asPage(), nil
		}
	}

	return nil, errors.New("page not found")
}

func (f *APIFake) GetPages(ctx context.Context, dtb string, tls []string, ops ...func(*url.Values)) (map[string]*Page, error) {
	pages := make(map[string]*Page)
	for _, t := range tls {
		for _, p := range f.pages {
			if p.Title == t {
				pages[t] = p.asPage()
			}
		}
	}

	return pages, nil
}

func (f *APIFake) GetPageHTML(ctx context.Context, dtb, ttl string, ops ...func(*url.Values)) (string, error) {
	for _, p := range f.pages {
		if p.Title == ttl {
			return p.HTML, nil
		}
	}

	return "", errors.New("page not found")
}

func (f *APIFake) GetPagesHTML(ctx context.Context, dtb string, tls []string, mxc int, ops ...func(*url.Values)) map[string]*PageHTML {
	pages := make(map[string]*PageHTML)
	for _, t := range tls {
		for _, p := range f.pages {
			if p.Title == t {
				pages[t] = p.asPageHTML()
			}
		}
	}

	return pages
}

func (f *APIFake) GetRevisionHTML(ctx context.Context, dtb, rid string, ops ...func(*url.Values)) (string, error) {
	revid, err := strconv.ParseInt(rid, 10, 64)
	if err != nil {
		return "", errors.New("invalid revid")
	}
	for _, p := range f.pages {
		if p.RevID == int(revid) {
			return p.HTML, nil
		}
	}

	return "", errors.New("revision not found")
}

func (f *APIFake) GetRevisionsHTML(ctx context.Context, dtb string, rvs []string, mxc int, ops ...func(*url.Values)) map[string]*PageHTML {
	pages := make(map[string]*PageHTML)
	for _, rid := range rvs {
		revid, err := strconv.ParseInt(rid, 10, 64)
		if err != nil {
			continue
		}
		for _, p := range f.pages {
			if p.RevID == int(revid) {
				pages[rid] = p.asPageHTML()
			}
		}
	}

	return pages
}

func (f *APIFake) GetPageByRevision(ctx context.Context, dtb string, revid int, ops ...func(*url.Values)) (*Page, error) {
	for _, p := range f.pages {
		if p.RevID == revid {
			return p.asPage(), nil
		}
	}

	return nil, errors.New("page not found")
}

func (f *APIFake) GetAllRevisions(ctx context.Context, dtb string, pageid int, ops ...func(*url.Values)) ([]*Revision, error) {
	result := []*Revision{}
	for _, p := range f.pages {
		if p.PageID == pageid {
			result = append(result, p.asRevision())
		}
	}
	return result, nil
}

func (f *APIFake) GetLanguages(ctx context.Context, dtb string, ops ...func(*url.Values)) ([]*Language, error) {
	var result []*Language
	for _, l := range f.languages {
		result = append(result, l)
	}
	return result, nil
}

func (f *APIFake) GetLanguage(ctx context.Context, dtb string) (*Language, error) {
	l, ok := f.languages[dtb]
	if !ok {
		return nil, errors.New("language not found")
	}
	return l, nil
}

func (f *APIFake) GetProjects(ctx context.Context, dtb string) ([]*Project, error) {
	var result []*Project
	for _, p := range f.projects {
		result = append(result, p)
	}
	return result, nil
}

func (f *APIFake) GetProject(ctx context.Context, dtb string) (*Project, error) {
	p, ok := f.projects[dtb]
	if !ok {
		return nil, errors.New("project not found")
	}
	return p, nil
}

func (f *APIFake) GetNamespaces(ctx context.Context, dtb string, ops ...func(*url.Values)) ([]*Namespace, error) {
	return f.namespaces, nil
}

func (f *APIFake) GetRandomPages(ctx context.Context, dtb string, ops ...func(*url.Values)) ([]*Page, error) {
	var result []*Page
	for _, p := range f.pages {
		result = append(result, p.asPage())
	}
	return result, nil
}

func (f *APIFake) GetUsers(ctx context.Context, dtb string, ids []int, ops ...func(*url.Values)) (map[int]*User, error) {
	result := map[int]*User{}
	for _, id := range ids {
		if u, ok := f.users[id]; ok {
			result[id] = u
		}
	}
	return result, nil
}

func (f *APIFake) GetUser(ctx context.Context, dtb string, id int, ops ...func(*url.Values)) (*User, error) {
	u, ok := f.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return u, nil
}

func (f *APIFake) GetScore(ctx context.Context, rev int, lng, prj, mdl string) (*Score, error) {
	return &Score{}, nil
}

func (f *APIFake) GetReferenceNeedScore(ctx context.Context, rev int, lng, prj string) (*ReferenceNeedScore, error) {
	return &ReferenceNeedScore{}, nil
}

func (f *APIFake) GetReferenceRiskScore(ctx context.Context, rev int, lng, prj string) (*ReferenceRiskScore, error) {
	return &ReferenceRiskScore{}, nil
}

func (f *APIFake) GetPageSummary(ctx context.Context, dtb, ttl string, ops ...func(*url.Values)) (*PageSummary, error) {
	for _, p := range f.pages {
		if p.Title == ttl {
			return p.asPageSummary(), nil
		}
	}

	return nil, errors.New("page not found")
}

func (f *APIFake) GetContributors(ctx context.Context, dtb string, pageid int, maxRegistered *int, ops ...func(*url.Values)) ([]*Page, bool, error) {
	pages := []*Page{}
	for _, p := range f.pages {
		if p.PageID == pageid {
			pages = append(pages, p.asPage())
		}
	}

	if len(pages) == 0 {
		return nil, false, errors.New("page not found")
	}

	return pages, true, nil
}

func (f *APIFake) GetContributorsCount(ctx context.Context, dtb string, pageid int, maxRegistered *int, ops ...func(*url.Values)) (*ContributorsCount, error) {
	pages, fetchedAll, err := f.GetContributors(ctx, dtb, pageid, maxRegistered, ops...)
	if err != nil {
		return nil, err
	}

	anon := 0
	reg := 0
	for _, page := range pages {
		// Don't add anonymous, every page has the full count.
		anon = page.AnonymousContributors

		reg += len(page.Contributors)
	}

	return &ContributorsCount{
		AnonymousCount:        anon,
		RegisteredCount:       reg,
		AllRegisteredIncluded: fetchedAll,
	}, nil
}

func (f *APIFake) DownloadFile(ctx context.Context, url string, ops ...func(*http.Request)) ([]byte, error) {
	data, ok := f.files[url]
	if !ok {
		return nil, fmt.Errorf("file not found")
	}
	return data, nil
}

func (f *APIFake) HeadFile(ctx context.Context, url string, ops ...func(*http.Request)) ([]byte, error) {
	return f.DownloadFile(ctx, url, ops...)
}

func (f *APIFake) GetAllPages(ctx context.Context, dtb string, cbk func([]*Page), ops ...func(*url.Values)) error {
	var result []*Page
	for _, p := range f.pages {
		result = append(result, p.asPage())
	}
	cbk(result)
	return nil
}
