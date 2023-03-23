#!/bin/bash

go-licenses report github.com/tbckr/sgpt/cmd/sgpt \
  --template .github/licenses.tmpl >licenses/licenses.md
