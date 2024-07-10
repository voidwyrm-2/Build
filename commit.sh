#! /bin/sh

if [ $# -ne 1 ]; then
    echo "expected '[sh ./commit.sh | ./commit.sh] [commit message]'"
    exit 1
fi

git add .
git commit . -m "$1"
git push