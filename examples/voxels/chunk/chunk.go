// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package chunk

import (
	"fmt"

	physics "github.com/tbogdala/cubez"
	physmath "github.com/tbogdala/cubez/math"
	"github.com/tbogdala/fizzle"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
	"github.com/tbogdala/glider"
)

const (
	// ChunkSize controls the cubic dimensions of Chunk structures
	ChunkSize = 16

	// ChunkSizeF is the size of the Chunk in float
	ChunkSizeF = float32(16.0)

	// ChunkSize2 is chunkSize squared
	ChunkSize2 = ChunkSize * ChunkSize

	// ChunkSize3 is chunkSize cubed
	ChunkSize3 = ChunkSize2 * ChunkSize
)

// TODO: these are temporary fixes until the chunk manager has a map of block types to descriptors.
const (
	BlockTypeEmpty  = BlockType(0)
	BlockTypeGrass  = BlockType(1)
	BlockTypeStones = BlockType(2)
	BlockTypeDirt   = BlockType(3)
)

// BlockType is a type alias for the variable that will determine
// what type a landscape Block is in the Chunk.
type BlockType uint32

// Chunk is the data structure that groups together Blocks.
type Chunk struct {
	// X is the location of Chunk along the X axis in multiples of ChunkSize.
	// (e.g. a value of Chunk.X==2 for ChunkSize==16 means a world position of 32)
	X int

	// Y is the location of Chunk along the Y axis in multiples of ChunkSize.
	// (e.g. a value of Chunk.Y==2 for ChunkSize==16 means a world position of 32)
	Y int

	// Z is the location of Chunk along the Z axis in multiples of ChunkSize.
	// (e.g. a value of Chunk.Z==2 for ChunkSize==16 means a world position of 32)
	Z int

	// Blocks is a flat array [y*x*z] Block values. You can think
	// of it in +Z first ordering, meaning it fills a depth of Z, then moves in +X
	// and fills another depth of Z.
	//
	// Viz:
	//  Y
	//  ^ Z
	//  |/__> X
	//
	// Further examples for RegionDepth of 16:
	//   (0,0,0) = [0]
	//   (0,0,5) = [5]
	//   (1,0,0) = [16]
	//   (1,0,5) = [21]
	//   (0,1,0) = [256] (16*16)
	Blocks [ChunkSize3]Block

	// Renderable is the drawable OpenGL object for the chunk. In a server-side
	// only server, this pointer will be nil.
	Renderable *fizzle.Renderable

	// AABBCollider is an axis aligned bounding box collider for the chunk for
	// use in simple collision checks with the space of the entire chunk,
	// regarldless of the contents (i.e. will still collide on 'empty' blocks
	// and contains the entierty of the chunk).
	// Note: coordinates used in the collider are in world space.
	AABBCollider *glider.AABBox

	// colliders is a slice of calculate physics colliders for the chunk geometry.
	// This is generated on calls to UpdateColliders() which should be called
	// when the Blocks array has value changes.
	Colliders []physics.Collider

	// BoxColliders is a slice of AABBoxes that mirrors the physics colliders.
	// This can be used for faster raycasts since the voxels aren't rotating.
	BoxColliders []*glider.AABBox

	// Owner is the owning landscape Manager class. This will be nil if the Chunk
	// has not been registered yet with a manager.
	Owner *Manager
}

// Block is the basic building block of the landscape terrain.
type Block struct {
	// Type determines the type of block that the Block struct represents
	// NOTE: a Type of 0 has special significance in that it's assumed
	// to be an empty space.
	Type BlockType
}

// NewChunk returns a newly created chunk
func NewChunk(chunkX, chunkY, chunkZ int) *Chunk {
	c := new(Chunk)
	c.X = chunkX
	c.Y = chunkY
	c.Z = chunkZ
	c.AABBCollider = glider.NewAABBox()
	c.AABBCollider.Offset = glider.Vec3{float32(c.X) * ChunkSize, float32(c.Y) * ChunkSize, float32(c.Z) * ChunkSize}
	c.AABBCollider.Min = glider.Vec3{0, 0, 0}
	c.AABBCollider.Max = glider.Vec3{ChunkSizeF, ChunkSizeF, ChunkSizeF}
	return c
}

