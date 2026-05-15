//go:build !linux

package installer

import "os"

func evalSymlinksReal() (string, error) {
	return os.Executable()
}
