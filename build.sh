#!/usr/bin/env bash

go build -i -gcflags "all=-trimpath=$GOPATH/src/github.com/atcharles/lotto-chart -dwarf=false" \
 -asmflags "all=-trimpath=$GOPATH/src/github.com/atcharles/lotto-chart" \
 -ldflags=all="-w -s" -o ./bin/mac_chart

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
 go build -gcflags "all=-trimpath=$GOPATH/src/github.com/atcharles/lotto-chart -dwarf=false" \
 -asmflags "all=-trimpath=$GOPATH/src/github.com/atcharles/lotto-chart" \
 -ldflags=all="-w -s" -o ./bin/lotto-chart
# upx
upx -9 ./bin/lotto-chart

#./bin/mac_chart init --push --push_url "http://chart.0755yicai.cn/push_file"