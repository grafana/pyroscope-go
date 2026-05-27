FROM golang:1.26.3@sha256:2d6c80227255c3112a4d08e67ba98e58efd3846daf15d9d7d4c389565d881b1a

WORKDIR /go/src/app

COPY main.go go.mod go.sum ./

RUN go get -d ./
RUN go build -o main .

CMD ["./main"]
