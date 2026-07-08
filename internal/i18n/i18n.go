// Package i18n provides translations for RadKeys using a single Go map.
// All languages live here — no JSON files, no embed. To add a string:
// add an entry to the messages map with a translation for each language.
// To add a language: add the code to Supported and fill translations below.
package i18n

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// Supported lists the language codes the UI accepts.
var Supported = []string{"en", "pt-BR", "pt-PT", "es", "fr", "de", "it"}

var bundle *i18n.Bundle
var current *i18n.Localizer

// messages is the single source of truth for all translations.
// Key = message ID. Value = map[language]translation.
// Missing translations fall back to English.
var messages = map[string]map[string]string{
	// ── Tabs ──────────────────────────────────────────────
	"tab.shortcuts": {
		"en": "Shortcuts", "pt-BR": "Atalhos", "pt-PT": "Atalhos",
		"es": "Atajos", "fr": "Raccourcis", "de": "Tastenkürzel", "it": "Scorciatoie",
	},
	"tab.settings": {
		"en": "Settings", "pt-BR": "Ajustes", "pt-PT": "Definições",
		"es": "Ajustes", "fr": "Réglages", "de": "Einstellungen", "it": "Impostazioni",
	},
	"tab.about": {
		"en": "About", "pt-BR": "Sobre", "pt-PT": "Sobre",
		"es": "Acerca de", "fr": "À propos", "de": "Über", "it": "Informazioni",
	},

	// ── Settings: sections ────────────────────────────────
	"settings.group_config": {
		"en": "Configuration File", "pt-BR": "Arquivo de Configuração",
		"pt-PT": "Ficheiro de Configuração", "es": "Archivo de Configuración",
		"fr": "Fichier de configuration", "de": "Konfigurationsdatei",
		"it": "File di Configurazione",
	},
	"settings.group_appearance": {
		"en": "Appearance", "pt-BR": "Aparência", "pt-PT": "Aparência",
		"es": "Apariencia", "fr": "Apparence", "de": "Erscheinungsbild", "it": "Aspetto",
	},
	"settings.group_device": {
		"en": "USB Device", "pt-BR": "Dispositivo USB", "pt-PT": "Dispositivo USB",
		"es": "Dispositivo USB", "fr": "Périphérique USB", "de": "USB-Gerät",
		"it": "Dispositivo USB",
	},

	// ── Settings: fields ──────────────────────────────────
	"settings.radiologist": {
		"en": "Radiologist", "pt-BR": "Radiologista", "pt-PT": "Radiologista",
		"es": "Radiólogo", "fr": "Radiologue", "de": "Radiologe", "it": "Radiologo",
	},
	"settings.language": {
		"en": "Language", "pt-BR": "Idioma", "pt-PT": "Idioma",
		"es": "Idioma", "fr": "Langue", "de": "Sprache", "it": "Lingua",
	},
	"settings.theme": {
		"en": "Theme", "pt-BR": "Tema", "pt-PT": "Tema",
		"es": "Tema", "fr": "Thème", "de": "Thema", "it": "Tema",
	},
	"settings.columns": {
		"en": "Columns", "pt-BR": "Colunas", "pt-PT": "Colunas",
		"es": "Columnas", "fr": "Colonnes", "de": "Spalten", "it": "Colonne",
	},
	"settings.rows": {
		"en": "Rows", "pt-BR": "Linhas", "pt-PT": "Linhas",
		"es": "Filas", "fr": "Lignes", "de": "Zeilen", "it": "Righe",
	},
	"settings.vid": {
		"en": "VID", "pt-BR": "VID", "pt-PT": "VID",
		"es": "VID", "fr": "VID", "de": "VID", "it": "VID",
	},
	"settings.pid": {
		"en": "PID", "pt-BR": "PID", "pt-PT": "PID",
		"es": "PID", "fr": "PID", "de": "PID", "it": "PID",
	},
	"settings.protocol": {
		"en": "Protocol", "pt-BR": "Protocolo", "pt-PT": "Protocolo",
		"es": "Protocolo", "fr": "Protocole", "de": "Protokoll", "it": "Protocollo",
	},
	"settings.config_file": {
		"en": "Path", "pt-BR": "Caminho", "pt-PT": "Caminho",
		"es": "Ruta", "fr": "Chemin", "de": "Pfad", "it": "Percorso",
	},
	"settings.browse": {
		"en": "Browse…", "pt-BR": "Procurar…", "pt-PT": "Procurar…",
		"es": "Examinar…", "fr": "Parcourir…", "de": "Durchsuchen…", "it": "Sfoglia…",
	},
	"settings.save": {
		"en": "Save", "pt-BR": "Salvar", "pt-PT": "Guardar",
		"es": "Guardar", "fr": "Enregistrer", "de": "Speichern", "it": "Salva",
	},
	"settings.icon": {
		"en": "Icon", "pt-BR": "Ícone", "pt-PT": "Ícone",
		"es": "Icono", "fr": "Icône", "de": "Symbol", "it": "Icona",
	},

	// ── Preview / keypad ──────────────────────────────────
	"preview.placeholder": {
		"en": "Select a phrase.", "pt-BR": "Selecione uma frase.",
		"pt-PT": "Selecione uma frase.", "es": "Seleccione una frase.",
		"fr": "Sélectionnez une phrase.", "de": "Wählen Sie einen Text.",
		"it": "Seleziona una frase.",
	},
	"button.copy": {
		"en": "Copy", "pt-BR": "Copiar", "pt-PT": "Copiar",
		"es": "Copiar", "fr": "Copier", "de": "Kopieren", "it": "Copia",
	},
	"button.paste": {
		"en": "Paste", "pt-BR": "Colar", "pt-PT": "Colar",
		"es": "Pegar", "fr": "Coller", "de": "Einfügen", "it": "Incolla",
	},
	"button.back": {
		"en": "Back", "pt-BR": "Voltar", "pt-PT": "Voltar",
		"es": "Volver", "fr": "Retour", "de": "Zurück", "it": "Indietro",
	},
	"button.home": {
		"en": "Home", "pt-BR": "Início", "pt-PT": "Início",
		"es": "Inicio", "fr": "Accueil", "de": "Start", "it": "Home",
	},

	// ── About ─────────────────────────────────────────────
	"about.version": {
		"en": "Version %s", "pt-BR": "Versão %s", "pt-PT": "Versão %s",
		"es": "Versión %s", "fr": "Version %s", "de": "Version %s", "it": "Versione %s",
	},
	"about.description": {
		"en":    "RadKeys is a cross-platform companion for radiology shortcut decks.",
		"pt-BR": "RadKeys é um companheiro multiplataforma para decks de atalhos em radiologia.",
		"pt-PT": "RadKeys é um companheiro multiplataforma para decks de atalhos em radiologia.",
		"es":    "RadKeys es un compañero multiplataforma para decks de atajos en radiología.",
		"fr":    "RadKeys est un compagnon multiplateforme pour les decks de raccourcis en radiologie.",
		"de":    "RadKeys ist ein plattformübergreifender Begleiter für Radiologie-Kurzbefehle.",
		"it":    "RadKeys è un compagno multipiattaforma per deck di scorciatoie radiologiche.",
	},
	"about.author": {
		"en": "Author: docg1701", "pt-BR": "Autor: docg1701", "pt-PT": "Autor: docg1701",
		"es": "Autor: docg1701", "fr": "Auteur : docg1701", "de": "Autor: docg1701",
		"it": "Autore: docg1701",
	},
	"about.license": {
		"en": "License: MIT", "pt-BR": "Licença: MIT", "pt-PT": "Licença: MIT",
		"es": "Licencia: MIT", "fr": "Licence : MIT", "de": "Lizenz: MIT",
		"it": "Licenza: MIT",
	},
	"about.repository": {
		"en": "Repository: ", "pt-BR": "Repositório: ", "pt-PT": "Repositório: ",
		"es": "Repositorio: ", "fr": "Dépôt : ", "de": "Repository: ",
		"it": "Repository: ",
	},
	"about.stack": {
		"en":    "Built with Go, Fyne, go-hid, go-i18n, and BurntSushi/toml.",
		"pt-BR": "Feito com Go, Fyne, go-hid, go-i18n e BurntSushi/toml.",
		"pt-PT": "Feito com Go, Fyne, go-hid, go-i18n e BurntSushi/toml.",
		"es":    "Hecho con Go, Fyne, go-hid, go-i18n y BurntSushi/toml.",
		"fr":    "Construit avec Go, Fyne, go-hid, go-i18n et BurntSushi/toml.",
		"de":    "Erstellt mit Go, Fyne, go-hid, go-i18n und BurntSushi/toml.",
		"it":    "Realizzato con Go, Fyne, go-hid, go-i18n e BurntSushi/toml.",
	},
	"about.i18n": {
		"en":    "Available in 7 languages.",
		"pt-BR": "Disponível em 7 idiomas.",
		"pt-PT": "Disponível em 7 idiomas.",
		"es":    "Disponible en 7 idiomas.",
		"fr":    "Disponible en 7 langues.",
		"de":    "Verfügbar in 7 Sprachen.",
		"it":    "Disponibile in 7 lingue.",
	},
}

func init() {
	bundle = i18n.NewBundle(language.English)

	for id, langs := range messages {
		for lang, text := range langs {
			bundle.AddMessages(language.Make(lang), &i18n.Message{
				ID:    id,
				Other: text,
			})
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
