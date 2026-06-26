FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
ARG BUILD_TIME=unknown
ARG GIT_COMMIT=unknown
RUN CGO_ENABLED=0 go build -ldflags "\
    -s -w \
    -X godis/version.Version=${VERSION} \
    -X godis/version.BuildTime=${BUILD_TIME} \
    -X godis/version.GitCommit=${GIT_COMMIT}" \
    -o /bin/godis .

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /bin/godis /usr/local/bin/godis
RUN mkdir -p /etc/godis /data /logs
COPY etc/godis.yaml /etc/godis/godis.yaml
EXPOSE 6379
VOLUME ["/data", "/logs"]
ENTRYPOINT ["godis", "--config", "/etc/godis/godis.yaml"]
