package controller

import (
	"path/filepath"
)

func (f *file) fullName() string {
	path := f.parent.path()
	path = append(path, f.name)
	return filepath.Join(path...)
}
