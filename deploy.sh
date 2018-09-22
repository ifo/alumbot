#!/bin/bash

mkdir deploy

cp remote-run.sh deploy/run.sh
env GOOS=linux GOARCH=amd64 go build -o deploy/alum-bot_linux_amd64
rsync -az ./deploy/ quine.space:~/alum-bot

rm deploy/alum-bot_linux_amd64
rm deploy/run.sh
rmdir deploy
