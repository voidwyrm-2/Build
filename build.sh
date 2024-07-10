#! /bin/sh

go build -o cbuild .
mv ./cbuild "$HOME/go/bin/cbuild"