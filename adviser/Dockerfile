FROM golang:1.15

RUN mkdir -p /go/src/app
WORKDIR /go/src/app/cmd/app

ADD . /go/src/app

RUN apt install bash

ARG GIT_PERSONAL_ACCESS_TOKEN
RUN echo "[url \"https://${GIT_PERSONAL_ACCESS_TOKEN}:@github.com/\"]\n\tinsteadOf = https://github.com/" >> /root/.gitconfig

RUN GO111MODULE=on go install

CMD /go/bin/app

EXPOSE 8080 8082
