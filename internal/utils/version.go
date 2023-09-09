// TODO: get runtime build/git info  see:
// https://github.com/lightningnetwork/lnd/blob/master/build/version.go#L66
package utils

import "fmt"

const (
	// AppMajor defines the major version of this binary.
	AppMajor uint = 0

	// AppMinor defines the minor version of this binary.
	AppMinor uint = 1

	// AppPatch defines the application patch for this binary.
	AppPatch uint = 0
)

func Version() string {
	return fmt.Sprintf("%d.%d.%d", AppMajor, AppMinor, AppPatch)
}
