VERDICT: BUGS_FOUND

Der Go-Testlauf `go test ./...` schlägt mit Exit-Code 1 fehl. Der Fehler `TestRoutesRegistered` zeigt, dass die Routen `/flags/{key}` (GET und DELETE) nicht als registriert erkannt werden und Anfragen stattdessen vom Catch-all-Handler beantwortet werden. Dies verstößt gegen AC-06 und AC-08.

**Bug: Feature-Flag-Routen `/flags/{key}` (GET/DELETE) werden nicht registriert — Anfragen landen im Catch-all**

- **Titel**: Feature-Flag-Routen `/flags/{key}` (GET/DELETE) werden nicht registriert
- **Symptom**: Ein Client, der `GET /flags/foo` oder `DELETE /flags/foo` aufruft, erhält eine 404-Antwort vom generischen `notFound`-Handler statt der spezifischen Handler-Logik. Die Endpunkte sind damit über den tatsächlichen HTTP-Router nicht erreichbar.
- **Repro**: `go test ./...` ausführen.
- **Evidence**:
  ```
  --- FAIL: TestRoutesRegistered (0.00s)
      --- FAIL: TestRoutesRegistered/flag_get (0.00s)
          routes_test.go:42: GET /flags/foo is not registered: got 404 (hit catch-all)
      --- FAIL: TestRoutesRegistered/flag_delete (0.00s)
          routes_test.go:42: DELETE /flags/foo is not registered: got 404 (hit catch-all)
  ```
- **Suspected file(s)**: `main.go`, insbesondere die Routenregistrierung `mux.HandleFunc("/flags/{key}", ...)`. Der verwendete `http.ServeMux` unterstützt die Wildcard-Syntax `{key}` erst ab Go 1.22. Falls die Go-Version im `go.mod` älter ist oder das Muster nicht korrekt interpretiert wird, wird der Pfad wörtlich behandelt und alle Anfragen an `/flags/...` (außer `/flags` selbst) fallen an den Catch-all `/` zurück. Der Fehler betrifft beide Methoden unter `/flags/{key}`; die Route `/flags` (POST/GET) ist offenbar korrekt registriert, daher liegt das Problem an der Wildcard-Behandlung.
- **Severity**: high