# Build image
FROM golang:1.10-alpine AS build-env

RUN apk add --no-cache git curl
RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 && chmod +x /usr/local/bin/dep

WORKDIR $GOPATH/src/github.com/natsflow/kube-nats

COPY . ./

RUN dep ensure -vendor-only
RUN CGO_ENABLED=0 go install

# Run image
FROM alpine:latest
COPY --from=build-env /go/bin/kube-nats .
ENTRYPOINT "/kube-nats"
