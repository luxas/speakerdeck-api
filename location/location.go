package location

import (
	"context"
	"regexp"
	"strings"

	"github.com/gocolly/colly"
	"github.com/luxaslabs/luxaslabs/generator/scraper"
	"github.com/luxaslabs/luxaslabs/generator/speakerdeck/types"
	log "github.com/sirupsen/logrus"
	"googlemaps.github.io/maps"
)

var locationRegexp = regexp.MustCompile(`Location: (.*)`)

var _ scraper.Extension = &LocationExtension{}

func NewLocationExtension(apiKey string) (*LocationExtension, error) {
	c, err := maps.NewClient(maps.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	return &LocationExtension{c}, nil
}

type LocationExtension struct {
	c *maps.Client
}

func (_ *LocationExtension) Name() string {
	return "LocationExtension"
}

func (le *LocationExtension) Hook() scraper.Hook {
	return scraper.Hook{
		DOMPath: ".deck-description.mb-4 p",
		Handler: le.onDescription,
	}
}

func (le *LocationExtension) onDescription(e *colly.HTMLElement, data interface{}) (*string, error) {
	// Fail fast, only consider descriptions with the "Location" substring
	if !strings.Contains(e.Text, "Location") {
		return nil, nil
	}

	switch data.(type) {
	case *types.Talk:
		// noop, allow this
	default:
		return nil, nil
	}

	t := data.(*types.Talk)

	if strings.Contains(e.Text, "Location: Online") {
		t.Location = &types.Location{
			RequestedAddress: "Online",
		}
		return nil, nil
	}

	locationStr := locationRegexp.FindStringSubmatch(e.Text)
	if len(locationStr) != 2 {
		log.Warnf("Couldn't find location for talk %s: %v %q", e.Request.URL, locationStr, e.Text)
		return nil, nil
	}

	l := &types.Location{
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
