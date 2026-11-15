# Changelog

Dieses Dokument enthält die wesentlichen Änderungen gegenüber dem ursprünglichen Übergabestand in `Intro.md`.

## Unveröffentlicht – Stand 15. November 2026

### Drucken über DNS-SD repariert

- Die Ursache des fehlgeschlagenen Drucks über die automatisch gefundene CUPS-Verbindung wurde behoben. Der Container konnte den vom Drucker gemeldeten Hostnamen `BRN001BA9D75C01.local` zuvor nicht auflösen.
- Das Debian-Paket `libnss-mdns` wurde dem Image hinzugefügt.
- Debian konfiguriert dadurch die Namensauflösung im Container als `hosts: files mdns4_minimal [NOTFOUND=return] dns`.
- Der Avahi-Socket des Hosts wird zusätzlich zum D-Bus-Socket in den Container eingebunden:
  - `/var/run/dbus:/var/run/dbus`
  - `/var/run/avahi-daemon/socket:/var/run/avahi-daemon/socket`
- Die Socket-Mounts wurden sowohl in `docker-compose.yml` als auch in `zimaos-compose.yml` ergänzt.
- Die automatisch gefundene Geräteadresse kann dadurch weiterhin verwendet werden: `dnssd://Brother%20MFC-7460DN._pdl-datastream._tcp.local/`.
- Die direkte Verbindung über `socket://<IP>:9100` bleibt als robuste manuelle Alternative erhalten.

### Einfache Einrichtungsoberfläche hinzugefügt

- Eine neue, eigenständige Weboberfläche wurde unter `setup-ui/` entwickelt.
- Die Oberfläche läuft standardmäßig auf Port `8631`.
- CUPS bleibt auf Port `631` für IPP/AirPrint und als Expertenansicht erreichbar.
- ZimaOS öffnet über `port_map: "8631"` jetzt direkt die neue Oberfläche statt der technischen CUPS-Oberfläche.
- Die Oberfläche ist responsiv und funktioniert auf Desktop, Tablet und Smartphone.
- Alle Styles, Skripte und SVG-Symbole sind im Image enthalten; es werden keine externen Webdienste oder Assets geladen.

Die Oberfläche bietet:

- optionaler Login mit demselben Admin-Benutzer und Passwort wie CUPS
- Betrieb ohne Login, wenn `ADMIN_PASSWORD` leer bleibt
- Anzeige bereits eingerichteter Drucker und ihres aktuellen Status
- Anzeige offener Druckaufträge
- automatische Suche nach Netzwerkdruckern über CUPS und DNS-SD
- verständliche Druckernamen statt technischer Geräteadressen
- Ausblendung kryptischer LPD-Duplikate, wenn derselbe Drucker bereits sinnvoll über DNS-SD gefunden wurde
- manuelle Einrichtung über die feste IP-Adresse des Druckers
- Auswahl aller im Image enthaltenen brlaser-Modelle
- automatische Vorauswahl des Modells bei eindeutigem Gerätenamen
- frei wählbaren Anzeigenamen
- automatische Erzeugung eines gültigen CUPS-Warteschlangennamens
- Unterstützung deutscher Umlaute bei der Erzeugung des internen Namens
- Auswahl zwischen A4 und US Letter
- Aktivierung oder Deaktivierung des Duplexdrucks
- Freigabe des eingerichteten Druckers im Netzwerk und per AirPrint
- Festlegung des neu eingerichteten Druckers als CUPS-Standarddrucker
- direkten Versand einer CUPS-Testseite
- Links zur jeweiligen CUPS-Druckerdetailseite
- Link zur vollständigen CUPS-Expertenansicht
- klare Hinweise, falls die automatische Druckersuche nicht funktioniert

### Sicherheit der Einrichtungsoberfläche

