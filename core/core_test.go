package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalTag(t *testing.T) {
	{
		tag, err := UnmarshalTag("")
		assert.NoError(t, err)
		assert.Equal(t, len(tag), 0)
	}
	{
		tag, err := UnmarshalTag("e.DrawBackground")
		assert.NoError(t, err)
		assert.Equal(t, len(tag), 1)
	}
	{
		tag, err := UnmarshalTag("colony, cursor")
		assert.NoError(t, err)
		assert.Equal(t, len(tag), 2)
	}
	{
		tag, err := UnmarshalTag("layer: LayerColonyBuildings")
		assert.NoError(t, err)
		assert.Equal(t, len(tag), 1)
	}
	{
		tag, err := UnmarshalTag("prepare: 'LayerPopulation, o.Draw'")
		assert.NoError(t, err)
		assert.Equal(t, len(tag), 1)
	}

	{
		tag, err := UnmarshalTag("new: e.Draw, prepare: { Layer: 'Named(LayerPopulation)' }")
		assert.NoError(t, err)
		assert.Equal(t, len(tag), 2)
	}
}
