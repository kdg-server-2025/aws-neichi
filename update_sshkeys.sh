#!/usr/bin/env bash
set -Ceux
set -o pipefail

# GitHub に登録した公開鍵をダウンロードする
GITHUB_USERNAME=neichi-haru
if wget --tries=50 --waitretry=10 --connect-timeout=10 -O authorized_keys https://github.com/$GITHUB_USERNAME.keys
then
  # SSH の公開鍵を配置する
  mkdir -p /home/ubuntu/.ssh/
  mv authorized_keys /home/ubuntu/.ssh/

  # sshd が使えるようにするために鍵のパーミッションを変更する
  chmod 700 /home/ubuntu/.ssh/
  chmod 600  /home/ubuntu/.ssh/authorized_keys
  chown -R ubuntu:ubuntu /home/ubuntu/.ssh/
  exit 0
else
  exit 1
fi