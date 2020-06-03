FROM golang:alpine

ARG CI=true
ARG VERSION=9000

ENV SCNNR src/github.com/selfup/scnnr

RUN mkdir -p go/src/github.com/selfup/scnnr

COPY . $GOPATH/$SCNNR

WORKDIR $GOPATH/$SCNNR

RUN go run cmd/pack/main.go
RUN go run cmd/checksum/main.go

CMD ["sleep", "infinity"]
