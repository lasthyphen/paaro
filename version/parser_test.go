// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDefaultVersionParser(t *testing.T) {
	p := NewDefaultParser()

	v, err := p.Parse("v1.2.3")

	assert.NoError(t, err)
	assert.NotNil(t, v)
	assert.Equal(t, "v1.2.3", v.String())
	assert.Equal(t, 1, v.Major())
	assert.Equal(t, 2, v.Minor())
	assert.Equal(t, 3, v.Patch())

	badVersions := []string{
		"",
		"1.2.3",
		"vz.2.3",
		"v1.z.3",
		"v1.2.z",
	}
	for _, badVersion := range badVersions {
		_, err := p.Parse(badVersion)
		assert.Error(t, err)
	}
}

func TestNewDefaultApplicationParser(t *testing.T) {
	p := NewDefaultApplicationParser()

	v, err := p.Parse("dijets/1.2.3")

	assert.NoError(t, err)
	assert.NotNil(t, v)
	assert.Equal(t, "dijets/1.2.3", v.String())
	assert.Equal(t, "dijets", v.App())
	assert.Equal(t, 1, v.Major())
	assert.Equal(t, 2, v.Minor())
	assert.Equal(t, 3, v.Patch())
	assert.NoError(t, v.Compatible(v))
	assert.False(t, v.Before(v))

	badVersions := []string{
		"",
		"dijets/",
		"dijets/z.0.0",
		"dijets/0.z.0",
		"dijets/0.0.z",
	}
	for _, badVersion := range badVersions {
		_, err := p.Parse(badVersion)
		assert.Error(t, err)
	}
}

func TestNewApplicationParser(t *testing.T) {
	p := NewApplicationParser(":", ",")

	v, err := p.Parse("dijets:1,2,3")

	assert.NoError(t, err)
	assert.NotNil(t, v)
	assert.Equal(t, "dijets:1,2,3", v.String())
	assert.Equal(t, "dijets", v.App())
	assert.Equal(t, 1, v.Major())
	assert.Equal(t, 2, v.Minor())
	assert.Equal(t, 3, v.Patch())
	assert.NoError(t, v.Compatible(v))
	assert.False(t, v.Before(v))
}
