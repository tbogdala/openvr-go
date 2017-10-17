// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package openvr

/*
#cgo CFLAGS: -I${SRCDIR}/vendored/openvr/headers -std=c99
#cgo windows,386 LDFLAGS: -L${SRCDIR}/vendored/openvr/bin/win32 -lopenvr_api
#cgo windows,amd64 LDFLAGS: -L${SRCDIR}/vendored/openvr/bin/win64 -lopenvr_api
#cgo linux,amd64 LDFLAGS: -L${SRCDIR}/vendored/openvr/bin/linux64 -lopenvr_api
#cgo linux,386 LDFLAGS: -L${SRCDIR}/vendored/openvr/bin/linux32 -lopenvr_api

#include <stdio.h>
#include <stdlib.h>
#include "openvr_capi.h"

#if defined(_WIN32)
    #define IMPORT __declspec(dllimport)
#else
    #define IMPORT
#endif

// lets declare some externals from the openvr_api.dll
IMPORT intptr_t VR_InitInternal( EVRInitError *peError, EVRApplicationType eType );
IMPORT const char * VR_GetVRInitErrorAsEnglishDescription( EVRInitError error );
IMPORT bool VR_IsInterfaceVersionValid(const char *interface);
IMPORT void VR_ShutdownInternal();
IMPORT intptr_t VR_GetGenericInterface( const char *pchInterfaceVersion, EVRInitError *peError );

// api tokens; set in initInternal()
intptr_t _iToken;
struct VR_IVRSystem_FnTable* _iSystem;
struct VR_IVRCompositor_FnTable* _iCompositor;
struct VR_IVRRenderModels_FnTable* _iRenderModels;
struct VR_IVRChaperone_FnTable* _iChaperone;


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

int chaperone_SetInternalInterface() {
    EVRInitError error = EVRInitError_VRInitError_None;
    if (_iChaperone == NULL) {
        char interfaceFnTable[256];
        sprintf(interfaceFnTable, "FnTable:%s", IVRChaperone_Version);
        _iChaperone = (struct VR_IVRChaperone_FnTable*) VR_GetGenericInterface(interfaceFnTable, &error);
        if (error != EVRInitError_VRInitError_None) {
            const char* msg = VR_GetVRInitErrorAsEnglishDescription(error);
            printf("Error on getting IVRChaperone: %s\n", msg);
            return error;
        }
    }
    return error;
}

*/
import "C"
import (
	"fmt"

	mgl "github.com/go-gl/mathgl/mgl32"
)

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
	C.VR_ShutdownInternal()
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

// GetChaperone returns a new IVRChaperone interface.
func GetChaperone() (*Chaperone, error) {
	e := C.chaperone_SetInternalInterface()
	if e == C.EVRInitError_VRInitError_None {
		chap := new(Chaperone)
		chap.ptr = C._iChaperone
		return chap, nil
	}
	cs := C.VR_GetVRInitErrorAsEnglishDescription(C.EVRInitError(e))
	return nil, fmt.Errorf("%s", C.GoString(cs))
}

// Mat34ToMat4 is a utility conversion function that takes a 3x4 matrix and outputs
// a 4x4 matrix with an identity fourth row of {0,0,0,1}.
func Mat34ToMat4(vrM34 *mgl.Mat3x4) (m4 mgl.Mat4) {
	m4[0] = vrM34[0]
	m4[1] = vrM34[1]
	m4[2] = vrM34[2]
	m4[3] = 0.0

	m4[4] = vrM34[3]
	m4[5] = vrM34[4]
	m4[6] = vrM34[5]
	m4[7] = 0.0

	m4[8] = vrM34[6]
	m4[9] = vrM34[7]
	m4[10] = vrM34[8]
	m4[11] = 0.0

	m4[12] = vrM34[9]
	m4[13] = vrM34[10]
	m4[14] = vrM34[11]
	m4[15] = 1.0
	return m4
}

var (
	// ShaderRenderModelV is the render model vertex shader
	ShaderRenderModelV = `#version 330
	uniform mat4 mvp;
	in vec4 position;
	in vec3 normal;
	in vec2 texCoord;

	out vec2 v2TexCoord;
	void main()
	{
	   v2TexCoord = texCoord;
	   gl_Position = mvp * vec4(position.xyz, 1);
	}`

	// ShaderRenderModelF is the render model fragment shader
	ShaderRenderModelF = `#version 330
	uniform sampler2D diffuse;
	in vec2 v2TexCoord;
	out vec4 outputColor;
	void main()
	{
		outputColor = texture(diffuse, v2TexCoord);
	}`

	// ShaderLensDistortionV is the lens distortion vertex shader
	ShaderLensDistortionV = `#version 330
        in vec4 position;
        in vec2 v2UVredIn;
        in vec2 v2UVGreenIn;
        in vec2 v2UVblueIn;
        noperspective  out vec2 v2UVred;
        noperspective  out vec2 v2UVgreen;
        noperspective  out vec2 v2UVblue;
        void main()
        {
        	v2UVred = v2UVredIn;
        	v2UVgreen = v2UVGreenIn;
        	v2UVblue = v2UVblueIn;
        	gl_Position = position;
        }`

	// ShaderLensDistortionF is the lens distortion fragment shader
	ShaderLensDistortionF = `#version 330
        uniform sampler2D mytexture;

        noperspective  in vec2 v2UVred;
        noperspective  in vec2 v2UVgreen;
        noperspective  in vec2 v2UVblue;

        out vec4 outputColor;

        void main()
        {
            float fBoundsCheck = ( (dot( vec2( lessThan( v2UVgreen.xy, vec2(0.05, 0.05)) ), vec2(1.0, 1.0))+dot( vec2( greaterThan( v2UVgreen.xy, vec2( 0.95, 0.95)) ), vec2(1.0, 1.0))) );
        	if( fBoundsCheck > 1.0 ) {
                outputColor = vec4( 0, 0, 1, 1.0 );
            } else {
        		float red = texture(mytexture, v2UVred).x;
        		float green = texture(mytexture, v2UVgreen).y;
        		float blue = texture(mytexture, v2UVblue).z;
        		outputColor = vec4( red, green, blue, 1.0  );
            }

        }`
)
