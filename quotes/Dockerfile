FROM golang:1.15

RUN mkdir -p /go/src/app
WORKDIR /go/src/app/cmd/app

ADD . /go/src/app

RUN GO111MODULE=on go install

CMD /go/bin/quotes_app

EXPOSE 8080 8082
