// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package openvr

/*
#include <stdio.h>
#include <stdlib.h>
#include "openvr_capi.h"

extern struct VR_IVRSystem_FnTable* _iSystem;

// .___ ____   ______________   _________                  __
// |   |\   \ /   /\______   \ /   _____/ ___.__.  _______/  |_   ____    _____
// |   | \   Y   /  |       _/ \_____  \ <   |  | /  ___/\   __\_/ __ \  /     \
// |   |  \     /   |    |   \ /        \ \___  | \___ \  |  |  \  ___/ |  Y Y  \
// |___|   \___/    |____|_  //_______  / / ____|/____  > |__|   \___  >|__|_|  /
//                         \/         \/  \/          \/             \/       \/

void system_GetRecommendedRenderTargetSize(struct VR_IVRSystem_FnTable* iSystem, uint32_t* width, uint32_t* height) {
    iSystem->GetRecommendedRenderTargetSize(width, height);
}

struct HmdMatrix44_t system_GetProjectionMatrix(struct VR_IVRSystem_FnTable* iSystem, EVREye eEye, float fNearZ, float fFarZ) {
    return iSystem->GetProjectionMatrix(eEye, fNearZ, fFarZ);
}

struct HmdMatrix34_t system_GetEyeToHeadTransform(struct VR_IVRSystem_FnTable* iSystem, EVREye eEye) {
    return iSystem->GetEyeToHeadTransform(eEye);
}

bool system_ComputeDistortion(struct VR_IVRSystem_FnTable* iSystem, EVREye eEye, float fU, float fV, struct DistortionCoordinates_t* dest) {
    return iSystem->ComputeDistortion(eEye, fU, fV, dest) != 0;
}

bool system_IsTrackedDeviceConnected(struct VR_IVRSystem_FnTable* iSystem, TrackedDeviceIndex_t unDeviceIndex) {
    return iSystem->IsTrackedDeviceConnected(unDeviceIndex);
}

ETrackedDeviceClass system_GetTrackedDeviceClass(struct VR_IVRSystem_FnTable* iSystem, TrackedDeviceIndex_t unDeviceIndex) {
    return iSystem->GetTrackedDeviceClass(unDeviceIndex);
}

bool system_IsInputFocusCapturedByAnotherProcess(struct VR_IVRSystem_FnTable* iSystem) {
    return iSystem->IsInputFocusCapturedByAnotherProcess();
}

uint32_t system_GetStringTrackedDeviceProperty(struct VR_IVRSystem_FnTable* iSystem, TrackedDeviceIndex_t unDeviceIndex, ETrackedDeviceProperty prop, char * pchValue, uint32_t unBufferSize, ETrackedPropertyError * pError) {
    return _iSystem->GetStringTrackedDeviceProperty(unDeviceIndex, prop, pchValue, unBufferSize, pError);
}

bool system_PollNextEvent(struct VR_IVRSystem_FnTable* iSystem, struct VREvent_t * pEvent, uint32_t uncbVREvent) {
    return iSystem->PollNextEvent(pEvent, uncbVREvent);
}

bool system_GetControllerState(struct VR_IVRSystem_FnTable* iSystem, TrackedDeviceIndex_t unControllerDeviceIndex, VRControllerState_t * pControllerState) {
    return iSystem->GetControllerState(unControllerDeviceIndex, pControllerState, sizeof(VRControllerState_t));
}

char* system_GetControllerAxisTypeNameFromEnum(struct VR_IVRSystem_FnTable* iSystem, EVRControllerAxisType eAxisType) {
	return iSystem->GetControllerAxisTypeNameFromEnum(eAxisType);
}

uint32_t system_GetInt32TrackedDeviceProperty(struct VR_IVRSystem_FnTable* iSystem, TrackedDeviceIndex_t unDeviceIndex, ETrackedDeviceProperty prop,ETrackedPropertyError * pError) {
    return _iSystem->GetInt32TrackedDeviceProperty(unDeviceIndex, prop, pError);
}

*/
import "C"
import (
	"unsafe"

	mgl "github.com/go-gl/mathgl/mgl32"
)

