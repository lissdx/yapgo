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

func (NaivePrimerFinder) endStubFn(monitoringStream pipeline.WriteOnlyStream) pipeline.ProcessFn {
	return func(inObj interface{}) (interface{}, error) {
		monitorMsg := fmt.Sprintf("endStubFn monitor : %v", time.Now().String())
		fmt.Printf("NaivePrimerFinder is pimer: %v\n", inObj)
		monitoringStream <- monitorMsg
		return inObj, nil
	}
}

func main() {
	// Register Int generation function
	randFn := func() interface{} { return rand.Intn(50000000) }
	// Create business-logic object
	primerFinder := NaivePrimerFinder{}
	naiveFinderLine := pipeline.NewPipeline()    // Create main pipeline
	monitoringPipeLine := pipeline.NewPipeline() // Create monitoring pipeline

	doneN := make(chan interface{}) // Create control channel for main pipeline
	doneT := make(chan interface{}) // Create control channel for monitor pipeline

	defer close(doneN)
	defer close(doneT)

	monitorStream := make(chan interface{}) // Create channel for monitoring data
	// Add Stage to monitoring pipeline
	// with slow monitoring function
	monitoringPipeLine.AddStageWithFanOut(func(inObj interface{}) (outObj interface{}, err error) {
		monitorMsg, ok := inObj.(string)
		time.Sleep(time.Second * 3)
		if ok {
			fmt.Printf("Monitoring Fun msg: %s\n", monitorMsg)
		} else {
			fmt.Printf("Monitoring Fun msg: %s\n", "error on assert")
			return nil, fmt.Errorf("can't assert inObj to string: %v", inObj)
		}
		return inObj, nil
	}, primerFinder.ErrorHandler, 100)

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
	naiveFinderLine.AddStageWithFanOut(primerFinder.NaivePrimer(), primerFinder.ErrorHandler, 10)
	naiveFinderLine.AddStage(primerFinder.endStubFn(monitorStream), primerFinder.ErrorHandler)
	intStream := pipeline.Take(doneN, pipeline.RepeatFn(doneN, randFn), 10)

	start := time.Now()
	// Start main pipeline
	doneCh := naiveFinderLine.RunPlug(doneN, intStream)

	// Wait for main pipeline
	<-doneCh
	// Don't forget close monitoring input
	close(monitorStream)
	// and wait monitoring...
	<-mDoneCh
	fmt.Printf("Search took: %v\n", time.Since(start))
}
