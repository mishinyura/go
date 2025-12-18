package ui

import "embed"

//go:embed redoc.html
var assets embed.FS

func ReDocHTML() ([]byte, error) {
	return assets.ReadFile("redoc.html")
}
