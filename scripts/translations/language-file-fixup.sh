#!/bin/bash
set -ex

go version
go get -u github.com/go-bindata/go-bindata/...
go install github.com/CharukaK/i18n4go/i18n4go

i18n4go fixup --ignore-regexp ".vscode/*|.github/*|cmd/*|strings/*"

