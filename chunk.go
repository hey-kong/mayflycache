package main

// Chunk implements the Value interface, as the value
// of the key-value entry in the cache, it's read-only.
type Chunk struct {
	b []byte
}

// NewChunk returns a new Chunk for a byte slice.
func NewChunk(b []byte) Chunk {
	return Chunk{b: cloneBytes(b)}
}

// Size returns the length of the byte slice in the chunk.
func (c Chunk) Size() int {
	return len(c.b)
}

// String returns a string of internal byte slice.
func (c Chunk) String() string {
	return string(c.b)
}

// ByteSlice returns a copy of the byte slice in the chunk.
func (c Chunk) ByteSlice() []byte {
	return cloneBytes(c.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
