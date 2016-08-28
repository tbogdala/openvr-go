// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package main

import (
	"fmt"
	vr "github.com/tbogdala/openvr-go"
	"runtime"
)

func init() {
	runtime.LockOSThread()
}

func main() {
	// attempt to initialize the system
	vrSystem, err := vr.Init()
	if err != nil {
		fmt.Printf("vr.Init() returned an error: %v\n", err)
	}

	if vrSystem == nil {
		panic("vrSystem is nil")
	}

	// print out the driver and display names
	fmt.Printf("About to test the driver and display names ...\n")
	driver, errInt := vrSystem.GetStringTrackedDeviceProperty(int(vr.TrackedDeviceIndexHmd), vr.PropTrackingSystemNameString)
	if errInt != vr.TrackedPropSuccess {
		panic("error getting driver name.")
	}
	display, errInt := vrSystem.GetStringTrackedDeviceProperty(int(vr.TrackedDeviceIndexHmd), vr.PropSerialNumberString)
	if errInt != vr.TrackedPropSuccess {
		panic("error getting display name.")
	}
	fmt.Printf("Connection Test: %s - %s\n", driver, display)

	// print out the recommended render target size
	w, h := vrSystem.GetRecommendedRenderTargetSize()
	fmt.Printf("Render target size: %d x %d\n", w, h)

	// print out the play area dimensions
	vrChaperone, err := vr.GetChaperone()
	if err != nil {
		panic("error getting IVRChaperone interface.")
	}
	calibrationState := vrChaperone.GetCalibrationState()
	fmt.Printf("Calibration state: %d (1==OK, 100's==Warning, 200's==Error)\n", calibrationState)
	playX, playZ := vrChaperone.GetPlayAreaSize()
	fmt.Printf("Play area size: %f x %f\n", playX, playZ)
	playRect := vrChaperone.GetPlayAreaRect()
	fmt.Printf("Play area corners:\n\t%v\n\t%v\n\t%v\n\t%v\n", playRect[0], playRect[1], playRect[2], playRect[3])
}