// DistortionCoordinates is used to return the post-distortion UVs for each color channel.
// UVs range from 0 to 1 with 0,0 in the upper left corner of the
// source render target. The 0,0 to 1,1 range covers a single eye.
type DistortionCoordinates struct {
	Red   mgl.Vec2
	Green mgl.Vec2
	Blue  mgl.Vec2
}

// System is an interface wrapper to IVRSystem.
type System struct {
	ptr *C.struct_VR_IVRSystem_FnTable
}

// GetRecommendedRenderTargetSize returns the suggested size for the intermediate render
// target that the distortion pulls from.
func (sys *System) GetRecommendedRenderTargetSize() (uint32, uint32) {
	var w, h C.uint32_t
	C.system_GetRecommendedRenderTargetSize(sys.ptr, &w, &h)
	return uint32(w), uint32(h)
}

// GetProjectionMatrix returns the projection matrix for the specified eye
func (sys *System) GetProjectionMatrix(eye int, near, far float32, dest *mgl.Mat4) {
	m44 := C.system_GetProjectionMatrix(sys.ptr, C.EVREye(eye), C.float(near), C.float(far))

	dest[0] = float32(m44.m[0][0])
	dest[4] = float32(m44.m[0][1])
	dest[8] = float32(m44.m[0][2])
	dest[12] = float32(m44.m[0][3])

	dest[1] = float32(m44.m[1][0])
	dest[5] = float32(m44.m[1][1])
	dest[9] = float32(m44.m[1][2])
	dest[13] = float32(m44.m[1][3])

	dest[2] = float32(m44.m[2][0])
	dest[6] = float32(m44.m[2][1])
	dest[10] = float32(m44.m[2][2])
	dest[14] = float32(m44.m[2][3])

	dest[3] = float32(m44.m[3][0])
	dest[7] = float32(m44.m[3][1])
	dest[11] = float32(m44.m[3][2])
	dest[15] = float32(m44.m[3][3])
}

// GetEyeToHeadTransform returns the transform from eye space to the head space. Eye space is the per-eye flavor of head
// space that provides stereo disparity. Instead of Model * View * Projection the sequence is Model * View * Eye^-1 * Projection.
// Normally View and Eye^-1 will be multiplied together and treated as View in your application.
func (sys *System) GetEyeToHeadTransform(eye int, dest *mgl.Mat3x4) {
	m34 := C.system_GetEyeToHeadTransform(sys.ptr, C.EVREye(eye))

	dest[0] = float32(m34.m[0][0])
	dest[3] = float32(m34.m[0][1])
	dest[6] = float32(m34.m[0][2])
	dest[9] = float32(m34.m[0][3])

	dest[1] = float32(m34.m[1][0])
	dest[4] = float32(m34.m[1][1])
	dest[7] = float32(m34.m[1][2])
	dest[10] = float32(m34.m[1][3])

	dest[2] = float32(m34.m[2][0])
	dest[5] = float32(m34.m[2][1])
	dest[8] = float32(m34.m[2][2])
	dest[11] = float32(m34.m[2][3])
}

// ComputeDistortion gets the result of the distortion function for the specified eye and input UVs. UVs go from 0,0 in
// the upper left of that eye's viewport and 1,1 in the lower right of that eye's viewport.
// Returns true for success. Otherwise, returns false, and distortion coordinates are not suitable.
func (sys *System) ComputeDistortion(eye int, u, v float32, dest *DistortionCoordinates) bool {
	var cDest C.struct_DistortionCoordinates_t
	if C.system_ComputeDistortion(sys.ptr, C.EVREye(eye), C.float(u), C.float(v), &cDest) == 0 {
		return false
	}

	dest.Red[0] = float32(cDest.rfRed[0])
	dest.Red[1] = float32(cDest.rfRed[1])
	dest.Green[0] = float32(cDest.rfGreen[0])
	dest.Green[1] = float32(cDest.rfGreen[1])
	dest.Blue[0] = float32(cDest.rfBlue[0])
	dest.Blue[1] = float32(cDest.rfBlue[1])
	return true
}

// IsTrackedDeviceConnected returns true if there is a device connected in this slot.
func (sys *System) IsTrackedDeviceConnected(deviceIndex uint32) bool {
	if C.system_IsTrackedDeviceConnected(sys.ptr, C.TrackedDeviceIndex_t(deviceIndex)) != 0 {
		return true
	}
	return false
}

