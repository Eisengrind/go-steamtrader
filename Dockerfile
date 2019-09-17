FROM golang:1.12 as build

COPY ./ /go/src/github.com/eisengrind/go-steamtrader/
WORKDIR /go/src/github.com/eisengrind/go-steamtrader/

RUN go install .

FROM alpine:3.6

LABEL maintainer "Vincent Heins"
LABEL type "public"
LABEL versioning "simple"

RUN apk add --no-cache libc6-compat curl
COPY --from=build /go/bin/go-steamtrader /usr/local/bin

ENTRYPOINT [ "go-steamtrader" ]