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

extern struct VR_IVRSystem_FnTable* _iSystem;

//   _____  _    _  ______   ______                     _                ______              _         _
//  (_____)| |  | |(_____ \ (_____ \                   | |              |  ___ \            | |       | |
//    _   | |  | | _____) ) _____) )  ____  ____    _ | |  ____   ____ | | _ | |  ___    _ | |  ____ | |  ___
//   | |   \ \/ / (_____ ( (_____ (  / _  )|  _ \  / || | / _  ) / ___)| || || | / _ \  / || | / _  )| | /___)
//  _| |_   \  /        | |      | |( (/ / | | | |( (_| |( (/ / | |    | || || || |_| |( (_| |( (/ / | ||___ |
// (_____)   \/         |_|      |_| \____)|_| |_| \____| \____)|_|    |_||_||_| \___/  \____| \____)|_|(___/

EVRRenderModelError rendermodels_LoadRenderModel_Async(struct VR_IVRRenderModels_FnTable* iRenderModels, char * pchRenderModelName, struct RenderModel_t ** ppRenderModel) {
    return iRenderModels->LoadRenderModel_Async(pchRenderModelName, ppRenderModel);
}

EVRRenderModelError rendermodels_LoadTexture_Async(struct VR_IVRRenderModels_FnTable* iRenderModels, TextureID_t textureId, struct RenderModel_TextureMap_t ** ppTexture) {
    return iRenderModels->LoadTexture_Async(textureId, ppTexture);
}

void rendermodels_FreeRenderModel(struct VR_IVRRenderModels_FnTable* iRenderModels, struct RenderModel_t * pRenderModel) {
    iRenderModels->FreeRenderModel(pRenderModel);
}

void rendermodels_FreeTexture(struct VR_IVRRenderModels_FnTable* iRenderModels, struct RenderModel_TextureMap_t * pTexture) {
    iRenderModels->FreeTexture(pTexture);
}

void flatenRenderModelsTextureData(const struct RenderModel_TextureMap_t* textureData, uint8_t* dest) {
    int max = textureData->unWidth * textureData->unHeight * 4;
    for (int i=0; i<max; i++) {
        dest[i] = textureData->rubTextureMapData[i];
    }
}

// loop through the renderModel's vertex data and copy it to a flat float array for Go to use.
void flatenRenderModelsVertexData(const struct RenderModel_t * renderModel, float* dest, unsigned int* destIndexes) {
    const int fsize = 8;
    for (int i=0; i<renderModel->unVertexCount; i++) {
        const struct RenderModel_Vertex_t *vert = &renderModel->rVertexData[i];
        dest[i*fsize] = vert->vPosition.v[0];
        dest[i*fsize+1] = vert->vPosition.v[1];
        dest[i*fsize+2] = vert->vPosition.v[2];

        dest[i*fsize+3] = vert->vNormal.v[0];
        dest[i*fsize+4] = vert->vNormal.v[1];
        dest[i*fsize+5] = vert->vNormal.v[2];

        dest[i*fsize+6] = vert->rfTextureCoord[0];
        dest[i*fsize+7] = vert->rfTextureCoord[1];
    }

    for (int i=0; i<renderModel->unTriangleCount * 3; i++) {
        destIndexes[i] = renderModel->rIndexData[i];
    }
}

*/
import "C"

import (
	"fmt"
	"runtime"
	"unsafe"
)

/*
type RenderModelTexture struct {
    Width uint16
    Height uint16
    Data []byte
}
*/

type RenderModel struct {
	VertexData    []float32
	Indexes       []uint32
	TriangleCount uint32

	TextureWidth  uint32
	TextureHeight uint32
	TextureBytes  []byte
}

func newRenderModel() *RenderModel {
	model := new(RenderModel)
	return model
}

// RenderModels is an interface wrapper to IVRRenderModels.
type RenderModels struct {
	ptr *C.struct_VR_IVRRenderModels_FnTable
}

