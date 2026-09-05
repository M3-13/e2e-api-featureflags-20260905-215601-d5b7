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
| `BIND_ADDR` | nein | `127.0.0.1` | Adresse, an die der HTTP-Server gebunden wird |
| `API_TOKEN` | optional (für Produktion erforderlich) | – | Bearer-Token für die Zugriffskontrolle; ist es gesetzt, ist ein gültiges `Authorization: Bearer <token>` Pflicht |
| `TLS_CERT` | optional | – | Pfad zum TLS-Serverzertifikat (PEM); wird gemeinsam mit `TLS_KEY` benötigt |
| `TLS_KEY` | optional | – | Pfad zum privaten TLS-Schlüssel (PEM); wird gemeinsam mit `TLS_CERT` benötigt |

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

## Datenschutz

Der Service verarbeitet ausschließlich die folgenden Daten:

- **Flag-Metadaten**: `key`, `enabled`, `description` und `rollout_percent` der
  Feature-Flags.
- **Query-Parameter `user`** (transient): bei `GET /flags/{key}/evaluate?user={id}`.

**Rechtsgrundlage** ist Art. 6 Abs. 1 lit. f DSGVO (berechtigtes Interesse) für
den Feature-Rollout. Der Rollout erfordert eine deterministische, pro Nutzer
stabile Zuordnung, ohne dass der konkrete Nutzer dafür identifiziert werden muss.

**Umgang mit `user`:** Der Wert des Query-Parameters `user` wird ausschließlich
transient im Arbeitsspeicher für die Berechnung des Hash-Werts (`stableHash`)
genutzt. Er wird weder gespeichert noch geloggt und verlässt den Prozess nicht.
Es reicht grundsätzlich ein pseudonymer Identifier (z. B. eine Nutzer- oder
Session-ID) aus; eine echte Identität ist nicht erforderlich.

**Verbindliche Vorgabe:** Das Feld `description` darf keine personenbezogenen
Daten enthalten.

## Sicherheit & Betrieb

- **TLS**: TLS muss in Produktion vorgelagert terminiert werden, z. B. über einen
  Reverse-Proxy oder Load-Balancer. Der Service selbst kann optional über
  `TLS_CERT`/`TLS_KEY` TLS anbieten; im Regelbetrieb erfolgt die Terminierung
  davor.
- **Zugriffskontrolle**: entweder über `API_TOKEN` (Bearer-Token, gesendet als
  `Authorization: Bearer <token>`) oder über einen authentifizierenden
  Reverse-Proxy, der den Zugriff auf den Service absichert.
- **Bindung**: Der Server bindet standardmäßig an `127.0.0.1` und ist damit
  nicht direkt aus dem Netzwerk erreichbar. Für einen öffentlichen Betrieb ist
  der Zugriff ausschließlich über den vorgelagerten, abgesicherten Endpunkt
  freizugeben.
- **Netzsegmentierung**: Der Service gehört in ein eigenes, restriktiv
  abgesichertes Netzwerksegment. Nur die Komponenten, die den Service tatsächlich
  benötigen, erhalten Netzwerkzugriff.
- **Update-/Patchprozess**: Abhängigkeiten (Standardbibliothek) und die
  Go-Version sind regelmäßig zu aktualisieren. Sicherheitsrelevante Updates sind
  zeitnah einzuspielen.
- **SBOM / Stückliste**: `go.mod` enthält keine externen Abhängigkeiten; die
  Software-Stückliste (SBOM) lässt sich jederzeit mit `go version -m .`
  erzeugen.

## Features

- CRUD-Endpunkte für Feature-Flags mit thread-sicherem In-Memory-Store
- Deterministische Rollout-Evaluierung pro Nutzer (`stableHash` auf Basis von `hash/fnv`)
- Health-Check unter `GET /healthz`
- Zugriffs-Logging-Middleware
- Einheitliche JSON-Fehlerobjekte (404/405/500) ohne Stacktraces oder interne Details
