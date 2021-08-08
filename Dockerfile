FROM golang:1.16 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN make test && make


FROM builder AS publisher

RUN mkdir publish
RUN rm -rf /tmp/go-link* && make clean && GOOS=darwin GOARCH=arm64 make build/blob && mv build/blob publish/darwin-arm64
RUN rm -rf /tmp/go-link* && make clean && GOOS=linux GOARCH=amd64 make build/blob && mv build/blob publish/linux-amd64
RUN rm -rf /tmp/go-link* && make clean && GOOS=darwin GOARCH=amd64 make build/blob && mv build/blob publish/darwin-amd64


FROM debian:10 AS runner

COPY --from=builder /app/build/blob /usr/bin/blob

VOLUME ["/src"]
