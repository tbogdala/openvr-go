// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package main

import (
	"fmt"
	"runtime"

	vr "github.com/tbogdala/openvr-go"
	fizzlevr "github.com/tbogdala/openvr-go/aux/fizzlevr"

	glfw "github.com/go-gl/glfw/v3.1/glfw"
	mgl "github.com/go-gl/mathgl/mgl32"

	fizzle "github.com/tbogdala/fizzle"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
	opengl "github.com/tbogdala/fizzle/graphicsprovider/opengl"
	input "github.com/tbogdala/fizzle/input/glfwinput"
	forward "github.com/tbogdala/fizzle/renderer/forward"
)

const (
	windowWidth     = 1280
	windowHeight    = 720
	basicShaderPath = "./basic"
)

type EyeFramebuffer struct {
	depthBuffer        graphics.Buffer
	renderTexture      graphics.Texture
	renderFramebuffer  graphics.Buffer
	resolveTexture     graphics.Texture
	resolveFramebuffer graphics.Buffer
}

var (
	gfx               graphics.GraphicsProvider
	mainWindow        *glfw.Window
	kbModel           *input.KeyboardModel
	renderer          *forward.ForwardRenderer
	basicShader       *fizzle.RenderShader
	cube              *fizzle.Renderable
	renderModelShader *fizzle.RenderShader
	deviceRenderables *fizzlevr.DeviceRenderables

	// interfaces for openvr
	vrSystem       *vr.System
	vrCompositor   *vr.Compositor
	vrRenderModels *vr.RenderModels

	// render surfaces and transforms
	renderWidth         uint32
	renderHeight        uint32
	projectionLeft      mgl.Mat4
	projectionRight     mgl.Mat4
	eyePositionLeft     mgl.Mat4
	eyePositionRight    mgl.Mat4
	eyeFramebufferLeft  EyeFramebuffer
	eyeFramebufferRight EyeFramebuffer
	hmdPose             mgl.Mat4
	hmdLoc              mgl.Vec3

	// lens values calculated in createDistortion()
	lensVAO        uint32
	lensVerts      graphics.Buffer
	lensIndices    graphics.Buffer
	lensShader     *fizzle.RenderShader
	lensIndexCount int32
)

func init() {
	runtime.LockOSThread()
}

func main() {
	// start off by initializing the GL and GLFW libraries and creating a window.
	mainWindow, gfx = initGraphics("OpenVR Hello Cube", windowWidth, windowHeight)

	// set the callback functions for key input
	kbModel = input.NewKeyboardModel(mainWindow)
	kbModel.BindTrigger(glfw.KeyEscape, setShouldClose)
	kbModel.SetupCallbacks()

	// attempt to initialize the system
	var err error
	vrSystem, err = vr.Init()
	if err != nil || vrSystem == nil {
		panic("vr.Init() returned an error: " + err.Error())
	}

	// print out some information about the headset as a good test that everything
	// is starting to work.
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
	// get the size of the render targets to make
	renderWidth, renderHeight = vrSystem.GetRecommendedRenderTargetSize()
	fmt.Printf("rec size: %d, %d\n", renderWidth, renderHeight)

	err = createShaders()
	if err != nil {
		panic(err.Error())
	}
	createScene(renderWidth, renderHeight)
	createEyeTransforms()
	createStereoRenderTargets(renderWidth, renderHeight)
	createDistortion(vrSystem)
	createRenderModels(vrSystem)

	vrCompositor, err = vr.GetCompositor()
	if err != nil {
		panic("Failed to get the compositor interface: " + err.Error())
	}

	// Main Loop
	for !mainWindow.ShouldClose() {
		handleInput()
		renderFrame()
	}

	vr.Shutdown()
}

