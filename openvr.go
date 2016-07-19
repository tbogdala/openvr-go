// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.


package openvr

/*
#cgo CPPFLAGS: -I${SRCDIR}/vendor/openvr/headers -std=c99
#cgo windows,386 LDFLAGS: -L${SRCDIR}/vendor/openvr/bin/win32 -lopenvr_api
#cgo windows,amd64 LDFLAGS: -L${SRCDIR}/vendor/openvr/bin/win64 -lopenvr_api

#include <stdio.h>
#include <stdlib.h>
#include "openvr_capi.h"

// lets declare some externals from the openvr_api.dll
__declspec( dllimport ) intptr_t VR_InitInternal( EVRInitError *peError, EVRApplicationType eType );
__declspec( dllimport ) const char * VR_GetVRInitErrorAsEnglishDescription( EVRInitError error );
__declspec( dllimport ) bool VR_IsInterfaceVersionValid(const char *interface);
__declspec( dllimport ) void VR_ShutdownInternal();
__declspec( dllimport ) intptr_t VR_GetGenericInterface( const char *pchInterfaceVersion, EVRInitError *peError );

// api tokens; set in initInternal()
intptr_t _iToken;
struct VR_IVRSystem_FnTable* _iSystem;
struct VR_IVRCompositor_FnTable* _iCompositor;
struct VR_IVRRenderModels_FnTable* _iRenderModels;

// gets the api token and makes sure the interface is valid
int initInternal(int appTypeEnum) {
    // get the api token
    EVRInitError error = EVRInitError_VRInitError_None;
    _iToken = VR_InitInternal(&error, appTypeEnum);
    if (error != EVRInitError_VRInitError_None) {
        const char* msg = VR_GetVRInitErrorAsEnglishDescription(error);
        printf("VR_InitInternal failed: %s\n", msg);
        return error;
    }

    bool icheck = VR_IsInterfaceVersionValid(IVRSystem_Version);
    if (!icheck) {
        printf("INVALID interface\n");
        VR_ShutdownInternal();
        return EVRInitError_VRInitError_Unknown;
    }

    char interfaceFnTable[256];
    sprintf(interfaceFnTable, "FnTable:%s", IVRSystem_Version);
    _iSystem = (struct VR_IVRSystem_FnTable*) VR_GetGenericInterface(interfaceFnTable, &error);
    if (error != EVRInitError_VRInitError_None) {
        const char* msg = VR_GetVRInitErrorAsEnglishDescription(error);
        printf("Error on getting IVRSystem: %s\n", msg);
        return error;
    }

    return EVRInitError_VRInitError_None;
}


int compositor_SetInternalInterface() {
    EVRInitError error = EVRInitError_VRInitError_None;
    if (_iCompositor == NULL) {
        char interfaceFnTable[256];
        sprintf(interfaceFnTable, "FnTable:%s", IVRCompositor_Version);
        _iCompositor = (struct VR_IVRCompositor_FnTable*) VR_GetGenericInterface(interfaceFnTable, &error);
        if (error != EVRInitError_VRInitError_None) {
            const char* msg = VR_GetVRInitErrorAsEnglishDescription(error);
            printf("Error on getting IVRCompositor: %s\n", msg);
            return error;
        }
    }
    return error;
}

int rendermodels_SetInternalInterface() {
    EVRInitError error = EVRInitError_VRInitError_None;
    if (_iRenderModels == NULL) {
        char interfaceFnTable[256];
        sprintf(interfaceFnTable, "FnTable:%s", IVRRenderModels_Version);
        _iRenderModels = (struct VR_IVRRenderModels_FnTable*) VR_GetGenericInterface(interfaceFnTable, &error);
        if (error != EVRInitError_VRInitError_None) {
            const char* msg = VR_GetVRInitErrorAsEnglishDescription(error);
            printf("Error on getting IVRRenderModels: %s\n", msg);
            return error;
        }
    }
    return error;
}

*/
import "C"
import (
    "fmt"
)

// Mat4 is a 4x4 matrix in column-major order
type Mat4 [16]float32

// Mat34 is a 3x4 matrix in column-major order
type Mat34 [12]float32

// Vec2 is a 2 diminensional vector of floats
type Vec2 [2]float32

// Vec3 is a 3 diminensional vector of floats
type Vec3 [3]float32

// Vec4 is a 4 diminensional vector of floats
type Vec4 [4]float32

// Init initializes the internal VR api structers and on success will
// return a System object with a valid IVRSystem interface ptr.
func Init() (*System, error) {
    // initialize the module _iToken value from the openvr api
    e := C.initInternal(C.EVRApplicationType_VRApplication_Scene)
    if e == C.EVRInitError_VRInitError_None {
        sys := new(System)
        sys.ptr = C._iSystem
        return sys, nil
    }

    errStr := GetErrorAsEnglish(int(e))
    return nil, fmt.Errorf("%s", errStr)
}

// Shutdown calls the ShutdownInternal function on the VR library.
func Shutdown() {
    C.VR_ShutdownInternal();
}

// GetErrorAsEnglish takes an EVRInitError enumeration value and returns a string.
func GetErrorAsEnglish(e int) string {
    cs := C.VR_GetVRInitErrorAsEnglishDescription(C.EVRInitError(e))
    // NOTE: does cs need to be freed somehow?
    return C.GoString(cs)
}

// GetCompositor returns a new IVRCompositor interface.
func GetCompositor() (*Compositor, error) {
    e := C.compositor_SetInternalInterface()
    if e == C.EVRInitError_VRInitError_None {
        comp := new(Compositor)
        comp.ptr = C._iCompositor
        return comp, nil
    }
    cs := C.VR_GetVRInitErrorAsEnglishDescription(C.EVRInitError(e))
    return nil, fmt.Errorf("%s", C.GoString(cs))
}

// GetRenderModels returns a new IVRRenderModels interface.
func GetRenderModels() (*RenderModels, error) {
    e := C.rendermodels_SetInternalInterface()
    if e == C.EVRInitError_VRInitError_None {
        rm := new(RenderModels)
        rm.ptr = C._iRenderModels
        return rm, nil
    }
    cs := C.VR_GetVRInitErrorAsEnglishDescription(C.EVRInitError(e))
    return nil, fmt.Errorf("%s", C.GoString(cs))
}
