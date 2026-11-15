package main

import (
	"crypto/rand"
	"crypto/subtle"
	"embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	sessionCookie = "brother_cups_session"
	flashCookie   = "brother_cups_flash"
	sessionTTL    = 12 * time.Hour
)

//go:embed web/*
var webFiles embed.FS

type session struct {
	CSRF      string
	ExpiresAt time.Time
}

type loginAttempt struct {
	Failures  int
	BlockedTo time.Time
}

type application struct {
	adminUser     string
	adminPassword string
	authEnabled   bool
	templates     *template.Template

	mu       sync.Mutex
	sessions map[string]session
	attempts map[string]loginAttempt

	modelMu      sync.Mutex
	cachedModels []model
}

type pageData struct {
	AdminUser   string
	AuthEnabled bool
	CSRF        string
	Flash       string
	Error       string
	Printers    []printer
	Models      []model
	CUPSBase    string
}

func main() {
	adminUser := envOr("ADMIN_USER", "admin")
	adminPassword := os.Getenv("CUPS_ADMIN_PASSWORD")

	funcs := template.FuncMap{
		"pathEscape": url.PathEscape,
		"hasPrefix":  strings.HasPrefix,
	}
	tmpl, err := template.New("pages").Funcs(funcs).ParseFS(webFiles, "web/*.html")
	if err != nil {
		log.Fatalf("Templates laden: %v", err)
	}

	app := &application{
		adminUser:     adminUser,
		adminPassword: adminPassword,
		authEnabled:   adminPassword != "",
		templates:     tmpl,
		sessions:      make(map[string]session),
		attempts:      make(map[string]loginAttempt),
	}

	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(webFiles)))
	mux.HandleFunc("GET /health", app.health)
	mux.HandleFunc("GET /login", app.loginPage)
	mux.HandleFunc("POST /login", app.login)
	mux.HandleFunc("POST /logout", app.requireSession(app.logout))
	mux.HandleFunc("GET /", app.requireSession(app.home))
	mux.HandleFunc("GET /api/devices", app.requireSession(app.devicesAPI))
	mux.HandleFunc("POST /printer/configure", app.requireSession(app.configurePrinter))
	mux.HandleFunc("POST /printer/test", app.requireSession(app.testPrinter))

	port := envOr("BROTHER_CUPS_UI_PORT", "8631")
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           app.securityHeaders(mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	if app.authEnabled {
		log.Printf("Brother-CUPS-Einrichtung läuft mit Login auf Port %s", port)
	} else {
		log.Printf("Brother-CUPS-Einrichtung läuft ohne Login auf Port %s", port)
	}
	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func (app *application) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self'; script-src 'self'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

func (app *application) health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("ok\n"))
}

