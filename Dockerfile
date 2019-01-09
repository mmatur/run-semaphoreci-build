FROM golang:alpine3.8 as builder

RUN apk --update upgrade \
&& apk --no-cache --no-progress add make git gcc musl-dev \
&& rm -rf /var/cache/apk/*

WORKDIR /go/src/github.com/mmatur/run-semaphoreci-build
COPY . .
RUN make build

FROM alpine:3.8
RUN apk update && apk add --no-cache --virtual ca-certificates
COPY --from=builder /go/src/github.com/mmatur/run-semaphoreci-build/run-semaphoreci-build /usr/bin/run-semaphoreci-build
ENTRYPOINT [ "/usr/bin/run-semaphoreci-build" ]