- Die Einrichtungsoberfläche besitzt bei gesetztem `ADMIN_PASSWORD` einen eigenen Formular-Login und verwendet dieselben Zugangsdaten wie CUPS.
- Wenn `ADMIN_PASSWORD` leer bleibt, erzeugt die Oberfläche automatisch eine anonyme Sitzung und überspringt die Anmeldung. Der CSRF-Schutz bleibt dabei aktiv.
- Das Admin-Passwort wird nicht in die Weboberfläche oder statische Dateien geschrieben.
- Sitzungen werden mit kryptografisch zufälligen Tokens erzeugt und laufen nach zwölf Stunden ab.
- Sitzungscookies sind `HttpOnly` und `SameSite=Strict`.
- Alle zustandsändernden Formulare sind mit einem CSRF-Token geschützt.
- Fehlgeschlagene Anmeldungen werden bei aktiviertem Login begrenzt; nach fünf Fehlversuchen wird die Quelladresse vorübergehend gesperrt.
- Die Anwendung setzt unter anderem eine Content Security Policy, `X-Content-Type-Options`, `X-Frame-Options` und eine restriktive Referrer Policy.
- HTTP-Server-Zeitlimits verhindern dauerhaft hängende Verbindungen.
- Druckermodelle werden serverseitig gegen die tatsächlich vorhandene brlaser-Liste geprüft.
- Manuell eingegebene IP-Adressen werden validiert.
- Geräteadressen werden auf die Protokolle DNS-SD, Socket, LPD, IPP und IPPS beschränkt.
- Systembefehle werden ohne Shell-Auswertung und mit getrennten Argumenten gestartet.
- Die bestehende Einschränkung bleibt dokumentiert: Beide Oberflächen laufen im Heimnetz über HTTP und dürfen nicht ins Internet freigegeben werden.

### Container und Startprozess erweitert

- Der `Dockerfile` verwendet nun einen offiziellen Go-Build-Stage, um die Einrichtungsoberfläche als kleine statisch gelinkte Binärdatei zu bauen.
- Das finale Laufzeit-Image bleibt `debian:stable-slim`.
- Die Go-Anwendung verwendet ausschließlich die Go-Standardbibliothek und besitzt keine externen Laufzeitabhängigkeiten.
- Das Image stellt jetzt die Ports `631` und `8631` bereit.
- `entrypoint.sh` startet nach CUPS auch die neue Einrichtungsoberfläche.
- Die Oberfläche erhält beim Start das gesetzte Admin-Passwort oder ausdrücklich einen leeren Wert, ohne ein zweites Passwortsystem einzuführen.
- Die frühere automatische Passworterzeugung wurde entfernt: leer bedeutet jetzt bewusst „ohne Login“.
- Die Signalbehandlung beendet bei einem Container-Stopp sowohl die Oberfläche als auch CUPS sauber.
- Die bisherige optionale automatische Einrichtung über `PRINTER_URI` und `PRINTER_MODEL` bleibt erhalten.
- Die neue Variable `BROTHER_CUPS_UI_PORT` erlaubt bei Bedarf einen anderen Port für die Einrichtungsoberfläche.

### Compose- und ZimaOS-Konfiguration aktualisiert

- `docker-compose.yml` und `zimaos-compose.yml` enthalten jetzt `BROTHER_CUPS_UI_PORT=8631`.
- `portainer-compose.yml` wurde als direkt importierbarer Portainer-Stack hinzugefügt.
- Portainer und ZimaOS laden `ghcr.io/vdite/brother-cups:latest` und bauen das Image nicht lokal.
- `pull_policy: always` sorgt bei einer erneuten Bereitstellung für den Abruf des aktuellen Images.
- Portainer verwendet für `/etc/cups` das persistente Volume `brother-cups-config`.
- `ADMIN_PASSWORD` ist in beiden Importvarianten optional; ein leerer Wert bedeutet „kein Login“.
- Die ZimaOS-Metadaten führen Port `8631` als einfache Druckereinrichtung auf.
- Port `631` ist in ZimaOS nun ausdrücklich als CUPS-Expertenansicht und AirPrint-Port beschrieben.
- Der Hauptlink der ZimaOS-App wurde von Port `631` auf Port `8631` umgestellt.
- Der für die `.local`-Auflösung erforderliche Avahi-Socket wurde in beide Compose-Dateien aufgenommen.

### Dokumentation aktualisiert

- Die README beschreibt jetzt die neue Oberfläche als normalen Einrichtungsweg.
- Die bisherige manuelle Einrichtung in CUPS wurde in einen aufklappbaren Abschnitt für technische Sonderfälle verschoben.
- Die Ports `631` und `8631` sowie ihre unterschiedlichen Aufgaben sind dokumentiert.
- Die Einrichtung über automatische Suche und über eine manuell eingetragene IP-Adresse ist beschrieben.
- Die neue Umgebungsvariable `BROTHER_CUPS_UI_PORT` wurde in die Variablentabelle aufgenommen.
- Die Sicherheitshinweise unterscheiden nun zwischen einfacher Einrichtungsoberfläche und CUPS-Expertenansicht.
- Die Installationsanleitung enthält einen eigenen Portainer-Ablauf.
- Die Passwortdokumentation beschreibt den optionalen Login und den Betrieb ohne Passwort.
- Vier Screenshots dokumentieren das Dashboard, den Einrichtungsassistenten sowie den erkannten Drucker und einen echten Druckauftrag unter macOS.
- `Intro.md` wurde ausdrücklich in `.gitignore` aufgenommen, damit die interne Übergabedatei nicht versehentlich committed oder veröffentlicht wird.

