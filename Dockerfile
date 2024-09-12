FROM golang:latest
MAINTAINER linimbus linimbus@126.com

WORKDIR /gopath/
ENV GOPATH=/gopath/
ENV GOOS=linux
ENV CGO_ENABLED=0

RUN go get -u -v github.com/linimbus/tcpproxy
WORKDIR /gopath/src/github.com/linimbus/tcpproxy
RUN go build .

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /opt/
COPY --from=0 /gopath/src/github.com/linimbus/tcpproxy/tcpproxy ./tcpproxy

RUN chmod +x *

ENTRYPOINT ["./tcpproxy"]
