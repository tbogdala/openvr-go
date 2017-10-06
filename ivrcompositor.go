// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package openvr

/*
#include <stdio.h>
#include <stdlib.h>
#include "openvr_capi.h"

extern struct VR_IVRSystem_FnTable* _iSystem;


// .___ ____   ______________ _________                                           .__   __
// |   |\   \ /   /\______   \\_   ___ \   ____    _____  ______    ____    ______|__|_/  |_   ____  _______
// |   | \   Y   /  |       _//    \  \/  /  _ \  /     \ \____ \  /  _ \  /  ___/|  |\   __\ /  _ \ \_  __ \
// |   |  \     /   |    |   \\     \____(  <_> )|  Y Y  \|  |_> >(  <_> ) \___ \ |  | |  |  (  <_> ) |  | \/
// |___|   \___/    |____|_  / \______  / \____/ |__|_|  /|   __/  \____/ /____  >|__| |__|   \____/  |__|
//                         \/         \/               \/ |__|                 \/



EVRCompositorError compositor_WaitGetPoses(struct VR_IVRCompositor_FnTable* iCompositor, struct TrackedDevicePose_t * pRenderPoseArray, uint32_t unRenderPoseArrayCount, struct TrackedDevicePose_t * pGamePoseArray, uint32_t unGamePoseArrayCount) {
    return iCompositor->WaitGetPoses(pRenderPoseArray, unRenderPoseArrayCount, pGamePoseArray, unGamePoseArrayCount);
}



EVRCompositorError compositor_Submit(struct VR_IVRCompositor_FnTable* iCompositor, EVREye eEye, struct Texture_t * pTexture, struct VRTextureBounds_t * pBounds, EVRSubmitFlags nSubmitFlags) {
    return iCompositor->Submit(eEye, pTexture, pBounds, nSubmitFlags);
}

EVRCompositorError compositor_SubmitSimple(struct VR_IVRCompositor_FnTable* iCompositor, EVREye eEye, intptr_t texture) {
    struct Texture_t tex;
    tex.handle = (void*) texture;
    tex.eType = ETextureType_TextureType_OpenGL;
    tex.eColorSpace = EColorSpace_ColorSpace_Gamma;
    return iCompositor->Submit(eEye, &tex, 0, EVRSubmitFlags_Submit_Default);
}

float compositor_GetFrameTimeRemaining(struct VR_IVRCompositor_FnTable* iCompositor) {
	return iCompositor->GetFrameTimeRemaining();
}

bool compositor_GetFrameTiming(struct VR_IVRCompositor_FnTable* iCompositor, struct Compositor_FrameTiming * pTiming, uint32_t unFramesAgo) {
	pTiming->m_nSize = sizeof(struct Compositor_FrameTiming);
	return iCompositor->GetFrameTiming(pTiming, unFramesAgo);
}


*/
import "C"

import (
	"bytes"
	"fmt"

	mgl "github.com/go-gl/mathgl/mgl32"
)

// TrackedDevicePose mirrors the OpenVR TrackedDevicePose_t structure.
type TrackedDevicePose struct {
	DeviceToAbsoluteTracking mgl.Mat3x4
	Velocity                 mgl.Vec3 // velocity in tracker space in m/s
	AngularVelocity          mgl.Vec3 // in radians/s
	TrackingResult           int      // ETrackingResult enum value
	PoseIsValid              bool

	// This indicates that there is a device connected for this spot in the pose array.
	// It could go from true to false if the user unplugs the device.
	DeviceIsConnected bool
}

// Compositor is an interface wrapper to IVRCompositor.
type Compositor struct {
	ptr *C.struct_VR_IVRCompositor_FnTable

	renderPoseArray [MaxTrackedDeviceCount]C.struct_TrackedDevicePose_t
	gamePoseArray   [MaxTrackedDeviceCount]C.struct_TrackedDevicePose_t
}

// WaitGetPoses updates the internal copy of pose(s) to use to render scene (and optionally poses predicted two frames out for gameplay).
func (comp *Compositor) WaitGetPoses(getPredictions bool) {
	if getPredictions {
		C.compositor_WaitGetPoses(comp.ptr, &comp.renderPoseArray[0], C.uint32_t(MaxTrackedDeviceCount), &comp.gamePoseArray[0], C.uint32_t(MaxTrackedDeviceCount))
	} else {
		C.compositor_WaitGetPoses(comp.ptr, &comp.renderPoseArray[0], C.uint32_t(MaxTrackedDeviceCount), nil, 0)
	}
}