func createRenderModels(vrSystem *vr.System) {
	var err error
	deviceRenderables, err = fizzlevr.CreateDeviceRenderables(vrSystem, renderModelShader)
	if err != nil {
		fmt.Printf("Failed to load renderables for the connected devices. " + err.Error() + "\n")
	}
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
	//mainWindow.SetSizeCallback(onWindowResize)
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

// createShaders will load the shaders necessary for the sample to run
func createShaders() error {
	// load the diffuse shader for the cube
	var err error
	basicShader, err = fizzle.LoadShaderProgramFromFiles(basicShaderPath, nil)
	if err != nil {
		return fmt.Errorf("Failed to compile and link the diffuse shader program!\n%v", err)
	}

	renderModelShader, err = fizzle.LoadShaderProgram(vr.ShaderRenderModelV, vr.ShaderRenderModelF, nil)
	if err != nil {
		return fmt.Errorf("Failed to compile and link the render model shader program!\n%v", err)
	}

	lensShader, err = fizzle.LoadShaderProgram(vr.ShaderLensDistortionV, vr.ShaderLensDistortionF, nil)
	if err != nil {
		return fmt.Errorf("Failed to compile and link the lens distortion shader program!\n%v", err)
	}

	return nil
}

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
	cube.Core.Shininess = 4.8
}

func createEyeTransforms() {
	const near = 0.1
	const far = 30.0

	var m vr.Mat4
	var m34 vr.Mat34

	vrSystem.GetProjectionMatrix(vr.EyeLeft, near, far, vr.APIOpenGL, &m)
	projectionLeft = mgl.Mat4(m)
	vrSystem.GetProjectionMatrix(vr.EyeRight, near, far, vr.APIOpenGL, &m)
	projectionRight = mgl.Mat4(m)
	vrSystem.GetEyeToHeadTransform(vr.EyeLeft, &m34)
	eyePositionLeft = vrMat34ToMat4(&m34).Inv()
	vrSystem.GetEyeToHeadTransform(vr.EyeRight, &m34)
	eyePositionRight = vrMat34ToMat4(&m34).Inv()
}

func createStereoRenderTargets(renderWidth, renderHeight uint32) {
	createFramebuffer(&eyeFramebufferLeft, renderWidth, renderHeight)
	createFramebuffer(&eyeFramebufferRight, renderWidth, renderHeight)
}

func createFramebuffer(eyeFB *EyeFramebuffer, width, height uint32) {
	eyeFB.renderFramebuffer = gfx.GenFramebuffer()
	gfx.BindFramebuffer(graphics.FRAMEBUFFER, eyeFB.renderFramebuffer)

	eyeFB.depthBuffer = gfx.GenRenderbuffer()
	gfx.BindRenderbuffer(graphics.RENDERBUFFER, eyeFB.depthBuffer)
	gfx.RenderbufferStorageMultisample(graphics.RENDERBUFFER, 4, graphics.DEPTH_COMPONENT, int32(width), int32(height))
	gfx.FramebufferRenderbuffer(graphics.FRAMEBUFFER, graphics.DEPTH_ATTACHMENT, graphics.RENDERBUFFER, eyeFB.depthBuffer)

	eyeFB.renderTexture = gfx.GenTexture()
	gfx.BindTexture(graphics.TEXTURE_2D_MULTISAMPLE, eyeFB.renderTexture)
	gfx.TexImage2DMultisample(graphics.TEXTURE_2D_MULTISAMPLE, 4, graphics.RGBA8, int32(width), int32(height), true)
	gfx.FramebufferTexture2D(graphics.FRAMEBUFFER, graphics.COLOR_ATTACHMENT0, graphics.TEXTURE_2D_MULTISAMPLE, eyeFB.renderTexture, 0)

	// ---------------------------------------------------------------------- //

	eyeFB.resolveFramebuffer = gfx.GenFramebuffer()
	gfx.BindFramebuffer(graphics.FRAMEBUFFER, eyeFB.resolveFramebuffer)

	eyeFB.resolveTexture = gfx.GenTexture()
	gfx.BindTexture(graphics.TEXTURE_2D, eyeFB.resolveTexture)
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_MIN_FILTER, graphics.LINEAR)
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_MAX_LEVEL, 0)
	gfx.TexImage2D(graphics.TEXTURE_2D, 0, graphics.RGBA8, int32(width), int32(height), 0, graphics.RGBA, graphics.UNSIGNED_BYTE, nil, 0)
	gfx.FramebufferTexture2D(graphics.FRAMEBUFFER, graphics.COLOR_ATTACHMENT0, graphics.TEXTURE_2D, eyeFB.resolveTexture, 0)

	gfx.BindFramebuffer(graphics.FRAMEBUFFER, 0)
}

