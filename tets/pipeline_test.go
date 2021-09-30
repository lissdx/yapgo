package main

import (
	"github.com/lissdx/yapgo/pkg/pipeline"
	"testing"
)

func TestGenerator(t *testing.T) {
	tests := []struct {
		name string
		args []interface{}
		want []interface{}
	}{
		{name: "empty", args: []interface{}{}, want: []interface{}{}},
		{name: "oneInt", args: []interface{}{1}, want: []interface{}{1}},
		{name: "twoInt", args: []interface{}{1, -1}, want: []interface{}{1, -1}},
		{name: "tenInt", args: []interface{}{1, -1, 23, 100, 26, 33, 78, 33, 221, 0}, want: []interface{}{1, -1, 23, 100, 26, 33, 78, 33, 221, 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			done := make(chan interface{})
			defer close(done)

			var got []interface{}
			valStream := pipeline.Generator(done, tt.args...)
			for num := range valStream {
				got = append(got, num)
			}

			if len(got) != len(tt.want) {
				t.Errorf("Generator = %v, want %v", got, tt.want)
			}

			for i, wantVal := range tt.want {
				if got[i] != wantVal {
					t.Errorf("Generator = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestTake(t *testing.T) {
	tests := []struct {
		name      string
		takeTimes int
		args      []interface{}
		want      []interface{}
	}{
		{name: "negativeTakeSrcEmpty", takeTimes: -1, args: []interface{}{}, want: []interface{}{}},
		{name: "zeroTakeSrcEmpty", takeTimes: 0, args: []interface{}{}, want: []interface{}{}},
		{name: "negativeTake", takeTimes: -1, args: []interface{}{1, 2, 3}, want: []interface{}{}},
		{name: "zeroTake", takeTimes: 0, args: []interface{}{1, 2, 3}, want: []interface{}{}},
		{name: "takeOne", takeTimes: 1, args: []interface{}{1, 2, 3}, want: []interface{}{1}},
		{name: "takeTwo", takeTimes: 2, args: []interface{}{1, 2, 3}, want: []interface{}{1, 2}},
		{name: "take5FromShortSrc", takeTimes: 5, args: []interface{}{1, 2, 3}, want: []interface{}{1, 2, 3}},
		//{name: "oneInt", args: []interface{}{1}, want: []interface{}{1}},
		//{name: "twoInt", args: []interface{}{1,-1}, want: []interface{}{1,-1}},
		//{name: "tenInt", args: []interface{}{1,-1, 23, 100, 26, 33, 78, 33, 221, 0}, want: []interface{}{1,-1, 23, 100, 26, 33, 78, 33, 221, 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			done := make(chan interface{})
			defer close(done)

			var got []interface{}
			valStream := pipeline.Take(done, pipeline.Generator(done, tt.args...), tt.takeTimes)
			for num := range valStream {
				got = append(got, num)
			}

			if len(got) != len(tt.want) {
				t.Errorf("Generator = %v, want %v", got, tt.want)
			}

			for i, wantVal := range tt.want {
				if got[i] != wantVal {
					t.Errorf("Generator = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
