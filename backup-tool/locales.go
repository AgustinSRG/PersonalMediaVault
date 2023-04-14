// Internationalization utilities

package main

import (
	"embed"
	"os"

	"github.com/jeandeaual/go-locale"
	"github.com/naoina/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed active.*.toml

var LocaleFS embed.FS
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
		lang, _ = locale.GetLanguage() // Default system language
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
	_, err = bundle.LoadMessageFileFS(LocaleFS, "active."+lang+".toml")

	if err != nil {
		panic(err)
	}

	Localizer = i18n.NewLocalizer(bundle, lang)
}
