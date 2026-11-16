package main

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestNormalizeQueueName(t *testing.T) {
	tests := map[string]string{
		"Brother MFC-7460DN":     "Brother_MFC-7460DN",
		"  Mein Drucker / Büro ": "Mein_Drucker_Buero",
		"---":                    "",
		"MFC_7460DN":             "MFC_7460DN",
	}
	for input, expected := range tests {
		if actual := normalizeQueueName(input); actual != expected {
			t.Errorf("normalizeQueueName(%q) = %q, erwartet %q", input, actual, expected)
		}
	}
}

func TestRequestedDeviceURI(t *testing.T) {
	t.Run("discovered", func(t *testing.T) {
		request := httptest.NewRequest("POST", "/printer/configure", strings.NewReader(url.Values{
			"connection": {"discovered"},
			"device_uri": {"dnssd://Brother%20MFC-7460DN._pdl-datastream._tcp.local/"},
		}.Encode()))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		_ = request.ParseForm()
		actual, err := requestedDeviceURI(request)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(actual, "dnssd://Brother") {
			t.Fatalf("unerwartete URI: %s", actual)
		}
	})

	t.Run("manual", func(t *testing.T) {
		request := httptest.NewRequest("POST", "/printer/configure", strings.NewReader(url.Values{
			"connection": {"manual"},
			"printer_ip": {"192.168.1.50"},
		}.Encode()))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		_ = request.ParseForm()
		actual, err := requestedDeviceURI(request)
		if err != nil {
			t.Fatal(err)
		}
		if actual != "socket://192.168.1.50:9100" {
			t.Fatalf("unerwartete URI: %s", actual)
		}
	})

	t.Run("rejects shell and file schemes", func(t *testing.T) {
		request := httptest.NewRequest("POST", "/printer/configure", strings.NewReader(url.Values{
			"connection": {"discovered"},
			"device_uri": {"file:///tmp/output"},
		}.Encode()))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		_ = request.ParseForm()
		if _, err := requestedDeviceURI(request); err == nil {
			t.Fatal("file:-URI wurde nicht abgelehnt")
		}
	})
}

func TestFriendlyDeviceName(t *testing.T) {
	actual := friendlyDeviceName("dnssd://Brother%20MFC-7460DN._pdl-datastream._tcp.local/")
	if actual != "Brother MFC-7460DN" {
		t.Fatalf("friendlyDeviceName = %q", actual)
	}
}

func TestParsePrinterURIs(t *testing.T) {
	parsed := parsePrinterURIs("device for Brother_MFC-7460DN: dnssd://Brother%20MFC-7460DN._pdl-datastream._tcp.local/\n")
	if parsed["Brother_MFC-7460DN"] == "" {
		t.Fatal("Drucker-URI wurde nicht erkannt")
	}
}

func TestParseJobCountsWithHyphenatedQueue(t *testing.T) {
	parsed := parseJobCounts("Brother_MFC-7460DN-12 root 1024 Sun Nov 15 13:57:51 2026\n")
	if parsed["Brother_MFC-7460DN"] != 1 {
		t.Fatalf("Job wurde falsch zugeordnet: %#v", parsed)
	}
}

func TestPreferFriendlyDevicesHidesLPDDuplicates(t *testing.T) {
	devices := []device{
		{URI: "dnssd://Brother%20MFC._pdl-datastream._tcp.local/", Name: "Brother MFC", Protocol: "DNSSD"},
		{URI: "lpd://BRN123/BINARY_P1", Name: "BRN123/BINARY P1", Protocol: "LPD"},
	}
	filtered := preferFriendlyDevices(devices)
	if len(filtered) != 1 || filtered[0].Protocol != "DNSSD" {
		t.Fatalf("LPD-Duplikat wurde nicht ausgeblendet: %#v", filtered)
	}
}

func TestPreferFriendlyDevicesKeepsLPDAsFallback(t *testing.T) {
	devices := []device{{URI: "lpd://BRN123/BINARY_P1", Name: "BRN123/BINARY P1", Protocol: "LPD"}}
	filtered := preferFriendlyDevices(devices)
	if len(filtered) != 1 {
		t.Fatalf("einziger LPD-Drucker wurde ausgeblendet: %#v", filtered)
	}
}