// Submit updates scene texture to display.
func (comp *Compositor) Submit(eye int, texture uint32) {
	C.compositor_SubmitSimple(comp.ptr, C.EVREye(eye), C.intptr_t(texture))
}

// IsPoseValid returns true if a render pose array at the given index has a valid pose.
func (comp *Compositor) IsPoseValid(i uint) bool {
	if comp.renderPoseArray[i].bPoseIsValid != false {
		return true
	}
	return false
}

// GetFrameTimeRemaining returns the time in seconds left in the current (as identified by
// FrameTiming's frameIndex) frame.  Due to "running start", this value may roll over
// to the next frame before ever reaching 0.0.
func (comp *Compositor) GetFrameTimeRemaining() float32 {
	return float32(C.compositor_GetFrameTimeRemaining(comp.ptr))
}

// GetRenderPose gets the render pose for a device at the given index.
func (comp *Compositor) GetRenderPose(i uint) (tdp TrackedDevicePose) {
	cTDP := comp.renderPoseArray[i]
	fillTrackedDevicePose(&tdp, &cTDP)
	return tdp
}

func fillTrackedDevicePose(tdp *TrackedDevicePose, cTDP *C.struct_TrackedDevicePose_t) {
	tdp.DeviceToAbsoluteTracking[0] = float32(cTDP.mDeviceToAbsoluteTracking.m[0][0])
	tdp.DeviceToAbsoluteTracking[3] = float32(cTDP.mDeviceToAbsoluteTracking.m[0][1])
	tdp.DeviceToAbsoluteTracking[6] = float32(cTDP.mDeviceToAbsoluteTracking.m[0][2])
	tdp.DeviceToAbsoluteTracking[9] = float32(cTDP.mDeviceToAbsoluteTracking.m[0][3])

	tdp.DeviceToAbsoluteTracking[1] = float32(cTDP.mDeviceToAbsoluteTracking.m[1][0])
	tdp.DeviceToAbsoluteTracking[4] = float32(cTDP.mDeviceToAbsoluteTracking.m[1][1])
	tdp.DeviceToAbsoluteTracking[7] = float32(cTDP.mDeviceToAbsoluteTracking.m[1][2])
	tdp.DeviceToAbsoluteTracking[10] = float32(cTDP.mDeviceToAbsoluteTracking.m[1][3])

	tdp.DeviceToAbsoluteTracking[2] = float32(cTDP.mDeviceToAbsoluteTracking.m[2][0])
	tdp.DeviceToAbsoluteTracking[5] = float32(cTDP.mDeviceToAbsoluteTracking.m[2][1])
	tdp.DeviceToAbsoluteTracking[8] = float32(cTDP.mDeviceToAbsoluteTracking.m[2][2])
	tdp.DeviceToAbsoluteTracking[11] = float32(cTDP.mDeviceToAbsoluteTracking.m[2][3])

	tdp.Velocity[0] = float32(cTDP.vVelocity.v[0])
	tdp.Velocity[1] = float32(cTDP.vVelocity.v[1])
	tdp.Velocity[2] = float32(cTDP.vVelocity.v[2])

	tdp.AngularVelocity[0] = float32(cTDP.vAngularVelocity.v[0])
	tdp.AngularVelocity[1] = float32(cTDP.vAngularVelocity.v[1])
	tdp.AngularVelocity[2] = float32(cTDP.vAngularVelocity.v[2])

	tdp.TrackingResult = int(cTDP.eTrackingResult)

	if cTDP.bPoseIsValid != false {
		tdp.PoseIsValid = true
	}

	if cTDP.bDeviceIsConnected != false {
		tdp.DeviceIsConnected = true
	}
}

