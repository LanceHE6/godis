package main

import (
	"embed"
	"io/fs"
)

//go:embed web/dist
var webDistFS embed.FS

func webAssets() fs.FS {
	fsys, _ := fs.Sub(webDistFS, "web/dist")
	return fsys
}
