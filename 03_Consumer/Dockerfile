FROM golang:latest

WORKDIR /app

COPY . .

RUN go mod download github.com/streadway/amqp

RUN go mod tidy

RUN go build -o consumer main.go

CMD ["./consumer"]