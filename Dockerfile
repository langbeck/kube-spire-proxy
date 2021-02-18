FROM golang:alpine AS builder
RUN mkdir /build
ADD . /repo
RUN cd /repo && go build -v -o /build ./cmd/authn-proxy


FROM alpine
COPY --from=builder /build/authn-proxy /usr/bin/authn-proxy
