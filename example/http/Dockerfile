FROM golang:1.21.5@sha256:672a2286da3ee7a854c3e0a56e0838918d0dbb1c18652992930293312de898a6

WORKDIR /go/src/app

COPY main.go go.mod go.sum ./

RUN go get -d ./
RUN go build -o main .

CMD ["./main"]
