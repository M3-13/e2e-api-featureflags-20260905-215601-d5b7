VERDICT: BUGS_FOUND

**Bug: TestRoutesRegistered interpretiert 404-Antworten der Flag-Handler fälschlich als fehlende Routenregistrierung**

- **Titel**  
  `TestRoutesRegistered` schlägt bei `GET /flags/foo` und `DELETE /flags/foo` fehl, weil der Test einen Status 404 als Beweis für den Catch-All-Handler wertet, obwohl die registrierten Handler bei unbekanntem Key ebenfalls 404 liefern.

- **Symptom**  
  Der Testlauf `go test ./...` bricht mit Exit-Status 1 ab. Die Untertests `flag_get` und `flag_delete` melden fälschlich, die Routen `GET /flags/{key}` bzw. `DELETE /flags/{key}` seien nicht registriert. Tatsächlich sind die Routen registriert und antworten korrekt mit 404, wenn kein Flag mit dem angegebenen Schlüssel existiert (gemäß AC-06 und AC-08). Der rote Testlauf blockiert die Abnahme, obwohl das Produktverhalten spezifikationskonform ist.

- **Repro**  
  `go test ./...` im Projektverzeichnis ausführen.

- **Evidence**  
  ```
  --- FAIL: TestRoutesRegistered (0.00s)
      --- FAIL: TestRoutesRegistered/flag_get (0.00s)
          routes_test.go:42: GET /flags/foo is not registered: got 404 (hit catch-all)
      --- FAIL: TestRoutesRegistered/flag_delete (0.00s)
          routes_test.go:42: DELETE /flags/foo is not registered: got 404 (hit catch-all)
  ```

- **Suspected file(s)**  
  `routes_test.go`, Zeile 42. Der Test prüft `rr.Code == http.StatusNotFound` und schließt daraus auf den Catch-All-Handler. Diese Annahme ist falsch: Auch die echten Handler `handleGetFlag` (in `handlers.go`) und `handleDeleteFlag` (in `handlers.go`) liefern bei unbekanntem Key den Status 404. Die Route ist in `main.go` über `mux.HandleFunc("/flags/{key}", …)` registriert. Der Test muss entweder vor dem Request ein passendes Flag anlegen oder den Status nicht als Indikator für die Registrierung verwenden (z. B. stattdessen prüfen, dass der Handler aufgerufen wird und die Antwort JSON mit `error`-Feld ist). Es handelt sich um einen Testfehler, nicht um einen Produktfehler.

- **Severity**  
  medium (fehlgeschlagener Testlauf blockiert die CI-Abnahme; die funktionale Implementierung ist gemäß Spezifikation korrekt)