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
		"button.select_all", "button.select_line", "button.line_start",
		"button.line_end", "button.backspace", "button.delete",
		"device_action.via_keypad_hint",
		"preview.placeholder",
		"status.mock_mode", "status.device_command_failed", "status.out_of_grid", "status.hid_read_failed",
		"error.config_title", "error.config_message", "error.config_fix", "error.open_file",
		"button.close",
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

// TestAllMessagesHaveAllLanguages ensures every message ID has a translation
// for every supported language, so a new language or key can't silently fall
// back to the raw ID in the UI.
func TestAllMessagesHaveAllLanguages(t *testing.T) {
	for id, langs := range messages {
		for _, lang := range Supported {
			if _, ok := langs[lang]; !ok {
				t.Errorf("message %q missing translation for language %q", id, lang)
			}
		}
	}
}