func (app *application) loginPage(w http.ResponseWriter, r *http.Request) {
	if !app.authEnabled {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if _, ok := app.currentSession(r); ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	app.render(w, "login.html", pageData{AdminUser: app.adminUser})
}

func (app *application) login(w http.ResponseWriter, r *http.Request) {
	if !app.authEnabled {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ungültige Anfrage", http.StatusBadRequest)
		return
	}

	clientIP := remoteIP(r)
	if wait := app.loginBlocked(clientIP); wait > 0 {
		w.Header().Set("Retry-After", fmt.Sprintf("%d", int(wait.Seconds())+1))
		w.WriteHeader(http.StatusTooManyRequests)
		app.render(w, "login.html", pageData{
			AdminUser: app.adminUser,
			Error:     "Zu viele Anmeldeversuche. Bitte kurz warten.",
		})
		return
	}

	userOK := subtle.ConstantTimeCompare([]byte(r.FormValue("username")), []byte(app.adminUser)) == 1
	passwordOK := subtle.ConstantTimeCompare([]byte(r.FormValue("password")), []byte(app.adminPassword)) == 1
	if !userOK || !passwordOK {
		app.recordLoginFailure(clientIP)
		w.WriteHeader(http.StatusUnauthorized)
		app.render(w, "login.html", pageData{
			AdminUser: app.adminUser,
			Error:     "Benutzername oder Passwort stimmen nicht.",
		})
		return
	}

	app.clearLoginFailures(clientIP)
	if err := app.createSession(w); err != nil {
		http.Error(w, "Sitzung konnte nicht erstellt werden", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) logout(w http.ResponseWriter, r *http.Request) {
	if !app.validCSRF(r) {
		http.Error(w, "Ungültige Sitzung", http.StatusForbidden)
		return
	}
	if cookie, err := r.Cookie(sessionCookie); err == nil {
		app.mu.Lock()
		delete(app.sessions, cookie.Value)
		app.mu.Unlock()
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteStrictMode})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	sess, _ := app.currentSession(r)
	printers, err := listPrinters()
	if err != nil {
		app.render(w, "index.html", pageData{
			AdminUser:   app.adminUser,
			AuthEnabled: app.authEnabled,
			CSRF:        sess.CSRF,
			Error:       "CUPS ist noch nicht bereit. Bitte lade die Seite gleich noch einmal.",
			CUPSBase:    cupsBaseURL(r),
		})
		return
	}
	models, err := app.models()
	if err != nil {
		log.Printf("Modelle lesen: %v", err)
	}

	app.render(w, "index.html", pageData{
		AdminUser:   app.adminUser,
		AuthEnabled: app.authEnabled,
		CSRF:        sess.CSRF,
		Flash:       readFlash(w, r),
		Printers:    printers,
		Models:      models,
		CUPSBase:    cupsBaseURL(r),
	})
}

func (app *application) devicesAPI(w http.ResponseWriter, _ *http.Request) {
	devices, err := discoverDevices()
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"error": "Die Druckersuche ist gerade nicht verfügbar.",
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"devices": devices})
}

