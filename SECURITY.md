VERDICT: CHANGES_REQUESTED

## Scanner-Abdeckung
Für den Projekttyp `go-backend` wurde kein automatisierter Security-Scanner ausgeführt. Es liegen daher keine Scanner-Befunde vor. Die folgende Bewertung basiert ausschließlich auf manueller Quellcode-Analyse.

## Sicherheitsbericht

### 1. Mittel — Fehlende/zwingend optionale Authentifizierung bei unsicherer Netzwerkbindung
**Betroffene Stellen:** `auth.go`, `main.go`

Der Dienst startet standardmäßig mit `API_TOKEN=""`, wodurch sämtliche `/flags`-Routen ohne Authentifizierung erreichbar sind. Standardmäßig wird zwar an `127.0.0.1` gebunden, aber sobald `BIND_ADDR=0.0.0.0` gesetzt wird, sind unauthentifiziertes Erstellen, Ändern und Löschen von Feature-Flags aus dem Netz möglich. Das kann zu Manipulation des Rollout-Verhaltens und unkontrolliertem Speicherwachstum führen.

**Fix:**
- Beim Start prüfen: Wird an eine Nicht-Loopback-Adresse gebunden und ist `API_TOKEN` leer, entweder mit klarer Fehlermeldung verweigern oder mindestens eine deutliche Warnung loggen.
- Beispiel in `main.go`:
  ```go
  token := os.Getenv("API_TOKEN")
  if token == "" && !isLoopback(bindAddr) {
      log.Fatal("API_TOKEN must be set when binding to a non-loopback address")
  }
  ```
- Zusätzlich in der Betriebsdokumentation festhalten, dass Production-Deployments immer `API_TOKEN` und TLS setzen müssen.

### 2. Mittel — Unbegrenzter In-Memory-Store und fehlende Ratenbegrenzung (Speicher-DoS)
**Betroffene Stellen:** `store.go`, `handlers.go`

Es gibt keine Begrenzung für die Anzahl der angelegten Flags. Jeder authentifizierte Client – bzw. bei leerem `API_TOKEN` jeder erreichbare Client – kann beliebig viele Flags mit bis zu 1 MiB großem Request-Body anlegen. Zwar ist der einzelne Body auf 1 MiB begrenzt, der Gesamtspeicher des Stores wächst jedoch unbegrenzt.

**Fix:**
- Maximale Anzahl Flags einführen, z. B. über Umgebungsvariable `MAX_FLAGS` mit einem konservativen Standardwert (etwa 10.000).
- Beim Überschreiten konsistent mit `429 Too Many Requests` oder `507 Insufficient Storage` als JSON-Fehlerobjekt antworten, ohne Flag zu speichern.
- Optional eine einfache Rate-Limit-Middleware vor die schreibenden Routen schalten.
- `maxKeyLength` einführen (z. B. 256) und in `handleCreateFlag` prüfen, um einzelne exzessive Keys zu verhindern.

### 3. Niedrig — Auth-Präfix zu breit
**Betroffene Stelle:** `auth.go`

`strings.HasPrefix(r.URL.Path, "/flags")` schützt nicht nur `/flags` und `/flags/...`, sondern auch beliebige andere Pfade, die mit `/flags` beginnen, etwa `/flagship`. Das ist derzeit kein direkter Exploit, aber eine unpräzise Zugriffsregel, die bei künftigen Routen zu unerwartetem Authentifizierungszwang führen kann.

**Fix:**
```go
if r.URL.Path != "/flags" && !strings.HasPrefix(r.URL.Path, "/flags/") {
    next.ServeHTTP(w, r)
    return
}
```

### 4. Niedrig — JSON-Decoder akzeptiert angehängte Zusatzdaten
**Betroffene Stelle:** `handlers.go`, Funktion `decodeBody`

`json.NewDecoder(r.Body).Decode(dst)` liest nur das erste JSON-Dokument. Ein Body wie
`{"key":"a","enabled":true}{"key":"b","enabled":true}` oder angehängter Nicht-JSON-Müll wird nach dem ersten Wert nicht geprüft. Das öffnet keine direkte RCE-Lücke, umgeht aber die beabsichtigte strikte Body-Validierung.

**Fix:**
Nach der ersten Dekodierung prüfen, dass kein weiteres Token folgt:
```go
if dec.Decode(&struct{}{}) != io.EOF {
    writeError(w, http.StatusBadRequest, "invalid request body")
    return false
}
```
`io` ist zusätzlich zu importieren.

### 5. Niedrig — Fehlende Transport-/Response-Härtung
**Betroffene Stellen:** `response.go`, `main.go`

- `writeJSON` setzt keinen `X-Content-Type-Options: nosniff`-Header, obwohl alle Antworten als `application/json` ausgeliefert werden.
- TLS ist optional; im Standardbetrieb läuft die API unverschlüsselt über HTTP. Die API überträgt im authentifizierten Betrieb den Bearer-Token im Klartext, sofern kein TLS aktiviert ist.

**Fix:**
```go
w.Header().Set("Content-Type", "application/json")
w.Header().Set("X-Content-Type-Options", "nosniff")
```
Für Produktion die Nutzung von `TLS_CERT`/`TLS_KEY` verbindlich dokumentieren bzw. beim Binden an Nicht-Loopback-Adressen TLS erzwingen.

### 6. Niedrig — Key-Validierung zu schwach
**Betroffene Stelle:** `handlers.go`, `handleCreateFlag`

Der Flag-Key wird nur auf leer geprüft. Keys mit `/`, Steuerzeichen oder sehr großer Länge können angelegt werden. Ein Key mit `/` ist über die späteren Route-Muster `GET /flags/{key}` nicht mehr adressierbar, was zu einem funktionalen Defekt führt und die Wartung erschwert.

**Fix:**
- Key zusätzlich auf eine URL-sichere Zeichenklasse beschränken, z. B. `[A-Za-z0-9._-]{1,256}`.
- Die Validierung zentral vor dem Schreiben im Handler durchführen, sodass ungültige Keys mit `400` und einheitlichem JSON-Fehlerobjekt abgewiesen werden.

## Nicht beanstandet
- Keine hartkodierten Secrets oder Token im Produktivcode.
- Keine SQL-, Command- oder Pfad-Injection erkennbar.
- `http.MaxBytesReader` begrenzt Request-Bodies korrekt auf 1 MiB und liefert 413.
- Evaluierung behandelt `user` ausschließlich im Request; der Wert wird nicht gespeichert.
- Logging protokolliert nur Methode, `r.URL.Path`, Status und Dauer; Query-Strings und User-IDs werden nicht geloggt.
- Panic-Recovery verhindert Stacktrace-Leaks in Antworten.