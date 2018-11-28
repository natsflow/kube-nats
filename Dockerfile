# Build image
FROM golang:1.11 AS builder

ADD https://github.com/golang/dep/releases/download/v0.5.0/dep-linux-amd64 /usr/bin/dep
RUN chmod +x /usr/bin/dep

WORKDIR $GOPATH/src/github.com/natsflow/kube-nats

COPY . ./

RUN dep ensure -vendor-only
RUN CGO_ENABLED=0 go install

# Run image
FROM alpine:latest
COPY --from=builder /go/bin/kube-nats .
ENTRYPOINT "/kube-nats"