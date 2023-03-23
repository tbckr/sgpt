#!/bin/bash

addlicense -c "Tim <tbckr>" -l MIT -s -check \
  -ignore ".github/**" \
  -ignore ".idea/**" \
  .
