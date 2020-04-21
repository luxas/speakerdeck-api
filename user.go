package speakerdeck

import (
	"fmt"
	"path"
	"sort"

	"github.com/gocolly/colly"
	"github.com/luxas/speakerdeck-api/scraper"
)

// ScrapeUser returns an user object based on the given user handle. In opts,
// you may specify possible scraping extensions, or log levels.
func ScrapeUser(userHandle string, opts *scraper.ScrapeOptions) (*User, error) {
	if len(userHandle) == 0 {
		return nil, fmt.Errorf("userHandle is mandatory!")
	}

	fullURL := fmt.Sprintf("%s/%s", speakerdeckRootURL, userHandle)

	data, err := scraper.Scrape(fullURL, &UserScraper{}, opts)
	if err != nil {
		return nil, err
	}
	user := data.(*User)
	sort.Sort(user.TalkPreviews)
	return user, nil
}

var _ scraper.Scraper = &UserScraper{}

// UserScraper implements scraper.Scraper
type UserScraper struct{}

// Name returns the name of the UserScraper
func (s *UserScraper) Name() string {
	return "UserScraper"
}

// Hooks returns mappings between DOM paths in the scraped web pages, and handler functions to extract data out
// of them
func (s *UserScraper) Hooks() []scraper.Hook {
	return []scraper.Hook{
		{
			DOMPath: ".sd-main > :first-child .row",
			Handler: onUserAuthor,
		},
		{
			DOMPath: ".deck-description p",
			Handler: onUserAbstract,
		},
		{
			DOMPath: ".container a[href][title]",
			Handler: onUserTalkFound,
		},
		{
			DOMPath: ".next .page-link[rel='next']",
			Handler: onUserNextPage,
		},
	}
}

// InitialData returns the struct pointer passed around between the handler functions registered in Hooks()
// This pointer is passed as the second argument to all handlers. The handlers can cast it from interface{}
// to its real type, and modify its data.
func (s *UserScraper) InitialData() interface{} {
	return NewUser()
}

func onUserAuthor(e *colly.HTMLElement, data interface{}) (*string, error) {
	u := data.(*User)
	u.Author.Link = e.Request.URL.String()
	u.Author.Name = e.ChildText("h1.m-0")
	u.Author.Handle = e.ChildText("div.text-muted")
	u.Author.AvatarLink = httpsPrefix + e.ChildAttr("img", "src")
	return nil, nil
}

func onUserAbstract(e *colly.HTMLElement, data interface{}) (*string, error) {
	u := data.(*User)
	u.Abstract = e.Text
	return nil, nil
}

func onUserTalkFound(e *colly.HTMLElement, data interface{}) (*string, error) {
	u := data.(*User)

	d, err := parseDate(e.ChildText(".deck-preview-meta > :nth-child(1)"))
	if err != nil {
		return nil, err
	}
	stars, err := parseNumber(e.ChildText(".deck-preview-meta > :nth-child(2)"))
	if err != nil {
		return nil, err
	}
	views, err := parseNumber(e.ChildText(".deck-preview-meta > :nth-child(3)"))
	if err != nil {
		return nil, err
	}

	t := TalkPreview{
		Title:  e.Attr("title"),
		Link:   sdPrefix(e.Attr("href")),
		DataID: e.ChildAttr("div.deck-preview", "data-id"),
		Date:   d,
		Views:  views,
		Stars:  stars,
	}
	t.ID = path.Base(t.Link)

	u.TalkPreviews = append(u.TalkPreviews, t)
	return nil, nil
}

func onUserNextPage(e *colly.HTMLElement, _ interface{}) (*string, error) {
	href := e.Attr("href")
	if len(href) > 0 {
		nextURL := sdPrefix(e.Attr("href"))
		return &nextURL, nil
	}
	return nil, nil
}
