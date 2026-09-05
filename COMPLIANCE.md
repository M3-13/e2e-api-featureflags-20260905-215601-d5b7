VERDICT: CHANGES_REQUESTED

**Prüfrahmen:** Feature-Flag-Service als REST-API in Go (`go-backend`). Geprüft wurden DSGVO, EU Cyber Resilience Act (CRA), KI-Verordnung, Pflichttexte/UI sowie Barrierefreiheit. Mangels Endnutzer-UI sind Pflichttexte, Cookie-/Consent-Pflichten und Barrierefreiheit **nicht anwendbar**. Eine KI-Funktion ist nicht vorhanden, daher ist die KI-Verordnung **nicht anwendbar**.

---

## 1. Datenschutz-Grundverordnung (DSGVO)

### DSGVO-1: Fehlende Authentifizierung/Autorisierung für alle Endpunkte
**Schweregrad:** hoch  
**Befund:** Sämtliche Endpunkte (`/flags`, `/flags/{key}`, `/flags/{key}/evaluate`) sind ohne jegliche Authentifizierung oder Autorisierung erreichbar. Das Feld `description` kann durch den Betreiber mit personenbezogenen Daten befüllt werden; ein unberechtigter Zugriff würde daher Daten offenlegen. Zudem kann jeder mit Netzzugriff Flags anlegen, ändern oder löschen. Das verletzt Art. 5 Abs. 1 lit. f, Art. 25 und Art. 32 DSGVO (Integrität und Vertraulichkeit).

**Konkrete Abhilfe:**  
- In `middleware.go` eine Authentifizierungs-Middleware ergänzen, die vor den Handlern einen API-Key oder Bearer-Token prüft.  
- Alternativ in `main.go` verbindlich dokumentieren und erzwingen, dass der Dienst ausschließlich hinter einem authentifizierenden Reverse-Proxy betrieben wird.  
- Rollen vorsehen (z. B. Lesen vs. Schreiben), um Datenzugriffe zu minimieren.

### DSGVO-2: Keine TLS-Terminierung
**Schweregrad:** hoch  
**Befund:** `main.go` startet den Server mit `server.ListenAndServe()` auf `":" + port` (alle Interfaces). Der Query-Parameter `user` aus `GET /flags/{key}/evaluate` – potenziell ein personenbezogenes Merkmal – wird bei aktivem HTTP unverschlüsselt übertragen. Auch Flag-Beschreibungen könnten personenbezogene Daten enthalten und im Klartext über das Netz gehen. Das widerspricht Art. 32 DSGVO (Verschlüsselung bei Übertragung).

**Konkrete Abhilfe:**  
- In `main.go` `ListenAndServeTLS` verwenden, wenn Zertifikat und Schlüssel als Umgebungsvariablen vorhanden sind.  
- Andernfalls einen Startabbruch mit klarer Fehlermeldung einbauen, wenn kein TLS konfiguriert ist, damit ein unsicherer Betrieb nicht versehentlich startet.  
- Mindestens in der Deployment-Dokumentation festlegen, dass TLS zwingend vorgelagert sein muss (Reverse-Proxy/TLS-Terminator).

### DSGVO-3: Log-Injection über nicht bereinigten Pfad
**Schweregrad:** mittel  
**Befund:** `middleware.go` loggt `r.URL.Path` direkt in `accessLogger.Printf`. Ein Angreifer kann über URL-Encoding Steuerzeichen (z. B. `%0a`) in den Pfad einbringen und dadurch Log-Einträge fälschen oder unübersichtlich machen. Das beeinträchtigt die Integrität der Protokolle (Art. 32 DSGVO).

**Konkrete Abhilfe:**  
In `middleware.go` vor dem Loggen eine Sanitizer-Funktion anwenden, z. B.:  
```go
logPath := strings.Map(func(r rune) rune {
    if r < 32 || r == 127 {
        return '?'
    }
    return r
}, r.URL.Path)
```  
Dann `logPath` statt `r.URL.Path` loggen. Die Funktion des Dienstes bleibt unverändert; nur die Logausgabe wird bereinigt.

### DSGVO-4: Möglicherweise personenbezogene Daten im Feld `description`
**Schweregrad:** niedrig  
**Befund:** `handlers.go` übernimmt `description` ungeprüft und `store.go` hält Flags dauerhaft im Speicher. Es gibt keine Vorgabe, dass dieses Feld keine personenbezogenen Daten enthalten darf. Das ist ein potenzielles Datenminimierungsrisiko (Art. 5 Abs. 1 lit. c DSGVO).

**Konkrete Abhilfe:**  
- Im `README.md` verbindlich dokumentieren: „`description` darf keine personenbezogenen Daten enthalten.“  
- Optional eine Längenbegrenzung oder Validierung für `description` in `handlers.go` ergänzen, um übermäßig große oder unerwartete Inhalte zu begrenzen.

### DSGVO-5: Rechtsgrundlage für die transiente Verarbeitung des `user`-Werts nicht dokumentiert
**Schweregrad:** mittel  
**Befund:** `evaluate.go` verarbeitet den Query-Parameter `user` zur Hash-Berechnung. Der Wert wird weder gespeichert noch geloggt – das ist positiv –, aber die Rechtsgrundlage ist nicht dokumentiert. Der Verantwortliche muss die Verarbeitung auf eine Rechtsgrundlage stützen (z. B. Art. 6 Abs. 1 lit. f DSGVO für Feature-Rollout).

