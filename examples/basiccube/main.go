// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package main

import (
	"fmt"
	"runtime"

	vr "github.com/tbogdala/openvr-go"
	fizzlevr "github.com/tbogdala/openvr-go/util/fizzlevr"

	glfw "github.com/go-gl/glfw/v3.1/glfw"
	mgl "github.com/go-gl/mathgl/mgl32"

	fizzle "github.com/tbogdala/fizzle"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
	opengl "github.com/tbogdala/fizzle/graphicsprovider/opengl"
	input "github.com/tbogdala/fizzle/input/glfwinput"
	forward "github.com/tbogdala/fizzle/renderer/forward"
)

const (
	basicShaderPath = "./basic"
	nearView        = 0.1
	farView         = 30.0
)

var (
	windowWidth  = int(1280)
	windowHeight = int(720)

	gfx               graphics.GraphicsProvider
	mainWindow        *glfw.Window
	kbModel           *input.KeyboardModel
	renderer          *forward.ForwardRenderer
	basicShader       *fizzle.RenderShader
	renderModelShader *fizzle.RenderShader
	lensShader        *fizzle.RenderShader
	cube              *fizzle.Renderable

	// interfaces for openvr
	vrSystem          *vr.System
	vrCompositor      *vr.Compositor
	deviceRenderables *fizzlevr.DeviceRenderables
	distortionLens    *fizzlevr.DistortionLens

	// render surfaces and transforms
	renderWidth         uint32
	renderHeight        uint32
	eyeTransforms       *vr.EyeTransforms
	eyeFramebufferLeft  *fizzlevr.EyeFramebuffer
	eyeFramebufferRight *fizzlevr.EyeFramebuffer
	hmdPose             mgl.Mat4
	hmdLoc              mgl.Vec3
)

func init() {
	runtime.LockOSThread()
}

