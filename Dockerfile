FROM golang as builder

WORKDIR /go/src/github.com/SinusBot/installer-script

COPY startup.go .
COPY test.go .

RUN go get -d -v

RUN CGO_ENABLED=0 go build startup.go
RUN CGO_ENABLED=0 go build test.go

FROM ubuntu

RUN apt-get update && \
    apt-get install wget -y

WORKDIR /root/

COPY sinusbot_installer.sh .
COPY test.sh .
COPY --from=builder /go/src/github.com/SinusBot/installer-script/startup .
COPY --from=builder /go/src/github.com/SinusBot/installer-script/test .

ENTRYPOINT ["bash", "test.sh"]
