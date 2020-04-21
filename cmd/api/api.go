package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	speakerdeck "github.com/luxas/speakerdeck-scraper"
	log "github.com/sirupsen/logrus"
)

// TODO: Add sample usage text at the "/" endpoint
// TODO: Add support for the location extension

const prefix = `/api`

var validPaths = regexp.MustCompile(`^` + prefix + `/(talks|users)/([a-zA-Z0-9/-]+)$`)

func makeHandler(fn func(http.ResponseWriter, *http.Request, string) (int, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

	talks, err := speakerdeck.ScrapeTalk(userID, talkID, nil)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	if err := encodeJSON(w, talks); err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

func main() {
	http.HandleFunc(prefix+"/users/", makeHandler(usersHandler))
	http.HandleFunc(prefix+"/talks/", makeHandler(talksHandler))

	log.Printf("Starting Speakerdeck API...")
	// TODO: Parameterize this
	log.Fatal(http.ListenAndServe(":8080", nil))
}
