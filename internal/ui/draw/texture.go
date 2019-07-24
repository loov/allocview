package draw

import (
	"image"
)

type Textures struct {
	TextureByID      map[TextureID]*Texture
	TextureByPointer map[*image.NRGBA]*Texture
	NextTextureID    TextureID
}

func NewTextures() *Textures {
	return &Textures{
		TextureByID:      map[TextureID]*Texture{},
		TextureByPointer: map[*image.NRGBA]*Texture{},
		NextTextureID:    1,
	}
}

type Texture struct {
	ID    TextureID
	Image *image.NRGBA
	GPU   map[DriverTag]DriverInfo
}

type DriverTag interface{}

type DriverInfo interface {
	Delete()
	Invalidate()
}

func (textures *Textures) IncludeTexture(m *image.NRGBA, dirty bool) TextureID {
	tex, ok := textures.TextureByPointer[m]
	if !ok {
		tex = NewTexture(m)
		tex.ID = textures.NextTextureID
		textures.NextTextureID++

		textures.TextureByID[tex.ID] = tex
		textures.TextureByPointer[m] = tex
	}

	if dirty {
		tex.Invalidate()
	}

	return tex.ID
}

func NewTexture(m *image.NRGBA) *Texture {
	return &Texture{
		Image: m,
		GPU:   map[DriverTag]DriverInfo{},
	}
}

func (texture *Texture) Invalidate() {
	for _, info := range texture.GPU {
		info.Invalidate()
	}
}

func (texture *Texture) Delete() {
	for _, info := range texture.GPU {
		info.Delete()
	}
}
