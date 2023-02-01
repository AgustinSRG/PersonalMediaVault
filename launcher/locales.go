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

// Load internationalization framework
func InitializeInternationalizationFramework() {
	// Available languages
	availableLanguages := map[string]bool{
		"en": true,
		"es": true,
	}

	// Get language
	lang := os.Getenv("LANGUAGE")
	if lang == "" {
		lang, _ = locale.GetLanguage() // Default system language
	}

	defaultLang, err := language.Parse(lang)

	if err != nil || !availableLanguages[lang] {
		defaultLang = language.English
	}

	// Create bundle
	bundle := i18n.NewBundle(defaultLang)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	// Add locale file
	bundle.LoadMessageFileFS(LocaleFS, "active."+lang+".toml")

	Localizer = i18n.NewLocalizer(bundle, lang)
}
