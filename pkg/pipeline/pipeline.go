package pipeline

// StageFn is a lower level function type that chains together multiple
// stages using channels.
type StageFn func(done <-chan interface{}, inChan <-chan interface{}, err chan <- interface{}) (outChan chan interface{})


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
type ProcessFn func(inObj interface{}) (outObj interface{}, err error)

// New is a convenience method that creates a new Pipeline
func New() Pipeline {
	return Pipeline{}
}

// AddStage is a convenience method for adding a stage with fanSize = 1.
// See AddStageWithFanOut for more information.
func (p *Pipeline) AddStage(inFunc ProcessFn) {
	*p = append(*p, stageFnFactory(inFunc))
}

// stageFnFactory makes a standard stage function from a given ProcessFn.
// StageFn functions types accept an inChan and return an outChan, allowing
// us to chain multiple functions into a pipeline.
func stageFnFactory(inFunc ProcessFn) (outFunc StageFn) {
	return func(done <-chan interface{}, inChan <-chan interface{}, errStream chan <- interface{}) (outChan chan interface{}) {
		outChan = make(chan interface{})
		go func() {
			defer close(outChan)
			for inObj := range OrDone(done, inChan) {
				//for inObj := range inChan{
				outObj, err := inFunc(inObj)
				if err != nil {
					errStream <- err
					continue
				}

				select {
				case <-done:
					return
				case outChan <- outObj:
				//case outChan <- inFunc(inObj):
					//default:
					//	if outObj := inFunc(inObj); outObj != nil {
					//		outChan <- outObj
					//	}
				}
			}
		}()
		return
	}
}

// Run starts the pipeline with all the stages that have been added. Run is not
// a blocking function and will return immediately with a doneChan. Consumers
// can wait on the doneChan for an indication of when the pipeline has completed
// processing.
//
// The pipeline runs until its `inChan` channel is open. Once the `inChan` is closed,
// the pipeline stages will sequentially complete from the first stage to the last.
// Once all stages are complete, the last outChan is drained and the doneChan is closed.
//
// Run() can be invoked multiple times to start multiple instances of a pipeline
// that will typically process different incoming channels.
func (p *Pipeline) Run(done <-chan interface{}, inChan <-chan interface{}, err  chan <- interface{}) <- chan interface{} {
	//func (p *Pipeline) Run(done <-chan interface{}, inChan <-chan interface{}) (doneChan chan struct{}) {
	for _, stage := range *p {
		inChan = stage(done, inChan, err)
	}

	//doneChan = make(chan struct{})
	//go func() {
	//	defer close(doneChan)
	//	for range inChan {
	//		// pull objects from inChan so that the gc marks them
	//	}
	//}()
	return inChan
}

func (p *Pipeline) RunPlug(done <-chan interface{}, inChan <-chan interface{}, err  chan <- interface{}) (doneChan chan interface{}) {
	for _, stage := range *p {
		inChan = stage(done, inChan, err)
	}

	doneChan = make(chan interface{})
	go func() {
		defer close(doneChan)
		for range inChan {
			// pull objects from inChan so that the gc marks them
		}
	}()

	return
}

// OrDone
// wrap our read from the channel with a select statement that
// also selects from a done channel.
// for val := range orDone(done, myChan) {
//   ...Do something with val
// }
func OrDone(done, c <-chan interface{}) <-chan interface{} {
	valStream := make(chan interface{})
	go func() {
		defer close(valStream)
		for {
			select {
			case <-done:
				return
			case v, ok := <-c:
				if ok == false {
					return
				}
				select {
				case valStream <- v:
				case <-done:
					return // I've added return here
				}
			}
		}
	}()
	return valStream
}

func Tee(done <-chan interface{}, in <-chan interface{}) (chan interface{}, chan interface{}) {

	out1 := make(chan interface{})
	out2 := make(chan interface{})
	go func() {
		defer close(out1)
		defer close(out2)
		for val := range OrDone(done, in) {
			var hOut1, hOut2 = out1, out2
			for i := 0; i < 2; i++ {
				select {
				case <-done:
				case hOut1 <- val:
					hOut1 = nil
				case hOut2 <- val:
					hOut2 = nil
				}
			}
		}
	}()
	return out1, out2
}