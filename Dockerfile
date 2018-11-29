# LiveRecord GoLang application

FROM golang:1.11
LABEL maintainer = "philipp@zoonman.com"

# WORKDIR /go/src/app

COPY . .

RUN go get -d -v ./...
RUN go build

ENTRYPOINT ["./lrs"]