package main

import (
	"fmt"
	"math/rand"
	"time"
	"yapgo/pkg/pipeline"
)

type Primer struct {
	IsPrimer bool
	integer  int
}

type NaivePrimerFinder struct{}

func (NaivePrimerFinder) isPrime(integer int) bool {
	isPrime := true
	for divisor := integer - 1; divisor > 1; divisor-- {
		if integer%divisor == 0 {
			isPrime = false
			break
		}
	}
	return isPrime
}

func (n NaivePrimerFinder) NaivePrimer() pipeline.ProcessFn {
	return func(inObj interface{}) interface{} {
		intVal, ok := inObj.(int)
		if !ok {
			return Primer{integer: intVal, IsPrimer: false}
		}
		return Primer{integer: intVal, IsPrimer: n.isPrime(intVal)}
	}
}

func (NaivePrimerFinder) endStubFn() pipeline.ProcessFn {
	return func(inObj interface{}) interface{} {
		fmt.Printf("NaivePrimerFinder is pimer: %v\n", inObj)
		return inObj
	}
}

func main() {
	rand := func() interface{} { return rand.Intn(50000000) }
	primerFinder := NaivePrimerFinder{}
	pLine := pipeline.Pipeline{}
	done := make(chan interface{})
	defer close(done)

	pLine.AddStageWithFanOut(primerFinder.NaivePrimer(), 10)
	pLine.AddStage(primerFinder.endStubFn())

	intStream := pipeline.Take(done, pipeline.RepeatFn(done, rand), 100)

	start := time.Now()
	doneCh := pLine.RunPlug(done, intStream)
	<-doneCh
	fmt.Printf("Search took: %v\n", time.Since(start))
}