// GetTrackedDeviceClass returns the device class of a tracked device. If there has not been a device connected in this slot
// since the application started this function will return TrackedDevice_Invalid. For previous detected
// devices the function will return the previously observed device class.
//
// To determine which devices exist on the system, just loop from 0 to k_unMaxTrackedDeviceCount and check
// the device class. Every device with something other than TrackedDevice_Invalid is associated with an
// actual tracked device.
func (sys *System) GetTrackedDeviceClass(deviceIndex int) int {
	result := C.system_GetTrackedDeviceClass(sys.ptr, C.TrackedDeviceIndex_t(deviceIndex))
	return int(result)
}

// IsInputFocusCapturedByAnotherProcess returns true if input focus is captured by another process.
func (sys *System) IsInputFocusCapturedByAnotherProcess() bool {
	if C.system_IsInputFocusCapturedByAnotherProcess(sys.ptr) != 0 {
		return true
	}
	return false
}

// GetStringTrackedDeviceProperty returns a string property. If the device index is not valid or the property is
// not a string type this function will an empty string. The int returned correspnds to the ETrackedPropertyError enumeration.
func (sys *System) GetStringTrackedDeviceProperty(deviceIndex int, property int) (string, int) {
	// attempt to get the size of the property first
	var cErrorVal C.ETrackedPropertyError
	size := C.system_GetStringTrackedDeviceProperty(sys.ptr, C.TrackedDeviceIndex_t(deviceIndex), C.ETrackedDeviceProperty(property), nil, 0, &cErrorVal)
	if size == 0 {
		return "", int(cErrorVal)
	}

	buffer := make([]byte, size)
	C.system_GetStringTrackedDeviceProperty(sys.ptr, C.TrackedDeviceIndex_t(deviceIndex), C.ETrackedDeviceProperty(property), (*C.char)(unsafe.Pointer(&buffer[0])), size, &cErrorVal)
	return string(buffer[:size]), int(cErrorVal)
}

// VREvent is an event posted by the server to all running applications
type VREvent struct {
	EventType          uint32 // EVREventType enum
	TrackedDeviceIndex uint32
	EventAgeSeconds    float32
	data               C.VREvent_Data_t
}

var (
	// eventBuffer is used as a temporary event item buffer
	eventBuffer C.struct_VREvent_t
)

// PollNextEvent returns true and fills the event with the next event on the queue if there is one.
// If there are no events this method returns false.
func (sys *System) PollNextEvent(event *VREvent) bool {
	result := C.system_PollNextEvent(sys.ptr, &eventBuffer, C.sizeof_struct_VREvent_t)

	if result != 0 {
		// update the event structure with a copy of the event
		event.EventType = uint32(eventBuffer.eventType)
		event.TrackedDeviceIndex = uint32(eventBuffer.trackedDeviceIndex)
		event.EventAgeSeconds = float32(eventBuffer.eventAgeSeconds)
		event.data = eventBuffer.data
		return true
	}
	return false
}

// ControllerAxis represents the state of joystick and track pads.
type ControllerAxis struct {
	X float32
	Y float32
}

// ControllerState is the current state of a controller
type ControllerState struct {
	PacketNum     uint32
	ButtonPressed uint64
	ButtonTouched uint64
	Axis          [ControllerStateAxisCount]ControllerAxis
}

var (
	// controllerStateBuffer is used as a temporary event item buffer
	controllerStateBuffer C.struct_VRControllerState_t
)

// GetControllerState fills the supplied struct with the current state of the controller.
// Returns false if the controller index is invalid.
func (sys *System) GetControllerState(deviceIndex int, state *ControllerState) bool {
	result := C.system_GetControllerState(sys.ptr, C.TrackedDeviceIndex_t(deviceIndex), &controllerStateBuffer)

	if result != 0 {
		state.PacketNum = uint32(controllerStateBuffer.unPacketNum)
		state.ButtonPressed = uint64(controllerStateBuffer.ulButtonPressed)
		state.ButtonTouched = uint64(controllerStateBuffer.ulButtonTouched)
		for i := uint(0); i < ControllerStateAxisCount; i++ {
			state.Axis[i].X = float32(controllerStateBuffer.rAxis[i].x)
			state.Axis[i].Y = float32(controllerStateBuffer.rAxis[i].y)
		}
		return true
	}
	return false
}

