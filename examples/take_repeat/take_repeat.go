package main

import (
	"fmt"
	"github.com/lissdx/yapgo/pkg/pipeline"
)

func main() {
	done := make(chan interface{})
	defer close(done)

	valStream := pipeline.Repeat(done, 1, 2, 3)
	for num := range pipeline.Take(done, valStream, 10) {
		fmt.Printf("%v ", num)
	}
}
