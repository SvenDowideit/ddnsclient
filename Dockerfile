#!/usr/bin/docker build -t ddnsclient .
FROM golang:onbuild
MAINTAINER Sven Dowideit <SvenDowideit@home.org.au>

ENTRYPOINT ["go-wrapper", "run"]
CMD ["-help"]
