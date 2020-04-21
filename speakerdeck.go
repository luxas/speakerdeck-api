package speakerdeck

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"
	"github.com/luxas/speakerdeck-scraper/scraper"
	log "github.com/sirupsen/logrus"
)

// TODO: Split this file into user.go, talk.go and common.go

const (
	speakerdeckRootURL = "https://speakerdeck.com"
	httpsPrefix        = "https:"
)

var linkRegexp = regexp.MustCompile(`http[s]?://[a-zA-Z-_/0-9\.#=&]*`)

func sdPrefix(s string) string {
	return fmt.Sprintf("%s%s", speakerdeckRootURL, s)
}

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

// ScrapeTalk returns either one sepecific talk if both userHandle and talkID are set, or a set of
// all the users' talks in detail if only userHandle is set. In opts you can set extensions
func ScrapeTalk(userHandle, talkID string, opts *scraper.ScrapeOptions) (Talks, error) {
	if len(userHandle) == 0 {
		return nil, fmt.Errorf("userHandle is mandatory!")
	}

	// If there was a specific talk given, look it up
	if len(talkID) > 0 {
		talkURL := fmt.Sprintf("%s/%s/%s", speakerdeckRootURL, userHandle, talkID)
		data, err := scraper.Scrape(talkURL, &TalkScraper{}, opts)
		if err != nil {
			return nil, err
		}

		talk := data.(*Talk)
		return []Talk{*talk}, nil
	}

	user, err := ScrapeUser(userHandle, opts)
	if err != nil {
		return nil, err
	}
	wg := &sync.WaitGroup{}
	wg.Add(len(user.TalkPreviews))

	mux := &sync.Mutex{}
	talks := make([]Talk, 0, len(user.TalkPreviews))

	for _, t := range user.TalkPreviews {
		go func(talkPreview TalkPreview) {
			defer wg.Done()

			talkList, err := ScrapeTalk(user.Author.Handle, talkPreview.ID, opts)
			if err != nil {
				log.Errorf("could not get speakerdeck talk %s/%s", user.Author.Handle, talkPreview.ID)
				return
			}
			mux.Lock()
			talks = append(talks, talkList...)
			mux.Unlock()
		}(t)
	}
	wg.Wait()

	sortedTalks := Talks(talks)
	sort.Sort(sortedTalks)

	return sortedTalks, nil
}

// TalkScraper implements scraper.Scraper
type TalkScraper struct{}

// Name returns the name of the TalkScraper
func (s *TalkScraper) Name() string {
	return "TalkScraper"
}

// Hooks returns mappings between DOM paths in the scraped web pages, and handler functions to extract data out
// of them
func (s *TalkScraper) Hooks() []scraper.Hook {
	return []scraper.Hook{
		{
			DOMPath: ".container h1.mb-4",
			Handler: onTalkTitle,
		},
		{
			DOMPath: ".col-auto.text-muted",
			Handler: onTalkDate,
		},
		{
			DOMPath: ".deck-description.mb-4 p",
			Handler: onTalkDescription,
		},
		{
			DOMPath: ".speakerdeck-embed",
			Handler: onTalkDataID,
		},
		{
			DOMPath: ".deck-meta .col-md-auto .row > div:nth-child(1) a",
			Handler: onTalkCategory,
		},
		{
			DOMPath: ".deck-meta .col-md-auto .row > div:nth-child(2) a",
			Handler: onTalkStars,
		},
		{
			DOMPath: ".deck-meta .col-md-auto .row > div:nth-child(3) span[title]",
			Handler: onTalkViews,
		},
		{
			DOMPath: ".deck-meta .col-md-auto .row > div:nth-child(4) a",
			Handler: onTalkDownloadLink,
		},
		{
			DOMPath: ".deck-meta .col-md-auto .row > a:nth-child(1)",
			Handler: onTalkAuthor,
		},
	}
}

