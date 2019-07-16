FROM golang:stretch

WORKDIR /usr/local/go/src/github.com/veonik/eokvin

COPY . .

RUN go get -v ./...

RUN go build -o eokvin ./cmd/eokvin/*.go

CMD ./eokvin
