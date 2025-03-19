FROM golang:1.24.1-alpine3.21 as builder

RUN apk add --no-cache gcc git make musl-dev

COPY . /src
ARG BININFO_BUILD_DATE BININFO_COMMIT_HASH BININFO_VERSION # provided to 'make install'
WORKDIR /src
RUN make -C /src install PREFIX=/pkg

################################################################################

FROM alpine:3.21
LABEL org.opencontainers.image.authors="Tilo Geissler <tilo.geissler@sap.com>"
LABEL source_repository="https://github.com/sapcc/schedules2slack"

RUN apk add --no-cache ca-certificates
COPY --from=builder /src/build/ /run/

ARG BININFO_BUILD_DATE BININFO_COMMIT_HASH BININFO_VERSION
LABEL source_repository="https://github.com/sapcc/schedules2slack" \
  org.opencontainers.image.url="https://github.com/sapcc/schedules2slack" \
  org.opencontainers.image.created=${BININFO_BUILD_DATE} \
  org.opencontainers.image.revision=${BININFO_COMMIT_HASH} \
  org.opencontainers.image.version=${BININFO_VERSION}

USER nobody:nobody
WORKDIR /var/empty
ENTRYPOINT [ "/run/schedules2slack", "-config", "/etc/config/config.yaml" ]
