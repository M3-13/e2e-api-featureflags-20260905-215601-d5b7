VERDICT: BLOCKED

## Sicherheitsbericht

### 1. Fehlende Authentifizierung / Autorisierung auf allen Flag-Endpunkten
- **Schweregrad:** Hoch
- **Betroffene Stelle:** `main.go` (`routes()` und `main()`), `handlers.go` (alle `/flags`-Handler), `evaluate.go`
- **Beschreibung:** Die REST-API besitzt keinerlei Authentifizierung oder Autorisierung. Jeder, der den Port erreichen kann, kann uneingeschränkt Flags anlegen, lesen, ändern und löschen sowie Rollout-Entscheidungen beeinflussen. Da Feature-Flags Produktionsverhalten steuern, ist der ungeschützte Verwaltungszugriff kritisch. Es handelt sich nicht um einen klassischen „Auth-Bypass“, sondern um vollständig fehlende Zugriffskontrolle.
- **Konkrete Korrektur:** Vor alle `/flags`-Routen eine Authentifizierungs-/Autorisierungs-Middleware schalten, z. B. ein konstanter Bearer-Token-/API-Key-Vergleich aus der Umgebungsvariable `API_TOKEN`.  
  Beispiel:
  ```go
  func withAuth(next http.Handler) http.Handler {
      expected := os.Getenv("API_TOKEN")
      return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          if expected == "" || !hmac.Equal([]byte(r.Header.Get("Authorization")), []byte("Bearer "+expected)) {
              writeError(w, http.StatusUnauthorized, "unauthorized")
              return
          }
          next.ServeHTTP(w, r)
      })
  }
  ```
  Alternativ den Dienst ausschließlich hinter einem authentifizierenden Reverse Proxy / API-Gateway betreiben. Die Tests müssen entsprechend einen `Authorization`-Header setzen oder die Middleware gezielt isoliert testen.

### 2. HTTP-Server ohne Timeouts
- **Schweregrad:** Mittel
- **Betroffene Stelle:** `main.go` (`server := &http.Server{Addr: ":" + port, Handler: api.routes()}`)
- **Beschreibung:** Es sind keine `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, `IdleTimeout` oder `MaxHeaderBytes` gesetzt. Der Server ist dadurch anfällig für Slowloris- und Ressourcenerschöpfungsangriffe.
- **Konkrete Korrektur:** Timeouts explizit setzen, z. B.:
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
  Dafür `time` importieren.

### 3. Standardmäßiges Binden an alle Netzwerkschnittstellen ohne TLS
- **Schweregrad:** Mittel
- **Betroffene Stelle:** `main.go` (`Addr: ":" + port`)
- **Beschreibung:** Der Dienst lauscht standardmäßig auf `0.0.0.0` und ist damit potenziell aus dem gesamten Netzwerk erreichbar. Es ist keine TLS-Absicherung vorgesehen. Flag-Metadaten und Steuerentscheidungen werden unverschlüsselt übertragen.
- **Konkrete Korrektur:** Standardmäßig nur auf `127.0.0.1` binden oder die Bind-Adresse explizit konfigurierbar machen:
  ```go
  bindAddr := os.Getenv("BIND_ADDR")
  if bindAddr == "" {
      bindAddr = "127.0.0.1"
  }
  server := &http.Server{
      Addr: bindAddr + ":" + port,
      // ...
  }
  ```
  In Produktion zusätzlich TLS (z. B. über `ListenAndServeTLS`) oder einen TLS-terminierenden Reverse Proxy verwenden.

### 4. Unbegrenzte Länge des `user`-Query-Parameters
- **Schweregrad:** Niedrig
- **Betroffene Stelle:** `evaluate.go` (`handleEvaluate`)
- **Beschreibung:** Der `user`-Wert aus `GET /flags/{key}/evaluate?user=...` hat keine applikationsseitige Längenbegrenzung. Die FNV-1a-Hash-Berechnung und Query-Verarbeitung sind linear zur Eingabelänge; sehr lange Werte können unnötig CPU verbrauchen. Der Wert wird zwar korrekt nicht geloggt oder gespeichert, die Eingabeverarbeitung ist jedoch nicht begrenzt.
- **Konkrete Korrektur:** Maximallänge prüfen, z. B.:
  ```go
  const maxUserLength = 256
  if len(user) > maxUserLength {
      writeError(w, http.StatusBadRequest, "user must be at most 256 characters")
      return
  }
  ```

## Zusammenfassung
- **Positiv:** Keine Secrets im Code. JSON-Parser mit 1-MiB-Body-Limit. Einheitliche JSON-Fehlerantworten. Kein Query-String-Logging. Keine externen dependencies mit bekannten Schwachstellen. RWMutex-Race-Schutz vorhanden.
- **Kritisch:** Fehlende Authentifizierung/Autorisierung auf allen mutierenden und lesenden Flag-Endpunkten. Dies ist der Grund für den Blocker.

Der Service muss vor einem produktiven Einsatz zwingend um eine Zugriffskontrolle und sichere Server-Konfiguration ergänzt werden.