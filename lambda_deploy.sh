#!/bin/bash
set -exo pipefail

# mktemp で作業用のディレクトリを作成 (カレントディレクトリが汚れないようにするため, 不要なファイルが zip に入らないようにするため)
TEMPDIR=$(mktemp -d)
# 各自のバケット名に書き換え
ARTIFACT_BUCKET="kdg-aws-2025-neichi-haru-lambda-artifacts"
cd function
go mod download
go mod tidy
cd -

# function で使うバイナリをzipファイルに追加
cp function/* "$TEMPDIR"
cd "$TEMPDIR"
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go
zip "$TEMPDIR"/deployment-package.zip -r ./bootstrap
cd -

# S3にアップロードしてlambda関数の参照を書き換える
aws s3 cp "$TEMPDIR"/deployment-package.zip s3://$ARTIFACT_BUCKET/

aws lambda update-function-code \
  --no-cli-pager \
  --function-name first-function \
  --s3-bucket "$ARTIFACT_BUCKET" \
  --s3-key deployment-package.zip \
  --publish

# デプロイ時の一時ファイルを削除
rm -rf "$TEMPDIR"

set +x
echo "INFO: デプロイ成功"