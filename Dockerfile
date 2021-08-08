FROM golang:1.16 AS builder

WORKDIR /app

COPY go.mod go.sum ./

# Extremely imperfect means of installing packages, but helps with Docker
#   build times
# RUN go get $(grep -zo 'require (\(.*\))' go.mod | sed '1d;$d;' | tr ' ' '@') 
RUN go mod download

COPY . .

RUN make test && make


FROM builder AS publisher

RUN \
    apt-get update && \
    apt-get install -y \
        build-essential && \
    apt-get clean

RUN mkdir publish
RUN rm -rf /tmp/go-link* && make clean && GOOS=darwin GOARCH=arm64 make build/blob && mv build/blob publish/darwin-arm64
RUN rm -rf /tmp/go-link* && make clean && GOOS=linux GOARCH=amd64 make build/blob && mv build/blob publish/linux-amd64
RUN rm -rf /tmp/go-link* && make clean && GOOS=darwin GOARCH=amd64 make build/blob && mv build/blob publish/darwin-amd64


FROM debian:10 AS runner

COPY --from=builder /app/build/blob /usr/bin/blob

VOLUME ["/src"]
