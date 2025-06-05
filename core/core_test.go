package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/packages"
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

func TestPkgAboveTag(t *testing.T) {
	p1 := Package{
		Pkg: &packages.Package{
			PkgPath: "github.com/igadmg/gogen/core",
		},
	}
	p2 := Package{
		Pkg: &packages.Package{
			PkgPath: "github.com/igadmg/gogen/core/gfx",
		},
	}
	p3 := Package{
		Pkg: &packages.Package{
			PkgPath: "github.com/igadmg/gogen/gridcity",
		},
	}

	assert.True(t, p1.Above(&p2))
	assert.False(t, p2.Above(&p1))
	assert.True(t, !p3.Above(&p1))
	assert.True(t, !p1.Above(&p3))
}
