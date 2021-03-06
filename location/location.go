package location

import (
	"context"
	"regexp"
	"strings"

	"github.com/gocolly/colly"
	"github.com/luxas/speakerdeck-api"
	"github.com/luxas/speakerdeck-api/scraper"
	log "github.com/sirupsen/logrus"
	"googlemaps.github.io/maps"
)

var locationRegexp = regexp.MustCompile(`Location: (.*)`)

var _ scraper.Extension = &LocationExtension{}

// NewLocationExtension creates a new LocationExtension using a Google Maps API Key with access
// to the Geocoding API.
func NewLocationExtension(apiKey string) (*LocationExtension, error) {
	c, err := maps.NewClient(maps.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	return &LocationExtension{c}, nil
}

// LocationExtension implements scraper.Extension, and adds geolocation features to the TalkScraper
// LocationExtension only works together with speakerdeck.TalkScraper at the moment.
type LocationExtension struct {
	c *maps.Client
}

// Name returns the LocationExtension name
func (_ *LocationExtension) Name() string {
	return "LocationExtension"
}

// Hook returns the hook for this extension
func (le *LocationExtension) Hook() scraper.Hook {
	return scraper.Hook{
		DOMPath: ".deck-description.mb-4 p",
		Handler: le.onDescription,
	}
}

// onDescription processes the location given in the Talk description field, and registers the geocoded
// response to the Talk object.
func (le *LocationExtension) onDescription(e *colly.HTMLElement, data interface{}) (*string, error) {
	// Fail fast, only consider descriptions with the "Location" substring
	if !strings.Contains(e.Text, "Location") {
		return nil, nil
	}

	switch data.(type) {
	case *speakerdeck.Talk:
		// noop, allow this
	default:
		return nil, nil
	}

	t := data.(*speakerdeck.Talk)

	if strings.Contains(e.Text, "Location: Online") {
		t.Location = &speakerdeck.Location{
			RequestedAddress: "Online",
		}
		return nil, nil
	}

	locationStr := locationRegexp.FindStringSubmatch(e.Text)
	if len(locationStr) != 2 {
		log.Warnf("Couldn't find location for talk %s: %v %q", e.Request.URL, locationStr, e.Text)
		return nil, nil
	}

	l := &speakerdeck.Location{
		RequestedAddress: locationStr[1],
	}

	r := &maps.GeocodingRequest{
		Address: l.RequestedAddress,
	}
	results, err := le.c.Geocode(context.Background(), r)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 { // no results
		log.Warnf("Found no geocode results for %q", l.RequestedAddress)
		return nil, nil
	}

	if len(results) > 1 {
		log.Warnf("Got more than one result for %q! Will only respect the first one.", l.RequestedAddress)
	}

	l.ResolvedAddress = results[0].FormattedAddress
	l.Lat = results[0].Geometry.Location.Lat
	l.Lng = results[0].Geometry.Location.Lng

	log.Infof("Found geolocation for %q: %f %f", l.RequestedAddress, l.Lat, l.Lng)
	t.Location = l
	return nil, nil
}
