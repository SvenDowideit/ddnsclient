#!/usr/bin/docker build -t ddnsclient .
FROM golang
MAINTAINER Sven Dowideit <SvenDowideit@home.org.au>

ENTRYPOINT ["/go/src/github.com/SvenDowideit/ddnsclient/ddnsclient"]
CMD ["-help"]

# pre-install known dependencies before the source, so we don't redownload them whenever the source changes
RUN go get github.com/vharitonsky/iniflags

WORKDIR /go/src/github.com/SvenDowideit/ddnsclient
COPY . /go/src/github.com/SvenDowideit/ddnsclient

#RUN go get -d -v github.com/SvenDowideit/ddnsclient
#RUN go install github.com/SvenDowideit/ddnsclient

#RUN go get -d -v
ARG VERSION
RUN go build -ldflags "-X main.Version $VERSION" -o ddnsclient main.go
#RUN go test github.com/SvenDowideit/ddnsclient/...



