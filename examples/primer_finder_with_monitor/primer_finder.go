package main

import (
	"fmt"
	"github.com/lissdx/yapgo/pkg/pipeline"
	"math/rand"
	"time"
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

func (NaivePrimerFinder) endStubFn(monitoringStream pipeline.WriteOnlyStream) pipeline.ProcessFn {
	return func(inObj interface{}) interface{} {
		monitorMsg := fmt.Sprintf("endStubFn monitor : %v", time.Now().String())
		fmt.Printf("NaivePrimerFinder is pimer: %v\n", inObj)
		monitoringStream <- monitorMsg
		return inObj
	}
}

func main() {
	// Register Int generation function
	randFn := func() interface{} { return rand.Intn(50000000) }
	// Create business-logic object
	primerFinder := NaivePrimerFinder{}
	naiveFinderLine := pipeline.New() // Create main pipeline
	monitoringPipeLine := pipeline.New() // Create monitoring pipeline

	doneN := make(chan interface{}) // Create control channel for main pipeline
	doneT := make(chan interface{})  // Create control channel for monitor pipeline


	defer close(doneN)
	defer close(doneT)

	monitorStream := make(chan interface{}) // Create channel for monitoring data
	// Add Stage to monitoring pipeline
	// with slow monitoring function
	monitoringPipeLine.AddStageWithFanOut(func(inObj interface{}) (outObj interface{}) {
		monitorMsg, ok := inObj.(string)
		time.Sleep(time.Second * 3)
		if ok {
			fmt.Printf("Monitoring Fun msg: %s\n", monitorMsg)
		}
		return inObj
	}, 100)

	//monitorStream := make(chan interface{}, 100) // Create channel for monitoring data
	//monitoringPipeLine.AddStage(func(inObj interface{}) (outObj interface{}) {
	//	monitorMsg, ok := inObj.(string)
	//	time.Sleep(time.Second * 3)
	//	if ok {
	//		fmt.Printf("Monitoring msg: %s\n", monitorMsg)
	//	}
	//	return inObj
	//})

	// Run monitoring pipeline and hold done channel
	// to wait the pipeline
	mDoneCh := monitoringPipeLine.RunPlug(doneT, monitorStream)

	// Create main working channel
	naiveFinderLine.AddStageWithFanOut(primerFinder.NaivePrimer(), 10)
	naiveFinderLine.AddStage(primerFinder.endStubFn(monitorStream))
	intStream := pipeline.Take(doneN, pipeline.RepeatFn(doneN, randFn), 10)

	start := time.Now()
	// Start main pipeline
	doneCh := naiveFinderLine.RunPlug(doneN, intStream)

	// Wait for main pipeline
	<-doneCh
	// Don't forget close monitoring input
	close(monitorStream)
	// and wait monitoring...
	<- mDoneCh
	fmt.Printf("Search took: %v\n", time.Since(start))
}