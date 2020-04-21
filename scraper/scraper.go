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

// HookFn is a callback function for processing HTML data at a given place in the DOM tree
// The first e argument gives access to the DOM, and the second data argument carries a pointer
// to the data struct you want to save important information in. You can cast data to what's
// returned by Scraper.InitialData(). The return values are an optional string which tells the
// scraper to also scrape an other page, and an error.
type HookFn func(e *colly.HTMLElement, data interface{}) (*string, error)

// Hook maps a handler of type HookFn to a DOMPath in the tree. The DOMPath can be any valid CSS selector.
type Hook struct {
	// DOMPath specifies one or many elements in the DOM tree using a CSS selector
	DOMPath string

	// Handler specifies the handler to be invoked for all of the elements on the HTML page matched by the CSS selector
	Handler HookFn
}

// Scraper is an interface which scraping implementations should implement.
// Any struct that satisfies this interface, may be passed to the generic Scrape function in this package.
type Scraper interface {
	// Name returns an user-friendly name of the scraper
	Name() string

	// Hooks returns the hooks for all HTML elements that should be matched and their handlers.
	Hooks() []Hook

	// InitialData returns the struct pointer which is then shared between/passed to all hook handlers.
	InitialData() interface{}
}

// Extension is an interface which allows for adding extensions on-demand to scraping implementations.
// Upon calling Scrape(), you may pass extra extension implementations in ScrapeOptions. The extension
// can register its own extra hook for processing the DOM. The extension shares/manipulates the same
// data as the Scraper it's used together with.
type Extension interface {
	// Name returns the name of the extension
	Name() string

	// Hook is the hook registered by this extension
	Hook() Hook
}

// ScrapeOptions contains extra parameters used when scraping
type ScrapeOptions struct {
	// Extensions allows registering extensions to a Scrape() call
	Extensions []Extension
	// LogLevel specifies the logrus log level for the Scrape() function
	LogLevel *log.Level
}

// Scrape takes in a Scraper struct, an URL to scrape, and optionally extra options.
// This function calls handlers from the the Scraper.Hooks() for the given DOM paths, and
// shares the Scraper.InitialData() struct pointer between them. The return value is that
// struct pointer, and/or possibly an error.
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
