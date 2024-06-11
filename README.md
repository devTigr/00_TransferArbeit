# Transferarbeit: Erstellen eines verteilten Systems

Alle informationen und Vorgaben zu der Transferarbeit sind im PDF "Auftrag_LB_PVS_01.pdf" zu finden.

Dieses Dokument dient ab hier also nur noch der Übersicht und Organisiation des Projekts und der Informationszusammentragung für die Durchführung und den Aufbau des Systems.

# Projektteile

## Producer (Stock-Publisher von gitHub)

Simmuliert Daten von Aktien und serndet diese an den MessageBroker

- [X] zur Verfügung gestellt -> https://github.com/SwitzerChees/stock-publisher
- [X] Dockerfile dazu schreiben (Go 1.22.2, port localhost:5672)
- [X] image erstellen
- [X] image tagen
- [X] image pushen
- [X] im docker-compose integrieren  -> devtigr/ta_publisher

## Rabbit MQ (Massage Broker)

Entkoppelt den Datentransfer vom Publisher zum Subscriber

- [X] im docker-compose integriert

## Consumer (Eigenentwicklung)

Verarbeitet die Daten die vom MessageBroker ge"Cued" werden.

- [X] Consumer in eigener Sprache schreiben
- [X] Dockerfile dazu erstellen
- [ ] image erstellen, tagen, pushen
- [ ] im docker-compose integrieren

## MongoDB (Cluster)

Speichert die verarbeiteten Daten in einer Hochverfügbaren ge"cluster"ten Datenbank

- [ ] im docker-compose integrieren

## Frontend (Stock-Liveview von GitHub)

Zeigt die verarbeiteten Daten in einer Liveview

- [X] zur Verfügung gestellt -> https://github.com/SwitzerChees/stock-liveview
- [X] Dockerfile dazu schreiben (node:22, npm install, npm start)
- [ ] image erstellen
- [ ] image tagen
- [ ] image pushen
- [ ] im Docker Compose integrieren

## NGINX (Load Balancer)

Organisiert die Anfragen der Frontends.

- [X] im Docker Compose integrieren
