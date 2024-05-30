# Transferarbeit: Erstellen eines verteilten Systems

Alle informationen und Vorgaben zu der Transferarbeit sind im PDF "Auftrag_LB_PVS_01.pdf" zu finden.

Dieses Dokument dient ab hier also nur noch der Übersicht und Organisiation des Projekts und der Informationszusammentragung für die Durchführung und den Aufbau des Systems.

# Projektteile

## Producer (Stock-Publisher von gitHub)

Simmuliert Daten von Aktien und serndet diese an den MessageBroker

- [ ] zur Verfügung gestellt -> https://github.com/SwitzerChees/stock-publisher
- [ ] Dockerfile dazu schreiben (Go 1.22.2, port localhost:5672)
- [ ] image erstellen
- [ ] image tagen
- [ ] image pushen
- [ ] im Docker Compose integrieren

## Rabbit MQ (Massage Broker)

Entkoppelt den Datentransfer vom Publisher zum Subscriber

- [ ] Dockerfile
- [ ] image erstellen, tagen, pushen
- [ ] im Docker Compose integrieren

## Consumer (Eigenentwicklung)

Verarbeitet die Daten die vom MessageBroker ge"Cued" werden.

- [ ] Consumer in eigener Sprache schreiben
- [ ] Dockerfile dazu erstellen
- [ ] image erstellen, tagen, pushen
- [ ] im Docker Compose integrieren

## MongoDB (Cluster)

Speichert die verarbeiteten Daten in einer Hochverfügbaren ge"cluster"ten Datenbank

## Frontend (Stock-Liveview von GitHub)

Zeigt die verarbeiteten Daten in einer Liveview


## NGINX (Load Balancer)

Organisiert die Anfragen der Frontends.
