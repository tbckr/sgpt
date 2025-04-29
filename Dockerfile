# Copyright (c) 2023 Tim <tbckr>
#
# Permission is hereby granted, free of charge, to any person obtaining a copy of
# this software and associated documentation files (the "Software"), to deal in
# the Software without restriction, including without limitation the rights to
# use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
# the Software, and to permit persons to whom the Software is furnished to do so,
# subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
# FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
# COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
# IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
# CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
#
# SPDX-License-Identifier: MIT

ARG BUILDPLATFORM=linux/amd64
ARG BASE_IMAGE_VERSION=golang:1.24@sha256:f52b85c1b58271ba213ac59878c4fe2fdbb41ac842c5c69f9e7a9d16d53c4ad8
FROM --platform=$BUILDPLATFORM ${BASE_IMAGE_VERSION} as build

WORKDIR /go/src/github.com/tbckr/sgpt
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o sgpt -v ./cmd/sgpt/main.go

FROM cgr.dev/chainguard/static:latest@sha256:2e3db1641bb4fe4e85d2210f4aadb79252e90d5fa745f53a3ffed6a1aab4f73b
ENV HOME /home/nonroot
VOLUME /home/nonroot
COPY --from=build /go/src/github.com/tbckr/sgpt/sgpt /sgpt
ENTRYPOINT ["/sgpt"]