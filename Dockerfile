FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /bin/godis .

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /bin/godis /usr/local/bin/godis
RUN mkdir -p /etc/godis /data /logs
COPY etc/godis.yaml /etc/godis/godis.yaml
EXPOSE 6379
VOLUME ["/data", "/logs"]
ENTRYPOINT ["godis", "--config", "/etc/godis/godis.yaml"]
