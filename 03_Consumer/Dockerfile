# Build stage
FROM golang:latest AS build

WORKDIR /app

COPY . .

RUN go mod download
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o consumer main.go

# Run stage
FROM scratch

WORKDIR /root/

COPY --from=build /app/consumer .

CMD ["./consumer"]
