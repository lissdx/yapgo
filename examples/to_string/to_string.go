package main

import (
	"fmt"
	"yapgo/pkg/pipeline"
)

// ToString small example that introduces a toString pipeline stage:
func ToString(doneCh pipeline.ReadOnlyStream, valueStream pipeline.ReadOnlyStream) <-chan string {
	stringStream := make(chan string)
	go func() {
		defer close(stringStream)
		for v := range valueStream {
			select {
			case <-doneCh:
				return
			case stringStream <- v.(string):
			}
		}
	}()
	return stringStream
}

func main() {
	done := make(chan interface{})
	defer close(done)

	valStream := pipeline.Repeat(done, "I", "am.")
	for num := range ToString(done, pipeline.Take(done, valStream, 10)) {
		fmt.Printf("%s \n", num)
	}
}
