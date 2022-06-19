# YAPGo
Yet Another Pipeline Golang - is a simple package provides one of possible  
GO's pipeline solution.

This project is kind of compilation of knowledge from
References:  
[Concurrency in Go.book](https://s1.phpcasts.org/Concurrency-in-Go_Tools-and-Techniques-for-Developers.pdf)  
[Go Dev Topic](https://go.dev/blog/pipelines)  
[Go Dev pkg src](https://github.com/hyfather/pipeline)  

### Description
Pipeline is a go library that helps you build pipelines without worrying about channel  
management and concurrency.  
It contains common **fan-in**, **fan-out** **filter** **marge** operations  
as well as useful utility funcs for generation splitting etc.

[Data Pipeline](pkg/pipeline/pipeline.go)  
[Data Pipeline Additional](pkg/pipeline/operator.go)

### How to use
More examples:  
[Data Pipeline Functional Approach](examples/prime_finder_no_fan_out/primer_finder.go)    
[Data Pipeline Obj. Approach ](examples/prime_finder_with_fan_out/primer_finder.go)  
[Data Pipeline Filter Approach ](examples/prime_finder_filter_with_fan_out/primer_finder.go)

```golang
package main

import (
	"fmt"
	"math/rand"
	"time"
)

// We can easily change approach from functional to obj
// see: prime_finder_with_fan_out
// Data pipeline simple example of func approach

type Primer struct {
	IsPrimer bool
	integer  int
}

type NaivePrimerFinder struct{}

func (n NaivePrimerFinder) ErrorHandler(err error) {
	fmt.Printf("pipeline error: %s \n", err.Error())
}

func (n NaivePrimerFinder) NaivePrimer() pipeline.ProcessFn {
	return func(inObj interface{}) (interface{}, error) {
		intVal, ok := inObj.(int)
		if !ok {
			return nil, fmt.Errorf("can't assert inObj: %v", Primer{integer: intVal, IsPrimer: false})
		}
		return Primer{integer: intVal, IsPrimer: n.isPrime(intVal)}, nil
	}
}

func (NaivePrimerFinder) endStubFn() pipeline.ProcessFn {
	return func(inObj interface{}) (interface{}, error) {
		fmt.Printf("NaivePrimerFinder is pimer: %v\n", inObj)
		return inObj, nil
	}
}

func main() {
	// Num generator func
	randNum := func() interface{} { return rand.Intn(50000000) }
	// New PrimeFinder (Finder) object
	primeFinder := prime.NewNaivePrimeFinder()
	// New Data Pipeline
	pLine := pipeline.NewPipeline()
	// Control channel
	done := make(chan interface{})
	defer close(done)

	// Add first stage to pipeline
	pLine.AddStage(naivePrimeStgFn(primeFinder), errorHandler)
	// Add last stage to pipeline
	pLine.AddStage(endStubFn(), errorHandler)
	// Start num generation
	// Repeat generator call and take
	// first 100 generated nums
	intStream := pipeline.Take(done, pipeline.RepeatFn(done, randNum), 100)

	start := time.Now()
	// Start pipeline
	doneCh := pLine.RunPlug(done, intStream)

	// Wait ...
	<-doneCh

	fmt.Printf("Search took: %v\n", time.Since(start))
}

// ---------------------------------------------
// Functional approach
// --------------------------------------------

// errorHandler just error handler we will use
// in or pipeline
func errorHandler(err error) {
	fmt.Printf("pipeline error: %s \n", err.Error())
}

// naivePrimeStgFn pipeline's staging function wraps
// our process function Finder.IsPrime
func naivePrimeStgFn(finder prime.Finder) pipeline.ProcessFn {
	return func(inObj interface{}) (interface{}, error) {
		intVal, ok := inObj.(int)
		if !ok {
			return nil, fmt.Errorf("can't assert inObj: %v", Primer{integer: intVal, IsPrimer: false})
		}
		return Primer{integer: intVal, IsPrimer: finder.IsPrime(intVal)}, nil
	}
}

// endStubFn just end stub func for report
func endStubFn() pipeline.ProcessFn {
	return func(inObj interface{}) (interface{}, error) {
		fmt.Printf("NaivePrimeFinder is pime: %v\n", inObj)
		return inObj, nil
	}
}
```