func createDistortion(vrSystem *vr.System) {
	const lensGridSegmentCountH = 43
	const lensGridSegmentCountV = 43

	w := float32(1.0) / float32(lensGridSegmentCountH-1)
	h := float32(1.0) / float32(lensGridSegmentCountV-1)

	u := float32(0.0)
	v := float32(0.0)

	var verts []float32
	var dc0 vr.DistortionCoordinates

	// left eye distortion verts
	Xoffset := float32(-1.0)
	for y := 0; y < lensGridSegmentCountV; y++ {
		for x := 0; x < lensGridSegmentCountH; x++ {
			u = float32(x) * w
			v = 1.0 - float32(y)*h
			vrSystem.ComputeDistortion(vr.EyeLeft, u, v, &dc0)
			verts = append(verts, Xoffset+u)
			verts = append(verts, -1.0+2.0*float32(y)*h)
			verts = append(verts, dc0.Red[0])
			verts = append(verts, 1.0-dc0.Red[1])
			verts = append(verts, dc0.Green[0])
			verts = append(verts, 1.0-dc0.Green[1])
			verts = append(verts, dc0.Blue[0])
			verts = append(verts, 1.0-dc0.Blue[1])
		}
	}

	// right eye distortion verts
	Xoffset = float32(0.0)
	for y := 0; y < lensGridSegmentCountV; y++ {
		for x := 0; x < lensGridSegmentCountH; x++ {
			u = float32(x) * w
			v = 1.0 - float32(y)*h
			vrSystem.ComputeDistortion(vr.EyeRight, u, v, &dc0)
			verts = append(verts, Xoffset+u)
			verts = append(verts, -1.0+2.0*float32(y)*h)
			verts = append(verts, dc0.Red[0])
			verts = append(verts, 1.0-dc0.Red[1])
			verts = append(verts, dc0.Green[0])
			verts = append(verts, 1.0-dc0.Green[1])
			verts = append(verts, dc0.Blue[0])
			verts = append(verts, 1.0-dc0.Blue[1])
		}
	}

	var vIndices []uint32
	var a, b, c, d uint32
	offset := 0
	for y := 0; y < lensGridSegmentCountV-1; y++ {
		for x := 0; x < lensGridSegmentCountH-1; x++ {
			a = uint32(lensGridSegmentCountH*y + x + offset)
			b = uint32(lensGridSegmentCountH*y + x + 1 + offset)
			c = uint32((y+1)*lensGridSegmentCountH + x + 1 + offset)
			d = uint32((y+1)*lensGridSegmentCountH + x + offset)
			vIndices = append(vIndices, a)
			vIndices = append(vIndices, b)
			vIndices = append(vIndices, c)

			vIndices = append(vIndices, a)
			vIndices = append(vIndices, c)
			vIndices = append(vIndices, d)
		}
	}

	offset = lensGridSegmentCountH * lensGridSegmentCountV
	for y := 0; y < lensGridSegmentCountV-1; y++ {
		for x := 0; x < lensGridSegmentCountH-1; x++ {
			a = uint32(lensGridSegmentCountH*y + x + offset)
			b = uint32(lensGridSegmentCountH*y + x + 1 + offset)
			c = uint32((y+1)*lensGridSegmentCountH + x + 1 + offset)
			d = uint32((y+1)*lensGridSegmentCountH + x + offset)
			vIndices = append(vIndices, a)
			vIndices = append(vIndices, b)
			vIndices = append(vIndices, c)

			vIndices = append(vIndices, a)
			vIndices = append(vIndices, c)
			vIndices = append(vIndices, d)
		}
	}
	lensIndexCount = int32(len(vIndices))

	const floatSize = 4
	const uintSize = 4

	// create the OpenGL objects
	lensVAO = gfx.GenVertexArray()
	gfx.BindVertexArray(lensVAO)

	lensVerts = gfx.GenBuffer()
	gfx.BindBuffer(graphics.ARRAY_BUFFER, lensVerts)
	gfx.BufferData(graphics.ARRAY_BUFFER, len(verts)*floatSize, gfx.Ptr(&verts[0]), graphics.STATIC_DRAW)

	lensIndices = gfx.GenBuffer()
	gfx.BindBuffer(graphics.ELEMENT_ARRAY_BUFFER, lensIndices)
	gfx.BufferData(graphics.ELEMENT_ARRAY_BUFFER, len(vIndices)*uintSize, gfx.Ptr(&vIndices[0]), graphics.STATIC_DRAW)

	const lensStride = 8 * floatSize
	const offsetPosition = 0
	const offsetRed = 2 * floatSize
	const offsetGreen = 4 * floatSize
	const offsetBlue = 6 * floatSize

	shaderPosition := lensShader.GetAttribLocation("position")
	gfx.EnableVertexAttribArray(uint32(shaderPosition))
	gfx.VertexAttribPointer(uint32(shaderPosition), 2, graphics.FLOAT, false, lensStride, gfx.PtrOffset(offsetPosition))

	shaderRed := lensShader.GetAttribLocation("v2UVredIn")
	gfx.EnableVertexAttribArray(uint32(shaderRed))
	gfx.VertexAttribPointer(uint32(shaderRed), 2, graphics.FLOAT, false, lensStride, gfx.PtrOffset(offsetRed))

	shaderGreen := lensShader.GetAttribLocation("v2UVGreenIn")
	gfx.EnableVertexAttribArray(uint32(shaderGreen))
	gfx.VertexAttribPointer(uint32(shaderGreen), 2, graphics.FLOAT, false, lensStride, gfx.PtrOffset(offsetGreen))

	shaderBlue := lensShader.GetAttribLocation("v2UVblueIn")
	gfx.EnableVertexAttribArray(uint32(shaderBlue))
	gfx.VertexAttribPointer(uint32(shaderBlue), 2, graphics.FLOAT, false, lensStride, gfx.PtrOffset(offsetBlue))

	gfx.BindVertexArray(0)
	gfx.BindBuffer(graphics.ARRAY_BUFFER, 0)
	gfx.BindBuffer(graphics.ELEMENT_ARRAY_BUFFER, 0)
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
	renderStereoTargets()
	renderDistortion()

	vrCompositor.Submit(vr.EyeLeft, uint32(eyeFramebufferLeft.resolveTexture))
	vrCompositor.Submit(vr.EyeRight, uint32(eyeFramebufferRight.resolveTexture))

	// draw the screen
	mainWindow.SwapBuffers()

	// update the HMD pose
	updateHMDPose()
}

