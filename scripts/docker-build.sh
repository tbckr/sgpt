#!/bin/bash
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


docker build \
  --pull \
  --platform=linux/amd64 \
  --label=org.opencontainers.image.title=sgpt \
  --label=org.opencontainers.image.description=sgpt \
  --label=org.opencontainers.image.url=https://github.com/tbckr/sgpt \
  --label=org.opencontainers.image.source=https://github.com/tbckr/sgpt \
  --label=org.opencontainers.image.licenses=MIT \
  --build-arg TARGETOS=linux \
  --build-arg TARGETARCH=amd64 \
  -t sgpt:latest \
  .
