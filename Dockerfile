FROM --platform=${BUILDPLATFORM} golang:alpine as base

RUN apk update
RUN apk add -U --no-cache ca-certificates && update-ca-certificates
RUN apk add git

RUN adduser -S -u 20000 -H inventory

WORKDIR /src
ENV CGO_ENABLED=0
COPY go.* .
RUN --mount=type=cache,target=/go/pkg/modx \
    go mod download

FROM base AS builder
ARG TARGETOS
ARG TARGETARCH

ARG VERSION
ARG COMMIT

RUN --mount=target= \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /out/k8s-inventory-client \
    -tags release \
    -ldflags "-s -w -X github.com/neticdk-k8s/k8s-inventory-client/collect/version.VERSION=${VERSION} -X github.com/neticdk-k8s/k8s-inventory-client/collect/version.COMMIT=${COMMIT}"

FROM scratch AS bin-unix
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /out/k8s-inventory-client /k8s-inventory-client
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

USER 20000

FROM bin-unix AS bin-linux
FROM bin-unix AS bin-darwin

FROM bin-${TARGETOS} as bin

EXPOSE 8086
ENTRYPOINT ["/k8s-inventory-client"]

ARG COMMIT=
ARG VERSION=

LABEL commit="$COMMIT" version="$VERSION"