// FrameTiming provides a single frame's timing information to the app.
type FrameTiming struct {
	FrameIndex              uint32
	NumFramePresents        uint32
	NumMisPresented         uint32
	NumDroppedFrames        uint32
	ReprojectionFlags       uint32
	SystemTimeInSeconds     float64
	PreSubmitGpuMs          float32
	PostSubmitGpuMs         float32
	TotalRenderGpuMs        float32
	CompositorRenderGpuMs   float32
	CompositorRenderCpuMs   float32
	CompositorIdleCpuMs     float32
	ClientFrameIntervalMs   float32
	PresentCallCpuMs        float32
	WaitForPresentCpuMs     float32
	SubmitFrameMs           float32
	WaitGetPosesCalledMs    float32
	NewPosesReadyMs         float32
	NewFrameReadyMs         float32
	CompositorUpdateStartMs float32
	CompositorUpdateEndMs   float32
	CompositorRenderStartMs float32
	HmdPose                 TrackedDevicePose
}

// GetFrameTiming teturns true if timing data is filled it.  Sets oldest timing info if framesAgo
// is larger than the stored history.
func (comp *Compositor) GetFrameTiming(timing *FrameTiming, framesAgo uint32) bool {
	var cTimingData C.struct_Compositor_FrameTiming
	cRet := C.compositor_GetFrameTiming(comp.ptr, &cTimingData, C.uint32_t(framesAgo))

	timing.FrameIndex = uint32(cTimingData.m_nFrameIndex)
	timing.NumFramePresents = uint32(cTimingData.m_nNumFramePresents)
	timing.NumMisPresented = uint32(cTimingData.m_nNumMisPresented)
	timing.NumDroppedFrames = uint32(cTimingData.m_nNumDroppedFrames)
	timing.ReprojectionFlags = uint32(cTimingData.m_nReprojectionFlags)
	timing.SystemTimeInSeconds = float64(cTimingData.m_flSystemTimeInSeconds)
	timing.PreSubmitGpuMs = float32(cTimingData.m_flPreSubmitGpuMs)
	timing.PostSubmitGpuMs = float32(cTimingData.m_flPostSubmitGpuMs)
	timing.TotalRenderGpuMs = float32(cTimingData.m_flTotalRenderGpuMs)
	timing.CompositorRenderGpuMs = float32(cTimingData.m_flCompositorRenderGpuMs)
	timing.CompositorRenderCpuMs = float32(cTimingData.m_flCompositorRenderCpuMs)
	timing.CompositorIdleCpuMs = float32(cTimingData.m_flCompositorIdleCpuMs)
	timing.ClientFrameIntervalMs = float32(cTimingData.m_flClientFrameIntervalMs)
	timing.PresentCallCpuMs = float32(cTimingData.m_flPresentCallCpuMs)
	timing.WaitForPresentCpuMs = float32(cTimingData.m_flWaitForPresentCpuMs)
	timing.SubmitFrameMs = float32(cTimingData.m_flSubmitFrameMs)
	timing.WaitGetPosesCalledMs = float32(cTimingData.m_flWaitGetPosesCalledMs)
	timing.NewPosesReadyMs = float32(cTimingData.m_flNewPosesReadyMs)
	timing.NewFrameReadyMs = float32(cTimingData.m_flNewFrameReadyMs)
	timing.CompositorUpdateStartMs = float32(cTimingData.m_flCompositorUpdateStartMs)
	timing.CompositorUpdateEndMs = float32(cTimingData.m_flCompositorUpdateEndMs)
	timing.CompositorRenderStartMs = float32(cTimingData.m_flCompositorRenderStartMs)

	fillTrackedDevicePose(&timing.HmdPose, &cTimingData.m_HmdPose)

	if cRet == false {
		return false
	}

	return true
}

