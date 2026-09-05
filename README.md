# Feature-Flag-Service

Ein Feature-Flag-Service als REST-API in Go. Der Service verwaltet Feature-Flags
in einem thread-sicheren In-Memory-Store (sync.RWMutex + map) und stellt CRUD-
Endpunkte, eine deterministische Rollout-Evaluierung pro Nutzer, einen
Health-Check sowie Zugriffs-Logging als Middleware bereit. Es werden
ausschließlich Module der Go-Standardbibliothek verwendet.

## Tech Stack

- **Sprache**: Go (1.22)
- **Framework**: `net/http` (Standardbibliothek, keine externen Abhängigkeiten)
- **Storage**: In-Memory mit `sync.RWMutex`
- **Testing**: `httptest`, `go test`

## Installation

Es sind keine externen Abhängigkeiten erforderlich. Ein installiertes Go
(>= 1.22) genügt:

```
go mod download
```

## Starten (Dev)

Der Server lauscht auf Port `8080` (überschreibbar über die Umgebungsvariable
`PORT`):

```
go run .
```

Der Server lauscht auf `:` + `PORT` (Standard `:8080`).

## Konfiguration (Env)

| Variable | Pflicht | Default | Beschreibung |
| --- | --- | --- | --- |
| `PORT` | nein | `8080` | Port, auf dem der HTTP-Server lauscht |

## Endpunkte

Alle Antworten sind JSON. Jede Fehlerantwort hat die Form `{"error": "<msg>"}`.

| Methode | Pfad | Beschreibung |
| --- | --- | --- |
| `GET` | `/healthz` | Health-Check → `200 {"status":"ok"}` |
| `POST` | `/flags` | Flag anlegen |
| `GET` | `/flags` | Alle Flags auflisten (nach `key` sortiert) |
| `GET` | `/flags/{key}` | Einzelnes Flag lesen |
| `PUT` | `/flags/{key}` | Flag aktualisieren |
| `DELETE` | `/flags/{key}` | Flag löschen |
| `GET` | `/flags/{key}/evaluate?user={id}` | Rollout-Entscheidung auswerten |

Nicht definierte Pfade antworten mit `404`, falsche HTTP-Methoden auf einem
definierten Pfad mit `405` — jeweils als JSON-Fehlerobjekt.

## Features

- CRUD-Endpunkte für Feature-Flags mit thread-sicherem In-Memory-Store
- Deterministische Rollout-Evaluierung pro Nutzer (`stableHash` auf Basis von `hash/fnv`)
- Health-Check unter `GET /healthz`
- Zugriffs-Logging-Middleware
- Einheitliche JSON-Fehlerobjekte (404/405/500) ohne Stacktraces oder interne Details