// EyeTransforms is a struct that contains the projection and translation
// matrix transforms for each eye in the HMD.
type EyeTransforms struct {
	ProjectionLeft  mgl.Mat4 // left eye projection
	ProjectionRight mgl.Mat4 // right eye projection
	PositionLeft    mgl.Mat4 // left eye offset
	PositionRight   mgl.Mat4 // right eye offset
}

// GetEyeTransforms returns a structure containing the projection and translation
// matrixes for both eyes given the near/far settings passed in.
func (sys *System) GetEyeTransforms(near, far float32) *EyeTransforms {
	transforms := new(EyeTransforms)
	var m mgl.Mat4
	var m34 mgl.Mat3x4

	sys.GetProjectionMatrix(EyeLeft, near, far, &m)
	transforms.ProjectionLeft = mgl.Mat4(m)

	sys.GetProjectionMatrix(EyeRight, near, far, &m)
	transforms.ProjectionRight = mgl.Mat4(m)

	sys.GetEyeToHeadTransform(EyeLeft, &m34)
	transforms.PositionLeft = mgl.Mat4(Mat34ToMat4(&m34))
	transforms.PositionLeft = transforms.PositionLeft.Inv()

	sys.GetEyeToHeadTransform(EyeRight, &m34)
	transforms.PositionRight = mgl.Mat4(Mat34ToMat4(&m34))
	transforms.PositionRight = transforms.PositionRight.Inv()

	return transforms
}

// GetControllerAxisTypeNameFromEnum returns the name of an EVRControllerAxisType enum value
func (sys *System) GetControllerAxisTypeNameFromEnum(axisType int) string {
	cAxisName := C.system_GetControllerAxisTypeNameFromEnum(sys.ptr, C.EVRControllerAxisType(axisType))
	axisName := C.GoString(cAxisName)
	return axisName
}

// GetInt32TrackedDeviceProperty returns a int32 property. If the device index is not valid or the property is
// not valid it will return 0. The second int returned correspnds to the ETrackedPropertyError enumeration.
func (sys *System) GetInt32TrackedDeviceProperty(deviceIndex int, property int) (int32, int) {
	var cErrorVal C.ETrackedPropertyError
	cInt32Prop := C.system_GetInt32TrackedDeviceProperty(sys.ptr, C.TrackedDeviceIndex_t(deviceIndex), C.ETrackedDeviceProperty(property), &cErrorVal)
	return int32(cInt32Prop), int(cErrorVal)
}

