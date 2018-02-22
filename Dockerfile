FROM docker:dind

RUN apk update && \
    apk add --no-cache \
      ca-certificates jq bash

ADD ./cmd/common.sh ./cmd/in.sh ./cmd/out.sh ./cmd/check.sh /opt/resource/
