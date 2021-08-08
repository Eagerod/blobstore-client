FROM golang:1.15 AS builder

WORKDIR /app

COPY go.mod go.sum ./

# Extremely imperfect means of installing packages, but helps with Docker
#   build times
# RUN go get $(grep -zo 'require (\(.*\))' go.mod | sed '1d;$d;' | tr ' ' '@') 
RUN go mod download

COPY . .

RUN make test && make


FROM debian:10 AS runner

COPY --from=builder /app/build/blob /usr/bin/blob

VOLUME ["/src"]
