package i18n

import (
	"testing"
)

func TestAllLanguagesLoad(t *testing.T) {
	for _, lang := range Supported {
		SetLanguage(lang)
		// Every language must translate at least the shortcut tab.
		if got := T("tab.shortcuts"); got == "tab.shortcuts" {
			t.Fatalf("language %q: missing translation for tab.shortcuts", lang)
		}
	}
	SetLanguage("en") // reset
}

func TestMissingKeyReturnsKey(t *testing.T) {
	got := T("this.key.does.not.exist")
	if got != "this.key.does.not.exist" {
		t.Fatalf("got %q, want the key itself", got)
	}
}

func TestCommonKeysPresent(t *testing.T) {
	keys := []string{
		"tab.shortcuts", "tab.settings", "tab.about",
		"settings.radiologist", "settings.language", "settings.theme",
		"settings.save", "settings.browse",
		"button.copy", "button.paste", "button.back", "button.home",
		"preview.placeholder",
	}
	for _, key := range keys {
		for _, lang := range Supported {
			SetLanguage(lang)
			if got := T(key); got == key {
				t.Fatalf("language %q: missing key %q", lang, key)
			}
		}
	}
	SetLanguage("en")
}
