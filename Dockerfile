FROM golang:1.22.1-alpine

WORKDIR /opt/drone

RUN apk add --update \
    bash \
    curl \
    && rm -rf /var/cache/apk/*

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /opt/drone/bin/drone ./src/drone

CMD ["./bin/drone"]
