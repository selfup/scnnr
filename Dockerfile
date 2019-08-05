FROM golang:alpine

ENV SCNNR src/github.com/selfup/scnnr

COPY . $GOPATH/$SCNNR

RUN cd $GOPATH/$SCNNR \
  && go run cmd/release/main.go \
  && cp scnnr_bins.zip $HOME/scnnr_bins.zip
