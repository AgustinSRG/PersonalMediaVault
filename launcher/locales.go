// Internationalization utilities

package main

import (
	"embed"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/naoina/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed active.*.toml

var LocaleFS embed.FS
var Language = "en"
var Localizer *i18n.Localizer

func InitializeInternationalizationFramework() {
	// Available languages
	availableLanguages := map[string]bool{
		"en": true,
		"es": true,
	}

	// Get language
	lang := os.Getenv("PMV_LANGUAGE")
	if lang == "" {
		ex, err := os.Executable()
		if err == nil {
			langFilePath := filepath.Join(filepath.Dir(ex), "lang")
			langBytes, err := ioutil.ReadFile(langFilePath)

			if err == nil {
				lang = string(langBytes)
			}
		}
	}

	defaultLang, err := language.Parse(lang)

	if err != nil || !availableLanguages[lang] {
		defaultLang = language.English
		lang = "en"
	}

	// Create bundle
	bundle := i18n.NewBundle(defaultLang)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	// Add locale file
	bundle.LoadMessageFileFS(LocaleFS, "active."+lang+".toml")

	Localizer = i18n.NewLocalizer(bundle, lang)
	Language = lang
}

func checkYesNoAnswer(answer string) bool {
	answer = strings.ToLower(answer)
	if answer == "y" {
		return true
	}

	yesPrefix, _ := Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YesPrefix",
			Other: "y",
		},
	})

	return strings.HasPrefix(answer, strings.ToLower(yesPrefix))
}
