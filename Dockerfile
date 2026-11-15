# Kleine, statisch gelinkte Einrichtungsoberfläche bauen.
FROM golang:1.24-bookworm AS ui-builder

WORKDIR /src
COPY setup-ui/go.mod ./
COPY setup-ui/*.go ./
COPY setup-ui/web ./web
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /brother-cups-ui .

# CUPS-Druckserver mit freiem brlaser-Treiber für ältere Brother-Laserdrucker.
# Basis ist ausschließlich das offizielle Debian-Image, alle Pakete kommen aus Debian.
FROM debian:stable-slim

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      cups \
      cups-filters \
      printer-driver-brlaser \
      libnss-mdns \
      ghostscript \
      poppler-utils \
 && mkdir -p /usr/share/doc/brother-cups \
 && dpkg-query -W -f='${binary:Package}\t${Version}\n' \
      | sort > /usr/share/doc/brother-cups/installed-packages.txt \
 && rm -rf /var/lib/apt/lists/*

# Drucken und Weboberfläche aus dem LAN erlauben, Drucker per Bonjour/AirPrint
# ankündigen. Änderungen in der Weboberfläche brauchen den Admin-Login.
# "Drucker freigeben" ist beim Hinzufügen standardmäßig angehakt.
RUN sed -i \
      -e 's/^Listen localhost:631/Port 631/' \
      -e 's/^Browsing .*/Browsing Yes/' \
      -e '/^<Location \/>/a\  Allow all' \
      -e '/^<Location \/admin/a\  Allow all' \
      /etc/cups/cupsd.conf \
 && echo 'ServerAlias *' >> /etc/cups/cupsd.conf \
 && echo 'DefaultEncryption IfRequested' >> /etc/cups/cupsd.conf \
 && find /usr/share/cups/templates -name add-printer.tmpl \
      -exec sed -i 's/{PRINTER_IS_SHARED=1?CHECKED:}/CHECKED/' {} + \
 && cp -a /etc/cups /etc/cups.default

COPY entrypoint.sh /entrypoint.sh
COPY --from=ui-builder /brother-cups-ui /usr/local/bin/brother-cups-ui
COPY LICENSE /usr/share/doc/brother-cups/LICENSE
COPY THIRD_PARTY_LICENSES.md /usr/share/doc/brother-cups/THIRD_PARTY_LICENSES.md
RUN chmod +x /entrypoint.sh

EXPOSE 631 8631
VOLUME ["/etc/cups"]
ENTRYPOINT ["/entrypoint.sh"]
