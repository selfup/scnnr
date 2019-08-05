FROM golang:alpine

ARG CI=""
ARG VERSION=""

ENV SCNNR src/github.com/selfup/scnnr

RUN mkdir -p go/src/github.com/selfup/scnnr

COPY . $GOPATH/$SCNNR

WORKDIR $GOPATH/$SCNNR

RUN go run cmd/release/main.go

CMD ["sleep", "1m"]
