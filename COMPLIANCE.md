VERDICT: CHANGES_REQUESTED

---

## 1. DSGVO / GDPR

| Schweregrad | Befund | Konkrete Abhilfe |
|---|---|---|
| Hoch | **Transportsicherheit fehlt als sicherer Standard.** Der Parameter `user` in `GET /flags/{key}/evaluate` kann personenbezogen sein, wird aber über `http.ListenAndServe` unverschlüsselt übertragen, wenn `BIND_ADDR` auf eine nicht-lokale Adresse gesetzt wird und `TLS_CERT`/`TLS_KEY` fehlen. | In `main.go` eine Startprüfung ergänzen: Wenn `bindAddr` nicht `127.0.0.1` oder `::1` ist und kein Zertifikat gesetzt ist, mit `log.Fatal("TLS_REQUIRED: TLS_CERT und TLS_KEY sind für nicht-lokale Binds verpflichtend")` abbrechen. Lokale Entwicklung bleibt weiterhin ohne TLS möglich. |
| Mittel | **Rechtsgrundlage und Transparenz nicht sichtbar.** Die Verarbeitung der `user`-ID erfolgt ausschließlich transient, ist aber personenbezogen. Im geprüften Code/Projektzustand fehlt eine dokumentierte Rechtsgrundlage und ein Datenschutzhinweis. | In `README.md` oder `COMPLIANCE.md` einen Abschnitt „Datenverarbeitung“ ergänzen: User-ID wird nur im RAM für die Hash-Berechnung verarbeitet, nicht gespeichert, nicht geloggt; Rechtsgrundlage Art. 6 Abs. 1 lit. f DSGVO (Feature-Rollout); Löschfrist = Ende des Requests. Zusätzlich externen Datenschutzhinweis sicherstellen. |
| Niedrig | **Mögliche PII in Flag-Keys oder Beschreibungen.** `Description` (max. 1024 Zeichen) und `Key` werden gespeichert und `Key` erscheint im Access-Log-Pfad. Enthält die Administration dort personenbezogene Daten, würde das geloggt/gespeichert. | In `handlers.go` eine Key-Validierung ergänzen, z. B. nur `^[a-z0-9._-]+$` zulassen; in `README.md` dokumentieren, dass `description` keine personenbezogenen Daten enthalten darf. |
| Positiv | Logging verzichtet auf Query-String und `user`-ID (`middleware.go`), der `user`-Wert wird nicht in den Store geschrieben (`evaluate.go`), Fehlerantworten enthalten keine Stacktraces. | Keine Maßnahme erforderlich. |

---

## 2. EU Cyber Resilience Act (CRA)

| Schweregrad | Befund | Konkrete Abhilfe |
|---|---|---|
| Hoch | **Sicherheit nicht „by default“. `withAuth` in `auth.go` lässt bei leerem `API_TOKEN` alle `/flags`-Routen unauthentifiziert passieren.** Das ist nur durch den Default-Bind `127.0.0.1` abgemildert; eine produktive Bereitstellung kann ohne Token und ohne TLS mit einem Konfigurationsfehler öffentlich erreichbar sein. | In `main.go` eine Prüfung vor dem Start: Ist `API_TOKEN` leer und `BIND_ADDR` nicht loopback, Start verweigern oder einen zufällig generierten Token setzen und loggen. Alternativ in `auth.go` bei leerem Token und nicht-loopback-Betrieb alle `/flags`-Anfragen mit `503`/`401` ablehnen. |
| Mittel | **SBOM nicht erkennbar.** Für CRA-relevante Produkte ist eine Software Bill of Materials erforderlich; im sichtbaren Projektzustand fehlt ein SBOM-Artefakt, obwohl `go.mod` nur Standardbibliothek nutzt. | CI-Pipeline (z. B. GitHub Actions) ergänzen: `syft . -o spdx-json > sbom.spdx.json` ausführen und als Release-Artefakt bereitstellen. |
| Niedrig | **Sicherheitseigenschaften und Update-Prozess nicht sichtbar dokumentiert.** `SECURITY.md`/`COMPLIANCE.md` sind vorhanden, aber ihre Inhalte nicht Teil des geprüften Zustands; der Code selbst zeigt keinen Versionierungs-/Update-Mechanismus über die `version`-Variable hinaus. | In `SECURITY.md` ausdrücklich dokumentieren: Sicherheitsannahmen, Patch- und Update-Weg für den Betreiber, Standardkonfiguration, Umgang mit Schwachstellenmeldungen. |
| Positiv | Body-Limit 1 MiB (`handlers.go`), Timeouts (`main.go`), Panic-Recovery ohne interne Leaks (`middleware.go`), thread-sicherer Store (`store.go`), einheitliche JSON-Fehler (`response.go`). | Keine Maßnahme erforderlich. |

---

## 3. EU AI Act

**Nicht einschlägig.** Der Dienst enthält keine KI-Funktion. Die deterministische Rollout-Berechnung per `fnv`-Hash ist eine regelbasierte Auswertung ohne Training, Inferenz oder Modelle.

---

## 4. Pflichttexte & UI

**Nicht einschlägig.** Es handelt sich um eine reine `go-backend`-REST-API ohne Browser-UI, Cookies, Warenkorb oder Verkaufsvorgang. Impressum, Cookie-Banner und Widerrufsbelehrung sind auf Produktebene nicht erforderlich. Die Datenschutzhinweise obliegen dem Betreiber (siehe DSGVO-Abschnitt).

---

## 5. Barrierefreiheit / EAA / BITV / WCAG

**Nicht einschlägig.** Es gibt keinen öffentlichen Web-UI-Endpunkt. Die JSON-Antworten sind maschinenlesbar; eine WCAG-Prüfung ist für diesen Projekttyp nicht vorgesehen.

---

### Gesamteinschätzung

Das Produkt ist funktional und in der Datenminimierung überwiegend sauber umgesetzt. Es verbleiben jedoch behebbare rechtliche und regulatorische Lücken, insbesondere fehlende sichere Standardkonfiguration (Auth/TLS) und fehlende Dokumentationsartefakte für DSGVO und CRA. Es liegt kein fundamentaler Verstoß mit unmittelbarem Blockadecharakter vor, daher **keine Blocker**, aber Änderungen sind vor einer Marktfreigabe erforderlich.