FROM --platform=$BUILDPLATFORM cgr.dev/chainguard/go:1.20 as build

WORKDIR /work

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG TARGETOS TARGETARCH TARGETVARIANT
RUN \
    if [ "${TARGETARCH}" = "arm" ] && [ -n "${TARGETVARIANT}" ]; then \
      export GOARM="${TARGETVARIANT#v}"; \
    fi; \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -v ./cmd/sgpt.go


FROM cgr.dev/chainguard/static:latest

ENV HOME /home/nonroot

COPY --from=build /work/sgpt /sgpt
ENTRYPOINT ["/sgpt"]