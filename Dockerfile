FROM golang:alpine AS builder
RUN apk update && apk add --no-cache git

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/release-redirector /app/cmd/release-redirector



FROM scratch

COPY --from=builder /go/bin/release-redirector /go/bin/release-redirector

ENTRYPOINT ["/go/bin/release-redirector"]
