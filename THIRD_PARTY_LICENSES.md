# Drittanbieter-Software und Lizenzen

Die selbst entwickelten Dateien dieses Repositories – insbesondere die Einrichtungsoberfläche, der `Dockerfile`, `entrypoint.sh` und die Compose-Dateien – stehen unter der MIT-Lizenz in `LICENSE`.

Das gebaute Container-Image enthält zusätzlich Debian und zahlreiche unveränderte Debian-Pakete. Diese Bestandteile stehen **nicht** unter der MIT-Lizenz, sondern jeweils unter ihren eigenen Lizenzen. Das Image darf daher nicht als ausschließlich MIT-lizenziert bezeichnet werden.

Zu den direkt installierten Hauptkomponenten gehören:

| Komponente | Wesentliche Lizenz laut Debian-Paket | Quellcode und Copyright |
|---|---|---|
| CUPS | Apache-2.0 mit GPL/LGPL-Ausnahmen sowie weitere Komponentenlizenzen | [Debian-Paket cups](https://packages.debian.org/stable/cups) |
| cups-filters | GPL-2+, GPL-2, LGPL-2 und weitere Komponentenlizenzen | [Debian-Quellpaket cups-filters](https://packages.debian.org/source/stable/cups-filters) |
| brlaser | GPL-2+ und BSL-1 | [Debian-Quellpaket brlaser](https://packages.debian.org/source/stable/brlaser) |
| Ghostscript/GhostPDL | überwiegend AGPL-3+ sowie weitere Komponentenlizenzen | [Debian-Quellpaket ghostscript](https://packages.debian.org/source/stable/ghostscript) |
| Poppler | GPL-2 oder GPL-3 sowie weitere Komponentenlizenzen | [Debian-Paket poppler-utils](https://packages.debian.org/stable/utils/poppler-utils) |
| nss-mdns | LGPL-2+ | [Debian-Paket libnss-mdns](https://packages.debian.org/stable/libnss-mdns) |

## Hinweise im Image

Debian installiert die vollständigen paketbezogenen Copyright- und Lizenzinformationen unter:

```text
/usr/share/doc/<paket>/copyright
```

Eine Liste aller tatsächlich im Image enthaltenen Binärpakete und Versionen liegt unter:

```text
/usr/share/doc/brother-cups/installed-packages.txt
```

Die MIT-Lizenz des eigenen Projektcodes und dieses Dokument liegen im Image unter:

```text
/usr/share/doc/brother-cups/LICENSE
/usr/share/doc/brother-cups/THIRD_PARTY_LICENSES.md
```

Die GitHub-Actions-Pipeline veröffentlicht zu jedem Image außerdem eine Software Bill of Materials (SBOM) und Build-Provenance. Die exakten Debian-Quellpakete lassen sich anhand der in der Paketliste bzw. SBOM dokumentierten Versionen über [Debian Sources](https://sources.debian.org/) und das [Debian-Paketarchiv](https://deb.debian.org/debian/) beziehen.

Dieses Dokument ist eine technische Lizenzübersicht und keine Rechtsberatung. Maßgeblich sind immer die im Image enthaltenen Copyright-Dateien und die jeweiligen Originallizenzen.
