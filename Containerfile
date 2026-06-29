ARG BASE_IMAGE_BUILDER=golang
ARG ALPINE_VERSION=3.22
ARG GO_VERSION=1.26

FROM ${BASE_IMAGE_BUILDER}:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder
ARG GOARCH=amd64
ARG GOARM
ARG VERSION
ARG VCS_REF
WORKDIR /tmp/gobuild
COPY ./ .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${GOARCH} GOARM=${GOARM} \
    go build -a -mod=vendor -o lazypodman \
    -tags=containers_image_openpgp,exclude_graphdriver_btrfs,exclude_graphdriver_devicemapper,remote \
    -ldflags="-s -w \
    -X main.commit=${VCS_REF} \
    -X main.version=${VERSION} \
    -X main.buildSource=Podman"

FROM alpine:${ALPINE_VERSION}
ARG BUILD_DATE
ARG VCS_REF
ARG VERSION
LABEL \
    org.opencontainers.image.authors="Clara Vanacker" \
    org.opencontainers.image.created=$BUILD_DATE \
    org.opencontainers.image.version=$VERSION \
    org.opencontainers.image.revision=$VCS_REF \
    org.opencontainers.image.url="https://github.com/ClaraVnk/lazypodman" \
    org.opencontainers.image.documentation="https://github.com/ClaraVnk/lazypodman" \
    org.opencontainers.image.source="https://github.com/ClaraVnk/lazypodman" \
    org.opencontainers.image.title="lazypodman" \
    org.opencontainers.image.description="A lazier way to manage Podman from your terminal"

# Bundle the Podman CLI so the image can drive a Podman engine out of the box.
RUN apk add --no-cache podman

ENTRYPOINT [ "/usr/local/bin/lazypodman" ]
COPY --from=builder /tmp/gobuild/lazypodman /usr/local/bin/lazypodman
