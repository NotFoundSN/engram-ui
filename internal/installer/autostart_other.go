//go:build !windows && !darwin && !linux

package installer

type otherAutostart struct{}

func (o *otherAutostart) Install(_ string) (Result, error) {
	return Result{Action: ActionUnsupportedPlatform}, ErrUnsupportedPlatform
}

func (o *otherAutostart) Remove() (Result, error) {
	return Result{Action: ActionUnsupportedPlatform}, ErrUnsupportedPlatform
}

// NewAutostartManager returns the unsupported-platform manager for all unrecognized OSes.
func NewAutostartManager() AutostartManager {
	return &otherAutostart{}
}
