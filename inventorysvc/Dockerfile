FROM golang AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# move to working directory /go-app
WORKDIR /build

# copy and download dependencies
COPY go.mod .
COPY go.sum .
RUN go mod download

# stage - 2
# ---------
FROM builder AS stage

WORKDIR /build
# copy the code into the container
COPY . .

# build the application
RUN go build -o inventorysvc cmd/main.go

# move to /dist directory as the place for resulting binary directory
WORKDIR /dist

# copy the binary from /build to /dist directory
RUN cp /build/inventorysvc .

# stage - 3
# ---------
# build a small image containing binary only
FROM alpine:3.13

COPY --from=builder /dist/inventorysvc /
RUN mkdir conf
COPY ./conf/dockerised_app_conf.json ./conf/conf.json

# command to run the application
ENTRYPOINT [ "/inventorysvc" ]
