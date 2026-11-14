//go:build embed

package kubelearn

import (
	"embed"
	"io/fs"
)

// The production build of the web interface. It only exists after
// "npm run build", so it is embedded only when the "embed" build tag is set
// (see "make build").
//
//go:embed all:kubelearn-frontend/build
var frontend embed.FS

// Frontend returns the web interface compiled into this binary.
func Frontend() (files fs.FS, ok bool) {
	sub, err := fs.Sub(frontend, "kubelearn-frontend/build")
	if err != nil {
		panic(err) // the directory is embedded above, so this cannot happen
	}
	return sub, true
}
