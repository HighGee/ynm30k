#!/bin/bash
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct
if [ ! -f /code/ynm30k ]; then
    cd /code/ && go build -o ynm30k *.go
fi
cd /code/ && ./ynm30k