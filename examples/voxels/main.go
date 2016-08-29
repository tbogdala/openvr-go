// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package main

import (
	"fmt"
	"math/rand"
	"os"
	"runtime"

	vr "github.com/tbogdala/openvr-go"
	chunk "github.com/tbogdala/openvr-go/examples/voxels/chunk"
	fizzlevr "github.com/tbogdala/openvr-go/util/fizzlevr"

	glfw "github.com/go-gl/glfw/v3.1/glfw"
	mgl "github.com/go-gl/mathgl/mgl32"

	fizzle "github.com/tbogdala/fizzle"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
	opengl "github.com/tbogdala/fizzle/graphicsprovider/opengl"
	input "github.com/tbogdala/fizzle/input/glfwinput"
	fizzlerenderer "github.com/tbogdala/fizzle/renderer"
	forward "github.com/tbogdala/fizzle/renderer/forward"
	noisey "github.com/tbogdala/noisey"
)

const (
	voxelShaderPath = "./assets/voxel"
	nearView        = 0.1
	farView         = 500.0
	worldChunkSize  = 24
	worldHeightGen  = 24
)

var (
	windowWidth  = int(1280)
	windowHeight = int(720)

	gfx               graphics.GraphicsProvider
	mainWindow        *glfw.Window
	kbModel           *input.KeyboardModel
	renderer          *forward.ForwardRenderer
	renderModelShader *fizzle.RenderShader
	lensShader        *fizzle.RenderShader
	voxelShader       *fizzle.RenderShader

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

	chunkMan       *chunk.Manager
	voxelTextures  *fizzle.TextureArray
	playerPosition = mgl.Vec3{
		float32(worldChunkSize / 2 * chunk.ChunkSize),
		float32(worldHeightGen + 2),
		float32(worldChunkSize / 2 * chunk.ChunkSize)}
)

func init() {
	runtime.LockOSThread()
}