**Konkrete Abhilfe:**  
- Im `README.md` einen Datenschutzabschnitt ergänzen: welche Daten verarbeitet werden, auf welcher Rechtsgrundlage, dass `user` nur transient im Arbeitsspeicher verarbeitet und nicht gespeichert wird.  
- Dies ist eine organisatorische Pflicht, nicht im Go-Code umsetzbar.

---

## 2. EU Cyber Resilience Act (CRA)

### CRA-1: Fehlende Timeouts am HTTP-Server
**Schweregrad:** hoch  
**Befund:** `main.go` erzeugt `http.Server{Addr: ":" + port, Handler: api.routes()}` ohne `ReadTimeout`, `ReadHeaderTimeout`, `WriteTimeout` oder `IdleTimeout`. Dadurch ist der Server anfällig für Slowloris-Angriffe und Ressourcenerschöpfung. Das verletzt die Anforderung „security by design/default“ des CRA.

**Konkrete Abhilfe:**  
In `main.go` Timeouts setzen, z. B.:  
```go
server := &http.Server{
    Addr:              ":" + port,
    Handler:           api.routes(),
    ReadHeaderTimeout: 5 * time.Second,
    ReadTimeout:       10 * time.Second,
    WriteTimeout:      10 * time.Second,
    IdleTimeout:       60 * time.Second,
}
```  
Dazu `time` importieren.

### CRA-2: Keine Panic-Recovery in der Middleware
**Schweregrad:** mittel  
**Befund:** `middleware.go` (`withLogging`) fängt keine Panics. Ein unbehandelter Panic in einem Handler kann den gesamten Serverprozess beenden und damit die Verfügbarkeit beeinträchtigen. Das widerspricht der CRA-Anforderung an robuste, ausfallsichere Software.

**Konkrete Abhilfe:**  
In `withLogging` einen `defer` mit `recover()` einbauen, der bei einem Panic eine JSON-Fehlerantwort mit Status 500 liefert und den Fehler nur intern loggt (ohne Stacktrace im Response-Body, passend zu AC-19):  
```go
defer func() {
    if rec := recover(); rec != nil {
        writeError(w, http.StatusInternalServerError, "internal server error")
        accessLogger.Printf("panic: %v", rec)
    }
}()
```  
Dabei sicherstellen, dass der Statusrecorder auf 500 gesetzt wird, damit die Logzeile den korrekten Status zeigt.

### CRA-3: Dokumentation der Sicherheitseigenschaften und SBOM nicht sichtbar
**Schweregrad:** mittel  
**Befund:** Die CRA verlangt dokumentierte Sicherheitseigenschaften und eine Stückliste (SBOM). `go.mod` (3 Zeilen) ist vorhanden, aber sein Inhalt wurde nicht vorgelegt; `README.md` existiert, aber der Inhalt ist nicht sichtbar. Eine Bestätigung dieser Pflichten ist daher derzeit nicht möglich.

**Konkrete Abhilfe:**  
- `README.md` um einen Abschnitt „Sicherheit & Betrieb“ ergänzen: verarbeitete Daten, Vertrauensgrenzen, Deployment-Anforderungen (TLS, Authentifizierung, Netzsegmentierung), Update-/Patchprozess.  
- SBOM erzeugen (z. B. mit `go version -m` oder `syft`) und im Repository ablegen.

### CRA-4: Versions- und Patch-Information fehlt
**Schweregrad:** niedrig  
**Befund:** Es gibt keinen Versionsendpunkt oder ein Versionsfeld, das im Betrieb den Patchstand ausweist. Für die CRA-Nachvollziehbarkeit von Updates ist eine Versionsangabe sinnvoll.

**Konkrete Abhilfe:**  
- In `main.go` eine Versionsvariable ergänzen (z. B. `var version = "dev"`, per `-ldflags` setzbar).  
- `/healthz` um ein `version`-Feld erweitern. Bestehende Tests tolerieren zusätzliche Felder, da sie nur `status` prüfen.

---

## 3. KI-Verordnung
Nicht anwendbar: Es ist keine KI-Funktion vorhanden.

## 4. Pflichttexte und UI
Nicht anwendbar: Reines Backend ohne Endnutzer-UI; keine Impressums-, Cookie- oder Consent-Pflichten.

## 5. Barrierefreiheit
Nicht anwendbar: Keine öffentliche Web-UI.

---

## Abgleich „eigene Auflagen nicht brechen“

Alle vorgeschlagenen Maßnahmen sind mit der Funktion des Dienstes vereinbar:
- **Authentifizierung/TLS** können vor dem Dienst oder im Dienst ergänzt werden; die legitimen API-Flows (`POST /flags`, `GET /flags`, `GET /flags/{key}/evaluate`, `PUT`, `DELETE`, `GET /healthz`) bleiben funktionsfähig.
- **Timeouts** betreffen nur den Serverbetrieb und brechen keine funktionale Anforderung.
- **Panic-Recovery** erhält den Dienstbetrieb und liefert weiterhin JSON-Fehler.
- **Log-Sanitisierung** verändert nur die Logzeile, nicht die API-Antworten.

---

**Positive Befunde:**  
- Maximale Body-Größe von 1 MiB wird durchgesetzt und mit 413 als JSON-Fehler beantwortet.  
- Fehlerantworten sind einheitliche JSON-Objekte und enthalten keine Stacktraces oder internen Pfade.  
- Die Logging-Middleware protokolliert keine Query-Strings und keine `user`-Werte.  
- Der `user`-Wert wird nicht im In-Memory-Store gespeichert; die Evaluierung ist deterministisch.  
- `sync.RWMutex` im Store sorgt für Race-Freiheit; Tests decken die Kernfunktionen ab.