func main() {
	////////////////////////////////////////////////////////////////////////////
	// start off by initializing the GL and GLFW libraries and creating a window.
	mainWindow, gfx = initGraphics("OpenVR Hello Cube", windowWidth, windowHeight)

	// set the callback functions for key input
	kbModel = input.NewKeyboardModel(mainWindow)
	kbModel.BindTrigger(glfw.KeyEscape, setShouldClose)
	kbModel.SetupCallbacks()

	////////////////////////////////////////////////////////////////////////////
	// attempt to initialize the system
	var err error
	vrSystem, err = vr.Init()
	if err != nil || vrSystem == nil {
		panic("vr.Init() returned an error: " + err.Error())
	}

	// print out some information about the headset as a good smoke test
	driver, errInt := vrSystem.GetStringTrackedDeviceProperty(int(vr.TrackedDeviceIndexHmd), vr.PropTrackingSystemNameString)
	if errInt != vr.TrackedPropSuccess {
		panic("error getting driver name.")
	}
	displaySerial, errInt := vrSystem.GetStringTrackedDeviceProperty(int(vr.TrackedDeviceIndexHmd), vr.PropSerialNumberString)
	if errInt != vr.TrackedPropSuccess {
		panic("error getting display name.")
	}
	fmt.Printf("Connected to %s %s\n", driver, displaySerial)

	////////////////////////////////////////////////////////////////////////////
	// setup VR specifics and the initial scene

	// get the size of the render targets to make
	renderWidth, renderHeight = vrSystem.GetRecommendedRenderTargetSize()
	fmt.Printf("rec size: %d, %d\n", renderWidth, renderHeight)

	// load up our shaders
	err = createShaders()
	if err != nil {
		panic(err.Error())
	}

	// create some objects and lights
	createScene(renderWidth, renderHeight)

	// get the eye transforms necessary for the VR HMD
	eyeTransforms = vrSystem.GetEyeTransforms(nearView, farView)

	// setup the framebuffers for the eyes
	eyeFramebufferLeft, eyeFramebufferRight = fizzlevr.CreateStereoRenderTargets(renderWidth, renderHeight)

	// create the lens distortion object which will be used to render the
	// eye framebuffers to the GLFW window.
	distortionLens = fizzlevr.CreateDistortionLens(vrSystem, lensShader, eyeFramebufferLeft, eyeFramebufferRight)

	// debug: do a little extra work right here to print out some debugging info.
	// this isn't required for any functionality but exists as a test of some API calls.
	// we even shoot 1 over on purpose in the loops to make sure the API call doesn't crash.
	vrRenderModels, err := vr.GetRenderModels()
	if err == nil {
		renderModelCount := vrRenderModels.GetRenderModelCount()
		fmt.Printf("Render Model count: %d\n", renderModelCount)
		for mi := uint32(0); mi <= renderModelCount; mi++ {
			modelName := vrRenderModels.GetRenderModelName(mi)
			fmt.Printf("\trender model %d: %s\n", mi, modelName)

			componentCount := vrRenderModels.GetComponentCount(modelName)
			if componentCount <= 0 {
				continue
			}
			fmt.Printf("\t\tcomponent count = %d\n", componentCount)
			for ci := uint32(0); ci <= componentCount; ci++ {
				componentName := vrRenderModels.GetComponentName(modelName, ci)
				fmt.Printf("\t\t%d = %s ", ci, componentName)
				if len(componentName) > 0 {
					componentRenderModelName := vrRenderModels.GetComponentRenderModelName(modelName, componentName)
					fmt.Printf("; render model = %s\n", componentRenderModelName)
					// try to load this thing
					componentModel, err2 := vrRenderModels.RenderModelLoad(componentRenderModelName)
					if componentModel != nil && err2 == nil {
						fmt.Printf("\t\tloaded model; %d faces\n", componentModel.TriangleCount)
					}
				} else {
					fmt.Printf("\n")
				}
			}
		}
	}

	// cache renderables for the connected devices
	deviceRenderables, err = fizzlevr.CreateDeviceRenderables(vrSystem, renderModelShader)
	if err != nil {
		fmt.Printf("Failed to load renderables for the connected devices. " + err.Error() + "\n")
	}

	// pull an interface to the compositor
	vrCompositor, err = vr.GetCompositor()
	if err != nil {
		panic("Failed to get the compositor interface: " + err.Error())
	}

	////////////////////////////////////////////////////////////////////////////
	// the main application loop
	for !mainWindow.ShouldClose() {
		handleInput()
		renderFrame()
	}

	vr.Shutdown()
}

// initGraphics creates an OpenGL window and initializes the required graphics libraries.
// It will either succeed or panic.
func initGraphics(title string, w int, h int) (*glfw.Window, graphics.GraphicsProvider) {
	// GLFW must be initialized before it's called
	err := glfw.Init()
	if err != nil {
		panic("Can't init glfw! " + err.Error())
	}

	// request a OpenGL 3.3 core context
	glfw.WindowHint(glfw.Samples, 4)
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)

	// do the actual window creation
	mainWindow, err = glfw.CreateWindow(w, h, title, nil, nil)
	if err != nil {
		panic("Failed to create the main window! " + err.Error())
	}
	mainWindow.SetSizeCallback(onWindowResize)
	mainWindow.MakeContextCurrent()

	// disable v-sync for max draw rate
	glfw.SwapInterval(0)

	// initialize OpenGL
	gfx, err := opengl.InitOpenGL()
	if err != nil {
		panic("Failed to initialize OpenGL! " + err.Error())
	}
	fizzle.SetGraphics(gfx)

	return mainWindow, gfx
}

// setShouldClose should be called to close the window and kill the app.
func setShouldClose() {
	mainWindow.SetShouldClose(true)
}

// onWindowResize should be called when the main window gets resized.
func onWindowResize(w *glfw.Window, width int, height int) {
	windowWidth = width
	windowHeight = height
}

