#!/bin/sh
set -e

# Empty volume? Copy the default configuration into it.
if [ ! -f /etc/cups/cupsd.conf ]; then
  echo "Creating default configuration in /etc/cups"
  cp -a /etc/cups.default/. /etc/cups/
fi

# Create the admin user. An empty ADMIN_PASSWORD disables the login of the
# simple setup page. Changes in the CUPS web interface still require a password.
ADMIN_USER="${ADMIN_USER:-admin}"
PW="${ADMIN_PASSWORD:-}"
rm -f /etc/cups/.admin-password
id "$ADMIN_USER" >/dev/null 2>&1 || useradd -M -s /usr/sbin/nologin "$ADMIN_USER"
usermod -a -G lpadmin "$ADMIN_USER"
if [ -n "$PW" ]; then
  echo "$ADMIN_USER:$PW" | chpasswd
  echo "The setup page is protected with the user $ADMIN_USER."
else
  passwd -l "$ADMIN_USER" >/dev/null 2>&1 || true
  echo "No ADMIN_PASSWORD set. The setup page runs without a login."
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

# Wait until CUPS is ready
i=0
until lpstat -r 2>/dev/null | grep -q "is running"; do
  i=$((i + 1))
  if [ "$i" -gt 30 ]; then echo "CUPS does not start" >&2; exit 1; fi
  sleep 1
done

# Create the printer automatically when PRINTER_URI is set
if [ -n "$PRINTER_URI" ]; then
  NAME="${PRINTER_NAME:-Brother}"
  MODEL="${PRINTER_MODEL:-}"
  if [ -z "$MODEL" ]; then
    echo "PRINTER_MODEL is missing, e.g. MFC-7460DN" >&2
  else
    case "$MODEL" in
      drv://*|*.ppd) ;;
      *)
        FOUND="$(lpinfo -m | grep -i brlaser | grep -i -- "$MODEL" | head -n 1 | cut -d ' ' -f 1)"
        if [ -z "$FOUND" ]; then
          echo "No brlaser driver found for \"$MODEL\". Available models:" >&2
          lpinfo -m | grep -i brlaser >&2
        fi
        MODEL="$FOUND"
        ;;
    esac
    if [ -n "$MODEL" ]; then
      # shellcheck disable=SC2086
      if lpadmin -p "$NAME" -E -v "$PRINTER_URI" -m "$MODEL" \
        -o printer-is-shared=true -o PageSize="${PAPER_SIZE:-A4}" ${PRINTER_OPTIONS:-}; then
        echo "Printer \"$NAME\" set up: $PRINTER_URI ($MODEL)"
      else
        echo "Printer \"$NAME\" could not be set up" >&2
      fi
    fi
  fi
fi

# Simple setup page. It uses the same admin login as CUPS.
CUPS_ADMIN_PASSWORD="$PW" /usr/local/bin/brother-cups-ui &
UI_PID=$!

wait "$CUPSD_PID"
