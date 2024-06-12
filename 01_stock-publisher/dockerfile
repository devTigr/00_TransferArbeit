FROM golang:latest

WORKDIR /app

COPY . .

RUN go build -o stock_publisher main.go

CMD ["./stock_publisher"]