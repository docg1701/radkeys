// Package i18n provides industry-standard internationalization for RadKeys
// using go-i18n/v2 with embedded JSON translation files. One file per locale.
// To add a language: create locales/<code>.json and add the code to Supported.
package i18n

import (
	"embed"
	"encoding/json"
	"fmt"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed locales/*.json
var localeFS embed.FS

// Supported lists the language codes the UI accepts.
var Supported = []string{"en", "pt-BR", "pt-PT", "es", "fr", "de", "it"}

var bundle *i18n.Bundle
var current *i18n.Localizer

func init() {
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	entries, err := localeFS.ReadDir("locales")
	if err != nil {
		panic(fmt.Sprintf("i18n: read embedded locales: %v", err))
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, err := localeFS.ReadFile("locales/" + e.Name())
		if err != nil {
			panic(fmt.Sprintf("i18n: read %s: %v", e.Name(), err))
		}
		if _, err := bundle.ParseMessageFileBytes(data, e.Name()); err != nil {
			panic(fmt.Sprintf("i18n: parse %s: %v", e.Name(), err))
		}
	}
	current = i18n.NewLocalizer(bundle, "en")
}

// SetLanguage switches the active locale. Falls back to English if unknown.
func SetLanguage(lang string) {
	if lang == "" {
		lang = "en"
	}
	current = i18n.NewLocalizer(bundle, lang)
}

// T translates a message ID to the active language.
func T(id string) string {
	msg, err := current.Localize(&i18n.LocalizeConfig{MessageID: id})
	if err != nil {
		return id
	}
	return msg
}
