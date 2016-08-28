// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package fizzlevr

import (
	fizzle "github.com/tbogdala/fizzle"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
	vr "github.com/tbogdala/openvr-go"
)

// EyeFramebuffer contains the render buffers and textures
// used to render each eye in VR.
type EyeFramebuffer struct {
	DepthBuffer        graphics.Buffer
	RenderTexture      graphics.Texture
	RenderFramebuffer  graphics.Buffer
	ResolveTexture     graphics.Texture
	ResolveFramebuffer graphics.Buffer
}

// CreateStereoRenderTargets returns new EyeFramebuffer structs, one for each eye,
// that are used for rendering.
func CreateStereoRenderTargets(width, height uint32) (left, right *EyeFramebuffer) {
	left = new(EyeFramebuffer)
	right = new(EyeFramebuffer)

	left.Init(width, height)
	right.Init(width, height)

	return left, right
}

// Init creates the necessary render buffers and render textures for an eye.
func (eyeFB *EyeFramebuffer) Init(width, height uint32) {
	gfx := fizzle.GetGraphics()

	eyeFB.RenderFramebuffer = gfx.GenFramebuffer()
	gfx.BindFramebuffer(graphics.FRAMEBUFFER, eyeFB.RenderFramebuffer)

	eyeFB.DepthBuffer = gfx.GenRenderbuffer()
	gfx.BindRenderbuffer(graphics.RENDERBUFFER, eyeFB.DepthBuffer)
	gfx.RenderbufferStorageMultisample(graphics.RENDERBUFFER, 4, graphics.DEPTH_COMPONENT, int32(width), int32(height))
	gfx.FramebufferRenderbuffer(graphics.FRAMEBUFFER, graphics.DEPTH_ATTACHMENT, graphics.RENDERBUFFER, eyeFB.DepthBuffer)

	eyeFB.RenderTexture = gfx.GenTexture()
	gfx.BindTexture(graphics.TEXTURE_2D_MULTISAMPLE, eyeFB.RenderTexture)
	gfx.TexImage2DMultisample(graphics.TEXTURE_2D_MULTISAMPLE, 4, graphics.RGBA8, int32(width), int32(height), true)
	gfx.FramebufferTexture2D(graphics.FRAMEBUFFER, graphics.COLOR_ATTACHMENT0, graphics.TEXTURE_2D_MULTISAMPLE, eyeFB.RenderTexture, 0)

	// ---------------------------------------------------------------------- //

	eyeFB.ResolveFramebuffer = gfx.GenFramebuffer()
	gfx.BindFramebuffer(graphics.FRAMEBUFFER, eyeFB.ResolveFramebuffer)

	eyeFB.ResolveTexture = gfx.GenTexture()
	gfx.BindTexture(graphics.TEXTURE_2D, eyeFB.ResolveTexture)
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_MIN_FILTER, graphics.LINEAR)
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_MAX_LEVEL, 0)
	gfx.TexImage2D(graphics.TEXTURE_2D, 0, graphics.RGBA8, int32(width), int32(height), 0, graphics.RGBA, graphics.UNSIGNED_BYTE, nil, 0)
	gfx.FramebufferTexture2D(graphics.FRAMEBUFFER, graphics.COLOR_ATTACHMENT0, graphics.TEXTURE_2D, eyeFB.ResolveTexture, 0)

	gfx.BindFramebuffer(graphics.FRAMEBUFFER, 0)
}

// DistortionLens is used to render the VR framebuffers out to a window
// based on lens distortionvalues calculated in Init().
type DistortionLens struct {
	VAO        uint32
	Verts      graphics.Buffer
	Indices    graphics.Buffer
	IndexCount int32
	Shader     *fizzle.RenderShader
	EyeRight   *EyeFramebuffer
	EyeLeft    *EyeFramebuffer
}

// CreateDistortionLens creates a DistortionLens object that can render
// the framebuffers for the left and right eye to a window.
func CreateDistortionLens(vrSystem *vr.System, lensShader *fizzle.RenderShader, eyeLeft, eyeRight *EyeFramebuffer) *DistortionLens {
	lens := new(DistortionLens)
	lens.Shader = lensShader
	lens.EyeLeft = eyeLeft
	lens.EyeRight = eyeRight

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
	lens.IndexCount = int32(len(vIndices))

	const floatSize = 4
	const uintSize = 4

	// create the OpenGL objects
	gfx := fizzle.GetGraphics()
	lens.VAO = gfx.GenVertexArray()
	gfx.BindVertexArray(lens.VAO)

	lens.Verts = gfx.GenBuffer()
	gfx.BindBuffer(graphics.ARRAY_BUFFER, lens.Verts)
	gfx.BufferData(graphics.ARRAY_BUFFER, len(verts)*floatSize, gfx.Ptr(&verts[0]), graphics.STATIC_DRAW)

	lens.Indices = gfx.GenBuffer()
	gfx.BindBuffer(graphics.ELEMENT_ARRAY_BUFFER, lens.Indices)
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

	return lens
}

// Render draws the distortion lens view of the left and right eye
// framebuffers to the window.
func (lens *DistortionLens) Render(windowWidth, windowHeight int32) {
	gfx := fizzle.GetGraphics()

	gfx.Disable(graphics.CULL_FACE)
	gfx.Disable(graphics.DEPTH_TEST)
	gfx.Viewport(0, 0, windowWidth, windowHeight)
	gfx.ClearColor(0.0, 0.0, 0.0, 1)
	gfx.Clear(graphics.COLOR_BUFFER_BIT | graphics.DEPTH_BUFFER_BIT)

	gfx.BindVertexArray(lens.VAO)
	gfx.UseProgram(lens.Shader.Prog)

	// render left lens
	gfx.ActiveTexture(graphics.TEXTURE0)
	gfx.BindTexture(graphics.TEXTURE_2D, lens.EyeLeft.ResolveTexture)
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_WRAP_S, graphics.CLAMP_TO_EDGE)
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_WRAP_T, graphics.CLAMP_TO_EDGE)
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_MAG_FILTER, graphics.LINEAR)
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_MIN_FILTER, graphics.LINEAR_MIPMAP_LINEAR)
	gfx.DrawElements(graphics.TRIANGLES, (lens.IndexCount / 2), graphics.UNSIGNED_INT, gfx.PtrOffset(0))

	// render right lens
	gfx.BindTexture(graphics.TEXTURE_2D, lens.EyeRight.ResolveTexture)
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_WRAP_S, graphics.CLAMP_TO_EDGE)
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_WRAP_T, graphics.CLAMP_TO_EDGE)
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_MAG_FILTER, graphics.LINEAR)
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_MIN_FILTER, graphics.LINEAR_MIPMAP_LINEAR)
	gfx.DrawElements(graphics.TRIANGLES, (lens.IndexCount / 2), graphics.UNSIGNED_INT, gfx.PtrOffset(int((lens.IndexCount/2)*4))) // uint32size

	gfx.BindVertexArray(0)
	gfx.UseProgram(0)
}
