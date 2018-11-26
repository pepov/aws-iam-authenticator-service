FROM golang:1.11
LABEL stage=intermediate

COPY . /go/src/github.com/hortonworks/aws-iam-authenticator-service
WORKDIR /go/src/github.com/hortonworks/aws-iam-authenticator-service

ENV GO111MODULE on

RUN make build-linux

FROM alpine:3.8
LABEL maintainer=Hortonworks

COPY --from=0 /go/src/github.com/hortonworks/aws-iam-authenticator-service/build/Linux/aias /usr/local/bin

EXPOSE 8080

ENTRYPOINT ["aias"]
