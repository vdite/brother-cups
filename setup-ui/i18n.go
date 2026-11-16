package main

import (
	"fmt"
	"net/http"
	"strings"
)

const (
	langEN     = "en"
	langDE     = "de"
	langCookie = "brother_cups_lang"
)

// messages holds every text of the setup interface. English is the fallback
// for missing keys. To add a language, add a map here and a dictionary in
// web/app.js.
var messages = map[string]map[string]string{
	langEN: {
		"page.title":           "Set up printer · Brother CUPS",
		"login.page_title":     "Log in · Brother CUPS",
		"brand.home":           "Brother CUPS home",
		"brand.tagline":        "AirPrint for older printers",
		"nav.account":          "Account menu",
		"nav.expert":           "Expert view",
		"nav.logout":           "Log out",
		"nav.no_password":      "No password",
		"hero.eyebrow":         "Print server for your home network",
		"hero.title":           "Your printer, back on the network.",
		"hero.text":            "Set it up, share it via AirPrint and print from Mac, iPhone or iPad, without the old vendor driver.",
		"ready.label":          "Ready to print",
		"ready.title":          "Your printers",
		"ready.airprint":       "AirPrint active",
		"jobs.none":            "No pending jobs",
		"jobs.one":             "1 job in the queue",
		"jobs.many":            "%d jobs in the queue",
		"action.test_page":     "Test page",
		"action.details":       "Details",
		"setup.label_more":     "Set up another printer",
		"setup.label_first":    "Ready in three steps",
		"setup.title":          "Add a Brother printer",
		"setup.text":           "We search your home network for your printer and only show matching free Brother drivers.",
		"step1.title":          "Choose printer",
		"step1.found":          "Printers found",
		"step1.searching":      "Searching the network …",
		"step1.hint":           "This can take a few seconds.",
		"manual.summary":       "My printer isn't listed",
		"manual.ip":            "Printer IP address",
		"manual.placeholder":   "e.g. 192.168.1.50",
		"manual.hint":          "You'll find the IP address in your router. Port 9100 is used automatically.",
		"step2.title":          "Model and settings",
		"model.label":          "Printer model",
		"model.choose":         "Choose model …",
		"model.hint":           "The matching free brlaser driver is used automatically.",
		"name.label":           "Display name",
		"paper.label":          "Paper size",
		"duplex.label":         "Enable two-sided printing",
		"duplex.hint":          "For models with an automatic duplex unit.",
		"advanced.summary":     "Advanced setting",
		"queue.label":          "Internal queue name",
		"queue.placeholder":    "generated automatically",
		"privacy.note":         "Setup stays within your home network.",
		"setup.submit":         "Set up printer",
		"footer.text":          "Brother CUPS uses CUPS and the free brlaser driver.",
		"footer.expert":        "Open CUPS expert view",
		"footer.language":      "Language",
		"login.title":          "Set up printer",
		"login.intro":          "Log in with the credentials you chose during installation.",
		"login.username":       "User name",
		"login.password":       "Password",
		"login.submit":         "Log in",
		"login.helper":         "The password is the value of the environment variable",
		"login.too_many":       "Too many login attempts. Please wait a moment.",
		"login.wrong":          "User name or password is incorrect.",
		"state.stopped":        "Stopped",
		"state.printing":       "Printing",
		"state.ready":          "Ready",
		"state.unknown":        "Status unknown",
		"err.bad_request":      "Invalid request",
		"err.invalid_session":  "Invalid session",
		"err.session_failed":   "Session could not be created",
		"err.cups_not_ready":   "CUPS isn't ready yet. Please reload the page in a moment.",
		"err.discovery":        "Printer discovery is not available right now.",
		"err.name_missing":     "Please enter a name for the printer.",
		"err.model":            "Please choose a supported Brother model.",
		"err.setup_failed":     "The printer could not be set up. Open the expert view for details.",
		"err.printer_gone":     "The selected printer no longer exists.",
		"err.test_failed":      "The test page could not be sent.",
		"err.invalid_ip":       "Please enter a valid printer IP address.",
		"err.choose_device":    "Please choose a printer that was found.",
		"err.unsupported_uri":  "This printer connection is not supported.",
		"flash.configured":     "%s is set up and shared via AirPrint.",
		"flash.test_sent":      "Test page sent to the printer.",
		"default.printer_name": "Brother printer",
	},
	langDE: {
		"page.title":           "Drucker einrichten · Brother CUPS",
		"login.page_title":     "Anmelden · Brother CUPS",
		"brand.home":           "Brother CUPS Startseite",
		"brand.tagline":        "AirPrint für ältere Drucker",
		"nav.account":          "Kontomenü",
		"nav.expert":           "Expertenansicht",
		"nav.logout":           "Abmelden",
		"nav.no_password":      "Ohne Passwort",
		"hero.eyebrow":         "Druckserver im Heimnetz",
		"hero.title":           "Dein Drucker, wieder im Netz.",
		"hero.text":            "Einrichten, per AirPrint freigeben und von Mac, iPhone oder iPad drucken – ohne alten Herstellertreiber.",
		"ready.label":          "Bereit zum Drucken",
		"ready.title":          "Eingerichtete Drucker",
		"ready.airprint":       "AirPrint aktiv",
		"jobs.none":            "Keine offenen Aufträge",
		"jobs.one":             "1 Auftrag in der Warteschlange",
		"jobs.many":            "%d Aufträge in der Warteschlange",
		"action.test_page":     "Testseite",
		"action.details":       "Details",
		"setup.label_more":     "Weiteren Drucker einrichten",
		"setup.label_first":    "In drei Schritten startklar",
		"setup.title":          "Brother-Drucker hinzufügen",
		"setup.text":           "Wir suchen im Heimnetz nach Deinem Drucker und zeigen nur passende freie Brother-Treiber an.",
		"step1.title":          "Drucker auswählen",
		"step1.found":          "Gefundene Drucker",
		"step1.searching":      "Suche im Netzwerk läuft …",
		"step1.hint":           "Das kann einige Sekunden dauern.",
		"manual.summary":       "Mein Drucker wird nicht gefunden",
		"manual.ip":            "IP-Adresse des Druckers",
		"manual.placeholder":   "z. B. 192.168.1.50",
		"manual.hint":          "Die IP-Adresse findest Du in Deinem Router. Port 9100 wird automatisch verwendet.",
		"step2.title":          "Modell und Einstellungen",
		"model.label":          "Druckermodell",
		"model.choose":         "Modell auswählen …",
		"model.hint":           "Der passende freie brlaser-Treiber wird automatisch verwendet.",
		"name.label":           "Anzeigename",
		"paper.label":          "Papierformat",
		"duplex.label":         "Beidseitigen Druck aktivieren",
		"duplex.hint":          "Für Modelle mit automatischer Duplexeinheit.",
		"advanced.summary":     "Erweiterte Einstellung",
		"queue.label":          "Interner Warteschlangenname",
		"queue.placeholder":    "wird automatisch erzeugt",
		"privacy.note":         "Die Einrichtung bleibt in Deinem Heimnetz.",
		"setup.submit":         "Drucker einrichten",
		"footer.text":          "Brother CUPS verwendet CUPS und den freien brlaser-Treiber.",
		"footer.expert":        "CUPS-Expertenansicht öffnen",
		"footer.language":      "Sprache",
		"login.title":          "Drucker einrichten",
		"login.intro":          "Melde Dich mit den Zugangsdaten an, die Du bei der Installation festgelegt hast.",
		"login.username":       "Benutzername",
		"login.password":       "Passwort",
		"login.submit":         "Anmelden",
		"login.helper":         "Das Passwort entspricht der Umgebungsvariable",
		"login.too_many":       "Zu viele Anmeldeversuche. Bitte kurz warten.",
		"login.wrong":          "Benutzername oder Passwort stimmen nicht.",
		"state.stopped":        "Angehalten",
		"state.printing":       "Druckt gerade",
		"state.ready":          "Bereit",
		"state.unknown":        "Status unbekannt",
		"err.bad_request":      "Ungültige Anfrage",
		"err.invalid_session":  "Ungültige Sitzung",
		"err.session_failed":   "Sitzung konnte nicht erstellt werden",
		"err.cups_not_ready":   "CUPS ist noch nicht bereit. Bitte lade die Seite gleich noch einmal.",
		"err.discovery":        "Die Druckersuche ist gerade nicht verfügbar.",
		"err.name_missing":     "Bitte gib einen Namen für den Drucker ein.",
		"err.model":            "Bitte wähle ein unterstütztes Brother-Modell aus.",
		"err.setup_failed":     "Der Drucker konnte nicht eingerichtet werden. Öffne die Expertenansicht für Details.",
		"err.printer_gone":     "Der ausgewählte Drucker existiert nicht mehr.",
		"err.test_failed":      "Die Testseite konnte nicht gesendet werden.",
		"err.invalid_ip":       "Bitte gib eine gültige IP-Adresse des Druckers ein.",
		"err.choose_device":    "Bitte wähle einen gefundenen Drucker aus.",
		"err.unsupported_uri":  "Diese Druckerverbindung wird nicht unterstützt.",
		"flash.configured":     "%s ist eingerichtet und per AirPrint freigegeben.",
		"flash.test_sent":      "Testseite wurde an den Drucker gesendet.",
		"default.printer_name": "Brother Drucker",
	},
}

