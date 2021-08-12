FROM golang:1.16 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN make test && make


FROM debian:10 AS runner

COPY --from=builder /app/build/blob /usr/bin/blob

VOLUME ["/src"]