// Show returns a formatted string with the timing information. If newlines
// is true, then each field will be written on its own line in the string.
func (ft *FrameTiming) Show(newlines bool) string {
	var b bytes.Buffer
	nl := " "
	if newlines {
		nl = "\n"
	}

	b.WriteString(fmt.Sprintf("FrameIndex: %v%s", ft.FrameIndex, nl))
	b.WriteString(fmt.Sprintf("NumFramePresent: %v%s", ft.NumFramePresents, nl))
	b.WriteString(fmt.Sprintf("NumMisPresented: %v%s", ft.NumMisPresented, nl))
	b.WriteString(fmt.Sprintf("NumDroppedFrames: %v%s", ft.NumDroppedFrames, nl))
	b.WriteString(fmt.Sprintf("ReprojectionFlags: %v%s", ft.ReprojectionFlags, nl))
	b.WriteString(fmt.Sprintf("SystemTimeInSeconds: %v%s", ft.SystemTimeInSeconds, nl))
	b.WriteString(fmt.Sprintf("PreSubmitGpuMs: %v%s", ft.PreSubmitGpuMs, nl))
	b.WriteString(fmt.Sprintf("PostSubmitGpuMs: %v%s", ft.PostSubmitGpuMs, nl))
	b.WriteString(fmt.Sprintf("TotalRenderGpuMs: %v%s", ft.TotalRenderGpuMs, nl))
	b.WriteString(fmt.Sprintf("CompositorRenderGpuMs: %v%s", ft.CompositorRenderGpuMs, nl))
	b.WriteString(fmt.Sprintf("CompositorRenderCpuMs: %v%s", ft.CompositorRenderCpuMs, nl))
	b.WriteString(fmt.Sprintf("CompositorIdleCpuMs: %v%s", ft.CompositorIdleCpuMs, nl))
	b.WriteString(fmt.Sprintf("ClientFrameIntervalMs: %v%s", ft.ClientFrameIntervalMs, nl))
	b.WriteString(fmt.Sprintf("PresentCallCpuMs: %v%s", ft.PresentCallCpuMs, nl))
	b.WriteString(fmt.Sprintf("WaitForPresentCpuMs: %v%s", ft.WaitForPresentCpuMs, nl))
	b.WriteString(fmt.Sprintf("SubmitFrameMs: %v%s", ft.SubmitFrameMs, nl))
	b.WriteString(fmt.Sprintf("WaitGetPosesCalledMs: %v%s", ft.WaitGetPosesCalledMs, nl))
	b.WriteString(fmt.Sprintf("NewPosesReadyMs: %v%s", ft.NewPosesReadyMs, nl))
	b.WriteString(fmt.Sprintf("NewFrameReadyMs: %v%s", ft.NewFrameReadyMs, nl))
	b.WriteString(fmt.Sprintf("CompositorUpdateStartMs: %v%s", ft.CompositorUpdateStartMs, nl))
	b.WriteString(fmt.Sprintf("CompositorUpdateEndMs: %v%s", ft.CompositorUpdateEndMs, nl))
	b.WriteString(fmt.Sprintf("CompositorRenderStartMs: %v%s", ft.CompositorRenderStartMs, nl))

	return b.String()
}

