package keychain

/*
#cgo LDFLAGS: -framework LocalAuthentication -framework Foundation
#include "touchid_darwin.h"
*/
import "C"

func VerifyTouchID() bool {
	return bool(C.verify_touch_id())
}
