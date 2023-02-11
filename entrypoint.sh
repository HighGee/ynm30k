#!/bin/bash
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct
cd /code/ && go build -o ynm30k *.go
cd /code/ && ./ynm30k