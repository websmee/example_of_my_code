FROM golang:1.15

RUN mkdir -p /go/src/app
WORKDIR /go/src/app/cmd/app

ADD . /go/src/app

RUN apt install bash

RUN GO111MODULE=on go install

CMD /go/bin/app

EXPOSE 8080 8082