func (app *application) configurePrinter(w http.ResponseWriter, r *http.Request) {
	if !app.validCSRF(r) {
		http.Error(w, "Ungültige Sitzung", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ungültige Anfrage", http.StatusBadRequest)
		return
	}

	description := strings.TrimSpace(r.FormValue("description"))
	if description == "" {
		description = "Brother Drucker"
	}
	queueName := normalizeQueueName(r.FormValue("queue_name"))
	if queueName == "" {
		queueName = normalizeQueueName(description)
	}
	if queueName == "" {
		failRedirect(w, r, "Bitte gib einen Namen für den Drucker ein.")
		return
	}

	deviceURI, err := requestedDeviceURI(r)
	if err != nil {
		failRedirect(w, r, err.Error())
		return
	}
	modelURI := r.FormValue("model_uri")
	if !app.knownModel(modelURI) {
		failRedirect(w, r, "Bitte wähle ein unterstütztes Brother-Modell aus.")
		return
	}

	paper := r.FormValue("paper_size")
	if paper != "A4" && paper != "Letter" {
		paper = "A4"
	}
	duplex := "None"
	if r.FormValue("duplex") == "yes" {
		duplex = "DuplexNoTumble"
	}

	args := []string{
		"-p", queueName,
		"-E",
		"-v", deviceURI,
		"-m", modelURI,
		"-D", description,
		"-o", "printer-is-shared=true",
		"-o", "PageSize=" + paper,
		"-o", "Duplex=" + duplex,
	}
	if _, err := runCommand("lpadmin", args...); err != nil {
		log.Printf("Drucker einrichten: %v", err)
		failRedirect(w, r, "Der Drucker konnte nicht eingerichtet werden. Öffne die Expertenansicht für Details.")
		return
	}
	_, _ = runCommand("lpadmin", "-d", queueName)
	setFlash(w, fmt.Sprintf("%s ist eingerichtet und per AirPrint freigegeben.", description))
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) testPrinter(w http.ResponseWriter, r *http.Request) {
	if !app.validCSRF(r) {
		http.Error(w, "Ungültige Sitzung", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ungültige Anfrage", http.StatusBadRequest)
		return
	}
	name := r.FormValue("printer")
	if !printerExists(name) {
		failRedirect(w, r, "Der ausgewählte Drucker existiert nicht mehr.")
		return
	}
	output, err := runCommand("lp", "-d", name, "/usr/share/cups/data/testprint")
	if err != nil {
		log.Printf("Testseite: %v", err)
		failRedirect(w, r, "Die Testseite konnte nicht gesendet werden.")
		return
	}
	message := "Testseite wurde an den Drucker gesendet."
	if strings.TrimSpace(output) != "" {
		message = strings.TrimSpace(output)
	}
	setFlash(w, message)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) requireSession(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := app.currentSession(r); !ok {
			if !app.authEnabled && (r.Method == http.MethodGet || r.Method == http.MethodHead) {
				if err := app.createSession(w); err != nil {
					http.Error(w, "Sitzung konnte nicht erstellt werden", http.StatusInternalServerError)
					return
				}
				http.Redirect(w, r, r.URL.RequestURI(), http.StatusSeeOther)
				return
			}
			if !app.authEnabled {
				http.Error(w, "Ungültige Sitzung", http.StatusForbidden)
				return
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

func (app *application) createSession(w http.ResponseWriter) error {
	token, err := randomToken(32)
	if err != nil {
		return err
	}
	csrf, err := randomToken(24)
	if err != nil {
		return err
	}

	app.mu.Lock()
	app.sessions[token] = session{CSRF: csrf, ExpiresAt: time.Now().Add(sessionTTL)}
	app.pruneSessionsLocked()
	app.mu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(sessionTTL.Seconds()),
	})
	return nil
}

func (app *application) currentSession(r *http.Request) (session, bool) {
	cookie, err := r.Cookie(sessionCookie)
	if err != nil {
		return session{}, false
	}
	app.mu.Lock()
	defer app.mu.Unlock()
	sess, ok := app.sessions[cookie.Value]
	if !ok || time.Now().After(sess.ExpiresAt) {
		delete(app.sessions, cookie.Value)
		return session{}, false
	}
	return sess, true
}

func (app *application) validCSRF(r *http.Request) bool {
	sess, ok := app.currentSession(r)
	if !ok {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(r.FormValue("csrf")), []byte(sess.CSRF)) == 1
}

func (app *application) models() ([]model, error) {
	app.modelMu.Lock()
	defer app.modelMu.Unlock()
	if len(app.cachedModels) > 0 {
		return append([]model(nil), app.cachedModels...), nil
	}
	models, err := listModels()
	if err != nil {
		return nil, err
	}
	app.cachedModels = append([]model(nil), models...)
	return models, nil
}

func (app *application) knownModel(uri string) bool {
	models, err := app.models()
	if err != nil {
		return false
	}
	for _, item := range models {
		if item.URI == uri {
			return true
		}
	}
	return false
}

func (app *application) loginBlocked(ip string) time.Duration {
	app.mu.Lock()
	defer app.mu.Unlock()
	attempt := app.attempts[ip]
	if time.Now().Before(attempt.BlockedTo) {
		return time.Until(attempt.BlockedTo)
	}
	return 0
}

func (app *application) recordLoginFailure(ip string) {
	app.mu.Lock()
	defer app.mu.Unlock()
	attempt := app.attempts[ip]
	attempt.Failures++
	if attempt.Failures >= 5 {
		attempt.BlockedTo = time.Now().Add(30 * time.Second)
		attempt.Failures = 0
	}
	app.attempts[ip] = attempt
}

func (app *application) clearLoginFailures(ip string) {
	app.mu.Lock()
	delete(app.attempts, ip)
	app.mu.Unlock()
}

func (app *application) pruneSessionsLocked() {
	now := time.Now()
	for token, sess := range app.sessions {
		if now.After(sess.ExpiresAt) {
			delete(app.sessions, token)
		}
	}
}

func (app *application) render(w http.ResponseWriter, name string, data pageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := app.templates.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("Template %s: %v", name, err)
	}
}

func requestedDeviceURI(r *http.Request) (string, error) {
	if r.FormValue("connection") == "manual" {
		ip := strings.TrimSpace(r.FormValue("printer_ip"))
		if net.ParseIP(ip) == nil {
			return "", errors.New("Bitte gib eine gültige IP-Adresse des Druckers ein.")
		}
		if strings.Contains(ip, ":") {
			ip = "[" + ip + "]"
		}
		return "socket://" + ip + ":9100", nil
	}

	rawURI := strings.TrimSpace(r.FormValue("device_uri"))
	if len(rawURI) > 2048 || strings.ContainsAny(rawURI, "\x00\r\n") {
		return "", errors.New("Bitte wähle einen gefundenen Drucker aus.")
	}
	scheme, target, found := strings.Cut(rawURI, "://")
	if !found || strings.Trim(target, "/") == "" {
		return "", errors.New("Bitte wähle einen gefundenen Drucker aus.")
	}
	allowed := map[string]bool{"dnssd": true, "socket": true, "lpd": true, "ipp": true, "ipps": true}
	if !allowed[strings.ToLower(scheme)] {
		return "", errors.New("Diese Druckerverbindung wird nicht unterstützt.")
	}
	return rawURI, nil
}

func normalizeQueueName(value string) string {
	value = strings.TrimSpace(value)
	var b strings.Builder
	separator := false
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
			separator = r == '_' || r == '-'
		case r == 'ä':
			b.WriteString("ae")
			separator = false
		case r == 'ö':
			b.WriteString("oe")
			separator = false
		case r == 'ü':
			b.WriteString("ue")
			separator = false
		case r == 'Ä':
			b.WriteString("Ae")
			separator = false
		case r == 'Ö':
			b.WriteString("Oe")
			separator = false
		case r == 'Ü':
			b.WriteString("Ue")
			separator = false
		case r == 'ß':
			b.WriteString("ss")
			separator = false
		case r == ' ', r == '.', r == '/':
			if b.Len() > 0 && !separator {
				b.WriteByte('_')
				separator = true
			}
		}
		if b.Len() >= 64 {
			break
		}
	}
	return strings.Trim(b.String(), "-_")
}

