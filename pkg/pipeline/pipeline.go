package pipeline

import "sync"

type ReadOnlyStream <-chan interface{}
type WriteOnlyStream chan<- interface{}
type BidirectionalStream chan interface{}

// StageFn is a lower level function type that chains together multiple
// stages using channels.
type StageFn func(doneCh, inStream ReadOnlyStream) ReadOnlyStream

// Pipeline type defines a pipeline to which processing "stages" can
// be added and configured to fan-out. Pipelines are meant to be long
// running as they continuously process data as it comes in.
//
// A pipeline can be simultaneously run multiple times with different
// input channels by invoking the Run() method multiple times.
// A running pipeline shouldn't be copied.
type Pipeline []StageFn

// ProcessFn are the primary function types defined by users of this
// package and passed in to instantiate a meaningful pipeline.
type ProcessFn func(inObj interface{}) (outObj interface{})

// New is a convenience method that creates a new Pipeline
func New() Pipeline {
	return Pipeline{}
}

// AddStage is a convenience method for adding a stage with fanSize = 1.
// See AddStageWithFanOut for more information.
func (p *Pipeline) AddStage(inFunc ProcessFn) {
	*p = append(*p, stageFnFactory(inFunc))
}

// AddStageWithFanOut adds a parallel fan-out ProcessFn to the pipeline. The
// fanSize number indicates how many instances of this stage will read from the
// previous stage and process the data flowing through simultaneously to take
// advantage of parallel CPU scheduling.
//
// Most pipelines will have multiple stages, and the order in which AddStage()
// and AddStageWithFanOut() is invoked matters -- the first invocation indicates
// the first stage and so forth.
//
// Since discrete goroutines process the inChan for FanOut > 1, the order of
// objects flowing through the FanOut stages can't be guaranteed.
func (p *Pipeline) AddStageWithFanOut(inFunc ProcessFn, fanSize uint64) {
	*p = append(*p, fanningStageFnFactory(inFunc, fanSize))
}

// fanningStageFnFactory makes a stage function that fans into multiple
// goroutines increasing the stage throughput depending on the CPU.
func fanningStageFnFactory(inFunc ProcessFn, fanSize uint64) (outFunc StageFn) {
	return func(done ReadOnlyStream, inChan ReadOnlyStream) (outChan ReadOnlyStream) {
		var channels []ReadOnlyStream
		for i := uint64(0); i < fanSize; i++ {
			//channels = append(channels, stageFnFactory(inFunc)(inChan))
			channels = append(channels, stageFnFactory(inFunc)(done, inChan))
		}
		outChan = MergeChannels(done, channels)
		return
	}
}

// MergeChannels merges an array of channels into a single channel. This utility
// function can also be used independently outside of a pipeline.
func MergeChannels(doneCh ReadOnlyStream, inChans []ReadOnlyStream) ReadOnlyStream {
	var wg sync.WaitGroup
	wg.Add(len(inChans))

	outChan := make(chan interface{})
	for _, inChan := range inChans {
		go func(ch <-chan interface{}) {
			defer wg.Done()
			for obj := range OrDone(doneCh, ch) {
				select {
				case <-doneCh:
					return
				case outChan <- obj:
				}

			}
		}(inChan)
	}

	go func() {
		defer close(outChan)
		wg.Wait()
	}()
	return outChan
}

// stageFnFactory makes a standard stage function from a given ProcessFn.
// StageFn functions types accept an inChan and return an outChan, allowing
// us to chain multiple functions into a pipeline.
func stageFnFactory(inFunc ProcessFn) (outFunc StageFn) {
	return func(doneCh, inChan ReadOnlyStream) ReadOnlyStream {
		outChan := make(chan interface{})
		go func() {
			defer close(outChan)
			for inObj := range OrDone(doneCh, inChan) {
				outObj := inFunc(inObj)
				select {
				case <-doneCh:
					return
				case outChan <- outObj:
				}
			}
		}()
		return outChan
	}
}

// Run starts the pipeline with all the stages that have been added. and returns
// the result channel with 'pipelined' data
// We be able to use the channel to connect another pipeline or use the pipelined data
func (p *Pipeline) Run(doneCh, inStream ReadOnlyStream) (outChan ReadOnlyStream) {

	for _, stage := range *p {
		inStream = stage(doneCh, inStream)
	}

	outChan = inStream
	return
}

// RunPlug starts the pipeline with all the stages that have been added. RunPlug is not
// a blocking function and will return immediately with a doneChan. Consumers
// can wait on the doneChan for an indication of when the pipeline has completed
// processing.
//
// The pipeline runs until its `inStream` channel is open. Once the `inStream` is closed,
// the pipeline stages will sequentially complete from the first stage to the last.
// Once all stages are complete, the last outChan is drained and the doneChan is closed.
//
// RunPlug() can be invoked multiple times to start multiple instances of a pipeline
// that will typically process different incoming channels.
func (p *Pipeline) RunPlug(doneCh, inStream ReadOnlyStream) ReadOnlyStream {

	for _, stage := range *p {
		inStream = stage(doneCh, inStream)
	}

	doneChan := make(chan interface{})
	go func() {
		defer close(doneChan)
		for range OrDone(doneCh, inStream) {
			// pull objects from inChan so that the gc marks them
		}
	}()

	return doneChan
}