// InitialData returns the struct pointer passed around between the handler functions registered in Hooks()
// This pointer is passed as the second argument to all handlers. The handlers can cast it from interface{}
// to its real type, and modify its data.
func (s *TalkScraper) InitialData() interface{} {
	return NewTalk()
}

func onTalkTitle(e *colly.HTMLElement, data interface{}) (*string, error) {
	t := data.(*Talk)
	t.Title = e.Text
	return nil, nil
}

func onTalkDataID(e *colly.HTMLElement, data interface{}) (*string, error) {
	t := data.(*Talk)
	t.DataID = e.Attr("data-id")
	return nil, nil
}

func onTalkDate(e *colly.HTMLElement, data interface{}) (*string, error) {
	t := data.(*Talk)

	d, err := parseDate(e.Text)
	if err != nil {
		return nil, err
	}
	t.Date = d
	return nil, nil
}

func parseDate(dateStr string) (time.Time, error) {
	// sanitize the text
	dateStr = strings.Trim(strings.ReplaceAll(strings.ReplaceAll(dateStr, ",", ""), "\n", ""), " ")
	// and parse it
	t, err := time.Parse("January 02 2006", dateStr)
	if err == nil {
		return t, nil
	}
	return time.Parse("Jan 2 2006", dateStr)
}

func onTalkDescription(e *colly.HTMLElement, data interface{}) (*string, error) {
	t := data.(*Talk)
	links := linkRegexp.FindStringSubmatch(e.Text)
	for _, link := range links {
		parsedLink, err := url.Parse(link)
		if err != nil {
			log.Warnf("Could not parse link %q", link)
			continue
		}
		t.ExtraLinks[parsedLink.Host] = append(t.ExtraLinks[parsedLink.Host], parsedLink.String())
	}

	if strings.Contains(e.Text, "Hide: true") {
		t.Hide = true
	}

	return nil, nil
}

func onTalkCategory(e *colly.HTMLElement, data interface{}) (*string, error) {
	t := data.(*Talk)
	t.CategoryLink = sdPrefix(e.Attr("href"))
	t.Category = strings.TrimSpace(e.Text)
	return nil, nil
}

func onTalkStars(e *colly.HTMLElement, data interface{}) (*string, error) {
	t := data.(*Talk)

	var err error
	t.Stars, err = parseNumber(e.Text)
	return nil, err
}

func onTalkViews(e *colly.HTMLElement, data interface{}) (*string, error) {
	t := data.(*Talk)

	viewsStr := strings.TrimSuffix(e.Attr("title"), " views")
	var err error
	t.Views, err = parseNumber(viewsStr)
	return nil, err
}

func parseNumber(numstr string) (uint32, error) {
	numstr = strings.TrimSpace(numstr)
	numstr = strings.ReplaceAll(numstr, ",", "")
	if len(numstr) == 0 {
		return 0, nil
	}

	// TODO: Handle this in a more sophisticated way :)?
	multiplier := float64(1)
	if strings.Contains(numstr, "k") {
		multiplier = 1000
		numstr = strings.ReplaceAll(numstr, "k", "")
	}

	n, err := strconv.ParseFloat(numstr, 64)
	if err != nil {
		return 0, err
	}
	return uint32(multiplier * n), nil
}

func onTalkDownloadLink(e *colly.HTMLElement, data interface{}) (*string, error) {
	t := data.(*Talk)
	t.DownloadLink = e.Attr("href")
	return nil, nil
}

func onTalkAuthor(e *colly.HTMLElement, data interface{}) (*string, error) {
	t := data.(*Talk)
	t.Link = e.Request.URL.String()
	t.ID = path.Base(t.Link)
	t.Author.Link = sdPrefix(e.Attr("href"))
	t.Author.Handle = path.Base(t.Author.Link)
	t.Author.Name = strings.TrimSpace(e.Text)
	t.Author.AvatarLink = httpsPrefix + e.ChildAttr("img", "src")
	return nil, nil
}
