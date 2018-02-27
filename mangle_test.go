package main

import "testing"

func TestMangleCtor(t *testing.T) {
	testcases := [][4]string{
		[4]string{"newValOf", "Val", "Int", "newIntOf"},
		[4]string{"make", "Val", "Int", "makeInt"},
		[4]string{"Make", "Val", "int", "MakeInt"},
		[4]string{"ValOfTea", "val", "Int", "IntOfTea"},
	}

	for _, tc := range testcases {
		got := MangleCtorName(tc[0], tc[1], tc[2])
		if got != tc[3] {
			t.Error(got, tc)
		}
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
		got := MangleDepTypeName(tc[0], tc[1], tc[2])
		if got != tc[3] {
			t.Error(got, tc)
		}
	}
}