func renderControllers(perspective mgl.Mat4, view mgl.Mat4, camera fizzle.Camera) {
	if vrSystem.IsInputFocusCapturedByAnotherProcess() {
		return
	}

	for i := vr.TrackedDeviceIndexHmd + 1; i < vr.MaxTrackedDeviceCount; i++ {
		// only draw controllers
		if vrSystem.GetTrackedDeviceClass(uint32(i)) != vr.TrackedDeviceClassController {
			continue
		}

		// make sure the pose is correct
		pose := vrCompositor.GetRenderPose(i)
		if !pose.PoseIsValid {
			continue
		}

		// get the renderable
		r, err := deviceRenderables.GetRenderableForTrackedDevice(int(i))
		if err != nil {
			fmt.Printf("renderControllers: failed to get the renderable for device #%d: %s\n", i, err.Error())
			continue
		}

		// calculate the mvp based off of the model's pose
		poseMat := vrMat34ToMat4(&pose.DeviceToAbsoluteTracking)
		mvp := perspective.Mul4(view).Mul4(poseMat)

		gfx.UseProgram(renderModelShader.Prog)
		gfx.BindVertexArray(r.Core.Vao)

		shaderMvp := renderModelShader.GetUniformLocation("mvp")
		if shaderMvp >= 0 {
			gfx.UniformMatrix4fv(shaderMvp, 1, false, mvp)
		}

		shaderPosition := renderModelShader.GetAttribLocation("position")
		if shaderPosition >= 0 {
			gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.VertVBO)
			gfx.EnableVertexAttribArray(uint32(shaderPosition))
			gfx.VertexAttribPointer(uint32(shaderPosition), 3, graphics.FLOAT, false, r.Core.VBOStride, gfx.PtrOffset(r.Core.VertVBOOffset))
		}

		shaderVertUv := renderModelShader.GetAttribLocation("texCoord")
		if shaderVertUv >= 0 {
			gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.UvVBO)
			gfx.EnableVertexAttribArray(uint32(shaderVertUv))
			gfx.VertexAttribPointer(uint32(shaderVertUv), 2, graphics.FLOAT, false, r.Core.VBOStride, gfx.PtrOffset(r.Core.UvVBOOffset))
		}

		shaderTex0 := renderModelShader.GetUniformLocation("diffuse")
		if shaderTex0 >= 0 {
			gfx.ActiveTexture(graphics.Texture(graphics.TEXTURE0))
			gfx.BindTexture(graphics.TEXTURE_2D, r.Core.Tex0)
			gfx.Uniform1i(shaderTex0, 0)
		}

		gfx.BindBuffer(graphics.ELEMENT_ARRAY_BUFFER, r.Core.ElementsVBO)
		gfx.DrawElements(graphics.Enum(graphics.TRIANGLES), int32(r.FaceCount*3), graphics.UNSIGNED_INT, gfx.PtrOffset(0))

		gfx.BindVertexArray(0)
	}
}

