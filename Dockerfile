FROM golang:1.12.1 as builder
WORKDIR /app/fetch
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/fetch ./cmd/fetch

FROM alpine:latest

WORKDIR /app
RUN apk update && apk upgrade && apk add ca-certificates
COPY --from=builder /app/fetch/bin/fetch .
CMD ["./fetch"]