// Clone creates a new Chunk object with the core data but does
// not duplicate the Renderable or physics colliders
func (c *Chunk) Clone() *Chunk {
	newChunk := NewChunk(c.X, c.Y, c.Z)
	newChunk.Blocks = c.Blocks
	return newChunk
}

// Destroy tells the chunk to release any special data.
func (c *Chunk) Destroy() {
	// destroy the renderable if there's one made for the landscape node
	if c.Renderable != nil {
		c.Renderable.Destroy()
	}
}

// GetTheRenderable returns the already crafted Renderable object or makes
// one, caches it and returns a pointer to it.
func (c *Chunk) GetTheRenderable(textureIndexes fizzle.TextureArrayIndexes) *fizzle.Renderable {
	// if we already have one made then return it
	if c.Renderable != nil {
		return c.Renderable
	}

	// build a new one
	c.Renderable = c.buildVoxelRenderable(textureIndexes)
	return c.Renderable
}

// BlockAt returns the Block object at a given offset within the Chunk.
// NOTE: not to be confused with coordinate -- this is the offset in the Cubes array.
func (c *Chunk) BlockAt(x, y, z int) *Block {
	return &c.Blocks[(y*ChunkSize2)+(x*ChunkSize)+z]
}

// SetBlock sets the attribute of a block at a given coordiante. X,Y,Z should
// be within range of [0..ChunkSize-1].
func (c *Chunk) SetBlock(x, y, z int, ty BlockType) {
	block := c.BlockAt(x, y, z)
	block.Type = ty
}

// IsVisible returns true if the block is a type of block that can be visualized normally.
// Basically: if it's not air.
func (b *Block) IsVisible() bool {
	if b.Type > 0 {
		return true
	}
	return false
}

// IsBlockVisible tests whether or not the block is visible. x,y,z are
// specified in local coordinates.
// NOTE: blocks on the outside of the chunk are always considered visible.
func (c *Chunk) IsBlockVisible(x, y, z int) bool {
	// at present, the test for block visibility is if Type > 0.

	// the block itself
	if !c.Blocks[y*ChunkSize2+(x*ChunkSize)+z].IsVisible() {
		return false
	}

	// blocks on the edges are always visible
	if x == 0 || y == 0 || z == 0 || x == ChunkSize-1 || y == ChunkSize-1 || z == ChunkSize-1 {
		return true
	}

	// up
	if !c.Blocks[(y+1)*ChunkSize2+(x*ChunkSize)+z].IsVisible() {
		return true
	}

	// down
	if !c.Blocks[(y-1)*ChunkSize2+(x*ChunkSize)+z].IsVisible() {
		return true
	}

	// left
	if !c.Blocks[y*ChunkSize2+((x+1)*ChunkSize)+z].IsVisible() {
		return true
	}

	// right
	if !c.Blocks[y*ChunkSize2+((x-1)*ChunkSize)+z].IsVisible() {
		return true
	}

	// front
	if !c.Blocks[y*ChunkSize2+(x*ChunkSize)+z+1].IsVisible() {
		return true
	}

	// back
	if !c.Blocks[y*ChunkSize2+(x*ChunkSize)+z-1].IsVisible() {
		return true
	}

	return false
}

// IsBlockMovementBlocking checks to see if a block location blocks movement
// from entities. Having this separate from IsVisible makes the mechanic
// being checked more explicit.
func (c *Chunk) IsBlockMovementBlocking(x, y, z int) bool {
	// Currently, the test is whether or not the type of block
	// is greater than 0 ... basically if there's any block present.
	block := c.BlockAt(x, y, z)
	if block.Type > 0 {
		return true
	}
	return false
}

