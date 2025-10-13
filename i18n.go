package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/labstack/echo"
	"github.com/pariz/gountries"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// xgettext --language=c --add-comments *.go -d languages/en

var matcher language.Matcher

func getSupportedLanguages() []language.Tag {
	return []language.Tag{
		language.English,
		language.Hindi,
		language.Marathi,
	}
}

func getSupportedLanguagesStrings() []string {
	var ss []string

	for _, t := range getSupportedLanguages() {
		b, _ := t.Base()
		ss = append(ss, b.String())
	}

	return ss
}

func getDefaultLanguageString() string {
	b, _ := language.English.Base()

	return b.String()
}

func initI18n() error {
	matcher = language.NewMatcher(getSupportedLanguages())

	file, err := os.OpenFile("languages/hi.json", os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}

	defer file.Close()

	bb, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	lines := make(map[string]string)

	err = json.Unmarshal(bb, &lines)
	if err != nil {
		return err
	}

	for from, to := range lines {
		err = message.SetString(language.Hindi, from, to)
		if err != nil {
			return err
		}
	}

	return nil
}

func gettext(s string, ctx echo.Context) string {
	if ctx == nil {
		return s
	}

	accept := ctx.Request().Header.Get("Accept-Language")
	tag, _ := language.MatchStrings(matcher, accept)

	p := message.NewPrinter(tag)

	return p.Sprint(s)
}

func getCountryName(country gountries.Country, ctx echo.Context) string {
	if ctx == nil {
		return country.Name.Common
	}

	accept := ctx.Request().Header.Get("Accept-Language")
	tag, _ := language.MatchStrings(matcher, accept)

	if tag == language.Hindi {
		hi, exists := country.Translations["HI"]
		if exists {
			return hi.Common
		}
	}

	return country.Name.Common
}

func getLanguageFromContext(ctx echo.Context) string {
	accept := ctx.Request().Header.Get("Accept-Language")
	tag, _ := language.MatchStrings(matcher, accept)
	b, _ := tag.Base()
	s := b.String()

	if !isOneOf(s, getSupportedLanguagesStrings()) {
		return getDefaultLanguageString()
	}

	return s
}
