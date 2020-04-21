package speakerdeck

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	speakerdeckRootURL = "https://speakerdeck.com"
	httpsPrefix        = "https:"
)

var linkRegexp = regexp.MustCompile(`http[s]?://[a-zA-Z-_/0-9\.#=&]*`)

func sdPrefix(s string) string {
	return fmt.Sprintf("%s%s", speakerdeckRootURL, s)
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
