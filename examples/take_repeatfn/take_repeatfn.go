package main

import (
	"fmt"
	"github.com/lissdx/yapgo/pkg/pipeline"
	"math/rand"
)

func main() {
	done := make(chan interface{})
	defer close(done)

	rand := func() interface{} { return rand.Int() }

	valStream := pipeline.RepeatFn(done, rand)
	for num := range pipeline.Take(done, valStream, 10) {
		fmt.Printf("%v \n", num)
	}
}
