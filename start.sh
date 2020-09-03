#!/usr/bin/env bash

set -e
set -x

go test
go build
./mediawiki-dl \
--site "https://wiki.osdev.org/index.php" \
