@ECHO OFF
go version
go get -u github.com/CharukaK/i18n4go/i18n4go
go install github.com/CharukaK/i18n4go/i18n4go

mkdir tmpdir

i18n4go -c extract-strings -e strings/exclude.json -o tmpdir -d ./ -r --ignore-regexp ".*test.go|i18n_resources.go"
i18n4go -c merge-strings -d tmpdir

i18n4go create-translations -v -f tmpdir/all.en.json --source-language en --languages "en_US" -o i18n/translations

