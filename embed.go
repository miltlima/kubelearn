// Package kubelearn carries the files that are compiled into the binary: the
// lab manifests and, in release builds, the web interface.
package kubelearn

import (
	"embed"
	"io/fs"
)

//go:embed manifests
var manifests embed.FS

// Manifests returns the lab manifests: the numbered files at the top level
// and the Gateway API extras under addons/.
func Manifests() fs.FS {
	sub, err := fs.Sub(manifests, "manifests")
	if err != nil {
		panic(err) // the directory is embedded above, so this cannot happen
	}
	return sub
}
