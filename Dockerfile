FROM golang:1.11
LABEL maintainer=Hortonworks

COPY . /go/src/github.com/mhmxs/aws-iam-authenticator-service
WORKDIR /go/src/github.com/mhmxs/aws-iam-authenticator-service

RUN make build-linux

FROM alpine

COPY --from=0 /go/src/github.com/mhmxs/aws-iam-authenticator-service/build/Linux/aias /usr/local/bin

EXPOSE 8080

ENTRYPOINT ["aias"]
