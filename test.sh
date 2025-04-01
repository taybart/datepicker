#!/usr/bin/env zsh

set -e

DATE=$(go run .)
echo "$DATE is a pretty cool day"