func renderStereoTargets() {
	gfx.Enable(graphics.CULL_FACE)
	gfx.ClearColor(0.15, 0.15, 0.18, 1.0) // nice background color, but not black

	// left eye
	gfx.Enable(graphics.MULTISAMPLE)
	gfx.BindFramebuffer(graphics.FRAMEBUFFER, eyeFramebufferLeft.renderFramebuffer)
	gfx.Viewport(0, 0, int32(renderWidth), int32(renderHeight))
	renderScene(vr.EyeLeft)
	gfx.BindFramebuffer(graphics.FRAMEBUFFER, 0)
	gfx.Disable(graphics.MULTISAMPLE)

	gfx.BindFramebuffer(graphics.READ_FRAMEBUFFER, eyeFramebufferLeft.renderFramebuffer)
	gfx.BindFramebuffer(graphics.DRAW_FRAMEBUFFER, eyeFramebufferLeft.resolveFramebuffer)
	gfx.BlitFramebuffer(0, 0, int32(renderWidth), int32(renderHeight), 0, 0, int32(renderWidth), int32(renderHeight), graphics.COLOR_BUFFER_BIT, graphics.LINEAR)
	gfx.BindFramebuffer(graphics.READ_FRAMEBUFFER, 0)
	gfx.BindFramebuffer(graphics.DRAW_FRAMEBUFFER, 0)

	// right eye
	gfx.Enable(graphics.MULTISAMPLE)
	gfx.BindFramebuffer(graphics.FRAMEBUFFER, eyeFramebufferRight.renderFramebuffer)
	gfx.Viewport(0, 0, int32(renderWidth), int32(renderHeight))
	renderScene(vr.EyeRight)
	gfx.BindFramebuffer(graphics.FRAMEBUFFER, 0)
	gfx.Disable(graphics.MULTISAMPLE)

	gfx.BindFramebuffer(graphics.READ_FRAMEBUFFER, eyeFramebufferRight.renderFramebuffer)
	gfx.BindFramebuffer(graphics.DRAW_FRAMEBUFFER, eyeFramebufferRight.resolveFramebuffer)
	gfx.BlitFramebuffer(0, 0, int32(renderWidth), int32(renderHeight), 0, 0, int32(renderWidth), int32(renderHeight), graphics.COLOR_BUFFER_BIT, graphics.LINEAR)
	gfx.BindFramebuffer(graphics.READ_FRAMEBUFFER, 0)
	gfx.BindFramebuffer(graphics.DRAW_FRAMEBUFFER, 0)

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

func renderScene(eye int) {
	gfx.Clear(graphics.COLOR_BUFFER_BIT | graphics.DEPTH_BUFFER_BIT)
	gfx.Enable(graphics.DEPTH_TEST)

	var perspective, view mgl.Mat4
	var camera FixedCamera
	if eye == vr.EyeLeft {
		view = eyePositionLeft.Mul4(hmdPose)
		perspective = projectionLeft
		camera.View = view
		camera.Position = hmdLoc
	} else {
		view = eyePositionRight.Mul4(hmdPose)
		perspective = projectionRight
		camera.View = view
		camera.Position = hmdLoc
	}

	// draw our cube as the main thing
	renderer.DrawRenderable(cube, nil, perspective, view, camera)

	// now draw any controllers that get rendered into the scene
	renderControllers(perspective, view, camera)
}

func updateHMDPose() {
	// WaitGetPoses is used as a sync point in the OpenVR API. This is on a timer to keep 90fps, so
	// the OpenVR gives you that much time to draw a frame. By calling WaitGetPoses() you wait the
	// remaining amount of time. If you only used 1ms it will wait 10ms here. If you used 5ms it will wait 6ms.
	// (approx.)
	vrCompositor.WaitGetPoses(false)
	if vrCompositor.IsPoseValid(vr.TrackedDeviceIndexHmd) {
		pose := vrCompositor.GetRenderPose(vr.TrackedDeviceIndexHmd)
		hmdPose = vrMat34ToMat4(&pose.DeviceToAbsoluteTracking).Inv()

		// FIXME: this is probably broken.
		hmdLoc[0] = pose.DeviceToAbsoluteTracking[9]
		hmdLoc[1] = pose.DeviceToAbsoluteTracking[10]
		hmdLoc[2] = pose.DeviceToAbsoluteTracking[11]
	}
}

func vrMat34ToMat4(vrM34 *vr.Mat34) (m4 mgl.Mat4) {
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

func renderDistortion() {
	gfx.Disable(graphics.CULL_FACE)
	gfx.Disable(graphics.DEPTH_TEST)
	gfx.Viewport(0, 0, windowWidth, windowHeight)
	gfx.ClearColor(0.0, 0.0, 0.0, 1)
	gfx.Clear(graphics.COLOR_BUFFER_BIT | graphics.DEPTH_BUFFER_BIT)

	gfx.BindVertexArray(lensVAO)
	gfx.UseProgram(lensShader.Prog)

	// render left lens
	gfx.ActiveTexture(graphics.TEXTURE0)
	gfx.BindTexture(graphics.TEXTURE_2D, eyeFramebufferLeft.resolveTexture)
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_WRAP_S, graphics.CLAMP_TO_EDGE)
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_WRAP_T, graphics.CLAMP_TO_EDGE)
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_MAG_FILTER, graphics.LINEAR)
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_MIN_FILTER, graphics.LINEAR_MIPMAP_LINEAR)
	gfx.DrawElements(graphics.TRIANGLES, (lensIndexCount / 2), graphics.UNSIGNED_INT, gfx.PtrOffset(0))

	// render right lens

	gfx.BindTexture(graphics.TEXTURE_2D, eyeFramebufferRight.resolveTexture)
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_WRAP_S, graphics.CLAMP_TO_EDGE)
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_WRAP_T, graphics.CLAMP_TO_EDGE)
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_MAG_FILTER, graphics.LINEAR)
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_MIN_FILTER, graphics.LINEAR_MIPMAP_LINEAR)
	gfx.DrawElements(graphics.TRIANGLES, (lensIndexCount / 2), graphics.UNSIGNED_INT, gfx.PtrOffset(int((lensIndexCount/2)*4))) // uint32size

	gfx.BindVertexArray(0)
	gfx.UseProgram(0)
}
