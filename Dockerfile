FROM golang:1.25-alpine AS test

WORKDIR /app
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY internals ./internals
COPY cmd ./cmd

RUN go test ./internals/... ./cmd/...

FROM golang:1.25-alpine AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY internals ./internals
COPY cmd ./cmd

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o app ./cmd/main.go

FROM alpine:3.20

WORKDIR /app
RUN apk add --no-cache ca-certificates

COPY --from=build /app/app /app/app

EXPOSE 8080 50051
CMD ["/app/app"]
