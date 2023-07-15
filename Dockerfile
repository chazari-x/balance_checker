FROM golang:1.20

WORKDIR /go/src/github.com/chazari-x/WEBSITE.chazari/tree/master

COPY go.mod go.sum .././
RUN go mod download && go mod verify

COPY ../.. .
RUN go build -o /usr/local/bin/app

CMD ["app"]