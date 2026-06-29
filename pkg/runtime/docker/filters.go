//go:build docker

package docker

import "github.com/docker/docker/api/types/filters"

// dockerFilters returns an empty filter set. We expose it as a helper so
// that the rest of the package never imports filters directly — keeping
// the SDK touch points minimal and easy to grep for.
func dockerFilters() filters.Args {
	return filters.NewArgs()
}
