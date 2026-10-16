package main

import (
	"reflect"
	"testing"
)

type inner struct {
	Promoted string `json:"promoted"`
}

type Named struct {
	Shared string `json:"shared"`
}

type Outer struct {
	inner
	Named
	Own string `json:"own"`
}

type Other struct {
	Named
}

func TestCollectEmbeddedStructs(t *testing.T) {
	rows := map[[2]string]bool{}
	seen := map[visit]bool{}
	collect("Outer", reflect.TypeOf(Outer{}), rows, seen)
	collect("Other", reflect.TypeOf(Other{}), rows, seen)
	for _, want := range [][2]string{
		{"Outer", "promoted"}, {"Outer", "shared"}, {"Outer", "own"}, {"Other", "shared"},
	} {
		if !rows[want] {
			t.Errorf("missing %v", want)
		}
	}
}