// RenderModelLoad syncrhonously loads the model.
func (rm *RenderModels) RenderModelLoad(name string) (*RenderModel, error) {
	var cModel *C.struct_RenderModel_t
	var result C.EVRRenderModelError
	csName := C.CString(name)
	for {
		result = C.rendermodels_LoadRenderModel_Async(rm.ptr, csName, &cModel)
		if result != VRRenderModelErrorLoading {
			break
		}
		runtime.Gosched()
	}
	C.free(unsafe.Pointer(csName))

	// we now have the model, right?
	if result != VRRenderModelErrorNone {
		return nil, fmt.Errorf("Failed to load render model for %s: %s", name, GetErrorAsEnglish(int(result)))
	}

	var cTexture *C.struct_RenderModel_TextureMap_t
	for {
		result = C.rendermodels_LoadTexture_Async(rm.ptr, cModel.diffuseTextureId, &cTexture)
		if result != VRRenderModelErrorLoading {
			break
		}
		runtime.Gosched()
	}

	// we now have the texture, right?
	if result != VRRenderModelErrorNone {
		return nil, fmt.Errorf("Failed to load render model texture for %s: %s", name, GetErrorAsEnglish(int(result)))
	}

	// create the render model with the data from the C structures
	model := newRenderModel()
	model.VertexData = make([]float32, int(cModel.unVertexCount*8))
	model.Indexes = make([]uint32, int(cModel.unTriangleCount)*3)
	model.TriangleCount = uint32(cModel.unTriangleCount)
	C.flatenRenderModelsVertexData(cModel, (*C.float)(unsafe.Pointer(&model.VertexData[0])), (*C.uint)(unsafe.Pointer(&model.Indexes[0])))

	model.TextureWidth = uint32(cTexture.unWidth)
	model.TextureHeight = uint32(cTexture.unHeight)
	model.TextureBytes = make([]byte, model.TextureWidth*model.TextureHeight*4)
	C.flatenRenderModelsTextureData(cTexture, (*C.uint8_t)(unsafe.Pointer(&model.TextureBytes[0])))

	C.rendermodels_FreeRenderModel(rm.ptr, cModel)
	C.rendermodels_FreeTexture(rm.ptr, cTexture)

	return model, nil
}

/*
TODO:

EVRRenderModelError (OPENVR_FNTABLE_CALLTYPE *LoadTextureD3D11_Async)(TextureID_t textureId, void * pD3D11Device, void ** ppD3D11Texture2D);
EVRRenderModelError (OPENVR_FNTABLE_CALLTYPE *LoadIntoTextureD3D11_Async)(TextureID_t textureId, void * pDstTexture);
void (OPENVR_FNTABLE_CALLTYPE *FreeTextureD3D11)(void * pD3D11Texture2D);
uint32_t (OPENVR_FNTABLE_CALLTYPE *GetRenderModelName)(uint32_t unRenderModelIndex, char * pchRenderModelName, uint32_t unRenderModelNameLen);
uint32_t (OPENVR_FNTABLE_CALLTYPE *GetRenderModelCount)();
uint32_t (OPENVR_FNTABLE_CALLTYPE *GetComponentCount)(char * pchRenderModelName);
uint32_t (OPENVR_FNTABLE_CALLTYPE *GetComponentName)(char * pchRenderModelName, uint32_t unComponentIndex, char * pchComponentName, uint32_t unComponentNameLen);
uint64_t (OPENVR_FNTABLE_CALLTYPE *GetComponentButtonMask)(char * pchRenderModelName, char * pchComponentName);
uint32_t (OPENVR_FNTABLE_CALLTYPE *GetComponentRenderModelName)(char * pchRenderModelName, char * pchComponentName, char * pchComponentRenderModelName, uint32_t unComponentRenderModelNameLen);
bool (OPENVR_FNTABLE_CALLTYPE *GetComponentState)(char * pchRenderModelName, char * pchComponentName, VRControllerState_t * pControllerState, struct RenderModel_ControllerMode_State_t * pState, struct RenderModel_ComponentState_t * pComponentState);
bool (OPENVR_FNTABLE_CALLTYPE *RenderModelHasComponent)(char * pchRenderModelName, char * pchComponentName);
uint32_t (OPENVR_FNTABLE_CALLTYPE *GetRenderModelThumbnailURL)(char * pchRenderModelName, char * pchThumbnailURL, uint32_t unThumbnailURLLen, EVRRenderModelError * peError);
uint32_t (OPENVR_FNTABLE_CALLTYPE *GetRenderModelOriginalPath)(char * pchRenderModelName, char * pchOriginalPath, uint32_t unOriginalPathLen, EVRRenderModelError * peError);
char * (OPENVR_FNTABLE_CALLTYPE *GetRenderModelErrorNameFromEnum)(EVRRenderModelError error);

*/
