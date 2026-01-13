// Package web provides embedded static assets for the web frontend.
package web

import "embed"

//go:embed dist/*
var DistFS embed.FS
