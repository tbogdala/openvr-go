// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package chunk

// Manager is the landscape manager type structure that manages access to landscape
// data.
type Manager struct {
	// Chunks is the internal splice of all of the chunks registered in the Manager.
	// Access to this will depend on the Size parameter.
	Chunks []*Chunk

	// Size is the number of chunks in each direction to manage.
	Size int

	// X is the location of Manager along the X axis in multiples of ChunkSize.
	// (e.g. a value of Manager.X==2 for ChunkSize==16 means a world position of 32)
	X int

	// Y is the location of Manager along the Y axis in multiples of ChunkSize.
	// (e.g. a value of Manager.Y==2 for ChunkSize==16 means a world position of 32)
	Y int

	// Z is the location of Manager along the Z axis in multiples of ChunkSize.
	// (e.g. a value of Manager.Z==2 for ChunkSize==16 means a world position of 32)
	Z int
}

// NewManager returns a new voxel manager object that manages a cube region
// of chunks.
func NewManager(length int, x, y, z int) *Manager {
	m := new(Manager)
	m.Size = length
	m.X = x
	m.Y = y
	m.Z = z
	m.Chunks = make([]*Chunk, length*length*length)

	return m
}

// RegisterChunk fits the new chunk into the chunks slice of the Manager
func (m *Manager) RegisterChunk(c *Chunk) bool {
	// see if the chunk actually fits within the geography of the Manager.
	localX := c.X - m.X
	localY := c.Y - m.Y
	localZ := c.Z - m.Z
	if localX < 0 || localX >= m.Size || localY < 0 || localY >= m.Size || localZ < 0 || localZ >= m.Size {
		return false
	}

	// calculate the offset in the splice
	offset := localY*m.Size*m.Size + localX*m.Size + localZ
	m.Chunks[offset] = c

	// pwn it
	c.Owner = m

	return true
}

// RegisterChunks registers all of the chunks in the slice passed in.
func (m *Manager) RegisterChunks(cs []*Chunk) {
	for _, chunk := range cs {
		m.RegisterChunk(chunk)
	}
}

// GetChunksFor returns the chunks that 'owns' the world space X,Z coordinate
// passed in. If no chunks are registered for this coordinate, then
// an empty slice is returned. Since the Y axis is not specified, multiple
// chunks can be returned that contain X,Z.
/*
func (m *Manager) GetChunksFor(worldX, worldZ int) []*Chunk {
	chunks :=[]*Chunk{}
	// a lame brute force search through all the chunks
	for _, cY := range m.Strips {
		for _, cX := range cY {
			for _, chunk := range cX {
				if chunk.X*ChunkSize <= worldX && worldX < chunk.X*ChunkSize+ChunkSize &&
					chunk.Z*ChunkSize <= worldZ && worldZ < chunk.Z*ChunkSize+ChunkSize {
					chunks = append(chunks, chunk)
				}
			}
		}
	}
	return chunks
}
*/

// GetHeightAt returns the local height of the land at the local coordinate X,Z passed in.
func (m *Manager) GetHeightAt(localX, localZ int) int {
	size2 := m.Size * m.Size
	size3 := size2 * m.Size
	offset := localX*m.Size + localZ
	answer := 0

	for i := 0; offset+i*size2 < size3; i++ {
		c := m.Chunks[offset+i*size2]
		if c != nil {
			cX := localX - c.X*ChunkSize
			cZ := localZ - c.Z*ChunkSize
			for y := 0; y < ChunkSize; y++ {
				b := c.BlockAt(cX, y, cZ)
				if b != nil && b.Type != 0 {
					answer = y + i*ChunkSize
				}
			}
		}
	}

	return answer
}

// GetBlockAt returns the block at a given coorindate and, as an added bonus
// to faithful callers, it will also return the chunk. A two for one!
// NOTE: restrictions apply! If world xyz doesn't exist in the manger, (nil, nil)
// is returned.
func (m *Manager) GetBlockAt(worldX, worldY, worldZ int) (*Block, *Chunk) {
	// see if the chunk actually fits within the geography of the Manager.
	totalSize := m.Size * ChunkSize
	localX := worldX - m.X*ChunkSize
	localY := worldY - m.Y*ChunkSize
	localZ := worldZ - m.Z*ChunkSize
	if localX < 0 || localX >= totalSize || localY < 0 || localY >= totalSize || localZ < 0 || localZ >= totalSize {
		return nil, nil
	}

	cX := localX / ChunkSize
	cY := localY / ChunkSize
	cZ := localZ / ChunkSize

	offset := cY*m.Size*m.Size + cX*m.Size + cZ
	ownerChunk := m.Chunks[offset]
	if ownerChunk == nil {
		return nil, nil
	}

	block := ownerChunk.BlockAt(localX-cX*ChunkSize, localY-cY*ChunkSize, localZ-cZ*ChunkSize)
	return block, ownerChunk
}

// SetBlockAt sets the block type and color for a given locaiton in the world. It
// will create a new chunk if necessary.
// It returns the chunk for the block that was set.
// NOTE: Does not update colliders. This was left to the calling client so that
// multiple updates at a time don't needlessly recreate colliders.
func (m *Manager) SetBlockAt(worldX, worldY, worldZ int, bt BlockType) *Chunk {
	cX := worldX / ChunkSize
	cY := worldY / ChunkSize
	cZ := worldZ / ChunkSize
	block, chunk := m.GetBlockAt(worldX, worldY, worldZ)

	if chunk != nil {
		block.Type = bt
	} else {
		// the chunk doesn't exist yet, so create a new one
		chunk = NewChunk(cX, cY, cZ)
		m.RegisterChunk(chunk)

		bX := worldX % ChunkSize
		bY := worldY % ChunkSize
		bZ := worldZ % ChunkSize
		chunk.SetBlock(bX, bY, bZ, bt)
	}

	return chunk
}