func (c *Chunk) buildVoxelRenderable(textureIndexes fizzle.TextureArrayIndexes) *fizzle.Renderable {
	var xmax, ymax, zmax float32 = 1.0, 1.0, 1.0
	var xmin, ymin, zmin float32 = 0.0, 0.0, 0.0

	/* Cube vertices are layed out like this:

	  +--------+           6          5
	/ |       /|
	+--------+ |        1          0        +Y
	| |      | |                            |___ +X
	| +------|-+           7          4    /
	|/       |/                           +Z
	+--------+          2          3

	*/

	lookupDirs := [...]int{
		// front
		0, 1, 1, 1, 0, 1, 1, 1, 1, // v0 (front+up, right+front, right+up+front)
		0, 1, 1, -1, 0, 1, -1, 1, 1, // v1 (front+up, left+front, left+up+front)
		0, -1, 1, -1, 0, 1, -1, -1, 1, // v2 (front+bottom, left+front, left+bottom+front)
		0, -1, 1, 1, 0, 1, 1, -1, 1, // v3 (front+bottom, right+front, right+bottom+front)

		// right
		1, 1, 0, 1, 0, -1, 1, 1, -1, // v5 (right+up, right+back, right+up+back)
		1, 1, 0, 1, 0, 1, 1, 1, 1, // v0 (right+up, right+front, right+up+front)
		1, -1, 0, 1, 0, 1, 1, -1, 1, // v3 (right+bottom, right+font, right+bottom+front)
		1, -1, 0, 1, 0, -1, 1, -1, -1, // v4 (right+bottom, right+back, right+bottom+back)

		// top
		0, 1, -1, 1, 1, 0, 1, 1, -1, // v5 (back+up, right+up, right+up+back)
		0, 1, -1, -1, 1, 0, -1, 1, -1, // v6 (back+up, left+up, left+up+back)
		0, 1, 1, -1, 1, 0, -1, 1, 1, // v1 (front+up, left+up, left+up+front)
		0, 1, 1, 1, 1, 0, 1, 1, 1, // v0 (front+up, right+up, right+up+front)

		// left
		-1, 1, 0, -1, 0, 1, -1, 1, 1, // v1 (left+up, left+front, left+up+front)
		-1, 1, 0, -1, 0, -1, -1, 1, -1, // v6 (left+up, left+back, left+up+back)
		-1, -1, 0, -1, 0, -1, -1, -1, -1, // v7 (left+bottom, left+back, left+bottom+back)
		-1, -1, 0, -1, 0, 1, -1, -1, 1, // v2 (left+bottom, left+front, left+bottom+front)

		// bottom
		0, -1, 1, 1, -1, 0, 1, -1, 1, // v3 (front+bottom, right+bottom, right+bottom+front)
		0, -1, 1, -1, -1, 0, -1, -1, 1, // v2 (front+bottom, left+bottom, left+bottom+front)
		0, -1, -1, -1, -1, 0, -1, -1, -1, // v7 (back+bottom, left+bottom, left+bottom+back)
		0, -1, -1, 1, -1, 0, 1, -1, -1, // v4 (back+bottom, right+bottom, right+bottom+back)

		// back
		0, 1, -1, -1, 0, -1, -1, 1, -1, // v6 (back+up, left+back, left+up+back)
		0, 1, -1, 1, 0, -1, 1, 1, -1, // v5 (back+up, right+back, right+up+back)
		0, -1, -1, 1, 0, -1, 1, -1, -1, // v4 (back+bottom, right+back, right+bottom+back)
		0, -1, -1, -1, 0, -1, -1, -1, -1, // v7 (back+bottom, left+back, left+bottom+back)
	}

	verts := [...]float32{
		xmax, ymax, zmax, xmin, ymax, zmax, xmin, ymin, zmax, xmax, ymin, zmax, // v0,v1,v2,v3 (front)
		xmax, ymax, zmin, xmax, ymax, zmax, xmax, ymin, zmax, xmax, ymin, zmin, // v5,v0,v3,v4 (right)
		xmax, ymax, zmin, xmin, ymax, zmin, xmin, ymax, zmax, xmax, ymax, zmax, // v5,v6,v1,v0 (top)
		xmin, ymax, zmax, xmin, ymax, zmin, xmin, ymin, zmin, xmin, ymin, zmax, // v1,v6,v7,v2 (left)
		xmax, ymin, zmax, xmin, ymin, zmax, xmin, ymin, zmin, xmax, ymin, zmin, // v3,v2,v7,v4 (bottom)
		xmin, ymax, zmin, xmax, ymax, zmin, xmax, ymin, zmin, xmin, ymin, zmin, // v6,v5,v4,v7 (back)
	}
	indexes := [...]uint32{
		0, 1, 2, 2, 3, 0,
	}
	uvs := [...]float32{
		1.0, 1.0, 0.0, 1.0, 0.0, 0.0, 1.0, 0.0,
	}

	normals := [...]float32{
		0, 0, 1, 0, 0, 1, 0, 0, 1, 0, 0, 1, // v0,v1,v2,v3 (front)
		1, 0, 0, 1, 0, 0, 1, 0, 0, 1, 0, 0, // v5,v0,v3,v4 (right)
		0, 1, 0, 0, 1, 0, 0, 1, 0, 0, 1, 0, // v5,v6,v1,v0 (top)
		-1, 0, 0, -1, 0, 0, -1, 0, 0, -1, 0, 0, // v1,v6,v7,v2 (left)
		0, -1, 0, 0, -1, 0, 0, -1, 0, 0, -1, 0, // v3,v2,v7,v4 (bottom)
		0, 0, -1, 0, 0, -1, 0, 0, -1, 0, 0, -1, // v6,v5,v4,v7 (back)
	}

	var faceCount uint32
	const comboFloatsPerGrid int = len(verts)
	const comboIntsPerGrid int = 2 // btype, mask
	sectorIndexes := make([]uint32, 0, len(indexes)*ChunkSize3)
	btCombo := make([]uint32, 0, comboIntsPerGrid*4*ChunkSize3)
	vnutBuffer := make([]float32, 0, (comboFloatsPerGrid*2+len(uvs))*ChunkSize3)

	// loop through each block
	for y := 0; y < ChunkSize; y++ {
		for x := 0; x < ChunkSize; x++ {
			for z := 0; z < ChunkSize; z++ {
				worldX := c.X*ChunkSize + x
				worldY := c.Y*ChunkSize + y
				worldZ := c.Z*ChunkSize + z

				// if the block itself is not visible, then just move on
				if c.IsBlockVisible(x, y, z) == false {
					continue
				}

				// get the block
				block := c.BlockAt(x, y, z)

				// process each face on the block separately
				currentFaceCount := 0
				for face := 0; face < 6; face++ {
					// do we need this face? check to see if there's an obstructing block.
					// NOTE: this is done with a lame implementation right now, because
					// going back to the manager isn't efficient.
					switch {
					case face == 0: // 0, 0, +1
						b, _ := c.Owner.GetBlockAt(worldX, worldY, worldZ+1)
						if b != nil && b.Type > 0 {
							continue
						}
					case face == 1: // +1, 0, 0
						b, _ := c.Owner.GetBlockAt(worldX+1, worldY, worldZ)
						if b != nil && b.Type > 0 {
							continue
						}
					case face == 2: // 0, +1, 0
						b, _ := c.Owner.GetBlockAt(worldX, worldY+1, worldZ)
						if b != nil && b.Type > 0 {
							continue
						}
					case face == 3: // -1, 0, 0
						b, _ := c.Owner.GetBlockAt(worldX-1, worldY, worldZ)
						if b != nil && b.Type > 0 {
							continue
						}
					case face == 4: // 0, -1, 0
						b, _ := c.Owner.GetBlockAt(worldX, worldY-1, worldZ)
						if b != nil && b.Type > 0 {
							continue
						}
					case face == 5: // 0, 0, -1
						b, _ := c.Owner.GetBlockAt(worldX, worldY, worldZ-1)
						if b != nil && b.Type > 0 {
							continue
						}
					}

					// time to make the vertices
					baseV := face * 12
					for iv := 0; iv < 4; iv++ {
						iv3 := iv * 3
						iv2 := iv * 2

						// add the vertex
						vnutBuffer = append(vnutBuffer, verts[baseV+iv3]+float32(x))
						vnutBuffer = append(vnutBuffer, verts[baseV+iv3+1]+float32(y))
						vnutBuffer = append(vnutBuffer, verts[baseV+iv3+2]+float32(z))

						// add the normal
						vnutBuffer = append(vnutBuffer, normals[baseV+iv3])
						vnutBuffer = append(vnutBuffer, normals[baseV+iv3+1])
						vnutBuffer = append(vnutBuffer, normals[baseV+iv3+2])

						// add the uv
						vnutBuffer = append(vnutBuffer, uvs[iv2])
						vnutBuffer = append(vnutBuffer, uvs[iv2+1])

						// setup the texture index for the face type, should be per
						// vertex since there's no per-face way of doing it otherwise.
						switch block.Type {
						case BlockTypeGrass:
							btCombo = append(btCombo, uint32(textureIndexes["Grass"]))
						case BlockTypeStones:
							btCombo = append(btCombo, uint32(textureIndexes["Stones"]))
						case BlockTypeDirt:
							btCombo = append(btCombo, uint32(textureIndexes["Dirt"]))
						default:
							fmt.Printf("ERROR: No mapping for block type (%v) to a texture index!\n", block.Type)
						}

						// do some fake AO checks based on the vertex and what face it's in
						vertBitFLags := 0
						lookupOffset := (face * 36) + (iv * 9) // 36 offset numbers per face, 9 offset numbers per vertex
						aoOffsetA := [3]int{lookupDirs[lookupOffset], lookupDirs[lookupOffset+1], lookupDirs[lookupOffset+2]}
						aoOffsetB := [3]int{lookupDirs[lookupOffset+3], lookupDirs[lookupOffset+4], lookupDirs[lookupOffset+5]}
						aoOffsetC := [3]int{lookupDirs[lookupOffset+6], lookupDirs[lookupOffset+7], lookupDirs[lookupOffset+8]}
						aoBlockA, _ := c.Owner.GetBlockAt(worldX+aoOffsetA[0], worldY+aoOffsetA[1], worldZ+aoOffsetA[2])
						aoBlockB, _ := c.Owner.GetBlockAt(worldX+aoOffsetB[0], worldY+aoOffsetB[1], worldZ+aoOffsetB[2])
						aoBlockC, _ := c.Owner.GetBlockAt(worldX+aoOffsetC[0], worldY+aoOffsetC[1], worldZ+aoOffsetC[2])
						if (aoBlockA != nil && aoBlockA.IsVisible()) || (aoBlockB != nil && aoBlockB.IsVisible()) {
							vertBitFLags = vertBitFLags | 0x01
						}
						if aoBlockC != nil && aoBlockC.IsVisible() {
							vertBitFLags = vertBitFLags | 0x02
						}
						btCombo = append(btCombo, uint32(vertBitFLags))

					} // iv

					// time to make the element indeces
					for iv := 0; iv < 6; iv++ {
						sectorIndexes = append(sectorIndexes, indexes[iv]+uint32(currentFaceCount*4)+uint32(faceCount*2))
					}

					// we're not skiping the face, so lets boost the count
					currentFaceCount++
				}

				faceCount += uint32(currentFaceCount) * 2
			} // z
		} // x
	} // y

	// if we didn't make any faces, just stop here and return an empty renderable
	gfx := fizzle.GetGraphics()

	r := fizzle.NewRenderable()
	r.Core = fizzle.NewRenderableCore()
	if faceCount < 1 {
		return r
	}

	r.FaceCount = faceCount
	r.BoundingRect.Top[0] = ChunkSize
	r.BoundingRect.Top[1] = ChunkSize
	r.BoundingRect.Top[2] = ChunkSize
	r.Location[0] = float32(c.X * ChunkSize)
	r.Location[1] = float32(c.Y * ChunkSize)
	r.Location[2] = float32(c.Z * ChunkSize)

	r.Core.DiffuseColor[0] = 1.0
	r.Core.DiffuseColor[1] = 1.0
	r.Core.DiffuseColor[2] = 1.0
	r.Core.DiffuseColor[3] = 1.0
	r.Core.SpecularColor[0] = 1.0
	r.Core.SpecularColor[1] = 1.0
	r.Core.SpecularColor[2] = 1.0
	r.Core.SpecularColor[3] = 1.0
	r.Core.Shininess = 0.00

	// calculate the memory size of floats used to calculate total memory size of float arrays
	const floatSize = 4
	const uintSize = 4

	r.Core.VertVBO = gfx.GenBuffer()
	r.Core.UvVBO = r.Core.VertVBO
	r.Core.NormsVBO = r.Core.VertVBO

	r.Core.VertVBOOffset = 0
	r.Core.NormsVBOOffset = floatSize * 3
	r.Core.UvVBOOffset = floatSize * 6
	r.Core.VBOStride = floatSize * (3 + 3 + 2) // vert / normal / uv
	gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.VertVBO)
	gfx.BufferData(graphics.ARRAY_BUFFER, floatSize*len(vnutBuffer), gfx.Ptr(&vnutBuffer[0]), graphics.STATIC_DRAW)

	// create a VBO to hold the combo data
	r.Core.ComboVBO1 = gfx.GenBuffer()
	gfx.BindBuffer(graphics.ARRAY_BUFFER, r.Core.ComboVBO1)
	gfx.BufferData(graphics.ARRAY_BUFFER, uintSize*len(btCombo), gfx.Ptr(&btCombo[0]), graphics.STATIC_DRAW)

	// create a VBO to hold the face indexes
	r.Core.ElementsVBO = gfx.GenBuffer()
	gfx.BindBuffer(graphics.ELEMENT_ARRAY_BUFFER, r.Core.ElementsVBO)
	gfx.BufferData(graphics.ELEMENT_ARRAY_BUFFER, uintSize*len(sectorIndexes), gfx.Ptr(&sectorIndexes[0]), graphics.STATIC_DRAW)

	return r
}

