FROM golang AS builder
LABEL stage="builder"

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# move to working directory /go-app
WORKDIR /build

# copy the code into the container
COPY setup-nats-js.go .

# build the application
RUN go build -o setup-nats-js setup-nats-js.go

# build a small image containing binary only
FROM natsio/nats-box

COPY --from=builder /build/setup-nats-js /usr/local/bin/
COPY ./consumer-configs /nats-js/consumer-configs
COPY ./stream-configs /nats-js/stream-configs

WORKDIR /nats-js

ENTRYPOINT [ "/bin/sh", "-l" ]