/* TODO List:

void (OPENVR_FNTABLE_CALLTYPE *GetProjectionRaw)(EVREye eEye, float * pfLeft, float * pfRight, float * pfTop, float * pfBottom);
bool (OPENVR_FNTABLE_CALLTYPE *GetTimeSinceLastVsync)(float * pfSecondsSinceLastVsync, uint64_t * pulFrameCounter);
int32_t (OPENVR_FNTABLE_CALLTYPE *GetD3D9AdapterIndex)();
void (OPENVR_FNTABLE_CALLTYPE *GetDXGIOutputInfo)(int32_t * pnAdapterIndex);
bool (OPENVR_FNTABLE_CALLTYPE *IsDisplayOnDesktop)();
bool (OPENVR_FNTABLE_CALLTYPE *SetDisplayVisibility)(bool bIsVisibleOnDesktop);
void (OPENVR_FNTABLE_CALLTYPE *GetDeviceToAbsoluteTrackingPose)(ETrackingUniverseOrigin eOrigin, float fPredictedSecondsToPhotonsFromNow, struct TrackedDevicePose_t * pTrackedDevicePoseArray, uint32_t unTrackedDevicePoseArrayCount);
void (OPENVR_FNTABLE_CALLTYPE *ResetSeatedZeroPose)();
struct HmdMatrix34_t (OPENVR_FNTABLE_CALLTYPE *GetSeatedZeroPoseToStandingAbsoluteTrackingPose)();
struct HmdMatrix34_t (OPENVR_FNTABLE_CALLTYPE *GetRawZeroPoseToStandingAbsoluteTrackingPose)();
uint32_t (OPENVR_FNTABLE_CALLTYPE *GetSortedTrackedDeviceIndicesOfClass)(ETrackedDeviceClass eTrackedDeviceClass, TrackedDeviceIndex_t * punTrackedDeviceIndexArray, uint32_t unTrackedDeviceIndexArrayCount, TrackedDeviceIndex_t unRelativeToTrackedDeviceIndex);
EDeviceActivityLevel (OPENVR_FNTABLE_CALLTYPE *GetTrackedDeviceActivityLevel)(TrackedDeviceIndex_t unDeviceId);
void (OPENVR_FNTABLE_CALLTYPE *ApplyTransform)(struct TrackedDevicePose_t * pOutputPose, struct TrackedDevicePose_t * pTrackedDevicePose, struct HmdMatrix34_t * pTransform);
TrackedDeviceIndex_t (OPENVR_FNTABLE_CALLTYPE *GetTrackedDeviceIndexForControllerRole)(ETrackedControllerRole unDeviceType);
ETrackedControllerRole (OPENVR_FNTABLE_CALLTYPE *GetControllerRoleForTrackedDeviceIndex)(TrackedDeviceIndex_t unDeviceIndex);
bool (OPENVR_FNTABLE_CALLTYPE *GetBoolTrackedDeviceProperty)(TrackedDeviceIndex_t unDeviceIndex, ETrackedDeviceProperty prop, ETrackedPropertyError * pError);
float (OPENVR_FNTABLE_CALLTYPE *GetFloatTrackedDeviceProperty)(TrackedDeviceIndex_t unDeviceIndex, ETrackedDeviceProperty prop, ETrackedPropertyError * pError);
uint64_t (OPENVR_FNTABLE_CALLTYPE *GetUint64TrackedDeviceProperty)(TrackedDeviceIndex_t unDeviceIndex, ETrackedDeviceProperty prop, ETrackedPropertyError * pError);
struct HmdMatrix34_t (OPENVR_FNTABLE_CALLTYPE *GetMatrix34TrackedDeviceProperty)(TrackedDeviceIndex_t unDeviceIndex, ETrackedDeviceProperty prop, ETrackedPropertyError * pError);
char * (OPENVR_FNTABLE_CALLTYPE *GetPropErrorNameFromEnum)(ETrackedPropertyError error);
bool (OPENVR_FNTABLE_CALLTYPE *PollNextEventWithPose)(ETrackingUniverseOrigin eOrigin, struct VREvent_t * pEvent, uint32_t uncbVREvent, TrackedDevicePose_t * pTrackedDevicePose);
char * (OPENVR_FNTABLE_CALLTYPE *GetEventTypeNameFromEnum)(EVREventType eType);
struct HiddenAreaMesh_t (OPENVR_FNTABLE_CALLTYPE *GetHiddenAreaMesh)(EVREye eEye, EHiddenAreaMeshType type);
bool (OPENVR_FNTABLE_CALLTYPE *GetControllerStateWithPose)(ETrackingUniverseOrigin eOrigin, TrackedDeviceIndex_t unControllerDeviceIndex, VRControllerState_t * pControllerState, uint32_t unControllerStateSize, struct TrackedDevicePose_t * pTrackedDevicePose);
void (OPENVR_FNTABLE_CALLTYPE *TriggerHapticPulse)(TrackedDeviceIndex_t unControllerDeviceIndex, uint32_t unAxisId, unsigned short usDurationMicroSec);
char * (OPENVR_FNTABLE_CALLTYPE *GetButtonIdNameFromEnum)(EVRButtonId eButtonId);
bool (OPENVR_FNTABLE_CALLTYPE *CaptureInputFocus)();
void (OPENVR_FNTABLE_CALLTYPE *ReleaseInputFocus)();
uint32_t (OPENVR_FNTABLE_CALLTYPE *DriverDebugRequest)(TrackedDeviceIndex_t unDeviceIndex, char * pchRequest, char * pchResponseBuffer, uint32_t unResponseBufferSize);
EVRFirmwareError (OPENVR_FNTABLE_CALLTYPE *PerformFirmwareUpdate)(TrackedDeviceIndex_t unDeviceIndex);
void (OPENVR_FNTABLE_CALLTYPE *AcknowledgeQuit_Exiting)();
void (OPENVR_FNTABLE_CALLTYPE *AcknowledgeQuit_UserPrompt)();
*/
