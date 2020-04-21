package speakerdeck

import (
	"time"
)

func NewUser() *User {
	return &User{}
}

// User represents a user on speakerdeck.com
type User struct {
	Author       Author       `json:"author"`
	Abstract     string       `json:"abstract"`
	TalkPreviews TalkPreviews `json:"talkPreviews"`
}

type Author struct {
	Name       string `json:"name"`
	Handle     string `json:"handle"`
	Link       string `json:"link"`
	AvatarLink string `json:"avatarLink"`
}

type TalkPreview struct {
	// Title describes the talk title
	Title string `json:"title"`

	ID string `json:"id"`

	Views uint32 `json:"views"`

	Stars uint32 `json:"stars"`
	// Date represents the talk presentation date
	Date time.Time `json:"date"`
	// Link describes the link to Speakerdeck
	Link string `json:"link"`
	// DataID represents the key used to embed Speakerdeck presentations on an other website
	DataID string `json:"dataID"`
}

// NewTalk returns a new, empty talk object
func NewTalk() *Talk {
	return &Talk{
		ExtraLinks: map[string][]string{},
	}
}

// Talk describes a presentation on Speakerdeck
type Talk struct {
	// Talk embeds all data that is visible from TalkPreview, too
	TalkPreview

	Author Author `json:"author"`

	Category     string `json:"category"`
	CategoryLink string `json:"categoryLink"`
	DownloadLink string `json:"downloadLink"`

	// ExtraLinks contains parsed URLs found in the description, mapped by their domain name
	ExtraLinks map[string][]string `json:"extraLinks"`
	// Hide is set to true if the talk description contains a "Hide: true" string, indicating it should not be scraped
	Hide bool `json:"hide"`
	// Location can be populated by the LocationExtension available in
	Location *Location `json:"location,omitempty"`
}

// Talks orders the Talk objects by time
type Talks []Talk

// Len implements sort.Interface
func (p Talks) Len() int {
	return len(p)
}

// Less implements sort.Interface
func (p Talks) Less(i, j int) bool {
	return p[i].Date.Before(p[j].Date)
}

// Swap implements sort.Interface
func (p Talks) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

// TalkPreviews orders the TalkPreview objects by time
type TalkPreviews []TalkPreview

// Len implements sort.Interface
func (p TalkPreviews) Len() int {
	return len(p)
}

// Less implements sort.Interface
func (p TalkPreviews) Less(i, j int) bool {
	return p[i].Date.Before(p[j].Date)
}

// Swap implements sort.Interface
func (p TalkPreviews) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

// Location describes a geographical location for the talk
// This field is populated by the LocationExtension, and is set based on
// a "Location: <address>" string in the talk description. For instance,
// if you put "Location: TUAS-talo, Aalto University" in the talk description,
// Location will be populated with coordinates based on Google Maps data.
type Location struct {
	// RequestedAddress is populated from the original "requested" address mentioned in the
	// talk description (e.g. "TUAS-talo, Aalto University")
	RequestedAddress string
	// ResolvedAddress is populated by the Google Maps Geocoding API, and is the official street address
	// (or similar) of the place. In the above example: "Maarintie 8, 02150 Espoo, Finland"
	ResolvedAddress string
	// Lat describes the latitude of the location
	Lat float64
	// Lng describes the longitude of the location
	Lng float64
}
