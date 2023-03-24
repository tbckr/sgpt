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

if ! command -v "addlicense" >/dev/null 2>&1; then
  echo "addlicense not found. Install via golang: go install github.com/google/addlicense@latest"
  exit 1
fi

addlicense -c "Tim <tbckr>" -l MIT -s -v \
  -ignore ".idea/**" \
  -ignore "dist/**" \
  -ignore ".github/licenses.tmpl" \
  -ignore "licenses/*" \
  .

# to check if all files have the correct license header:
# addlicense -c "Tim <tbckr>" -l MIT -s -check \
#   -ignore ".idea/**" \
#   -ignore "dist/" \
#   -ignore ".github/licenses.tmpl" \
#   -ignore "licenses/*" \
#   .
