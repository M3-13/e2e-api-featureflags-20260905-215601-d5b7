VERDICT: BUGS_FOUND

- **Titel**: `TestRoutesRegistered` schlägt fehl, weil er 404 bei unbekanntem Flag als fehlende Route interpretiert
- **Symptom**: `go test ./...` läuft auf Rot; die Untertests `TestRoutesRegistered/flag_get` und `TestRoutesRegistered/flag_delete` scheitern, obwohl die Handler korrekt 404 für nicht vorhandene Flags liefern (gemäß AC-06/AC-08). Dadurch ist die CI rot, ohne dass ein Produktfehler vorliegt.
- **Repro**: `go test ./...` ausführen; die beiden Untertests schlagen fehl.
- **Evidence**:
  ```
  --- FAIL: TestRoutesRegistered (0.00s)
      --- FAIL: TestRoutesRegistered/flag_get (0.00s)
          routes_test.go:42: GET /flags/foo is not registered: got 404 (hit catch-all)
      --- FAIL: TestRoutesRegistered/flag_delete (0.00s)
          routes_test.go:42: DELETE /flags/foo is not registered: got 404 (hit catch-all)
  ```
- **Suspected file(s)**: `routes_test.go` — die Annahme in `TestRoutesRegistered`, ein 404 stamme nur vom Catch-all, trifft nicht zu, wenn die echten Handler für unbekannte Schlüssel 404 zurückgeben. Die tatsächlichen Handler `handleGetFlag` und `handleDeleteFlag` in `handlers.go` liefern für unbekannte Keys korrekt 404; der Test müsste entweder vorher ein Flag anlegen oder die 404-Antwort als legitime Handler-Antwort akzeptieren. `handlers.go` selbst ist hier nicht defekt.
- **Severity**: medium (Build rot, aber Produktverhalten korrekt)