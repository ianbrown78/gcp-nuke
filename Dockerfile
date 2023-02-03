FROM golang:1.19-alpine as builder

RUN apk add --no-cache git make curl openssl

# Configure Go
ENV GOPATH=/go PATH=/go/bin:$PATH CGO_ENABLED=0 GO111MODULE=on
RUN mkdir -p ${GOPATH}/src ${GOPATH}/bin

WORKDIR /src

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN set -x \
 && make build \
 && cp /src/dist/gcp-nuke /usr/local/bin/

FROM alpine:latest
RUN apk add --no-cache ca-certificates

COPY --from=builder /usr/local/bin/* /usr/local/bin/

RUN adduser -D gcp-nuke
USER gcp-nuke

ENTRYPOINT ["/usr/local/bin/gcp-nuke"]