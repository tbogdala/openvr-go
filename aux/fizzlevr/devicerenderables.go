// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package fizzlevr

import (
	"fmt"

	fizzle "github.com/tbogdala/fizzle"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
	vr "github.com/tbogdala/openvr-go"
)

// DeviceRenderables creates Renderable objects for connected devices.
type DeviceRenderables struct {
	// vrSystem is the cached reference to the ISystem interface
	vrSystem *vr.System

	// vrRenderModels is the cached reference to the IRenderModels interface
	vrRenderModels *vr.RenderModels

	// Shader is the render model shader to use
	Shader *fizzle.RenderShader

	// renderables are the loaded renderables for devices
	renderables map[string]*fizzle.Renderable
}

// CreateDeviceRenderables creates a new DeviceRenderables object which creates
// Renderable objects for each connected device.
func CreateDeviceRenderables(vrSystem *vr.System, shader *fizzle.RenderShader) (*DeviceRenderables, error) {
	deviceRenderables := new(DeviceRenderables)
	deviceRenderables.Shader = shader
	deviceRenderables.vrSystem = vrSystem

	// get the render models interface
	var err error
	deviceRenderables.vrRenderModels, err = vr.GetRenderModels()
	if err != nil {
		return nil, err
	}

	// create the map to cache the renderables
	deviceRenderables.renderables = make(map[string]*fizzle.Renderable)

	// loop through all possible devices besides the first, which is the HMD,
	// and try to load the model.
	/*
		for i := vr.TrackedDeviceIndexHmd + 1; i < vr.MaxTrackedDeviceCount; i++ {
			if vrSystem.IsTrackedDeviceConnected(uint32(i)) {
				_, err := deviceRenderables.GetRenderableForTrackedDevice(int(i))
				if err != nil {
					return nil, fmt.Errorf("Failed to load renderable for device index %d; %v\n", i, err)
				}
			}
		}
	*/

	return deviceRenderables, nil
}

// GetRenderableForTrackedDevice will look up the tracked device and create a
// renderable if one hasn't been cached already.
func (dr *DeviceRenderables) GetRenderableForTrackedDevice(deviceIndex int) (*fizzle.Renderable, error) {
	// sanity check
	if uint(deviceIndex) >= vr.MaxTrackedDeviceCount {
		return nil, fmt.Errorf("Device index out of range.")
	}

	// get the name of the device
	fmt.Printf("dr == %v\n\n", dr)
	rendermodelName, errInt := dr.vrSystem.GetStringTrackedDeviceProperty(deviceIndex, vr.PropRenderModelNameString)
	if errInt != vr.TrackedPropSuccess {
		return nil, fmt.Errorf("%s", vr.GetErrorAsEnglish(errInt))
	}

	// return a cached copy if there is one
	existingRenderable, okay := dr.renderables[rendermodelName]
	if okay {
		return existingRenderable, nil
	}

	// no cached copy, so load a new one
	renderModel, err := dr.vrRenderModels.RenderModelLoad(rendermodelName)
	if err != nil {
		return nil, err
	}

	// as a test, make a renderable with the data
	const floatSize = 4
	const uintSize = 4
	r := fizzle.NewRenderable()
	r.Core = fizzle.NewRenderableCore()
	r.FaceCount = renderModel.TriangleCount
	r.Core.Shader = dr.Shader

	// create a VBO to hold the vertex data
	gfx := fizzle.GetGraphics()
	r.Core.VertVBO = gfx.GenBuffer()
	r.Core.UvVBO = r.Core.VertVBO
	r.Core.NormsVBO = r.Core.VertVBO
	r.Core.VertVBOOffset = 0
	r.Core.NormsVBOOffset = floatSize * 3
	r.Core.UvVBOOffset = floatSize * 6
	r.Core.VBOStride = floatSize * (3 + 3 + 2) // vert / normal / uv
	gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.VertVBO)
	gfx.BufferData(graphics.ARRAY_BUFFER, floatSize*len(renderModel.VertexData), gfx.Ptr(&renderModel.VertexData[0]), graphics.STATIC_DRAW)

	// create a VBO to hold the face indexes
	r.Core.ElementsVBO = gfx.GenBuffer()
	gfx.BindBuffer(graphics.ELEMENT_ARRAY_BUFFER, r.Core.ElementsVBO)
	gfx.BufferData(graphics.ELEMENT_ARRAY_BUFFER, uintSize*len(renderModel.Indexes), gfx.Ptr(&renderModel.Indexes[0]), graphics.STATIC_DRAW)

	// upload the texture
	r.Core.Tex0 = gfx.GenTexture()
	gfx.ActiveTexture(graphics.TEXTURE0)
	gfx.BindTexture(graphics.TEXTURE_2D, r.Core.Tex0)

	gfx.TexImage2D(graphics.TEXTURE_2D, 0, graphics.RGBA, int32(renderModel.TextureWidth), int32(renderModel.TextureHeight),
		0, graphics.RGBA, graphics.UNSIGNED_BYTE, gfx.Ptr(renderModel.TextureBytes), len(renderModel.TextureBytes))

	// If this renders black ask yourself what's wrong. ;p
	gfx.GenerateMipmap(graphics.TEXTURE_2D)

	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_MAG_FILTER, graphics.LINEAR)
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_MIN_FILTER, graphics.LINEAR_MIPMAP_LINEAR)
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_WRAP_S, graphics.CLAMP_TO_EDGE)
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_WRAP_T, graphics.CLAMP_TO_EDGE)
	/*
		GLfloat fLargest;
		glGetFloatv( GL_MAX_TEXTURE_MAX_ANISOTROPY_EXT, &fLargest );
		glTexParameterf( GL_TEXTURE_2D, GL_TEXTURE_MAX_ANISOTROPY_EXT, fLargest );
	*/
	gfx.BindTexture(graphics.TEXTURE_2D, 0)

	// store the renderable
	dr.renderables[rendermodelName] = r

	// return the new renderable
	return r, nil
}
