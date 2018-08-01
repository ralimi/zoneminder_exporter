FROM golang:alpine as builder
MAINTAINER Richard Alimi <rich@velvetsea.net>

# Add packages necessary to build
RUN apk add --no-cache \
    git \
    make

# Setup the build directory
WORKDIR /go/src/github.com/ralimi/zoneminder_exporter
COPY . /go/src/github.com/ralimi/zoneminder_exporter

# Build the binary for ARM
RUN GOOS=linux GOARCH=arm make build

# Create minimal image with built binary
FROM alpine:latest
COPY --from=builder /go/src/github.com/ralimi/zoneminder_exporter/zoneminder_exporter /
EXPOSE 9180/tcp
ENTRYPOINT [ "/zoneminder_exporter" ]
