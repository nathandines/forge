FROM golang:latest AS build

WORKDIR /go/src/github.com/nathandines/forge
COPY . .

RUN make clean && \
  make deps && \
  make

FROM alpine:latest

RUN apk add --no-cache ca-certificates

COPY --from=build /go/src/github.com/nathandines/forge/bin/forge /usr/bin/forge

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
