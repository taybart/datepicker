#!/usr/bin/env zsh

set -e

DATE=$(go run . -o "Jan 02, 2006" -p "Pick a cool date" )
echo "$DATE is a pretty cool day"
