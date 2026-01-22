# build stage
FROM golang:latest AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o backend ./cmd/server

FROM alpine:latest
WORKDIR /app

RUN apk add --no-cache bash

COPY --from=build /app/backend .

COPY --from=build /app/.env .

EXPOSE 8080
CMD ["./backend"]
