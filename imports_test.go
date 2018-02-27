package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImportsAdd(t *testing.T) {
	im := Imports{}
	assert.NoError(t, im.Add("n1", "p1"))
	assert.NoError(t, im.Add("n1", "p1"))
	assert.Error(t, im.Add("n1", "p2"))
	assert.NoError(t, im.Add("n2", "p2"))
	assert.Error(t, im.Add(".", "some"))
	assert.Error(t, im.Add("", "some"))

	assert.Equal(t, im.p2n["p2"], "n2")
	assert.Equal(t, im.n2p["n1"], "p1")
	assert.Equal(t, im.Named("n1"), "p1")
}

func TestImportsMerge(t *testing.T) {
	im1 := Imports{}
	im1.Add("n1", "p1")
	im1.Add("n2", "p1")

	im2 := Imports{}
	im2.Add("n1", "p3")

	rename1 := im1.Merge(im2)

	assert.Len(t, rename1, 1)
	_, has := rename1["n1"]
	assert.True(t, has)
	_, has = im1.p2n["p3"]
	assert.True(t, has)
}
