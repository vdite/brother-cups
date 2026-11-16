# Brother CUPS: AirPrint für ältere Brother-Laserdrucker

[English](README.md) · **Deutsch**

Ein kleiner CUPS-Druckserver im Docker-Container, mit dem freien Treiber [brlaser](https://github.com/Owl-Maintain/brlaser). Er macht ältere Brother-Laserdrucker, die selbst kein AirPrint können, im Heimnetz als **AirPrint-Drucker** verfügbar. Mac, iPhone und iPad drucken dann ohne Herstellertreiber und ohne Rosetta.

## Warum?

Viele Brother-Treiber für macOS sind reiner Intel-Code und laufen auf Macs mit Apple-Chip nur über Rosetta 2. macOS 27 entfernt Rosetta beim Update, mit macOS 28 fällt es für normale Programme ganz weg. Ein Drucker wie der MFC-7460DN (Baujahr um 2011) funktioniert dann am Mac nicht mehr, obwohl er technisch einwandfrei ist.

Dieser Container übernimmt das Drucken im Netz. Dem Mac ist es dann egal, welcher Treiber dahintersteckt.

Hintergrund: [macOS 27 und alte Drucker](https://mizine.de/apple/macos-27-drucker-rosetta/) auf mizine.de.

## Was ist drin?

- Basis: offizielles Image `debian:stable-slim`
- Pakete aus Debian: `cups`, `cups-filters`, `printer-driver-brlaser`, `libnss-mdns`, `ghostscript`, `poppler-utils`
- Eine einfache Einrichtungsoberfläche auf Deutsch und Englisch, die Drucker im Netz findet, passende brlaser-Modelle anbietet und Testseiten druckt
- Ein kurzes Startskript (`entrypoint.sh`), das den Admin-Zugang anlegt und auf Wunsch auch den Drucker automatisch einrichtet

Sonst nichts. Das Image wird per GitHub Actions direkt aus diesem Repo gebaut.

Fertiges Image für AMD64 und ARM64: `ghcr.io/vdite/brother-cups:latest`

## So sieht es aus

Die Oberfläche richtet sich nach der Sprache des Browsers (Deutsch oder Englisch). Umschalten lässt sie sich unten auf der Seite.

![Brother-CUPS-Dashboard mit eingerichtetem Brother MFC-7460DN](docs/screenshots/dashboard.jpg)

Die übersichtliche Einrichtung sucht selbstständig nach Brother-Druckern im Heimnetz und bietet die passenden freien brlaser-Treiber an:

![Einrichtungsassistent mit Druckersuche und Modellauswahl](docs/screenshots/setup.jpg)

Danach erscheint der Drucker ganz normal in macOS und nimmt Druckaufträge per AirPrint an:

| Drucker in macOS | Erfolgreicher Druckauftrag |
|---|---|
| ![Brother MFC-7460DN in den macOS-Druckereinstellungen](docs/screenshots/macos-printer.jpg) | ![Druckauftrag an Brother CUPS in der macOS-Druckerwarteschlange](docs/screenshots/macos-print-job.jpg) |

## Unterstützte Drucker

Alle Modelle, die brlaser 6.2.7 (die Version in Debian 13 „trixie“) unterstützt:

| Serie | Modelle |
|---|---|
| **DCP** | DCP-1510, DCP-1600, DCP-1610W, DCP-7010, DCP-7020, DCP-7030, DCP-7040, DCP-7055, DCP-7055W, DCP-7060D, DCP-7065DN, DCP-7070DW, DCP-7080, DCP-7080D, DCP-8065DN, DCP-B7500D, DCP-L2500D, DCP-L2510D, DCP-L2520D, DCP-L2520DW, DCP-L2537DW, DCP-L2540DW, DCP-L2550DW, DCP-L2560DW |
| **HL** | HL-1110, HL-1200, HL-2030, HL-2130, HL-2140, HL-2220, HL-2230, HL-2240, HL-2240D, HL-2250DN, HL-2260, HL-2260D, HL-2270DW, HL-2280DW, HL-5030, HL-5040, HL-5140, HL-5370DW, HL-5450DN, HL-L2300D, HL-L2305, HL-L2310D, HL-L2320D, HL-L2335D, HL-L2340D, HL-L2350DW, HL-L2360D, HL-L2370DN, HL-L2375DW, HL-L2380DW, HL-L2390DW, HL-L2400DW, HL-L2402D, HL-L2405W, HL-L5000D |
| **MFC** | MFC-1810, MFC-1910W, MFC-7240, MFC-7320, MFC-7340, MFC-7360N, MFC-7365DN, MFC-7420, MFC-7440N, MFC-7460DN, MFC-7860DW, MFC-8440, MFC-8710DW, MFC-8860DN, MFC-9160, MFC-L2690DW, MFC-L2700DN, MFC-L2700DW, MFC-L2710DN, MFC-L2710DW, MFC-L2750DW |
| **FAX** | FAX-2820, FAX-2840 |
| **Andere Marken** | Fuji Xerox DocuPrint P265 dw, Lenovo LJ2650DN |

Viele Einträge stehen für eine ganze Serie (etwa „HL-L2350DW series“). Selbst getestet: MFC-7460DN. Wenn Dein Modell funktioniert, freue ich mich über ein Issue, dann kann ich es als getestet markieren.

Brother-Drucker, die schon selbst AirPrint können, wie der MFC-9332CDW, brauchen diesen Container nicht. Die fügst Du am Mac direkt mit „AirPrint“ hinzu.

## Voraussetzungen

- Ein Linux-Rechner mit Docker, der dauerhaft läuft (z. B. ZimaOS, Raspberry Pi, Mini-PC)
- Avahi auf dem Host (bei ZimaOS schon vorhanden), damit der Drucker per Bonjour auftaucht
- Der Drucker hängt per LAN im Netz und hat eine feste IP (Adressreservierung im Router)
- Die Ports 631 (AirPrint/CUPS) und 8631 (einfache Einrichtung) sind auf dem Host frei

## Installation unter ZimaOS

1. Im App Store auf **Benutzerdefiniert installieren** und dann **Importieren** gehen.
2. Den Inhalt von [`zimaos-compose.yml`](zimaos-compose.yml) einfügen.
3. Optional ein `ADMIN_PASSWORD` eintragen. Bleibt es leer, öffnet sich die Einrichtungsoberfläche ohne Login.
4. Installieren und anschließend die App öffnen. ZimaOS führt direkt zur einfachen Einrichtungsseite.

## Installation in Portainer

1. **Stacks** > **Add stack** > **Web editor** öffnen.
2. Den Inhalt von [`portainer-compose.yml`](portainer-compose.yml) einfügen.
3. Unter **Environment variables** optional `ADMIN_PASSWORD` mit dem gewünschten Passwort anlegen. Ohne diese Variable läuft die einfache Oberfläche ohne Login.
4. **Deploy the stack** anklicken und danach `http://<IP-des-Servers>:8631` öffnen.

Portainer lädt das fertige Image aus der GitHub Container Registry. Ein lokaler Build ist dafür nicht erforderlich.

## Installation mit Docker Compose

```
git clone https://github.com/vdite/brother-cups.git
cd brother-cups
docker compose up -d --build
```

Diese Variante baut das Image bewusst lokal. Wer nur das fertige GitHub-Image starten möchte, verwendet stattdessen `portainer-compose.yml` oder `zimaos-compose.yml`.

Danach `http://<IP-des-Servers>:8631` öffnen.

## Drucker einrichten

1. Die Einrichtungsseite unter `http://<IP-des-Servers>:8631` öffnen. Wenn `ADMIN_PASSWORD` gesetzt ist, mit `admin` und diesem Passwort anmelden.
2. Kurz warten, bis die Druckersuche abgeschlossen ist, dann den Brother-Drucker auswählen.
3. Das passende Modell wählen. Bei eindeutigem Modellnamen wird es bereits vorausgewählt.
4. Papierformat und Duplex prüfen und **Drucker einrichten** anklicken.
5. Über **Testseite** direkt kontrollieren, ob alles funktioniert.

Wird der Drucker nicht automatisch gefunden, lässt sich seine feste IP-Adresse eintragen. Die Oberfläche verwendet dann `socket://<IP>:9100`.

Die bisherige CUPS-Oberfläche bleibt für technische Details und seltene Spezialfälle unter `http://<IP-des-Servers>:631` erreichbar. Normale Nutzer brauchen sie für die Einrichtung nicht mehr.

<details>
<summary>Drucker direkt in der CUPS-Expertenansicht einrichten</summary>

1. `http://<IP-des-Servers>:631` im Browser öffnen.
2. **Verwaltung** > **Drucker hinzufügen**, mit `admin` und dem gesetzten `ADMIN_PASSWORD` anmelden.
3. Den Drucker aus der Liste der gefundenen Netzwerkdrucker wählen, oder **AppSocket/HP JetDirect** mit `socket://<IP-des-Druckers>:9100`.
4. „Drucker im Netzwerk freigeben“ ist schon angehakt, so lassen.
5. Hersteller **Brother**, dann Dein Modell mit **„using brlaser“** wählen.
6. Unter **Standardeinstellungen** das Papierformat auf A4 setzen und eine Testseite drucken.

Mehrere Drucker sind kein Problem, einfach nacheinander hinzufügen.

Ohne `ADMIN_PASSWORD` ist die einfache Oberfläche vollständig nutzbar; administrative Änderungen direkt in CUPS erfordern dagegen ein gesetztes Passwort.

</details>

## Oder: Drucker automatisch anlegen

Wer lieber alles in der Compose-Datei hat, setzt `PRINTER_URI` und `PRINTER_MODEL`. Dann legt der Container den Drucker bei jedem Start selbst an.

| Variable | Beispiel | Bedeutung |
|---|---|---|
| `ADMIN_USER` | `admin` | Benutzername für die Weboberfläche |
| `ADMIN_PASSWORD` | | Optionales Passwort, leer = Einrichtungsoberfläche ohne Login |
| `BROTHER_CUPS_UI_PORT` | `8631` | Port der einfachen Einrichtungsoberfläche |
| `PRINTER_URI` | `socket://192.168.1.50:9100` | Adresse des Druckers (Port 9100 = RAW/JetDirect) |
| `PRINTER_MODEL` | `MFC-7460DN` | Modell laut brlaser-Liste, alternativ der volle `drv:///…`-Pfad |
| `PRINTER_NAME` | `Brother` | Name der Druckwarteschlange (ohne Leerzeichen) |
| `PAPER_SIZE` | `A4` | Standard-Papierformat |
| `PRINTER_OPTIONS` | `-o Duplex=DuplexNoTumble` | Weitere Optionen für `lpadmin` (optional) |

Welche Modelle brlaser unterstützt:

```
docker exec brother-cups lpinfo -m | grep brlaser
```

Findet das Startskript Dein Modell nicht, schreibt es die Liste ins Container-Log.

## Am Mac, iPhone und iPad hinzufügen

**Mac:** Systemeinstellungen > Drucker & Scanner > Drucker hinzufügen. Der Drucker taucht mit dem Namen des Servers auf. Bei „Verwenden“ **AirPrint** wählen.

**iPhone und iPad:** Teilen > Drucken. Der Drucker erscheint automatisch.

## Was nicht geht

- **Scannen und Fax:** Der Container kümmert sich nur ums Drucken.
- **Drucker, die brlaser nicht kennt:** siehe [Unterstützte Drucker](#unterstützte-drucker).
- CUPS meldet beim Einrichten, dass Druckertreiber „deprecated“ sind. Mit CUPS 2.x funktionieren sie weiterhin.

## Sicherheit

- Drucken und Ansehen der CUPS-Expertenansicht ist für alle Geräte im Netz erlaubt.
- Wenn `ADMIN_PASSWORD` gesetzt ist, schützt es die einfache Einrichtungsoberfläche und administrative Änderungen in CUPS.
- Ohne `ADMIN_PASSWORD` ist die einfache Oberfläche für alle Geräte im Heimnetz erreichbar. CSRF-Schutz und Sitzungen bleiben aktiv, ersetzen aber keine Anmeldung. In Netzen mit nicht vertrauenswürdigen Teilnehmern sollte deshalb ein Passwort gesetzt werden.
- Beide Oberflächen laufen im Heimnetz über HTTP. Ein gesetztes Passwort geht daher unverschlüsselt durchs LAN. Für ein privates Netz ist das üblich, für Firmennetze nicht.
- Den Container nicht ins Internet freigeben.

## Lizenz

Der selbst entwickelte Projektcode und die Repo-Dateien stehen unter der [MIT-Lizenz](LICENSE).

Das gebaute Image als Ganzes ist nicht ausschließlich MIT-lizenziert. Es enthält unveränderte Debian-Pakete mit eigenen Lizenzen, unter anderem CUPS, cups-filters, brlaser, Ghostscript und Poppler. Die Details, Quellcode-Links und Fundstellen im Image stehen in [THIRD_PARTY_LICENSES.md](THIRD_PARTY_LICENSES.md). GitHub Actions veröffentlicht zusätzlich SBOM und Build-Provenance.

Brother ist eine Marke der Brother Industries, Ltd. Dieses Projekt steht in keiner Verbindung zu Brother.
