// Package i18n provides translations for RadKeys using a single Go map.
// All languages live here — no JSON files, no embed. To add a string:
// add an entry to the messages map with a translation for each language.
// To add a language: add the code to Supported and fill translations below.
package i18n

// Supported lists the language codes the UI accepts.
var Supported = []string{"en", "pt-BR", "pt-PT", "es", "fr", "de", "it"}

var current = "en"

// messages is the single source of truth for all translations.
// Key = message ID. Value = map[language]translation.
// Missing translations fall back to English.
var messages = map[string]map[string]string{
	// ── Theme names (i18n) ─────────────────────────────────
	"theme.system": {
		"en": "System default", "pt-BR": "Padrão do sistema", "pt-PT": "Padrão do sistema",
		"es": "Predeterminado del sistema", "fr": "Défaut système", "de": "Systemstandard",
		"it": "Predefinito di sistema",
	},
	"theme.dracula": {
		"en": "Dracula", "pt-BR": "Dracula", "pt-PT": "Dracula",
		"es": "Drácula", "fr": "Dracula", "de": "Dracula", "it": "Dracula",
	},
	"theme.solarized_dark": {
		"en": "Solarized Dark", "pt-BR": "Solarized Escuro", "pt-PT": "Solarized Escuro",
		"es": "Solarized Oscuro", "fr": "Solarized Sombre", "de": "Solarized Dunkel",
		"it": "Solarized Scuro",
	},
	"theme.monokai": {
		"en": "Monokai", "pt-BR": "Monokai", "pt-PT": "Monokai",
		"es": "Monokai", "fr": "Monokai", "de": "Monokai", "it": "Monokai",
	},
	"theme.gruvbox_dark": {
		"en": "Gruvbox Dark", "pt-BR": "Gruvbox Escuro", "pt-PT": "Gruvbox Escuro",
		"es": "Gruvbox Oscuro", "fr": "Gruvbox Sombre", "de": "Gruvbox Dunkel",
		"it": "Gruvbox Scuro",
	},
	"theme.nord": {
		"en": "Nord", "pt-BR": "Nord", "pt-PT": "Nord",
		"es": "Nord", "fr": "Nord", "de": "Nord", "it": "Nord",
	},
	"theme.one_dark": {
		"en": "One Dark", "pt-BR": "One Dark", "pt-PT": "One Dark",
		"es": "One Dark", "fr": "One Dark", "de": "One Dark", "it": "One Dark",
	},
	"theme.tokyo_night": {
		"en": "Tokyo Night", "pt-BR": "Tokyo Night", "pt-PT": "Tokyo Night",
		"es": "Tokyo Night", "fr": "Tokyo Night", "de": "Tokyo Night", "it": "Tokyo Night",
	},
	"theme.catppuccin_mocha": {
		"en": "Catppuccin Mocha", "pt-BR": "Catppuccin Mocha", "pt-PT": "Catppuccin Mocha",
		"es": "Catppuccin Mocha", "fr": "Catppuccin Mocha", "de": "Catppuccin Mocha",
		"it": "Catppuccin Mocha",
	},
	"theme.solarized_light": {
		"en": "Solarized Light", "pt-BR": "Solarized Claro", "pt-PT": "Solarized Claro",
		"es": "Solarized Claro", "fr": "Solarized Clair", "de": "Solarized Hell",
		"it": "Solarized Chiaro",
	},
	"theme.gruvbox_light": {
		"en": "Gruvbox Light", "pt-BR": "Gruvbox Claro", "pt-PT": "Gruvbox Claro",
		"es": "Gruvbox Claro", "fr": "Gruvbox Clair", "de": "Gruvbox Hell",
		"it": "Gruvbox Chiaro",
	},
	"theme.light_gray": {
		"en": "Light Gray", "pt-BR": "Cinza Claro", "pt-PT": "Cinzento Claro",
		"es": "Gris Claro", "fr": "Gris Clair", "de": "Hellgrau", "it": "Grigio Chiaro",
	},
	"theme.dark_gray": {
		"en": "Dark Gray", "pt-BR": "Cinza Escuro", "pt-PT": "Cinzento Escuro",
		"es": "Gris Oscuro", "fr": "Gris Foncé", "de": "Dunkelgrau", "it": "Grigio Scuro",
	},

	// ── Action labels (canonical i18n keys) ─────────────
	"action.text": {
		"en": "Text", "pt-BR": "Texto", "pt-PT": "Texto", "es": "Texto",
		"fr": "Texte", "de": "Text", "it": "Testo",
	},
	"action.exec": {
		"en": "Execute command", "pt-BR": "Executar comando", "pt-PT": "Executar comando",
		"es": "Ejecutar comando", "fr": "Exécuter commande", "de": "Befehl ausführen",
		"it": "Esegui comando",
	},
	"action.copy": {
		"en": "Copy", "pt-BR": "Copiar", "pt-PT": "Copiar",
		"es": "Copiar", "fr": "Copier", "de": "Kopieren", "it": "Copia",
	},
	"action.paste": {
		"en": "Paste", "pt-BR": "Colar", "pt-PT": "Colar",
		"es": "Pegar", "fr": "Coller", "de": "Einfügen", "it": "Incolla",
	},
	"action.prev": {
		"en": "Back", "pt-BR": "Voltar", "pt-PT": "Voltar",
		"es": "Volver", "fr": "Retour", "de": "Zurück", "it": "Indietro",
	},
	"action.home": {
		"en": "Home", "pt-BR": "Início", "pt-PT": "Início",
		"es": "Inicio", "fr": "Accueil", "de": "Start", "it": "Home",
	},
	"action.navigate": {
		"en": "Navigate", "pt-BR": "Navegar", "pt-PT": "Navegar", "es": "Navegar",
		"fr": "Naviguer", "de": "Navigieren", "it": "Naviga",
	},
	"action.select_all": {
		"en": "Select All", "pt-BR": "Selecionar Tudo", "pt-PT": "Selecionar Tudo",
		"es": "Seleccionar Todo", "fr": "Sélectionner Tout", "de": "Alle auswählen",
		"it": "Seleziona Tutto",
	},
	"action.select_line": {
		"en": "Select Line", "pt-BR": "Selecionar Linha", "pt-PT": "Selecionar Linha",
		"es": "Seleccionar Línea", "fr": "Sélectionner Ligne", "de": "Zeile auswählen",
		"it": "Seleziona Riga",
	},
	"action.line_start": {
		"en": "Line Start", "pt-BR": "Início da Linha", "pt-PT": "Início da Linha",
		"es": "Inicio de Línea", "fr": "Début de Ligne", "de": "Zeilenanfang",
		"it": "Inizio Riga",
	},
	"action.line_end": {
		"en": "Line End", "pt-BR": "Fim da Linha", "pt-PT": "Fim da Linha",
		"es": "Fin de Línea", "fr": "Fin de Ligne", "de": "Zeilenende",
		"it": "Fine Riga",
	},
	"action.backspace": {
		"en": "Backspace", "pt-BR": "Backspace", "pt-PT": "Backspace",
		"es": "Retroceso", "fr": "Retour arrière", "de": "Rücktaste",
		"it": "Backspace",
	},
	"action.delete": {
		"en": "Delete", "pt-BR": "Delete", "pt-PT": "Delete",
		"es": "Suprimir", "fr": "Supprimer", "de": "Entf", "it": "Canc",
	},

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
	"button.close": {
		"en": "Close", "pt-BR": "Fechar", "pt-PT": "Fechar",
		"es": "Cerrar", "fr": "Fermer", "de": "Schließen", "it": "Chiudi",
	},
	"button.select_all": {
		"en": "Select All", "pt-BR": "Selecionar Tudo", "pt-PT": "Selecionar Tudo",
		"es": "Seleccionar Todo", "fr": "Sélectionner Tout", "de": "Alle auswählen",
		"it": "Seleziona Tutto",
	},
	"button.select_line": {
		"en": "Select Line", "pt-BR": "Selecionar Linha", "pt-PT": "Selecionar Linha",
		"es": "Seleccionar Línea", "fr": "Sélectionner Ligne", "de": "Zeile auswählen",
		"it": "Seleziona Riga",
	},
	"button.line_start": {
		"en": "Line Start", "pt-BR": "Início da Linha", "pt-PT": "Início da Linha",
		"es": "Inicio de Línea", "fr": "Début de Ligne", "de": "Zeilenanfang",
		"it": "Inizio Riga",
	},
	"button.line_end": {
		"en": "Line End", "pt-BR": "Fim da Linha", "pt-PT": "Fim da Linha",
		"es": "Fin de Línea", "fr": "Fin de Ligne", "de": "Zeilenende",
		"it": "Fine Riga",
	},
	"button.backspace": {
		"en": "Backspace", "pt-BR": "Backspace", "pt-PT": "Backspace",
		"es": "Retroceso", "fr": "Retour arrière", "de": "Rücktaste",
		"it": "Backspace",
	},
	"button.delete": {
		"en": "Delete", "pt-BR": "Delete", "pt-PT": "Delete",
		"es": "Suprimir", "fr": "Supprimer", "de": "Entf", "it": "Canc",
	},
	"button.exec": {
		"en": "Execute command", "pt-BR": "Executar comando", "pt-PT": "Executar comando",
		"es": "Ejecutar comando", "fr": "Exécuter commande", "de": "Befehl ausführen",
		"it": "Esegui comando",
	},
	"device_action.via_keypad_hint": {
		"en":    "Use the physical keypad for %s — clicking here would send the keystroke into RadKeys itself.",
		"pt-BR": "Use o teclado físico para %s — clicar aqui enviaria a tecla para o próprio RadKeys.",
		"pt-PT": "Use o teclado físico para %s — clicar aqui enviaria a tecla para o próprio RadKeys.",
		"es":    "Use el teclado físico para %s — hacer clic aquí enviaría la tecla al propio RadKeys.",
		"fr":    "Utilisez le clavier physique pour %s — cliquer ici enverrait la touche dans RadKeys lui-même.",
		"de":    "Physisches Keypad für %s nutzen — Klicken hier würde die Taste in RadKeys selbst senden.",
		"it":    "Usa il keypad fisico per %s — cliccare qui invierebbe il tasto in RadKeys stesso.",
	},

	// ── Status messages ───────────────────────────────────
	"status.mock_mode": {
		"en":    "No HID device found — running in mock mode (use on-screen buttons).",
		"pt-BR": "Dispositivo HID não encontrado — executando em modo mock (use os botões na tela).",
		"pt-PT": "Dispositivo HID não encontrado — a executar em modo mock (use os botões no ecrã).",
		"es":    "Dispositivo HID no encontrado — ejecutando en modo mock (use los botones en pantalla).",
		"fr":    "Aucun périphérique HID trouvé — mode mock actif (utilisez les boutons à l'écran).",
		"de":    "Kein HID-Gerät gefunden — Mock-Modus aktiv (Verwenden Sie die Bildschirmschaltflächen).",
		"it":    "Nessun dispositivo HID trovato — modalità mock attiva (usa i pulsanti a schermo).",
	},
	"status.device_command_failed": {
		"en":    "Device command failed: %s",
		"pt-BR": "Comando do dispositivo falhou: %s",
		"pt-PT": "Comando do dispositivo falhou: %s",
		"es":    "Error del comando del dispositivo: %s",
		"fr":    "Échec de la commande du périphérique : %s",
		"de":    "Gerätebefehl fehlgeschlagen: %s",
		"it":    "Comando dispositivo non riuscito: %s",
	},
	"status.out_of_grid": {
		"en":    "Device event out of grid bounds (row=%d, col=%d) for %dx%d.",
		"pt-BR": "Evento do dispositivo fora dos limites da grade (linha=%d, coluna=%d) para %dx%d.",
		"pt-PT": "Evento do dispositivo fora dos limites da grelha (linha=%d, coluna=%d) para %dx%d.",
		"es":    "Evento del dispositivo fuera de los límites de la cuadrícula (fila=%d, columna=%d) para %dx%d.",
		"fr":    "Événement périphérique hors limites de la grille (ligne=%d, colonne=%d) pour %dx%d.",
		"de":    "Geräteereignis außerhalb des Rasters (Zeile=%d, Spalte=%d) für %dx%d.",
		"it":    "Evento dispositivo fuori dalla griglia (riga=%d, colonna=%d) per %dx%d.",
	},
	"status.hid_read_failed": {
		"en":    "HID read failed. Hardware may be disconnected.",
		"pt-BR": "Falha na leitura HID. O hardware pode estar desconectado.",
		"pt-PT": "Falha na leitura HID. O hardware pode estar desligado.",
		"es":    "Error de lectura HID. El hardware puede estar desconectado.",
		"fr":    "Échec de lecture HID. Le matériel est peut-être déconnecté.",
		"de":    "HID-Lesefehler. Hardware möglicherweise getrennt.",
		"it":    "Lettura HID fallita. L'hardware potrebbe essere disconnesso.",
	},

	// ── Firmware version warning (one-shot at connect) ─────
	"firmware.outdated_title": {
		"en": "Firmware Update Required", "pt-BR": "Atualização de Firmware Necessária",
		"pt-PT": "Atualização de Firmware Necessária", "es": "Actualización de Firmware Requerida",
		"fr": "Mise à jour du firmware requise", "de": "Firmware-Aktualisierung erforderlich",
		"it": "Aggiornamento Firmware Necessario",
	},
	"firmware.outdated_message": {
		"en":    "Device firmware is v%d.%d, but v%d.%d or later is required.",
		"pt-BR": "O firmware do dispositivo é v%d.%d, mas v%d.%d ou superior é necessário.",
		"pt-PT": "O firmware do dispositivo é v%d.%d, mas v%d.%d ou superior é necessário.",
		"es":    "El firmware del dispositivo es v%d.%d, pero se requiere v%d.%d o posterior.",
		"fr":    "Le firmware du périphérique est v%d.%d, mais v%d.%d ou ultérieur est requis.",
		"de":    "Geräte-Firmware ist v%d.%d, aber v%d.%d oder neuer ist erforderlich.",
		"it":    "Il firmware del dispositivo è v%d.%d, ma è richiesta v%d.%d o successiva.",
	},
	"firmware.unknown_message": {
		"en":    "Firmware version unknown — update to v%d.%d or later.",
		"pt-BR": "Versão de firmware desconhecida — atualize para v%d.%d ou superior.",
		"pt-PT": "Versão de firmware desconhecida — atualize para v%d.%d ou superior.",
		"es":    "Versión de firmware desconocida — actualice a v%d.%d o posterior.",
		"fr":    "Version du firmware inconnue — mettez à jour vers v%d.%d ou ultérieur.",
		"de":    "Firmware-Version unbekannt — aktualisieren Sie auf v%d.%d oder neuer.",
		"it":    "Versione firmware sconosciuta — aggiornare a v%d.%d o successiva.",
	},

	// ── Config error dialog ───────────────────────────────
	"error.config_title": {
		"en": "RadKeys — Config Error", "pt-BR": "RadKeys — Erro de Configuração",
		"pt-PT": "RadKeys — Erro de Configuração", "es": "RadKeys — Error de Configuración",
		"fr": "RadKeys — Erreur de configuration", "de": "RadKeys — Konfigurationsfehler",
		"it": "RadKeys — Errore di Configurazione",
	},
	"error.config_message": {
		"en":    "The configuration file contains an error:",
		"pt-BR": "O arquivo de configuração contém um erro:",
		"pt-PT": "O ficheiro de configuração contém um erro:",
		"es":    "El archivo de configuración contiene un error:",
		"fr":    "Le fichier de configuration contient une erreur :",
		"de":    "Die Konfigurationsdatei enthält einen Fehler:",
		"it":    "Il file di configurazione contiene un errore:",
	},
	"error.config_fix": {
		"en":    "Fix the error above and restart RadKeys.",
		"pt-BR": "Corrija o erro acima e reinicie o RadKeys.",
		"pt-PT": "Corrija o erro acima e reinicie o RadKeys.",
		"es":    "Corrija el error anterior y reinicie RadKeys.",
		"fr":    "Corrigez l'erreur ci-dessus et redémarrez RadKeys.",
		"de":    "Beheben Sie den obigen Fehler und starten Sie RadKeys neu.",
		"it":    "Correggi l'errore sopra e riavvia RadKeys.",
	},
	"error.open_file": {
		"en":    "Open file to edit",
		"pt-BR": "Abrir arquivo para editar",
		"pt-PT": "Abrir ficheiro para editar",
		"es":    "Abrir archivo para editar",
		"fr":    "Ouvrir le fichier pour modifier",
		"de":    "Datei zum Bearbeiten öffnen",
		"it":    "Apri file per modificare",
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
		"en": "License: RadKeys Source-Available v1.0", "pt-BR": "Licença: RadKeys Source-Available v1.0", "pt-PT": "Licença: RadKeys Source-Available v1.0",
		"es": "Licencia: RadKeys Source-Available v1.0", "fr": "Licence : RadKeys Source-Available v1.0", "de": "Lizenz: RadKeys Source-Available v1.0",
		"it": "Licenza: RadKeys Source-Available v1.0",
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

	// ── Editor ────────────────────────────────────────────
	"editor.title": {
		"en": "RadKeys Config Editor", "pt-BR": "Editor de Configuração do RadKeys",
		"pt-PT": "Editor de Configuração do RadKeys", "es": "Editor de Configuración de RadKeys",
		"fr": "Éditeur de configuration RadKeys", "de": "RadKeys-Konfigurationseditor",
		"it": "Editor di configurazione RadKeys",
	},
	"editor.tab_app_settings": {
		"en": "App Settings", "pt-BR": "Configurações do App",
		"pt-PT": "Definições da Aplicação", "es": "Configuración de la App",
		"fr": "Paramètres de l'application", "de": "App-Einstellungen",
		"it": "Impostazioni App",
	},
	"editor.tab_buttons": {
		"en": "Buttons & Layers", "pt-BR": "Botões e Camadas",
		"pt-PT": "Botões e Camadas", "es": "Botones y Capas",
		"fr": "Boutons et couches", "de": "Tasten und Ebenen",
		"it": "Pulsanti e Livelli",
	},
	"editor.layer": {
		"en": "Layer", "pt-BR": "Camada", "pt-PT": "Camada", "es": "Capa",
		"fr": "Couche", "de": "Ebene", "it": "Livello",
	},
	"editor.add_layer": {
		"en": "Add layer", "pt-BR": "Adicionar camada", "pt-PT": "Adicionar camada",
		"es": "Añadir capa", "fr": "Ajouter une couche", "de": "Ebene hinzufügen",
		"it": "Aggiungi livello",
	},
	"editor.remove_layer": {
		"en": "Remove layer", "pt-BR": "Remover camada", "pt-PT": "Remover camada",
		"es": "Eliminar capa", "fr": "Supprimer la couche", "de": "Ebene entfernen",
		"it": "Rimuovi livello",
	},
	"editor.rename_layer": {
		"en": "Rename", "pt-BR": "Renomear", "pt-PT": "Renomear", "es": "Renombrar",
		"fr": "Renommer", "de": "Umbenennen", "it": "Rinomina",
	},
	"editor.layer_id": {
		"en": "Layer ID", "pt-BR": "ID da camada", "pt-PT": "ID da camada",
		"es": "ID de capa", "fr": "ID de couche", "de": "Ebenen-ID",
		"it": "ID livello",
	},
	"editor.layer_name": {
		"en": "Layer name", "pt-BR": "Nome da camada", "pt-PT": "Nome da camada",
		"es": "Nombre de capa", "fr": "Nom de la couche", "de": "Ebenenname",
		"it": "Nome livello",
	},
	"editor.label": {
		"en": "Label", "pt-BR": "Rótulo", "pt-PT": "Rótulo", "es": "Etiqueta",
		"fr": "Libellé", "de": "Beschriftung", "it": "Etichetta",
	},
	"editor.action": {
		"en": "Action", "pt-BR": "Ação", "pt-PT": "Ação", "es": "Acción",
		"fr": "Action", "de": "Aktion", "it": "Azione",
	},
	"editor.target": {
		"en": "Target", "pt-BR": "Destino", "pt-PT": "Destino", "es": "Destino",
		"fr": "Cible", "de": "Ziel", "it": "Destinazione",
	},
	"editor.content": {
		"en": "Content", "pt-BR": "Conteúdo", "pt-PT": "Conteúdo", "es": "Contenido",
		"fr": "Contenu", "de": "Inhalt", "it": "Contenuto",
	},
	"editor.remove": {
		"en": "Remove", "pt-BR": "Remover", "pt-PT": "Remover", "es": "Eliminar",
		"fr": "Supprimer", "de": "Entfernen", "it": "Rimuovi",
	},
	"editor.action_text": {
		"en": "Text", "pt-BR": "Texto", "pt-PT": "Texto", "es": "Texto",
		"fr": "Texte", "de": "Text", "it": "Testo",
	},
	"editor.action_navigate": {
		"en": "Navigate", "pt-BR": "Navegar", "pt-PT": "Navegar", "es": "Navegar",
		"fr": "Naviguer", "de": "Navigieren", "it": "Naviga",
	},
	"editor.action_exec": {
		"en": "Execute command", "pt-BR": "Executar comando", "pt-PT": "Executar comando",
		"es": "Ejecutar comando", "fr": "Exécuter commande", "de": "Befehl ausführen",
		"it": "Esegui comando",
	},
	"editor.empty_cell": {
		"en": "+", "pt-BR": "+", "pt-PT": "+", "es": "+", "fr": "+",
		"de": "+", "it": "+",
	},
	"editor.click_to_edit": {
		"en":    "Click a button to edit it.",
		"pt-BR": "Clique em um botão para editá-lo.",
		"pt-PT": "Clique num botão para editá-lo.",
		"es":    "Haga clic en un botón para editarlo.",
		"fr":    "Cliquez sur un bouton pour le modifier.",
		"de":    "Klicken Sie auf eine Taste, um sie zu bearbeiten.",
		"it":    "Clicca un pulsante per modificarlo.",
	},
	"editor.click_to_add": {
		"en":    "Click an empty cell to add a button.",
		"pt-BR": "Clique em uma célula vazia para adicionar um botão.",
		"pt-PT": "Clique numa célula vazia para adicionar um botão.",
		"es":    "Haga clic en una celda vacía para añadir un botón.",
		"fr":    "Cliquez sur une cellule vide pour ajouter un bouton.",
		"de":    "Klicken Sie auf eine leere Zelle, um eine Taste hinzuzufügen.",
		"it":    "Clicca una cella vuota per aggiungere un pulsante.",
	},
	"editor.save": {
		"en": "Save", "pt-BR": "Salvar", "pt-PT": "Guardar", "es": "Guardar",
		"fr": "Enregistrer", "de": "Speichern", "it": "Salva",
	},
	"editor.save_as": {
		"en": "Save As…", "pt-BR": "Salvar como…", "pt-PT": "Guardar como…",
		"es": "Guardar como…", "fr": "Enregistrer sous…", "de": "Speichern unter…",
		"it": "Salva come…",
	},
	"editor.open": {
		"en": "Open…", "pt-BR": "Abrir…", "pt-PT": "Abrir…", "es": "Abrir…",
		"fr": "Ouvrir…", "de": "Öffnen…", "it": "Apri…",
	},
	"editor.new": {
		"en": "New", "pt-BR": "Novo", "pt-PT": "Novo", "es": "Nuevo",
		"fr": "Nouveau", "de": "Neu", "it": "Nuovo",
	},
	"editor.unsaved_title": {
		"en": "*", "pt-BR": "*", "pt-PT": "*", "es": "*", "fr": "*", "de": "*",
		"it": "*",
	},
	"editor.unsaved": {
		"en": "unsaved", "pt-BR": "não salvo", "pt-PT": "não guardado",
		"es": "sin guardar", "fr": "non enregistré", "de": "ungespeichert",
		"it": "non salvato",
	},
	"editor.close_file": {
		"en": "Close file", "pt-BR": "Fechar arquivo", "pt-PT": "Fechar ficheiro",
		"es": "Cerrar archivo", "fr": "Fermer le fichier", "de": "Datei schließen",
		"it": "Chiudi file",
	},
	"editor.quit": {
		"en": "Quit", "pt-BR": "Sair", "pt-PT": "Sair", "es": "Salir",
		"fr": "Quitter", "de": "Beenden", "it": "Esci",
	},
	"editor.confirm_discard": {
		"en":    "Discard unsaved changes?",
		"pt-BR": "Descartar alterações não salvas?",
		"pt-PT": "Descartar alterações não guardadas?",
		"es":    "¿Descartar cambios no guardados?",
		"fr":    "Ignorer les modifications non enregistrées ?",
		"de":    "Ungespeicherte Änderungen verwerfen?",
		"it":    "Scartare le modifiche non salvate?",
	},
	"editor.confirm_discard_title": {
		"en": "Unsaved Changes", "pt-BR": "Alterações Não Salvas",
		"pt-PT": "Alterações Não Guardadas", "es": "Cambios Sin Guardar",
		"fr": "Modifications non enregistrées", "de": "Nicht gespeicherte Änderungen",
		"it": "Modifiche Non Salvate",
	},
	"editor.problems_title": {
		"en": "Problems", "pt-BR": "Problemas", "pt-PT": "Problemas", "es": "Problemas",
		"fr": "Problèmes", "de": "Probleme", "it": "Problemi",
	},
	"editor.no_problems": {
		"en": "No problems", "pt-BR": "Sem problemas", "pt-PT": "Sem problemas",
		"es": "Sin problemas", "fr": "Aucun problème", "de": "Keine Probleme",
		"it": "Nessun problema",
	},
	"editor.out_of_grid": {
		"en":    "Button %q is outside the grid — move it or remove it.",
		"pt-BR": "O botão %q está fora da grade — mova-o ou remova-o.",
		"pt-PT": "O botão %q está fora da grelha — mova-o ou remova-o.",
		"es":    "El botón %q está fuera de la cuadrícula — muévalo o elimínelo.",
		"fr":    "Le bouton %q est hors de la grille — déplacez-le ou supprimez-le.",
		"de":    "Die Taste %q befindet sich außerhalb des Rasters — verschieben oder entfernen Sie sie.",
		"it":    "Il pulsante %q è fuori dalla griglia — spostalo o rimuovilo.",
	},
	"editor.duplicate_pos": {
		"en":    "Buttons %q and %q share the same cell.",
		"pt-BR": "Os botões %q e %q compartilham a mesma célula.",
		"pt-PT": "Os botões %q e %q partilham a mesma célula.",
		"es":    "Los botones %q y %q comparten la misma celda.",
		"fr":    "Les boutons %q et %q partagent la même cellule.",
		"de":    "Die Tasten %q und %q teilen sich dieselbe Zelle.",
		"it":    "I pulsanti %q e %q condividono la stessa cella.",
	},
	"editor.bad_target": {
		"en":    "Target %q does not exist.",
		"pt-BR": "O destino %q não existe.",
		"pt-PT": "O destino %q não existe.",
		"es":    "El destino %q no existe.",
		"fr":    "La cible %q n'existe pas.",
		"de":    "Das Ziel %q existiert nicht.",
		"it":    "La destinazione %q non esiste.",
	},
	"editor.label_required": {
		"en": "Label is required.", "pt-BR": "Rótulo obrigatório.", "pt-PT": "Rótulo obrigatório.",
		"es": "Etiqueta obligatoria.", "fr": "Libellé requis.", "de": "Beschriftung erforderlich.",
		"it": "Etichetta obbligatoria.",
	},
	"editor.content_required": {
		"en":    "Content is required for Text actions.",
		"pt-BR": "Conteúdo obrigatório para ações de Texto.",
		"pt-PT": "Conteúdo obrigatório para ações de Texto.",
		"es":    "Contenido obligatorio para acciones de Texto.",
		"fr":    "Un contenu est requis pour les actions Texte.",
		"de":    "Inhalt erforderlich für Text-Aktionen.",
		"it":    "Contenuto obbligatorio per le azioni Testo.",
	},
	"editor.target_required": {
		"en":    "Target is required for Navigate actions.",
		"pt-BR": "Destino obrigatório para ações de Navegar.",
		"pt-PT": "Destino obrigatório para ações de Navegar.",
		"es":    "Destino obligatorio para acciones de Navegar.",
		"fr":    "Une cible est requise pour les actions Naviguer.",
		"de":    "Ziel erforderlich für Navigieren-Aktionen.",
		"it":    "Destinazione obbligatoria per le azioni Naviga.",
	},
	"editor.cannot_remove_root_screen": {
		"en":    "The root layer cannot be removed.",
		"pt-BR": "A camada raiz não pode ser removida.",
		"pt-PT": "A camada raiz não pode ser removida.",
		"es":    "La capa raíz no se puede eliminar.",
		"fr":    "La couche racine ne peut pas être supprimée.",
		"de":    "Die Root-Ebene kann nicht entfernt werden.",
		"it":    "Il layer radice non può essere rimosso.",
	},
	"editor.new_layer_name": {
		"en": "New Layer", "pt-BR": "Nova Camada", "pt-PT": "Nova Camada",
		"es": "Nueva Capa", "fr": "Nouvelle couche", "de": "Neue Ebene",
		"it": "Nuovo Livello",
	},
	"editor.file_menu": {
		"en": "File", "pt-BR": "Arquivo", "pt-PT": "Ficheiro", "es": "Archivo",
		"fr": "Fichier", "de": "Datei", "it": "File",
	},
	"editor.action_rejects_target": {
		"en":    "Action %q does not use a target.",
		"pt-BR": "A ação %q não usa destino.",
		"pt-PT": "A ação %q não usa destino.",
		"es":    "La acción %q no usa destino.",
		"fr":    "L'action %q n'utilise pas de cible.",
		"de":    "Die Aktion %q verwendet kein Ziel.",
		"it":    "L'azione %q non usa destinazione.",
	},
	"editor.action_rejects_content": {
		"en":    "Action %q does not use content.",
		"pt-BR": "A ação %q não usa conteúdo.",
		"pt-PT": "A ação %q não usa conteúdo.",
		"es":    "La acción %q no usa contenido.",
		"fr":    "L'action %q n'utilise pas de contenu.",
		"de":    "Die Aktion %q verwendet keinen Inhalt.",
		"it":    "L'azione %q non usa contenuto.",
	},
	"editor.invalid_action": {
		"en":    "Action %q is not valid.",
		"pt-BR": "A ação %q não é válida.",
		"pt-PT": "A ação %q não é válida.",
		"es":    "La acción %q no es válida.",
		"fr":    "L'action %q n'est pas valide.",
		"de":    "Die Aktion %q ist ungültig.",
		"it":    "L'azione %q non è valida.",
	},
	"editor.save_blocked_title": {
		"en": "Cannot Save", "pt-BR": "Não é possível salvar",
		"pt-PT": "Não é possível guardar", "es": "No se puede guardar",
		"fr": "Impossible d'enregistrer", "de": "Speichern nicht möglich",
		"it": "Impossibile salvare",
	},
	"editor.save_blocked_message": {
		"en":    "Fix the problems before saving.",
		"pt-BR": "Corrija os problemas antes de salvar.",
		"pt-PT": "Corrija os problemas antes de guardar.",
		"es":    "Corrija los problemas antes de guardar.",
		"fr":    "Corrigez les problèmes avant d'enregistrer.",
		"de":    "Beheben Sie die Probleme vor dem Speichern.",
		"it":    "Risolve i problemi prima di salvare.",
	},
	"editor.cannot_remove_last_screen": {
		"en":    "Cannot remove the last layer.",
		"pt-BR": "Não é possível remover a última camada.",
		"pt-PT": "Não é possível remover a última camada.",
		"es":    "No se puede eliminar la última capa.",
		"fr":    "Impossible de supprimer la dernière couche.",
		"de":    "Die letzte Ebene kann nicht entfernt werden.",
		"it":    "Impossibile rimuovere l'ultimo livello.",
	},
	"editor.cannot_remove_targeted_screen": {
		"en":    "Cannot remove a layer that is targeted by a navigate button.",
		"pt-BR": "Não é possível remover uma camada que é destino de um botão de navegação.",
		"pt-PT": "Não é possível remover uma camada que é destino de um botão de navegação.",
		"es":    "No se puede eliminar una capa que es destino de un botón de navegación.",
		"fr":    "Impossible de supprimer une couche ciblée par un bouton de navigation.",
		"de":    "Eine Ebene, die Ziel eines Navigieren-Buttons ist, kann nicht entfernt werden.",
		"it":    "Impossibile rimuovere un livello destinazione di un pulsante di navigazione.",
	},
	"editor.confirm_remove_screen": {
		"en":    "Remove this layer and all its buttons?",
		"pt-BR": "Remover esta camada e todos os seus botões?",
		"pt-PT": "Remover esta camada e todos os seus botões?",
		"es":    "¿Eliminar esta capa y todos sus botones?",
		"fr":    "Supprimer cette couche et tous ses boutons ?",
		"de":    "Diese Ebene und alle ihre Tasten entfernen?",
		"it":    "Rimuovere questo livello e tutti i suoi pulsanti?",
	},
	"editor.cancel": {
		"en": "Cancel", "pt-BR": "Cancelar", "pt-PT": "Cancelar", "es": "Cancelar",
		"fr": "Annuler", "de": "Abbrechen", "it": "Annulla",
	},
	"editor.discard": {
		"en": "Discard", "pt-BR": "Descartar", "pt-PT": "Descartar", "es": "Descartar",
		"fr": "Ignorer", "de": "Verwerfen", "it": "Scarta",
	},
	"editor.select_target": {
		"en": "Select target…", "pt-BR": "Selecione o destino…",
		"pt-PT": "Selecione o destino…", "es": "Seleccione destino…",
		"fr": "Sélectionnez la cible…", "de": "Ziel auswählen…",
		"it": "Seleziona destinazione…",
	},
	"editor.grid_size": {
		"en": "Grid size", "pt-BR": "Tamanho da grade", "pt-PT": "Tamanho da grelha",
		"es": "Tamaño de la cuadrícula", "fr": "Taille de la grille",
		"de": "Rastergröße", "it": "Dimensione griglia",
	},
	"editor.app_name": {
		"en": "App name", "pt-BR": "Nome do app", "pt-PT": "Nome da aplicação",
		"es": "Nombre de la app", "fr": "Nom de l'application", "de": "App-Name",
		"it": "Nome app",
	},
	"editor.hex_format": {
		"en": "Hex (e.g. 0x1234)", "pt-BR": "Hex (ex. 0x1234)",
		"pt-PT": "Hex (ex. 0x1234)", "es": "Hex (p. ej. 0x1234)",
		"fr": "Hex (ex. 0x1234)", "de": "Hex (z. B. 0x1234)",
		"it": "Hex (es. 0x1234)",
	},
	"settings.invalid_hex": {
		"en": "Invalid hex value", "pt-BR": "Valor hexadecimal inválido",
		"pt-PT": "Valor hexadecimal inválido", "es": "Valor hexadecimal inválido",
		"fr": "Valeur hexadécimale invalide", "de": "Ungültiger Hexadezimalwert",
		"it": "Valore esadecimale non valido",
	},
}

// SetLanguage switches the active locale. Falls back to English on empty input.
func SetLanguage(lang string) {
	if lang == "" {
		lang = "en"
	}
	current = lang
}

// IsSupported reports whether lang is one of the supported language codes.
func IsSupported(lang string) bool {
	for _, s := range Supported {
		if s == lang {
			return true
		}
	}
	return false
}

// T translates a message ID to the active language. Falls back to English,
// then to the raw ID.
func T(id string) string {
	msg, ok := messages[id]
	if !ok {
		return id
	}
	if text, ok := msg[current]; ok {
		return text
	}
	if text, ok := msg["en"]; ok {
		return text
	}
	return id
}