### Lizenztransparenz und Lieferkette

- Die MIT-Lizenz gilt ausdrücklich für den selbst entwickelten Projektcode und die Repo-Dateien.
- `THIRD_PARTY_LICENSES.md` dokumentiert, dass das Container-Image zusätzliche Komponenten unter Apache-, GPL-, LGPL-, AGPL- und weiteren Lizenzen enthält.
- Die MIT-Lizenz und die Drittanbieterübersicht werden nach `/usr/share/doc/brother-cups/` in das Image kopiert.
- Das Image enthält unter `/usr/share/doc/brother-cups/installed-packages.txt` eine vollständige Liste der installierten Debian-Pakete und Versionen.
- Die vollständigen Debian-Copyright-Dateien bleiben unter `/usr/share/doc/<paket>/copyright` erhalten.
- Die GitHub-Actions-Pipeline erzeugt SBOM und Build-Provenance für die veröffentlichten Multi-Arch-Images.

### Tests und Verifikation

- Die neue Go-Anwendung wurde mit automatisierten Tests für folgende Bereiche geprüft:
  - Erzeugung gültiger CUPS-Warteschlangennamen
  - Behandlung deutscher Umlaute
  - validierte DNS-SD-Geräteadressen
  - manuelle Socket-Verbindungen über IP-Adresse
  - Ablehnung nicht erlaubter Geräteprotokolle
  - lesbare Namen für gefundene DNS-SD-Drucker
  - Auswertung von CUPS-Geräteadressen
  - korrekte Zuordnung von Druckaufträgen bei Warteschlangennamen mit Bindestrichen
  - Ausblendung doppelter LPD-Verbindungen mit Erhalt von LPD als Fallback
- `go test ./...` ist erfolgreich.
- `go vet ./...` ist erfolgreich.
- `entrypoint.sh` wurde mit ShellCheck geprüft.
- Beide Compose-Dateien wurden erfolgreich mit `docker compose config` validiert.
- Das vollständige Image wurde lokal für ARM64 gebaut und gestartet.
- Das vollständige Image wurde auf ZimaOS-2 nativ für AMD64 gebaut und gestartet.
- Login, Sitzungscookie, Seitenrendering, Styles und geschützte API wurden per HTTP geprüft.
- Ein vollständiger lokaler Einrichtungsablauf wurde mit einer reservierten Test-IP durchgespielt.
- Ein Testaufruf ohne CSRF-Token wurde erwartungsgemäß mit HTTP `403` abgewiesen.
- Die Live-Druckersuche auf ZimaOS findet den Brother MFC-7460DN und den Brother MFC-9332CDW mit verständlichen DNS-SD-Namen.
- `BRN001BA9D75C01.local` wird im AMD64-Container korrekt zu `192.168.178.53` aufgelöst.
- Die Warteschlange `Brother_MFC-7460DN` bleibt nach Container- und Imagewechseln erhalten und ist bereit.
- Der Mac erkennt weiterhin `Brother_MFC-7460DN @ ZimaOS-2` als AirPrint-Dienst.
- Eine echte CUPS-Testseite wurde über die zuvor fehlerhafte DNS-SD-Verbindung vollständig verarbeitet und kam am Brother MFC-7460DN heraus. Der physische Ausdruck wurde bestätigt.

### Veröffentlichung

- Die Veröffentlichung ist als öffentliches Repository `vdite/brother-cups` vorgesehen.
- GitHub Actions baut und veröffentlicht das Multi-Arch-Image für AMD64 und ARM64 unter `ghcr.io/vdite/brother-cups:latest`.
- Portainer und ZimaOS laden dieses fertige Image direkt; auf dem Zielsystem ist kein Image-Build erforderlich.
- Das persistente CUPS-Volume mit Druckerwarteschlange und Admin-Konfiguration bleibt bei Image-Aktualisierungen erhalten.
