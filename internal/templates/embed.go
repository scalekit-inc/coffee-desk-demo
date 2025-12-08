package templates

import (
	"embed"
	"html/template"
	"io/fs"
)

//go:embed all:assets
var assets embed.FS

func Assets() (fs.FS, error) {
	return fs.Sub(assets, "assets")
}

//go:embed all:uploads
var uploads embed.FS

func Uploads() (fs.FS, error) {
	return fs.Sub(uploads, "uploads")
}

//go:embed robots.txt placeholder.svg
var rootFiles embed.FS

func RootFiles() (fs.FS, error) {
	return rootFiles, nil
}

//go:embed *.html
var templateFS embed.FS

var Templates = template.Must(template.ParseFS(templateFS, "*.html"))