// createShaders will load the shaders necessary for the sample to run.
func createShaders() error {
	// load the diffuse shader for the cube
	var err error
	basicShader, err = fizzle.LoadShaderProgramFromFiles(basicShaderPath, nil)
	if err != nil {
		return fmt.Errorf("Failed to compile and link the diffuse shader program!\n%v", err)
	}

	// load the shader used to draw the connected devices
	renderModelShader, err = fizzle.LoadShaderProgram(vr.ShaderRenderModelV, vr.ShaderRenderModelF, nil)
	if err != nil {
		return fmt.Errorf("Failed to compile and link the render model shader program!\n%v", err)
	}

	// load the shader used to render the framebuffers to a window for viewing
	lensShader, err = fizzle.LoadShaderProgram(vr.ShaderLensDistortionV, vr.ShaderLensDistortionF, nil)
	if err != nil {
		return fmt.Errorf("Failed to compile and link the lens distortion shader program!\n%v", err)
	}

	return nil
}

// createScene creates a simple test scene to render.
func createScene(renderWidth, renderHeight uint32) {
	// create a new renderer
	renderer = forward.NewForwardRenderer(gfx)
	renderer.ChangeResolution(int32(renderWidth), int32(renderHeight))

	// put a light in there
	light := renderer.NewDirectionalLight(mgl.Vec3{1.0, -0.5, -1.0})
	light.DiffuseIntensity = 0.70
	light.SpecularIntensity = 0.10
	light.AmbientIntensity = 0.3
	renderer.ActiveLights[0] = light

	// create a 1 ft. cube to render
	const cubeSize = 0.30 * 0.5
	cube = fizzle.CreateCube(-cubeSize, -cubeSize, -cubeSize, cubeSize, cubeSize, cubeSize)
	cube.Core.Shader = basicShader
	cube.Core.DiffuseColor = mgl.Vec4{0.9, 0.05, 0.05, 1.0}
	cube.Core.SpecularColor = mgl.Vec4{1.0, 1.0, 1.0, 1.0}
	cube.Core.Shininess = 10.0
}

func handleInput() {
	// advise GLFW to poll for input. without this the window appears to hang.
	glfw.PollEvents()

	// handle any keyboard input
	kbModel.CheckKeyPresses()

	var event vr.VREvent
	for vrSystem.PollNextEvent(&event) {
		proccessVREvent(&event)
	}

	// TODO: update controller states
}

func proccessVREvent(event *vr.VREvent) {
	switch event.EventType {
	case vr.VREventTrackedDeviceActivated:
		// TODO: setup render model
		fmt.Printf("Device %d attached.\n", event.TrackedDeviceIndex)
	case vr.VREventTrackedDeviceDeactivated:
		fmt.Printf("Device %d detached.\n", event.TrackedDeviceIndex)
	case vr.VREventTrackedDeviceUpdated:
		fmt.Printf("Device %d updated.\n", event.TrackedDeviceIndex)
	}
}

func renderFrame() {
	// draw the framebuffers
	renderStereoTargets()

	// draw the framebuffers to the window
	distortionLens.Render(int32(windowWidth), int32(windowHeight))

	// send the framebuffer textures out to the compositor for rendering to the HMD
	vrCompositor.Submit(vr.EyeLeft, uint32(eyeFramebufferLeft.ResolveTexture))
	vrCompositor.Submit(vr.EyeRight, uint32(eyeFramebufferRight.ResolveTexture))

	// draw the screen
	mainWindow.SwapBuffers()

	// update the HMD pose, which causes a wait to vsync the HMD
	updateHMDPose()
}

type FixedCamera struct {
	View     mgl.Mat4
	Position mgl.Vec3
}

func (c FixedCamera) GetViewMatrix() mgl.Mat4 {
	return c.View
}
func (c FixedCamera) GetPosition() mgl.Vec3 {
	return c.Position
}

// renderScene gets called for each eye and is responsible for
// rendering the entire scene.
func renderScene(eye int) {
	gfx.Clear(graphics.COLOR_BUFFER_BIT | graphics.DEPTH_BUFFER_BIT)
	gfx.Enable(graphics.DEPTH_TEST)

	var perspective, view mgl.Mat4
	var camera FixedCamera
	if eye == vr.EyeLeft {
		view = eyeTransforms.PositionLeft.Mul4(hmdPose)
		perspective = eyeTransforms.ProjectionLeft
		camera.View = view
		camera.Position = hmdLoc
	} else {
		view = eyeTransforms.PositionRight.Mul4(hmdPose)
		perspective = eyeTransforms.ProjectionRight
		camera.View = view
		camera.Position = hmdLoc
	}

	// draw our cube as the main thing
	renderer.DrawRenderable(cube, nil, perspective, view, camera)

	// now draw any devices that get rendered into the scene
	deviceRenderables.RenderDevices(vrCompositor, perspective, view, camera)
}

