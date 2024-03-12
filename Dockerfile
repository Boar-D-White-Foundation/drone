FROM golang:1.22.1-alpine as builder

WORKDIR /opt/drone

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o ./bin/drone ./drone


FROM golang:1.22.1-alpine

WORKDIR /opt/drone

RUN apk add --update \
    bash \
    curl \
    && rm -rf /var/cache/apk/*

COPY --from=builder /opt/drone/bin/drone /opt/drone/bin/drone

CMD ["/opt/drone/bin/drone"]
