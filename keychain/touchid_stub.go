//go:build !darwin || !cgo

package keychain

// VerifyTouchID is a stub for non-darwin or non-cgo builds.
func VerifyTouchID() bool {
	return false
}
