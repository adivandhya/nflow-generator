FROM golang

MAINTAINER Adivandhya B R <adivandhya@gmail.com>

ADD . /go/src/github.com/admin/nflow-generator

RUN go get github.com/Sirupsen/logrus \
    && go get github.com/jessevdk/go-flags \
    && go install github.com/admin/nflow-generator

ENTRYPOINT ["/go/bin/nflow-generator"]
