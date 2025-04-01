#!/usr/bin/env zsh

set -e

DATE=$(go run . -o "Jan 02, 2006" -p "Pick a cool date" -s)
echo "$DATE is a pretty cool day"