func cupsBaseURL(r *http.Request) string {
	host := r.Host
	if parsedHost, _, err := net.SplitHostPort(host); err == nil {
		host = parsedHost
	} else if strings.Contains(host, ":") && !strings.HasPrefix(host, "[") {
		host = strings.Trim(host, "[]")
	}
	if strings.Contains(host, ":") {
		host = "[" + host + "]"
	}
	return "http://" + host + ":631"
}

func printerExists(name string) bool {
	printers, err := listPrinters()
	if err != nil {
		return false
	}
	for _, item := range printers {
		if item.Name == name {
			return true
		}
	}
	return false
}

func randomToken(bytes int) (string, error) {
	buffer := make([]byte, bytes)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func remoteIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func setFlash(w http.ResponseWriter, message string) {
	http.SetCookie(w, &http.Cookie{
		Name:     flashCookie,
		Value:    base64.RawURLEncoding.EncodeToString([]byte(message)),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   60,
	})
}

func readFlash(w http.ResponseWriter, r *http.Request) string {
	cookie, err := r.Cookie(flashCookie)
	if err != nil {
		return ""
	}
	http.SetCookie(w, &http.Cookie{Name: flashCookie, Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteStrictMode})
	decoded, err := base64.RawURLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return ""
	}
	return string(decoded)
}

func failRedirect(w http.ResponseWriter, r *http.Request, message string) {
	setFlash(w, "Fehler: "+message)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func sortModels(models []model) {
	sort.Slice(models, func(i, j int) bool {
		return strings.ToLower(models[i].Name) < strings.ToLower(models[j].Name)
	})
}