// UpdateColliders should be called whenever the Blocks of the Chunk
// change which could result in the pathing being different.
func (c *Chunk) UpdateColliders() {
	c.Colliders, c.BoxColliders = c.buildColliders()
}

// collisionBlock is a temporary data structure used to create collision cubes for landscape blocks
type collisionBlock struct {
	X, Y, StartZ, EndZ int
}

// buildColliders will generate the landscape physics and AABB colliders
func (c *Chunk) buildColliders() ([]physics.Collider, []*glider.AABBox) {
	// the value used to see if the z-tracker location is not set
	const unsetZ = -1

	// create the slice to return
	colliders := []collisionBlock{}

	for y := 0; y < ChunkSize; y++ {
		for x := 0; x < ChunkSize; x++ {
			// keep track of the colliders made along the z-axis
			startZ := unsetZ
			for z := 0; z < ChunkSize; z++ {
				if c.IsBlockMovementBlocking(x, y, z) == false {
					// if we never started a block, don't start now, just keep going.
					if startZ == unsetZ {
						continue
					}

					// we end the block of landscape and create the collider cube
					colliders = append(colliders, collisionBlock{x, y, startZ, z - 1})

					// reset the start locator
					startZ = unsetZ
				} else {
					// if we get here it's movement blocking, so set the start position
					// if it's not set already.
					if startZ == unsetZ {
						startZ = z
					}

					// are we at the end of the z-axis? if so, create a collider
					if z == ChunkSize-1 {
						colliders = append(colliders, collisionBlock{x, y, startZ, z})

						// reset the start locator
						startZ = unsetZ
					}
				}
			} // z
		} // x
	} // y

	results := make([]physics.Collider, 0, len(colliders))
	boxResults := make([]*glider.AABBox, 0, len(colliders))
	for _, cb := range colliders {
		cb.X += c.X * ChunkSize
		cb.Y += c.Y * ChunkSize
		cb.StartZ += c.Z * ChunkSize
		cb.EndZ += c.Z * ChunkSize
		results = append(results, buildCollider(&cb))
		boxResults = append(boxResults, buildBoxCollider(&cb))
	}

	return results, boxResults
}

