package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	speakerdeck "github.com/luxas/speakerdeck-api"
	"github.com/luxas/speakerdeck-api/location"
	"github.com/luxas/speakerdeck-api/scraper"
	log "github.com/sirupsen/logrus"
)

const (
	prefix      = `/api`
	welcomeText = `
<h1>Welcome to this Speakerdeck API server!</h1>
<span>Available paths are:</span>
<ul>
	<li>/api/users/{user-handle}</li>
	<li>/api/talks/{user-handle}</li>
	<li>/api/talks/{user-handle}/{talk-id}</li>
</ul>
<br />
<span>Created by Lucas Käldström. Source code at: <a href="https://github.com/luxas/speakerdeck-api">github.com/luxas/speakerdeck-api</a></span>
`
)

var (
	validPaths = regexp.MustCompile(`^` + prefix + `/(talks|users)/([a-zA-Z0-9/-]+)$`)

	address    = flag.String("address", "0.0.0.0", "What address to expose the API on")
	port       = flag.Int("port", 8080, "What port to expose the API on")
	mapsAPIKey = flag.String("maps-api-key", "", "Google Maps API key with the Geocoding API usage set")

	locationExt *location.LocationExtension
)

func main() {
	flag.Parse()
	http.HandleFunc("/", makeHandler(helpHandler))
	http.HandleFunc(prefix+"/users/", makeHandler(usersHandler))
	http.HandleFunc(prefix+"/talks/", makeHandler(talksHandler))

	if len(*mapsAPIKey) > 0 {
		var err error
		locationExt, err = location.NewLocationExtension(*mapsAPIKey)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Initialized the LocationExtension!")
	}

	addrPort := fmt.Sprintf("%s:%d", *address, *port)
	log.Printf("Starting Speakerdeck API on %s...", addrPort)
	log.Fatal(http.ListenAndServe(addrPort, nil))
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string) (int, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			helpHandler(w, r, "")
			return
		}
		m := validPaths.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		t := time.Now()
		code, err := fn(w, r, m[2])
		if err != nil {
			http.Error(w, err.Error(), code)
		} else {
			log.Infof("Handler for path %q responded in %s", r.URL.Path, time.Since(t))
		}
	}
}

func encodeJSON(w io.Writer, data interface{}) error {
	e := json.NewEncoder(w)
	e.SetEscapeHTML(true)
	e.SetIndent("", "  ")
	return e.Encode(data)
}

func helpHandler(w http.ResponseWriter, r *http.Request, _ string) (int, error) {
	w.Write([]byte(welcomeText))
	return http.StatusOK, nil
}

func usersHandler(w http.ResponseWriter, r *http.Request, userID string) (int, error) {
	if strings.Contains(userID, "/") {
		return http.StatusBadRequest, fmt.Errorf("invalid user name, can't contain /")
	}

	user, err := speakerdeck.ScrapeUser(userID, nil)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	if err := encodeJSON(w, user); err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

func talksHandler(w http.ResponseWriter, r *http.Request, talkStr string) (int, error) {
	parts := strings.Split(talkStr, "/")
	if len(parts) > 2 {
		return http.StatusBadRequest, fmt.Errorf("invalid talk name, argument should be of form {user} or {user}/{talk}")
	}
	userID := parts[0]
	talkID := ""
	if len(parts) == 2 {
		talkID = parts[1]
	}

	var opts *scraper.ScrapeOptions
	if locationExt != nil {
		opts = &scraper.ScrapeOptions{
			Extensions: []scraper.Extension{locationExt},
		}
	}

	talks, err := speakerdeck.ScrapeTalks(userID, talkID, opts)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	if err := encodeJSON(w, talks); err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}
