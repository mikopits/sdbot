package sdbot

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

// Returns a new string with non-alphanumeric characters removed. This is
// particularly useful when it comes to identifying unique usernames. Keep in
// mind that Pokemon Showdown usernames are _not_ case sensitive, so a
// downcased and sanitized username is a unique identifier.
func Sanitize(s string) string {
	reg := regexp.MustCompile("[^A-Za-z0-9]")

	return strings.ToLower(reg.ReplaceAllString(s, ""))
}

// Public API for errorchecking. Logs to only os.Stderr.
func CheckErr(err error) {
	if err != nil {
		Error(&Log, err)
	}
}

// Public API for errorchecking. Logs to every logger in ActiveLoggers.
func CheckErrAll(err error) {
	if err != nil {
		ErrorAll(ActiveLoggers, err)
	}
}

func Inspect(i interface{}) {
	Debugf(&Log, "%+v", i)
}

type HasteKey struct {
	Key string
}

// Upload to Hastebin and return the response URL.
func Haste(buf io.Reader, bodyType string) (string, error) {
	data, err := ioutil.ReadAll(buf)
	if err != nil {
		Error(&Log, err)
		return "", err
	}

	trimmedData := strings.TrimRight(string(data), "\r\n")

	res, err := http.Post("http://hastebin.com/documents", bodyType, bytes.NewBufferString(trimmedData))
	if err != nil {
		Error(&Log, err)
		return "", err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		Error(&Log, err)
		return "", err
	}

	var hk HasteKey
	err = json.Unmarshal(body, &hk)
	if err != nil {
		Error(&Log, err)
		return "", err
	}

	return "http://hastebin.com/" + hk.Key, nil
}
