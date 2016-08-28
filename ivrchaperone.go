// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package openvr

/*
#include <stdio.h>
#include <stdlib.h>
#include "openvr_capi.h"

extern struct VR_IVRSystem_FnTable* _iSystem;



// _____   _   _  ______  _____ _
// |_   _|| | | || ___ \/  __ \| |
//  | |   | | | || |_/ /| /  \/| |__    __ _  _ __    ___  _ __   ___   _ __    ___
//  | |   | | | ||    / | |    | '_ \  / _` || '_ \  / _ \| '__| / _ \ | '_ \  / _ \
// _| |_  \ \_/ /| |\ \ | \__/\| | | || (_| || |_) ||  __/| |   | (_) || | | ||  __/
// \___/   \___/ \_| \_| \____/|_| |_| \__,_|| .__/  \___||_|    \___/ |_| |_| \___|
//                                           | |
//                                           |_|

int chaperone_GetCalibrationState(struct VR_IVRChaperone_FnTable* iChaperone) {
    return (int)iChaperone->GetCalibrationState();
}

void chaperone_GetPlayAreaSize(struct VR_IVRChaperone_FnTable* iChaperone, float* x, float* y) {
    iChaperone->GetPlayAreaSize(x, y);
}

void chaperone_GetPlayAreaRect(struct VR_IVRChaperone_FnTable* iChaperone, struct HmdQuad_t* rect) {
    iChaperone->GetPlayAreaRect(rect);
}


*/
import "C"

import (
	mgl "github.com/go-gl/mathgl/mgl32"
)

// Chaperone is an interface wrapper to IVRChaperone.
type Chaperone struct {
	ptr *C.struct_VR_IVRChaperone_FnTable
}

// GetCalibrationState returns a ChaperoneCalibrationState enumeration
// value indicating the current calibration state.
// Note: Tis can change at any time during a session.
func (chap *Chaperone) GetCalibrationState() int {
	result := int(C.chaperone_GetCalibrationState(chap.ptr))
	return result
}

// GetPlayAreaSize returns the width and depth of the play area.
func (chap *Chaperone) GetPlayAreaSize() (float32, float32) {
	var cx, cz C.float
	C.chaperone_GetPlayAreaSize(chap.ptr, &cx, &cz)
	return float32(cx), float32(cz)
}

// GetPlayAreaRect returns vectors for the 4 corners of the play area
// in a counter-clockwise order. (0,0,0) is the center of the play area.
// The height of every corner should be 0Y.
func (chap *Chaperone) GetPlayAreaRect() [4]mgl.Vec3 {
	var crekt C.struct_HmdQuad_t
	C.chaperone_GetPlayAreaRect(chap.ptr, &crekt)

	var result [4]mgl.Vec3
	result[0][0] = float32(crekt.vCorners[0].v[0])
	result[0][1] = float32(crekt.vCorners[0].v[1])
	result[0][2] = float32(crekt.vCorners[0].v[2])

	result[1][0] = float32(crekt.vCorners[1].v[0])
	result[1][1] = float32(crekt.vCorners[1].v[1])
	result[1][2] = float32(crekt.vCorners[1].v[2])

	result[2][0] = float32(crekt.vCorners[2].v[0])
	result[2][1] = float32(crekt.vCorners[2].v[1])
	result[2][2] = float32(crekt.vCorners[2].v[2])

	result[3][0] = float32(crekt.vCorners[3].v[0])
	result[3][1] = float32(crekt.vCorners[3].v[1])
	result[3][2] = float32(crekt.vCorners[3].v[2])
	return result
}

/*
  TODO:

	void (OPENVR_FNTABLE_CALLTYPE *ReloadInfo)();
	void (OPENVR_FNTABLE_CALLTYPE *SetSceneColor)(struct HmdColor_t color);
	void (OPENVR_FNTABLE_CALLTYPE *GetBoundsColor)(struct HmdColor_t * pOutputColorArray, int nNumOutputColors, float flCollisionBoundsFadeDistance, struct HmdColor_t * pOutputCameraColor);
	bool (OPENVR_FNTABLE_CALLTYPE *AreBoundsVisible)();
	void (OPENVR_FNTABLE_CALLTYPE *ForceBoundsVisible)(bool bForce);
*/