func buildBoxCollider(cb *collisionBlock) *glider.AABBox {
	box := glider.NewAABBox()

	length := float32(cb.EndZ - cb.StartZ + 1)
	halfLength := float32(length) / 2.0

	box.SetOffset3f(float32(cb.X)+0.5, float32(cb.Y)+0.5, float32(cb.StartZ)+halfLength)
	box.Min = glider.Vec3{-0.5, -0.5, -halfLength}
	box.Max = glider.Vec3{0.5, 0.5, halfLength}
	return box
}

func buildCollider(cb *collisionBlock) physics.Collider {
	length := physmath.Real(cb.EndZ - cb.StartZ + 1)
	halfLength := physmath.Real(length) / 2.0

	// create the collision box for the the cube
	cubeCollider := physics.NewCollisionCube(nil, physmath.Vector3{0.5, 0.5, halfLength})
	cubeCollider.Body.Position = physmath.Vector3{physmath.Real(cb.X) + 0.5, physmath.Real(cb.Y) + 0.5, physmath.Real(cb.StartZ) + halfLength}
	cubeCollider.Body.SetInfiniteMass()
	cubeCollider.Body.CanSleep = false
	cubeCollider.Body.CalculateDerivedData()
	cubeCollider.CalculateDerivedData()

	return cubeCollider
}
