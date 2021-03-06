// +build windows

package openvr

/*
#include <stdio.h>
#include <stdlib.h>
#include "openvr_capi.h"
*/
import "C"

// convertCBool2Int is a bandaid over the way the bool types end up working out
// with go re: openvr_capi.h. On windows, openvr_capi.h will setup a typedef
// mapping bool to char; on any other platform stdbool.h is used which will boil
// down to using C._Bool in go.
//
// In order to cope with the different types, this function was made to return
// an integer version of the bool value. It is conditionally compiled based on
// platform so it shouldn't create a duplication error.
func convertCBool2Int(b C.bool) int {
	if b == C.bool(0) {
		return 0
	}

	return 1
}
