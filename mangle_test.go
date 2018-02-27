package main

import "testing"
import "github.com/stretchr/testify/assert"

func TestMangleCtor(t *testing.T) {
	testcases := [][4]string{
		[4]string{"newValOf", "Val", "Int", "newIntOf"},
		[4]string{"make", "Val", "Int", "makeInt"},
		[4]string{"Make", "Val", "int", "MakeInt"},
		[4]string{"ValOfTea", "val", "Int", "IntOfTea"},
	}

	for _, tc := range testcases {
		assert.Equal(t, tc[3], MangleCtorName(tc[0], tc[1], tc[2]))
	}
}

func TestMangleDepType(t *testing.T) {
	testcases := [][4]string{
		[4]string{"valEntry", "Val", "Int", "intEntry"},
		[4]string{"entry", "val", "int", "intEntry"},
		[4]string{"ValEntry", "Val", "zzz", "ZzzEntry"},
		[4]string{"EntryValName", "Val", "zzz", "EntryZzzName"},
		[4]string{"entryValName", "Val", "X", "entryXName"},
	}

	for _, tc := range testcases {
		assert.Equal(t, tc[3], MangleDepTypeName(tc[0], tc[1], tc[2]))
	}
}
