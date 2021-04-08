#!/bin/sh

go tool cgo -godefs -- -I../../dkms $1 | sed 2d > $2
