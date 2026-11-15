#!/bin/sh
set -e

# Leeres Volume? Dann Standardkonfiguration hineinkopieren.
if [ ! -f /etc/cups/cupsd.conf ]; then
  echo "Lege Standardkonfiguration in /etc/cups an"
  cp -a /etc/cups.default/. /etc/cups/
fi

# Admin-Benutzer anlegen. Ein leeres ADMIN_PASSWORD deaktiviert den Login der
# einfachen Oberfläche. Für Änderungen in der CUPS-Expertenansicht ist dagegen
# ein gesetztes Passwort erforderlich.
ADMIN_USER="${ADMIN_USER:-admin}"
PW="${ADMIN_PASSWORD:-}"
rm -f /etc/cups/.admin-password
id "$ADMIN_USER" >/dev/null 2>&1 || useradd -M -s /usr/sbin/nologin "$ADMIN_USER"
usermod -a -G lpadmin "$ADMIN_USER"
if [ -n "$PW" ]; then
  echo "$ADMIN_USER:$PW" | chpasswd
  echo "Die Einrichtungsoberfläche ist mit dem Benutzer $ADMIN_USER geschützt."
else
  passwd -l "$ADMIN_USER" >/dev/null 2>&1 || true
  echo "Kein ADMIN_PASSWORD gesetzt. Die Einrichtungsoberfläche läuft ohne Login."
fi

mkdir -p /run/cups
/usr/sbin/cupsd -f &
CUPSD_PID=$!
UI_PID=""
shutdown() {
  if [ -n "$UI_PID" ]; then
    kill -TERM "$UI_PID" 2>/dev/null || true
  fi
  kill -TERM "$CUPSD_PID" 2>/dev/null || true
  wait "$CUPSD_PID" 2>/dev/null || true
}
trap shutdown TERM INT

# Warten, bis CUPS bereit ist
i=0
until lpstat -r 2>/dev/null | grep -q "is running"; do
  i=$((i + 1))
  if [ "$i" -gt 30 ]; then echo "CUPS startet nicht" >&2; exit 1; fi
  sleep 1
done

# Drucker automatisch anlegen, wenn PRINTER_URI gesetzt ist
if [ -n "$PRINTER_URI" ]; then
  NAME="${PRINTER_NAME:-Brother}"
  MODEL="${PRINTER_MODEL:-}"
  if [ -z "$MODEL" ]; then
    echo "PRINTER_MODEL fehlt, z. B. MFC-7460DN" >&2
  else
    case "$MODEL" in
      drv://*|*.ppd) ;;
      *)
        FOUND="$(lpinfo -m | grep -i brlaser | grep -i -- "$MODEL" | head -n 1 | cut -d ' ' -f 1)"
        if [ -z "$FOUND" ]; then
          echo "Kein brlaser-Treiber für \"$MODEL\" gefunden. Verfügbare Modelle:" >&2
          lpinfo -m | grep -i brlaser >&2
        fi
        MODEL="$FOUND"
        ;;
    esac
    if [ -n "$MODEL" ]; then
      # shellcheck disable=SC2086
      if lpadmin -p "$NAME" -E -v "$PRINTER_URI" -m "$MODEL" \
        -o printer-is-shared=true -o PageSize="${PAPER_SIZE:-A4}" ${PRINTER_OPTIONS:-}; then
        echo "Drucker \"$NAME\" eingerichtet: $PRINTER_URI ($MODEL)"
      else
        echo "Drucker \"$NAME\" konnte nicht eingerichtet werden" >&2
      fi
    fi
  fi
fi

# Verständliche Einrichtungsoberfläche. Sie verwendet denselben Admin-Login wie CUPS.
CUPS_ADMIN_PASSWORD="$PW" /usr/local/bin/brother-cups-ui &
UI_PID=$!

wait "$CUPSD_PID"
