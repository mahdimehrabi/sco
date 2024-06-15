FROM golang:1.22-alpine


RUN apk add build-base


RUN echo $GOPATH

copy . /app

RUN set -x \
    && apk add --no-cache git \
    && git clone --branch "v4.17.1" --depth 1 --single-branch https://github.com/golang-migrate/migrate /tmp/go-migrate

WORKDIR /tmp/go-migrate

RUN set -x \
    && CGO_ENABLED=0 go build -tags 'postgres' -ldflags="-s -w" -o ./migrate ./cmd/migrate \
    && ./migrate -version

RUN cp /tmp/go-migrate/migrate /usr/bin/migrate

WORKDIR /app

RUN go build /app/cmd/main.go

CMD /app/main
