package main

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type ListItem struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type ListItems []ListItem

func (ll ListItems) Len() int {
	return len(ll)
}

func (ll ListItems) Less(i, j int) bool {
	var err error

	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

	x := strings.ToLower(ll[i].Label)
	x, _, err = transform.String(t, x)
	if err != nil {
		x = ll[i].Label
	}

	y := strings.ToLower(ll[j].Label)
	y, _, err = transform.String(t, y)
	if err != nil {
		y = ll[j].Label
	}

	return x < y
}

func (ll ListItems) Swap(i, j int) {
	ll[i], ll[j] = ll[j], ll[i]
}

func createList(ss ...[]string) ListItems {
	ll := make([]ListItem, 0, len(ss))

	for _, s := range ss {
		len := len(s)
		if len == 0 {
			continue
		}

		l := ListItem{}
		l.Label = s[0]

		if len == 2 {
			l.Value = s[1]
		} else {
			l.Value = l.Label
		}

		ll = append(ll, l)
	}

	return ll
}

func (ll ListItems) Contains(value string) bool {
	for _, l := range ll {
		if l.Value != value {
			continue
		}

		return true
	}

	return false
}

func (ll ListItems) GetMap() map[string]string {
	m := make(map[string]string)

	for _, l := range ll {
		m[l.Value] = l.Label
	}

	return m
}
