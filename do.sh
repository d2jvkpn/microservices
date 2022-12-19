#! /usr/bin/env bash
set -eu -o pipefail
_wd=$(pwd)
_path=$(dirname $0 | xargs -i readlink -f {})

mkdir microservices && cd microservices

git init
# touch .gitignore

git remote add origin git@github.com:d2jvkpn/microservices.git
git pull git@github.com:d2jvkpn/microservices.git master
# git push --set-upstream origin master
git branch --set-upstream-to=origin/master master
git branch -M master

git config user.name d2jvkpn
git config user.email chenbin2018@protonmail.com

git remote set-url --add origin git@gitlab.com:d2jvkpn/microservices.git

git add -A
git commit -m "init"
git push --set-upstream origin master
git push
