FROM golang:1.26.2@sha256:b54cbf583d390341599d7bcbc062425c081105cc5ef6d170ced98ef9d047c716

WORKDIR /go/src/app

COPY main.go go.mod go.sum ./

RUN go get -d ./
RUN go build -o main .

CMD ["./main"]
