package web

import (
	"embed"
	"io/fs"
)

//go:embed dist/*
var content embed.FS

func NewStatic() (fs.FS, error) {
	subFs, err := fs.Sub(&content, "dist")
	if err != nil {
		return nil, err
	}

	return subFs, nil
}
