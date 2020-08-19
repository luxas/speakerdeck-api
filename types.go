package speakerdeck

import (
	"time"
)

// NewUser creates a new User object
func NewUser() *User {
	return &User{}
}

// User represents a user on as browsed on the user page (i.e. https://speakerdeck.com/{user-handle})
type User struct {
	// Author describes the Speakerdeck profile of the person that's created the presentations
	Author Author `json:"author"`

	// Abstract contains a short description of the user
	Abstract string `json:"abstract"`

	// TalkPreviews is a list of TalkPreview objects, containing information about each talk
	// as seen on the user page
	TalkPreviews []TalkPreview `json:"talkPreviews"`
}

// Author describes the SD profile of the person that's created the presentations
type Author struct {
	// Name describes the human-friendly name of the author
	Name string `json:"name"`

	// Handle describes the preferred nickname of the author. This is used in the URL
	Handle string `json:"handle"`

	// Link describes the link to the author's page
	Link string `json:"link"`

	// AvatarLink describes the link to the avatar of the author
	AvatarLink string `json:"avatarLink"`
}

// TalkPreview contains the information about a talk that can be seen on the user page
type TalkPreview struct {
	// Title describes the talk title
	Title string `json:"title"`

	// ID describes the URL-encoded descriptor for the talk, used in the URL as
	// https://speakerdeck.com/{user-handle}/{talk-id}. The ID is unique per user, but
	// not globally across users
	ID string `json:"id"`

	// Views describes how many views a talk has got
	Views uint32 `json:"views"`

	// Stars describes how many other users have starred this talk
	Stars uint32 `json:"stars"`

	// Link describes the link to the talk at Speakerdeck
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
	// TalkPreview is embedded here as it contains all the data we want to display here, too
	TalkPreview

	// Date represents the talk presentation date
	Date time.Time `json:"date"`

	// Author describes the Speakerdeck profile of the person that's created the presentations
	Author Author `json:"author"`

	// Category is a string of Speakerdeck-specific categories you can choose from when uploading a
	// presentation, e.g. "Technology"
	Category string `json:"category"`

	// CategoryLink is the link to other presentations belonging to the same category, e.g.
	// https://speakerdeck.com/c/technology.
	CategoryLink string `json:"categoryLink"`

	// DownloadLink is the link from where you can download the underlying PDF
	DownloadLink string `json:"downloadLink"`

	// ExtraLinks contains parsed URLs found in the talk description, mapped by their domain name
	ExtraLinks map[string][]string `json:"extraLinks"`

	// Hide is set to true if the talk description contains a "Hide: true" string,
	// indicating it should not be visible for the API. If a talk has set Hide=true,
	// it will be returned as normal, and the user of the API can choose whether to ignore
	// or respect that pledge
	Hide bool `json:"hide"`

	// Location describes a geographical location for the talk
	// This field is populated by the LocationExtension, and is set based on
	// a "Location: <address>" string in the talk description. For instance,
	// if you put "Location: TUAS-talo, Aalto University" in the talk description,
	// Location will be populated with coordinates based on Google Maps data.
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

// Location describes a geographical location for the talk
// This struct is populated by the LocationExtension, and is set based on
// a "Location: <address>" string in the talk description. For instance,
// if you put "Location: TUAS-talo, Aalto University" in the talk description,
// Location will be populated with coordinates based on Google Maps data.
type Location struct {
	// RequestedAddress is populated from the original "requested" address mentioned in the
	// talk description (e.g. "TUAS-talo, Aalto University")
	RequestedAddress string `json:"requestedAddress"`

	// ResolvedAddress is populated by the Google Maps Geocoding API, and is the official street address
	// (or similar) of the place. In the above example: "Maarintie 8, 02150 Espoo, Finland"
	ResolvedAddress string `json:"resolvedAddress"`

	// Lat describes the latitude of the location
	Lat float64 `json:"lat"`

	// Lng describes the longitude of the location
	Lng float64 `json:"lng"`
}
