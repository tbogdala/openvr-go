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

void chaperone_GetPlayAreaSize(struct VR_IVRChaperone_FnTable* iChaperone, float* x, float* y) {
    iChaperone->GetPlayAreaSize(x, y);
}

*/
import "C"

// Chaperone is an interface wrapper to IVRChaperone.
type Chaperone struct {
	ptr *C.struct_VR_IVRChaperone_FnTable
}

// GetPlayAreaSize returns the width and depth of the play area.
func (chap *Chaperone) GetPlayAreaSize() (float32, float32) {
	var cx, cz C.float
	C.chaperone_GetPlayAreaSize(chap.ptr, &cx, &cz)
	return float32(cx), float32(cz)
}

/*
  TODO:

	ChaperoneCalibrationState (OPENVR_FNTABLE_CALLTYPE *GetCalibrationState)();
	bool (OPENVR_FNTABLE_CALLTYPE *GetPlayAreaSize)(float * pSizeX, float * pSizeZ);
	bool (OPENVR_FNTABLE_CALLTYPE *GetPlayAreaRect)(struct HmdQuad_t * rect);
	void (OPENVR_FNTABLE_CALLTYPE *ReloadInfo)();
	void (OPENVR_FNTABLE_CALLTYPE *SetSceneColor)(struct HmdColor_t color);
	void (OPENVR_FNTABLE_CALLTYPE *GetBoundsColor)(struct HmdColor_t * pOutputColorArray, int nNumOutputColors, float flCollisionBoundsFadeDistance, struct HmdColor_t * pOutputCameraColor);
	bool (OPENVR_FNTABLE_CALLTYPE *AreBoundsVisible)();
	void (OPENVR_FNTABLE_CALLTYPE *ForceBoundsVisible)(bool bForce);
*/
