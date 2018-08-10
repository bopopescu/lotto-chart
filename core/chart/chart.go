package chart

import (
	"github.com/atcharles/gof/gofutils"
	"github.com/elazarl/go-bindata-assetfs"
)

const (
	Version    = "v0.0.1"
	ServerName = "lotto-chart"
)

var (
	RootDir = gofutils.SelfDir() + "/"
)

//go:generate go-bindata-assetfs -pkg chart mysql_files/...
func AssetFS() *assetfs.AssetFS {
	return assetFS()
}
