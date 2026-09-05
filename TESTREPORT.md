VERDICT: BUGS_FOUND

**Bug:**

- **Titel:** TestRoutesRegistered schlägt fehl – GET/DELETE /flags/{key} wird fälschlich als nicht registriert (404) gewertet
- **Symptom:** Die Go-Testsuite läuft nicht grün; `go test ./...` endet mit Exit-Status 1. Zwei Untertests von `TestRoutesRegistered` schlagen fehl, weil die Anfragen an `GET /flags/foo` und `DELETE /flags/foo` einen 404-Status liefern und der Test diesen 404 als Treffer des Catch-all-Routeninterpretiert („not registered“), obwohl die Endpunkte laut Spezifikation existieren und die Handler für unbekannte Keys ebenfalls 404 liefern. Dadurch ist die geforderte Akzeptanzbedingung AC-15 („go test läuft grün durch“) verletzt und der Build ist rot.
- **Repro:** Im Projektverzeichnis `go test ./...` ausführen. Der Test `TestRoutesRegistered` (genauer die Untertests `flag_get` und `flag_delete`) schlägt fehl.
- **Evidence:**
  ```
  --- FAIL: TestRoutesRegistered (0.00s)
      --- FAIL: TestRoutesRegistered/flag_get (0.00s)
          routes_test.go:42: GET /flags/foo is not registered: got 404 (hit catch-all)
      --- FAIL: TestRoutesRegistered/flag_delete (0.00s)
          routes_test.go:42: DELETE /flags/foo is not registered: got 404 (hit catch-all)
  FAIL
  FAIL	featureflags	0.406s
  ```
- **Suspected file(s):** `routes_test.go` (Testlogik wertet jeden 404 als fehlende Registrierung, obwohl `handleGetFlag` und `handleDeleteFlag` bei unbekanntem key 404 liefern) oder `main.go` (falls die Routing-Muster `/flags/{key}` nicht korrekt registriert werden). Da die übrigen Untertests von `TestRoutesRegistered` und die spezifischen Handler-Tests (`TestGetFlagNotFound`, `TestDelete`-Logik) offenbar bestehen, liegt die Ursache vermutlich in der Testannahme und nicht in der Produktregistrierung.
- **Severity:** medium