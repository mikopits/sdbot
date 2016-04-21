package sdbot

import (
	"regexp"
	"strings"
)

// Returns a new string with non-alphanumeric characters removed. This is
// particularly useful when it comes to identifying unique usernames. Keep in
// mind that Pokemon Showdown usernames are _not_ case sensitive, so a
// downcased and sanitized username is a unique identifier.
func Sanitize(s string) string {
	reg, err := regexp.Compile("[^A-Za-z0-9]")
	if err != nil {
		Error(&Log, err)
	}

	return strings.ToLower(reg.ReplaceAllString(s, ""))
}