// renderStereoTargets renders each of the left and right eye framebuffers
// calling renderScene to do the rendering for the scene.
func renderStereoTargets() {
	gfx.Enable(graphics.CULL_FACE)
	gfx.ClearColor(0.15, 0.15, 0.18, 1.0) // nice background color, but not black

	// left eye
	gfx.Enable(graphics.MULTISAMPLE)
	gfx.BindFramebuffer(graphics.FRAMEBUFFER, eyeFramebufferLeft.RenderFramebuffer)
	gfx.Viewport(0, 0, int32(renderWidth), int32(renderHeight))
	renderScene(vr.EyeLeft)
	gfx.BindFramebuffer(graphics.FRAMEBUFFER, 0)
	gfx.Disable(graphics.MULTISAMPLE)

	gfx.BindFramebuffer(graphics.READ_FRAMEBUFFER, eyeFramebufferLeft.RenderFramebuffer)
	gfx.BindFramebuffer(graphics.DRAW_FRAMEBUFFER, eyeFramebufferLeft.ResolveFramebuffer)
	gfx.BlitFramebuffer(0, 0, int32(renderWidth), int32(renderHeight), 0, 0, int32(renderWidth), int32(renderHeight), graphics.COLOR_BUFFER_BIT, graphics.LINEAR)
	gfx.BindFramebuffer(graphics.READ_FRAMEBUFFER, 0)
	gfx.BindFramebuffer(graphics.DRAW_FRAMEBUFFER, 0)

	// right eye
	gfx.Enable(graphics.MULTISAMPLE)
	gfx.BindFramebuffer(graphics.FRAMEBUFFER, eyeFramebufferRight.RenderFramebuffer)
	gfx.Viewport(0, 0, int32(renderWidth), int32(renderHeight))
	renderScene(vr.EyeRight)
	gfx.BindFramebuffer(graphics.FRAMEBUFFER, 0)
	gfx.Disable(graphics.MULTISAMPLE)

	gfx.BindFramebuffer(graphics.READ_FRAMEBUFFER, eyeFramebufferRight.RenderFramebuffer)
	gfx.BindFramebuffer(graphics.DRAW_FRAMEBUFFER, eyeFramebufferRight.ResolveFramebuffer)
	gfx.BlitFramebuffer(0, 0, int32(renderWidth), int32(renderHeight), 0, 0, int32(renderWidth), int32(renderHeight), graphics.COLOR_BUFFER_BIT, graphics.LINEAR)
	gfx.BindFramebuffer(graphics.READ_FRAMEBUFFER, 0)
	gfx.BindFramebuffer(graphics.DRAW_FRAMEBUFFER, 0)

}

func updateHMDPose() {
	// WaitGetPoses is used as a sync point in the OpenVR API. This is on a timer to keep 90fps, so
	// the OpenVR gives you that much time to draw a frame. By calling WaitGetPoses() you wait the
	// remaining amount of time. If you only used 1ms it will wait 10ms here. If you used 5ms it will wait 6ms.
	// (approx.)
	vrCompositor.WaitGetPoses(false)
	if vrCompositor.IsPoseValid(vr.TrackedDeviceIndexHmd) {
		pose := vrCompositor.GetRenderPose(vr.TrackedDeviceIndexHmd)
		hmdPose = mgl.Mat4(vr.Mat34ToMat4(&pose.DeviceToAbsoluteTracking)).Inv()

		// FIXME: this is probably broken.
		hmdLoc[0] = pose.DeviceToAbsoluteTracking[9]
		hmdLoc[1] = pose.DeviceToAbsoluteTracking[10]
		hmdLoc[2] = pose.DeviceToAbsoluteTracking[11]
	}
}
