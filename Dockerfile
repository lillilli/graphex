# stage for caching modules
FROM golang:1.12 as build_base

# build envs
ENV GOOS linux
ENV GOARCH amd64

# service configs
ENV PKG github.com/lillilli/graphex

RUN mkdir -p /go/src/${PKG}
WORKDIR /go/src/${PKG}

COPY go.mod go.sum Makefile ./
RUN make setup


# build main stage
FROM build_base as service_builder

COPY . .
RUN make setup && make build_scratch


# result container
FROM alpine as service_runner

ENV SERVICE_NAME websocket_api
ENV PKG github.com/lillilli/graphex

WORKDIR /root/

COPY --from=service_builder /go/src/${PKG}/frontend ./frontend
COPY --from=service_builder /go/src/${PKG}/shared ./shared

COPY --from=service_builder /go/src/${PKG}/prod.yml .
COPY --from=service_builder /go/src/${PKG}/cmd/${SERVICE_NAME}/${SERVICE_NAME} .

EXPOSE 8081
ENTRYPOINT ["./websocket_api", "-config=prod.yml"]