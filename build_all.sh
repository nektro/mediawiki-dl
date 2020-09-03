#!/usr/bin/env bash

build_template() {
    export CGO_ENABLED=0
    export GOOS=$1
    export GOARCH=$2
    ext=$3
    date=$(date +'%Y.%m.%d')
    version=${CIRCLE_BUILD_NUM-$date}
    tag=v$version-$(git log --format=%h -1)
    echo $tag-$GOOS-$GOARCH
    go build -ldflags="-s -w" -o ./bin/SiteRipper-$tag-$GOOS-$GOARCH$ext
}

# From `go tool dist list`

build_template aix ppc64
build_template darwin 386
build_template darwin amd64
build_template dragonfly amd64
build_template freebsd 386
build_template freebsd amd64
build_template freebsd arm
build_template linux 386
build_template linux amd64
build_template linux arm
build_template linux arm64
build_template linux ppc64
build_template linux ppc64le
build_template linux mips
build_template linux mipsle
build_template linux mips64
build_template linux mips64le
build_template linux s390x
build_template netbsd 386
build_template netbsd amd64
build_template netbsd arm
build_template openbsd 386
build_template openbsd amd64
build_template openbsd arm
build_template plan9 386
build_template plan9 amd64
build_template plan9 arm
build_template solaris amd64
build_template windows 386 .exe
build_template windows amd64 .exe
build_template windows arm .exe