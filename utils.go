package main

import (
	"fmt"
	"html"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/lib/pq"
	"github.com/microcosm-cc/bluemonday"
	"golang.org/x/net/publicsuffix"
)

func maybeTruncate(s string, length uint) string {
	if length > 0 && utf8.RuneCountInString(s) > int(length) {
		// https://stackoverflow.com/questions/23466497/how-to-truncate-a-string-in-a-golang-template
		var runeCount = 0

		for idx := range s {
			runeCount++
			if runeCount > int(length) {
				return s[:idx]
			}
		}
	}

	return s
}

func sanitizeHTML(s string) string {
	p := bluemonday.UGCPolicy()
	s = p.Sanitize(s)
	s = html.UnescapeString(s)

	return s
}

func sanitizeSlug(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)

	reg := regexp.MustCompile("[^a-z0-9]")
	s = reg.ReplaceAllString(s, "-")

	reg = regexp.MustCompile("-+")
	s = reg.ReplaceAllString(s, "-")

	return maybeTruncate(s, 80)
}

func sanitizePhoneNumber(s string) string {
	re := regexp.MustCompile("[0-9]+")

	sa := re.FindAllString(s, -1)
	s = strings.Join(sa, "")

	return s
}

func getDomainFromURL(d string) (string, error) {
	d = strings.ToLower(d)

	if strings.HasPrefix(d, "http://") {
		d = strings.Replace(d, "http://", "", 1)
	}

	if strings.HasPrefix(d, "https://") {
		d = strings.Replace(d, "https://", "", 1)
	}

	u, err := url.ParseRequestURI("https://" + d)
	if err != nil {
		return "", fmt.Errorf("'%s' is an invalid domain", d)
	}

	return u.Hostname(), err
}

func isOneOf(value string, allowedValues []string) bool {
	for _, allowedValue := range allowedValues {
		if value == allowedValue {
			return true
		}
	}

	return false
}

func isOneOfUint(value uint, allowedValues []uint) bool {
	for _, allowedValue := range allowedValues {
		if value == allowedValue {
			return true
		}
	}

	return false
}

func isOneOfInt64(value int64, allowedValues []int64) bool {
	for _, allowedValue := range allowedValues {
		if value == allowedValue {
			return true
		}
	}

	return false
}

func isIncluded(needles []string, haystack []string) bool {
	for _, n := range needles {
		if isOneOf(n, haystack) {
			return true
		}
	}

	return false
}

func isAllIncluded(needles []int64, haystack []int64) bool {
	for _, n := range needles {
		if !isOneOfInt64(n, haystack) {
			return false
		}
	}

	return true
}

func isOneIncluded(needles []int64, haystack []int64) bool {
	for _, n := range needles {
		if isOneOfInt64(n, haystack) {
			return true
		}
	}

	return false
}

func getUnique(ss []string) []string {
	m := make(map[string]bool, len(ss))

	for _, s := range ss {
		if s == "" {
			continue
		}

		m[s] = true
	}

	vv := []string{}
	for s := range m {
		vv = append(vv, s)
	}

	return vv
}

func getStringArrayFromPqStringArray(pp pq.StringArray) []string {
	ss := make([]string, 0, len(pp))
	for _, p := range pp {
		ss = append(ss, p)
	}

	return ss
}

func validateDate(s string, format string) (time.Time, error) {

	if format != "" {
		return time.Parse(format, s)
	}
	return time.Parse("2006-01-02 15:04:05", s)
}

func getTLDPlusOne(s string) (string, error) {
	if s == "" {
		return "", nil
	}

	s = strings.TrimSpace(s)
	s = strings.ToLower(s)

	if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		s = "http://" + s
	}

	u, err := url.ParseRequestURI(s)
	if err != nil {
		return "", err
	}

	s = u.Hostname()

	domain, err := publicsuffix.EffectiveTLDPlusOne(s)
	if err != nil {
		return "", err
	}

	return domain, nil
}

func sanitizeText(s string, length uint) string {
	p := bluemonday.StrictPolicy()
	s = p.Sanitize(s)
	s = html.UnescapeString(s)
	s = strings.TrimSpace(s)

	if length > 0 && utf8.RuneCountInString(s) > int(length) {
		// https://stackoverflow.com/questions/23466497/how-to-truncate-a-string-in-a-golang-template
		runeCount := 0

		for idx := range s {
			runeCount++
			if runeCount > int(length) {
				return s[:idx]
			}
		}
	}

	return s
}

func sanitizeHost(host string) string {
	if host == "" {
		return os.Getenv("VM_DOMAIN")
	}

	hosts := strings.Split(os.Getenv("VM_ALLOWED_HOSTS"), ",")

	if isOneOf(host, hosts) {
		return host
	}

	return os.Getenv("VM_DOMAIN")
}

func getDefaultTimeout() time.Duration {
	return time.Duration(15)
}

func getEnvFilePath(f string) string {
	env := os.Getenv("VM_ENVIRONMENT")

	return "./environment/" + env + "/" + f
}
