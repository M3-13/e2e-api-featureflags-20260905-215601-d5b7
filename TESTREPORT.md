VERDICT: BUGS_FOUND

- **Title:** Testfall `TestRoutesRegistered` verwechselt Handler-404 mit fehlender Route und lässt `go test` scheitern
- **Symptom:** Der Testlauf schlägt fehl, obwohl die angefragten Routen tatsächlich registriert sind. `GET /flags/foo` und `DELETE /flags/foo` erreichen den jeweiligen Handler; dort wird für den unbekannten Key korrekt 404 mit `{"error":"flag not found"}` zurückgegeben. Der Test interpretiert diesen legitimen fachlichen 404 fälschlich als „Catch-all erreicht“ und meldet die Route als nicht registriert.
- **Repro:** `go test ./...` im Projektverzeichnis ausführen.
- **Evidence:**
  ```
  --- FAIL: TestRoutesRegistered (0.00s)
      --- FAIL: TestRoutesRegistered/flag_get (0.00s)
          routes_test.go:42: GET /flags/foo is not registered: got 404 (hit catch-all)
      --- FAIL: TestRoutesRegistered/flag_delete (0.00s)
          routes_test.go:42: DELETE /flags/foo is not registered: got 404 (hit catch-all)
  FAIL
  FAIL	featureflags	0.403s
  ```
- **Suspected file(s):** `routes_test.go` – `TestRoutesRegistered` prüft lediglich, ob `rr.Code != http.StatusNotFound` bzw. `!= http.StatusMethodNotAllowed`. Ein 404 eines realen Handlers (unbekannter Flag-Key, AC-06) ist von einem 404 der `notFound`-Catch-all-Route nicht zu unterscheiden. Der Test muss entweder vor dem Abruf einen passenden Flag-Key anlegen oder ein eindeutigeres Merkmal der Registrierung prüfen (z. B. Response-Body, spezifische Fehlermeldung oder ein erfolgreicher Status nach Flag-Anlage).
- **Severity:** high