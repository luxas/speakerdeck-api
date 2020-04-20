/*
The scraper package contains generic, high-level scraper functionality built on top of github.com/gocolly/colly
In order to use it, create a struct (e.g. MyScraper) that embeds the BaseScraper, and implements the Scraper
interface, e.g.

	type MyScraper struct {}

	func (ms *MyScraper) InitialData() interface{} {
		return &MyScrapedData{}
	}

	func (ms *MyScraper) Hooks() []scraper.Hook {
		return []scraper.Hook{
			{
				DOMPath: "#my-awesome-element",
				Handler: extractImportantMessage,
			},
		}
	}

	func extractImportantMessage(e *colly.HTMLElement, data interface{}) (*string, error) {
		myData := data.(*MyScrapedData)
		myData.ImportantMessage = e.Text
		return nil, nil
	}

	type MyScrapedData struct {
		ImportantMessage string
	}

	func main() {
		s := &MyScraper{}
		data, err := scraper.Scrape("example.com", s, nil)
		if err != nil {
			log.Fatal(err)
		}
		myData := data.(*MyScrapedData)
		log.Printf("Important message is: %s", myData.ImportantMessage)
	}

*/
package scraper

import (
	"fmt"
	"sync"

	"github.com/gocolly/colly"
	log "github.com/sirupsen/logrus"
)

type HookFn func(e *colly.HTMLElement, data interface{}) (*string, error)

type Hook struct {
	DOMPath string
	Handler HookFn
}

type Scraper interface {
	Name() string

	Hooks() []Hook

	InitialData() interface{}
}

type Extension interface {
	Name() string
	Hook() Hook
}

type ScrapeOptions struct {
	Extensions []Extension
	LogLevel   *log.Level
}

func Scrape(url string, s Scraper, opts *ScrapeOptions) (interface{}, error) {
	c := colly.NewCollector()
	mux := &sync.Mutex{}
	data := s.InitialData()
	logger := log.New()
	allHooks := s.Hooks()

	if opts != nil {
		if opts.LogLevel != nil {
			logger.SetLevel(*opts.LogLevel)
		}

		for _, ext := range opts.Extensions {
			allHooks = append(allHooks, ext.Hook())
		}
	}

	errs := []error{}
	for _, h := range allHooks {
		func(hook Hook) {
			c.OnHTML(hook.DOMPath, func(e *colly.HTMLElement) {
				mux.Lock()

				logger.Debugf("DOMPath: %q, URL: %q", hook.DOMPath, e.Request.URL)
				next, err := hook.Handler(e, data)
				if err != nil {
					logger.Errorf("error while handling dompath %q for request %q: %v", hook.DOMPath, e.Request.URL, err)
					errs = append(errs, err)
				}
				mux.Unlock()

				if next != nil {
					c.Visit(*next)
				}
			})
		}(h)
	}
	c.OnRequest(func(r *colly.Request) {
		logger.Infof("%s visiting page %q", s.Name(), r.URL)
	})

	if err := c.Visit(url); err != nil {
		return nil, err
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("errors occured during scraping: %v", errs)
	}
	return data, nil
}