/* TODO:

struct VR_IVRCompositor_FnTable
{
	void (OPENVR_FNTABLE_CALLTYPE *SetTrackingSpace)(ETrackingUniverseOrigin eOrigin);
	ETrackingUniverseOrigin (OPENVR_FNTABLE_CALLTYPE *GetTrackingSpace)();
	EVRCompositorError (OPENVR_FNTABLE_CALLTYPE *GetLastPoses)(struct TrackedDevicePose_t * pRenderPoseArray, uint32_t unRenderPoseArrayCount, struct TrackedDevicePose_t * pGamePoseArray, uint32_t unGamePoseArrayCount);
	EVRCompositorError (OPENVR_FNTABLE_CALLTYPE *GetLastPoseForTrackedDeviceIndex)(TrackedDeviceIndex_t unDeviceIndex, struct TrackedDevicePose_t * pOutputPose, struct TrackedDevicePose_t * pOutputGamePose);
	EVRCompositorError (OPENVR_FNTABLE_CALLTYPE *Submit)(EVREye eEye, struct Texture_t * pTexture, struct VRTextureBounds_t * pBounds, EVRSubmitFlags nSubmitFlags);
	void (OPENVR_FNTABLE_CALLTYPE *ClearLastSubmittedFrame)();
	void (OPENVR_FNTABLE_CALLTYPE *PostPresentHandoff)();
	uint32_t (OPENVR_FNTABLE_CALLTYPE *GetFrameTimings)(struct Compositor_FrameTiming * pTiming, uint32_t nFrames);
	void (OPENVR_FNTABLE_CALLTYPE *GetCumulativeStats)(struct Compositor_CumulativeStats * pStats, uint32_t nStatsSizeInBytes);
	void (OPENVR_FNTABLE_CALLTYPE *FadeToColor)(float fSeconds, float fRed, float fGreen, float fBlue, float fAlpha, bool bBackground);
	struct HmdColor_t (OPENVR_FNTABLE_CALLTYPE *GetCurrentFadeColor)(bool bBackground);
	void (OPENVR_FNTABLE_CALLTYPE *FadeGrid)(float fSeconds, bool bFadeIn);
	float (OPENVR_FNTABLE_CALLTYPE *GetCurrentGridAlpha)();
	EVRCompositorError (OPENVR_FNTABLE_CALLTYPE *SetSkyboxOverride)(struct Texture_t * pTextures, uint32_t unTextureCount);
	void (OPENVR_FNTABLE_CALLTYPE *ClearSkyboxOverride)();
	void (OPENVR_FNTABLE_CALLTYPE *CompositorBringToFront)();
	void (OPENVR_FNTABLE_CALLTYPE *CompositorGoToBack)();
	void (OPENVR_FNTABLE_CALLTYPE *CompositorQuit)();
	bool (OPENVR_FNTABLE_CALLTYPE *IsFullscreen)();
	uint32_t (OPENVR_FNTABLE_CALLTYPE *GetCurrentSceneFocusProcess)();
	uint32_t (OPENVR_FNTABLE_CALLTYPE *GetLastFrameRenderer)();
	bool (OPENVR_FNTABLE_CALLTYPE *CanRenderScene)();
	void (OPENVR_FNTABLE_CALLTYPE *ShowMirrorWindow)();
	void (OPENVR_FNTABLE_CALLTYPE *HideMirrorWindow)();
	bool (OPENVR_FNTABLE_CALLTYPE *IsMirrorWindowVisible)();
	void (OPENVR_FNTABLE_CALLTYPE *CompositorDumpImages)();
	bool (OPENVR_FNTABLE_CALLTYPE *ShouldAppRenderWithLowResources)();
	void (OPENVR_FNTABLE_CALLTYPE *ForceInterleavedReprojectionOn)(bool bOverride);
	void (OPENVR_FNTABLE_CALLTYPE *ForceReconnectProcess)();
	void (OPENVR_FNTABLE_CALLTYPE *SuspendRendering)(bool bSuspend);
	EVRCompositorError (OPENVR_FNTABLE_CALLTYPE *RequestScreenshot)(EVRScreenshotType type, char * pchDestinationFileName, char * pchVRDestinationFileName);
	EVRScreenshotType (OPENVR_FNTABLE_CALLTYPE *GetCurrentScreenshotType)();
	EVRCompositorError (OPENVR_FNTABLE_CALLTYPE *GetMirrorTextureD3D11)(EVREye eEye, void * pD3D11DeviceOrResource, void ** ppD3D11ShaderResourceView);
	EVRCompositorError (OPENVR_FNTABLE_CALLTYPE *GetMirrorTextureGL)(EVREye eEye, glUInt_t * pglTextureId, glSharedTextureHandle_t * pglSharedTextureHandle);
	bool (OPENVR_FNTABLE_CALLTYPE *ReleaseSharedGLTexture)(glUInt_t glTextureId, glSharedTextureHandle_t glSharedTextureHandle);
	void (OPENVR_FNTABLE_CALLTYPE *LockGLSharedTextureForAccess)(glSharedTextureHandle_t glSharedTextureHandle);
	void (OPENVR_FNTABLE_CALLTYPE *UnlockGLSharedTextureForAccess)(glSharedTextureHandle_t glSharedTextureHandle);
	uint32_t (OPENVR_FNTABLE_CALLTYPE *GetVulkanInstanceExtensionsRequired)(char * pchValue, uint32_t unBufferSize);
	uint32_t (OPENVR_FNTABLE_CALLTYPE *GetVulkanDeviceExtensionsRequired)(struct VkPhysicalDevice_T * pPhysicalDevice, char * pchValue, uint32_t unBufferSize);
	void (OPENVR_FNTABLE_CALLTYPE *SetExplicitTimingMode)(bool bExplicitTimingMode);
+	EVRCompositorError (OPENVR_FNTABLE_CALLTYPE *SubmitExplicitTimingData)();
};
*/
