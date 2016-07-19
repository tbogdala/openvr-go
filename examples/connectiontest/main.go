// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.


package main

import (
    "fmt"
    "runtime"
    vr "github.com/tbogdala/openvr-go"
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

    w,h := vrSystem.GetRecommendedRenderTargetSize()
    fmt.Printf("rec size: %d, %d\n", w, h)

    fmt.Printf("About to test the driver and display names ...\n")

    driver, errInt := vrSystem.GetStringTrackedDeviceProperty(int(vr.TrackedDeviceIndexHmd), vr.PropTrackingSystemNameString)
    if errInt != vr.TrackedPropSuccess {
        panic("error getting driver name.")
    }

    display, errInt := vrSystem.GetStringTrackedDeviceProperty(int(vr.TrackedDeviceIndexHmd), vr.PropSerialNumberString)
    if errInt != vr.TrackedPropSuccess {
        panic("error getting display name.")
    }

    fmt.Printf("Connection Test - %s %s\n", driver, display)
}
