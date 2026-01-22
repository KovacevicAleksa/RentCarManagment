FROM golang:latest AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o car ./cmd/server

FROM alpine:latest
WORKDIR /app

COPY --from=build /app/car .

EXPOSE 8090
CMD ["./car"]