func main() {
	////////////////////////////////////////////////////////////////////////////
	// start off by initializing the GL and GLFW libraries and creating a window.
	mainWindow, gfx = initGraphics("Voxels", windowWidth, windowHeight)

	// set the callback functions for key input
	kbModel = input.NewKeyboardModel(mainWindow)
	kbModel.BindTrigger(glfw.KeyEscape, setShouldClose)
	kbModel.SetupCallbacks()

	////////////////////////////////////////////////////////////////////////////
	// attempt to initialize the system
	var err error
	vrSystem, err = vr.Init()
	if err != nil || vrSystem == nil {
		fmt.Printf("vr.Init() returned an error: %v\n", err)
		os.Exit(1)
	}

	// print out some information about the headset as a good smoke test
	driver, errInt := vrSystem.GetStringTrackedDeviceProperty(int(vr.TrackedDeviceIndexHmd), vr.PropTrackingSystemNameString)
	if errInt != vr.TrackedPropSuccess {
		fmt.Printf("error getting driver name: %v\n", err)
		os.Exit(1)
	}
	displaySerial, errInt := vrSystem.GetStringTrackedDeviceProperty(int(vr.TrackedDeviceIndexHmd), vr.PropSerialNumberString)
	if errInt != vr.TrackedPropSuccess {
		fmt.Printf("error getting display name: %v\n", err)
		os.Exit(1)
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
		fmt.Printf("Error loading shaders: %v\n", err)
		os.Exit(1)
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

	// cache renderables for the connected devices
	deviceRenderables, err = fizzlevr.CreateDeviceRenderables(vrSystem, renderModelShader)
	if err != nil {
		fmt.Printf("Failed to load renderables for the connected devices. " + err.Error() + "\n")
	}

	// pull an interface to the compositor
	vrCompositor, err = vr.GetCompositor()
	if err != nil {
		fmt.Printf("Failed to get the compositor interface: %v\n", err)
		os.Exit(1)
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
	var err error

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

	voxelShader, err = fizzle.LoadShaderProgramFromFiles(voxelShaderPath, nil)
	if err != nil {
		return fmt.Errorf("Failed to compile and link the voxel shader program!\n%v", err)
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

	const TextureSize = 32
	voxelTextureFiles := make(map[string]string)
	voxelTextureFiles["Grass"] = "./assets/textures/default_grass.png"
	voxelTextureFiles["Dirt"] = "./assets/textures/default_dirt.png"
	voxelTextureFiles["Stones"] = "./assets/textures/default_stone_block.png"
	// create the texture array object
	voxelTextures = fizzle.NewTextureArray(TextureSize, int32(len(voxelTextureFiles)))
	err := voxelTextures.LoadImagesFromFiles(voxelTextureFiles, TextureSize, 0)
	if err != nil {
		fmt.Printf("Failed to load the voxel textures!\n%v", err)
		os.Exit(1)
	}

	createVoxels(worldChunkSize, worldHeightGen)
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
	playerPosition := mgl.Translate3D(-playerPosition[0], -playerPosition[1], -playerPosition[2])
	worldHmdPose := hmdPose.Mul4(playerPosition)
	if eye == vr.EyeLeft {
		view = eyeTransforms.PositionLeft.Mul4(worldHmdPose)
		perspective = eyeTransforms.ProjectionLeft
		camera.View = view
		camera.Position = hmdLoc
	} else {
		view = eyeTransforms.PositionRight.Mul4(worldHmdPose)
		perspective = eyeTransforms.ProjectionRight
		camera.View = view
		camera.Position = hmdLoc
	}

	// draw the voxels
	for _, c := range chunkMan.Chunks {
		if c == nil {
			continue
		}
		r := c.GetTheRenderable(voxelTextures.TextureIndexes)
		renderer.DrawRenderableWithShader(r, voxelShader, customVoxelBinder, perspective, view, camera)
	}

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

// ==========================================================================

// createVoxels creates the chunk landscape data.
func createVoxels(chunkLength, landScale int) {
	const seed1 = 1
	const seed2 = 2
	voxelGen := NewVoxelGenerator(seed1, seed2)
	chunkMan = chunk.NewManager(chunkLength, 0, 0, 0)

	// generate the basic land mass
	for y := 0; y < landScale/chunk.ChunkSize; y++ {
		yOffset := y * chunk.ChunkSize
		for x := 0; x < chunkLength; x++ {
			xOffset := x * chunk.ChunkSize
			for z := 0; z < chunkLength; z++ {
				zOffset := z * chunk.ChunkSize
				// create the chunk at the offset
				newChunk := chunk.NewChunk(x, y, z)

				// now loop through the Blocks of the chunk
				for cX := 0; cX < chunk.ChunkSize; cX++ {
					for cZ := 0; cZ < chunk.ChunkSize; cZ++ {
						noise := voxelGen.GetFBM(float32(cX+xOffset), float32(cZ+zOffset))
						heightF := (noise*0.5 + 0.5) * float32(landScale)
						heightI := int(heightF)
						for cY := 0; cY < chunk.ChunkSize; cY++ {
							// index the chunk's Blocks array
							chunkI := (cY * chunk.ChunkSize2) + (cX * chunk.ChunkSize) + cZ

							if cY+yOffset == 0 {
								// always have a stone floor
								newChunk.Blocks[chunkI].Type = chunk.BlockTypeStones
							} else if heightI > cY+yOffset {
								// if we're under the noise height, default to grass
								newChunk.Blocks[chunkI].Type = chunk.BlockTypeGrass

								if heightI-1 > cY+yOffset {
									// if we're under the top layer make dirt
									newChunk.Blocks[chunkI].Type = chunk.BlockTypeDirt
								} else if heightI-3 > cY+yOffset {
									// if we're sufficiently deep from the top layer, make stones
									newChunk.Blocks[chunkI].Type = chunk.BlockTypeStones
								}
							}
						} // cY
					} // cZ
				} // cX

				newChunk.UpdateColliders()
				chunkMan.RegisterChunk(newChunk)
				//fmt.Printf("Chunk registered @ %d, %d, %d\n", xOffset, yOffset, zOffset)
			} // z
		} // x
	} // y
}

// VoxelGenerator is the structure that contains the random generators for the voxels.
type VoxelGenerator struct {
	r1 noisey.RandomSource
	r2 noisey.RandomSource

	simplex1 noisey.OpenSimplexGenerator
	simplex2 noisey.OpenSimplexGenerator

	hifreq  noisey.FBMGenerator2D
	lofreq  noisey.FBMGenerator2D
	flatter noisey.Scale2D
	control noisey.FBMGenerator2D
	mixer   noisey.Select2D
}

// NewVoxelGenerator returns a new VoxelGenerator structure
func NewVoxelGenerator(seed1, seed2 int64) *VoxelGenerator {
	lg := new(VoxelGenerator)

	// setup the random sources
	lg.r1 = rand.New(rand.NewSource(seed1))
	lg.r2 = rand.New(rand.NewSource(seed2))

	// setup the noise sources
	lg.simplex1 = noisey.NewOpenSimplexGenerator(lg.r1)
	lg.simplex2 = noisey.NewOpenSimplexGenerator(lg.r2)

	// now setup the fBm noise generators and the land selector
	lg.lofreq = noisey.NewFBMGenerator2D(&lg.simplex2, 2, 0.15, 1.8, 1.1)
	lg.hifreq = noisey.NewFBMGenerator2D(&lg.simplex1, 5, 0.75, 2.1, 1.33)

	lg.flatter = noisey.NewScale2D(&lg.lofreq, 0.4, 0.1, -1, 1)
	lg.control = noisey.NewFBMGenerator2D(&lg.simplex2, 2, 0.5, 2.0, 1.0)
	lg.mixer = noisey.NewSelect2D(&lg.flatter, &lg.lofreq, &lg.control, 0.4, 100, 0.2)
	return lg
}

// Get3D returns the density of a particaular coordinate
func (gen *VoxelGenerator) Get3D(x, y, z float32) float32 {
	v := gen.simplex1.Get3D(float64(x)*0.1, float64(y)*0.1, float64(z)*0.1)
	return float32(v)
}

// GetFBM returns a fract brownian motion noise level for the coordinate
func (gen *VoxelGenerator) GetFBM(x, z float32) float32 {
	v := gen.mixer.Get2D(float64(x)*0.1, float64(z)*0.1)
	return float32(v)
}

func customVoxelBinder(renderer fizzlerenderer.Renderer, r *fizzle.Renderable, shader *fizzle.RenderShader, texturesBound *int32) {
	const uintSize = 4
	gfx := fizzle.GetGraphics()

	shaderTexArray := shader.GetUniformLocation("VOXEL_TEXTURES")
	if shaderTexArray >= 0 {
		gfx.ActiveTexture(graphics.Texture(graphics.TEXTURE0 + uint32(*texturesBound)))
		gfx.BindTexture(graphics.TEXTURE_2D_ARRAY, voxelTextures.Texture)
		gfx.Uniform1i(shaderTexArray, *texturesBound)
		*texturesBound = *texturesBound + 1
	}

	// Note: The following will get turned to floats in the shaders because OpenGL ES 2.0 is
	// miserable to work with an can't have int attributes coming from VBO since
	// VertexAttribIPointer isn't implemented in v2.0.

	shaderCombo1 := shader.GetAttribLocation("VERTEX_TEXTURE_INDEX")
	if shaderCombo1 >= 0 {
		gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.ComboVBO1)
		gfx.EnableVertexAttribArray(uint32(shaderCombo1))
		gfx.VertexAttribPointer(uint32(shaderCombo1), 1, graphics.UNSIGNED_INT, false, 2*uintSize, gfx.PtrOffset(0))
	}

	shaderFakeAO := shader.GetAttribLocation("VERTEX_VOXEL_BF")
	if shaderFakeAO >= 0 {
		gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.ComboVBO1)
		gfx.EnableVertexAttribArray(uint32(shaderFakeAO))
		gfx.VertexAttribPointer(uint32(shaderFakeAO), 1, graphics.UNSIGNED_INT, false, 2*uintSize, gfx.PtrOffset(uintSize))
	}
}