func TestNoPasswordCreatesAnonymousCSRFSesssion(t *testing.T) {
	app := &application{
		authEnabled: false,
		sessions:    make(map[string]session),
		attempts:    make(map[string]loginAttempt),
	}
	handler := app.requireSession(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	first := httptest.NewRecorder()
	handler(first, httptest.NewRequest(http.MethodGet, "/", nil))
	if first.Code != http.StatusSeeOther {
		t.Fatalf("erster Aufruf = %d, erwartet Redirect", first.Code)
	}
	cookies := first.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != sessionCookie {
		t.Fatalf("keine anonyme Sitzung gesetzt: %#v", cookies)
	}

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.AddCookie(cookies[0])
	second := httptest.NewRecorder()
	handler(second, request)
	if second.Code != http.StatusNoContent {
		t.Fatalf("Aufruf mit anonymer Sitzung = %d", second.Code)
	}
}

func TestPasswordModeRedirectsToLogin(t *testing.T) {
	app := &application{
		authEnabled: true,
		sessions:    make(map[string]session),
		attempts:    make(map[string]loginAttempt),
	}
	handler := app.requireSession(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	response := httptest.NewRecorder()
	handler(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if response.Code != http.StatusSeeOther || response.Header().Get("Location") != "/login" {
		t.Fatalf("unerwartete Antwort: %d %q", response.Code, response.Header().Get("Location"))
	}
}

func TestRequestLang(t *testing.T) {
	tests := map[string]string{
		"de-DE,de;q=0.9,en;q=0.8": "de",
		"en-US,en;q=0.9":          "en",
		"fr-FR,fr;q=0.9,de;q=0.8": "de",
		"fr-FR":                   "en",
		"":                        "en",
	}
	for header, expected := range tests {
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.Header.Set("Accept-Language", header)
		if actual := requestLang(request); actual != expected {
			t.Errorf("requestLang(%q) = %q, want %q", header, actual, expected)
		}
	}

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Accept-Language", "de-DE")
	request.AddCookie(&http.Cookie{Name: langCookie, Value: "en"})
	if actual := requestLang(request); actual != "en" {
		t.Errorf("language cookie ignored, got %q", actual)
	}
}

func TestMessagesHaveSameKeys(t *testing.T) {
	for lang, table := range messages {
		for other, otherTable := range messages {
			for key := range table {
				if _, ok := otherTable[key]; !ok {
					t.Errorf("key %q exists in %q but not in %q", key, lang, other)
				}
			}
		}
	}
}

func TestTemplatesRenderInAllLanguages(t *testing.T) {
	tmpl, err := template.New("pages").Funcs(template.FuncMap{
		"pathEscape": url.PathEscape,
		"t":          tr,
	}).ParseFS(webFiles, "web/*.html")
	if err != nil {
		t.Fatal(err)
	}
	printers := []printer{
		{Name: "A", DisplayName: "A", State: "state.ready", StateClass: "ready"},
		{Name: "B", DisplayName: "B", State: "state.printing", StateClass: "working", Jobs: 1},
		{Name: "C", DisplayName: "C", State: "state.stopped", StateClass: "warning", Jobs: 3},
	}
	for lang := range messages {
		for _, page := range []string{"index.html", "login.html"} {
			var out strings.Builder
			data := pageData{Lang: lang, AuthEnabled: true, Printers: printers, Flash: "x", FlashError: true}
			if err := tmpl.ExecuteTemplate(&out, page, data); err != nil {
				t.Fatalf("%s/%s: %v", lang, page, err)
			}
			for key := range messages[langEN] {
				if strings.Contains(out.String(), ">"+key+"<") || strings.Contains(out.String(), "\""+key+"\"") {
					t.Errorf("%s/%s shows untranslated key %q", lang, page, key)
				}
			}
			if !strings.Contains(out.String(), `lang="`+lang+`"`) {
				t.Errorf("%s/%s: html lang not set", lang, page)
			}
		}
	}
}
