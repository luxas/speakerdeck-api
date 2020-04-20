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
	"github.com/luxas/speakerdeck-scraper/types"
	log "github.com/sirupsen/logrus"
)

// TODO: Split this file into user.go, talk.go and common.go

const (
	SpeakerdeckRootURL = "https://speakerdeck.com"
	httpsPrefix        = "https:"
)

func sdPrefix(s string) string {
	return fmt.Sprintf("%s%s", SpeakerdeckRootURL, s)
}

var linkRegexp = regexp.MustCompile(`http[s]?://[a-zA-Z-_/0-9\.#=&]*`)

var _ scraper.Scraper = &UserScraper{}

func NewUserScraper() *UserScraper {
	return &UserScraper{}
}

type UserScraper struct{}

func (s *UserScraper) Name() string {
	return "UserScraper"
}

func (s *UserScraper) ScrapeUser(userID string, opts *scraper.ScrapeOptions) (*types.User, error) {
	fullURL := fmt.Sprintf("%s/%s", SpeakerdeckRootURL, userID)

	data, err := scraper.Scrape(fullURL, s, opts)
	if err != nil {
		return nil, err
	}
	user := data.(*types.User)
	sort.Sort(user.TalkPreviews)
	return user, nil
}

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

func (s *UserScraper) InitialData() interface{} {
	return types.NewUser()
}

func onUserAuthor(e *colly.HTMLElement, data interface{}) (*string, error) {
	u := data.(*types.User)
	u.Author.Link = e.Request.URL.String()
	u.Author.Name = e.ChildText("h1.m-0")
	u.Author.Handle = e.ChildText("div.text-muted")
	u.Author.AvatarLink = httpsPrefix + e.ChildAttr("img", "src")
	return nil, nil
}

func onUserAbstract(e *colly.HTMLElement, data interface{}) (*string, error) {
	u := data.(*types.User)
	u.Abstract = e.Text
	return nil, nil
}

func onUserTalkFound(e *colly.HTMLElement, data interface{}) (*string, error) {
	u := data.(*types.User)

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

	t := types.TalkPreview{
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

func NewTalkScraper() *TalkScraper {
	return &TalkScraper{}
}

type TalkScraper struct{}

func (s *TalkScraper) Name() string {
	return "TalkScraper"
}

func (s *TalkScraper) ScrapeTalk(userID, talkID string, opts *scraper.ScrapeOptions) (types.Talks, error) {
	if len(userID) == 0 {
		return nil, fmt.Errorf("userID is mandatory!")
	}

	if len(talkID) > 0 {
		talkURL := fmt.Sprintf("%s/%s/%s", SpeakerdeckRootURL, userID, talkID)
		data, err := scraper.Scrape(talkURL, s, opts)
		if err != nil {
			return nil, err
		}

		talk := data.(*types.Talk)
		return []types.Talk{*talk}, nil
	}

	us := NewUserScraper()
	user, err := us.ScrapeUser(userID, opts)
	if err != nil {
		return nil, err
	}
	wg := &sync.WaitGroup{}
	wg.Add(len(user.TalkPreviews))

	mux := &sync.Mutex{}
	talks := make([]types.Talk, 0, len(user.TalkPreviews))

	for _, t := range user.TalkPreviews {
		go func(talkPreview types.TalkPreview) {
			defer wg.Done()

			talkList, err := s.ScrapeTalk(user.Author.Handle, talkPreview.ID, opts)
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

	sortedTalks := types.Talks(talks)
	sort.Sort(sortedTalks)

	return sortedTalks, nil
}

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

func (s *TalkScraper) InitialData() interface{} {
	return types.NewTalk()
}

func onTalkTitle(e *colly.HTMLElement, data interface{}) (*string, error) {
	t := data.(*types.Talk)
	t.Title = e.Text
	return nil, nil
}

func onTalkDataID(e *colly.HTMLElement, data interface{}) (*string, error) {
	t := data.(*types.Talk)
	t.DataID = e.Attr("data-id")
	return nil, nil
}

func onTalkDate(e *colly.HTMLElement, data interface{}) (*string, error) {
	t := data.(*types.Talk)

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
	t := data.(*types.Talk)
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
	t := data.(*types.Talk)
	t.CategoryLink = sdPrefix(e.Attr("href"))
	t.Category = strings.TrimSpace(e.Text)
	return nil, nil
}

func onTalkStars(e *colly.HTMLElement, data interface{}) (*string, error) {
	t := data.(*types.Talk)

	var err error
	t.Stars, err = parseNumber(e.Text)
	return nil, err
}

func onTalkViews(e *colly.HTMLElement, data interface{}) (*string, error) {
	t := data.(*types.Talk)

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
	t := data.(*types.Talk)
	t.DownloadLink = e.Attr("href")
	return nil, nil
}

func onTalkAuthor(e *colly.HTMLElement, data interface{}) (*string, error) {
	t := data.(*types.Talk)
	t.Link = e.Request.URL.String()
	t.ID = path.Base(t.Link)
	t.Author.Link = sdPrefix(e.Attr("href"))
	t.Author.Handle = path.Base(t.Author.Link)
	t.Author.Name = strings.TrimSpace(e.Text)
	t.Author.AvatarLink = httpsPrefix + e.ChildAttr("img", "src")
	return nil, nil
}