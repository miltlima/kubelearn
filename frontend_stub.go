//go:build !embed

package kubelearn

import "io/fs"

// Frontend reports that this binary was built without the web interface.
// Development builds serve only the API; the interface runs from the dev
// server ("make frontend-start").
func Frontend() (files fs.FS, ok bool) {
	return nil, false
}