// tr returns the text for key in the given language, formatted with args.
func tr(lang, key string, args ...any) string {
	text, ok := messages[lang][key]
	if !ok {
		text, ok = messages[langEN][key]
	}
	if !ok {
		return key
	}
	if len(args) > 0 {
		return fmt.Sprintf(text, args...)
	}
	return text
}

func supportedLang(value string) bool {
	_, ok := messages[value]
	return ok
}

// requestLang picks the language from the language cookie, then from the
// browser's Accept-Language header. Without a match it falls back to English.
func requestLang(r *http.Request) string {
	if cookie, err := r.Cookie(langCookie); err == nil && supportedLang(cookie.Value) {
		return cookie.Value
	}
	for _, part := range strings.Split(r.Header.Get("Accept-Language"), ",") {
		tag := strings.ToLower(strings.TrimSpace(strings.SplitN(part, ";", 2)[0]))
		if tag == "" || tag == "*" {
			continue
		}
		primary, _, _ := strings.Cut(tag, "-")
		if supportedLang(primary) {
			return primary
		}
	}
	return langEN
}

// setLangFromQuery stores a language chosen via ?lang= in a cookie.
// It reports whether the request carried a valid choice.
func setLangFromQuery(w http.ResponseWriter, r *http.Request) bool {
	choice := r.URL.Query().Get("lang")
	if !supportedLang(choice) {
		return false
	}
	http.SetCookie(w, &http.Cookie{
		Name:     langCookie,
		Value:    choice,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   365 * 24 * 60 * 60,
	})
	return true
}
