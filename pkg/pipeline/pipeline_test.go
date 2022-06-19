package pipeline

import (
	"fmt"
	"sort"
	"sync"
	"testing"
	"time"
	bufferLogger "yapgo/pkg/test/infra"
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
			valStream := Generator(done, tt.args...)
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			done := make(chan interface{})
			defer close(done)

			var got []interface{}
			valStream := Take(done, Generator(done, tt.args...), tt.takeTimes)
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

func TestMergeChannels(t *testing.T) {
	tests := []struct {
		name string
		args [][]interface{}
		want []interface{}
	}{
		{name: "empty nil", args: [][]interface{}{}, want: []interface{}{}},
		{name: "empty 1", args: [][]interface{}{{}}, want: []interface{}{}},
		{name: "empty 2", args: [][]interface{}{{}, {}}, want: []interface{}{}},
		{name: "empty 4", args: [][]interface{}{{}, {}, {}, {}}, want: []interface{}{}},
		{name: "Elements 1 in 1", args: [][]interface{}{{1}}, want: []interface{}{1}},
		{name: "Elements 2 in 1", args: [][]interface{}{{1, 2}}, want: []interface{}{1, 2}},
		{name: "Elements 5 in 1", args: [][]interface{}{{1, 2, 3, 4, 5}}, want: []interface{}{1, 2, 3, 4, 5}},
		{name: "Elements 1 in 2", args: [][]interface{}{{1}, {-1}}, want: []interface{}{-1, 1}},
		{name: "Elements nil one", args: [][]interface{}{{}, {-1}}, want: []interface{}{-1}},
		{name: "Elements nil one nil", args: [][]interface{}{{}, {-1}}, want: []interface{}{-1}},
		{name: "Elements nil one nil", args: [][]interface{}{{}, {-1}}, want: []interface{}{-1}},
		{name: "Elements nil nil one", args: [][]interface{}{{}, {}, {-1}}, want: []interface{}{-1}},
		{name: "Elements one one", args: [][]interface{}{{1}, {-1}}, want: []interface{}{-1, 1}},
		{name: "Elements one nil one, nil", args: [][]interface{}{{1}, {}, {-1}, {}}, want: []interface{}{-1, 1}},
		{name: "Elements one nil one, one", args: [][]interface{}{{10}, {}, {-1}, {0}}, want: []interface{}{-1, 0, 10}},
		{name: "Elements 5 nil 3, 1", args: [][]interface{}{{2, 4, 6, 8, 10}, {}, {-1, -3, -2}, {100}}, want: []interface{}{-3, -2, -1, 2, 4, 6, 8, 10, 100}},
		{name: "Multi Elements",
			args: [][]interface{}{{2, 4, 6, 8, 10}, {}, {-1, -3, -2}, {100}, {}, {1, 7, 100, 99, -20, -33, 1000, 300, -1}},
			want: []interface{}{-33, -20, -3, -2, -1, -1, 1, 2, 4, 6, 7, 8, 10, 99, 100, 100, 300, 1000}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			done := make(chan interface{})
			defer close(done)

			var got []int
			{
			}
			var valStreams = make([]ReadOnlyStream, 0, 1)
			for _, args := range tt.args {
				valStream := Generator(done, args...)
				valStreams = append(valStreams, valStream)
			}

			for num := range MergeChannels(done, valStreams) {
				got = append(got, (num).(int))
			}

			sort.Ints(got)

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

func TestRepeat(t *testing.T) {
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

			var got []interface{}
			valStream := Repeat(done, tt.args...)

			go func() {
				defer close(done)
				time.Sleep(time.Microsecond * 1000)
			}()

			for num := range OrDone(done, valStream) {
				got = append(got, num)
			}

			if len(got) < len(tt.want) {
				t.Errorf("%s = %v, want %v", t.Name(), got, tt.want)
			}

			for i, wantVal := range tt.want {
				if got[i] != wantVal {
					t.Errorf("%s = %v, want %v", t.Name(), got, tt.want)
				}
			}
		})
	}
}

func TestMergeRepeatFn(t *testing.T) {
	tests := []struct {
		name string
		args func() interface{}
		want []interface{}
	}{
		{name: "empty", args: func() interface{} {
			return nil
		}, want: []interface{}{}},
		{name: "oneInt", args: func() interface{} {
			return 1
		}, want: []interface{}{1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			done := make(chan interface{})

			var got []interface{}
			valStream := RepeatFn(done, tt.args)

			go func() {
				defer close(done)
				time.Sleep(time.Microsecond * 100)
			}()

			for num := range OrDone(done, valStream) {
				got = append(got, num)
			}

			if len(got) < len(tt.want) {
				t.Errorf("%s = %v, want %v", t.Name(), got, tt.want)
			}

			for i, wantVal := range tt.want {
				if got[i] != wantVal {
					t.Errorf("%s = %v, want %v", t.Name(), got, tt.want)
				}
			}
		})
	}
}

func TestTee(t *testing.T) {
	tests := []struct {
		name      string
		takeTimes int
		args      []interface{}
		want      []interface{}
	}{
		{name: "Empty data", args: []interface{}{}, want: []interface{}{}},
		{name: "One element", args: []interface{}{1}, want: []interface{}{1}},
		{name: "Two elements", args: []interface{}{1, 2}, want: []interface{}{1, 2}},
		{name: "Multi element", args: []interface{}{1, 2, 3, 4, 5, -1}, want: []interface{}{1, 2, 3, 4, 5, -1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var wg = sync.WaitGroup{}
			done := make(chan interface{})
			defer close(done)

			var lGot []interface{}
			var rGot []interface{}

			lChan, rChan := Tee(done, Generator(done, tt.args...))

			wg.Add(2)
			go func(got *[]interface{}, wg *sync.WaitGroup) {
				defer wg.Done()
				for num := range lChan {
					*got = append(*got, num)
				}
			}(&lGot, &wg)

			go func(got *[]interface{}, wg *sync.WaitGroup) {
				defer wg.Done()
				for num := range rChan {
					*got = append(*got, num)
				}
			}(&rGot, &wg)

			wg.Wait()

			if len(lGot) != len(tt.want) || len(rGot) != len(tt.want) {
				t.Errorf("%s: lGot=%v rGot=%v, want %v", tt.name, lGot, rGot, tt.want)
			}

			for i, wantVal := range tt.want {
				if lGot[i] != wantVal {
					t.Errorf("%s = %v, want %v", tt.name, lGot, tt.want)
				}
				if rGot[i] != wantVal {
					t.Errorf("%s = %v, want %v", tt.name, rGot, tt.want)
				}

			}
		})
	}
}

type PipelineTest struct {
	logger bufferLogger.Logger
}

func (pt *PipelineTest) testStg() ProcessFn {
	nameFn := "testStg"
	return func(inObj interface{}) (interface{}, error) {
		if fmt.Sprintf("%v", inObj) == "1" {
			return nil, fmt.Errorf("%s : %v", nameFn, inObj)
		}
		pt.logger.Debug(fmt.Sprintf("%s: %v", nameFn, inObj))
		return inObj, nil
	}
}

func (pt *PipelineTest) testFanOutStg() ProcessFn {
	nameFn := "testFanOutStg"
	return func(inObj interface{}) (interface{}, error) {
		pt.logger.Debug(fmt.Sprintf("%s : %v", nameFn, inObj))
		return inObj, nil
	}
}

func (pt *PipelineTest) testFilterStg() FilterFn {
	nameFn := "testFilterStg"
	return func(inObj interface{}) (interface{}, bool) {
		pt.logger.Debug(fmt.Sprintf("%s : %v", nameFn, inObj))
		return inObj, true
	}
}

func (pt *PipelineTest) testFunOutFilterStg() FilterFn {
	nameFn := "testFunOutFilterStg"
	return func(inObj interface{}) (interface{}, bool) {
		pt.logger.Debug(fmt.Sprintf("%s : %v", nameFn, inObj))
		return inObj, true
	}
}

func (pt *PipelineTest) errorHandler() ErrorProcessFn {
	nameFn := "errorHandler"
	return func(err error) {
		errorFrom := fmt.Sprintf("stg %s error: %v", nameFn, err)
		pt.logger.Error(errorFrom)
	}
}

var want = `ERROR  [stg errorHandler error: testStg : 1]
DEBUG  [testStg: 2]
DEBUG  [testFanOutStg : 2]
DEBUG  [testFilterStg : 2]
DEBUG  [testFunOutFilterStg : 2]
`

func TestPipeline(t *testing.T) {
	done := make(chan interface{})
	defer close(done)
	pTest := PipelineTest{
		logger: bufferLogger.NewBufferLogger(true),
	}

	testPipeline := NewPipeline()

	testPipeline.AddStage(pTest.testStg(), pTest.errorHandler())
	testPipeline.AddStageWithFanOut(pTest.testFanOutStg(), pTest.errorHandler(), 3)
	testPipeline.AddFilterStage(pTest.testFilterStg())
	testPipeline.AddFilterStageWithFanOut(pTest.testFunOutFilterStg(), 3)

	prDone := testPipeline.RunPlug(done, Generator(done, 1, 2))

	<-prDone

	if pTest.logger.(*bufferLogger.BufferLogger).GetBufferedString() != want {
		t.Errorf("%s got \n%s, want \n%s\n", t.Name(), pTest.logger.(*bufferLogger.BufferLogger).GetBufferedString(), want)
	}
}

func BenchmarkMerge(b *testing.B) {
	tests := []struct {
		name string
		args [][]interface{}
	}{

		{name: "Multi Elements",
			args: [][]interface{}{{2, 4, 6, 8, 10}, {}, {-1, -3, -2}, {100}, {}, {1, 7, 100, 99, -20, -33, 1000, 300, -1}},
		},
	}

	for _, tt := range tests {

		b.Run(tt.name, func(b *testing.B) {
			done := make(chan interface{})
			defer close(done)

			var valStreams = make([]ReadOnlyStream, 0, 1)
			for _, args := range tt.args {
				valStream := Generator(done, args...)
				valStreams = append(valStreams, valStream)
			}

			for i := 0; i < b.N; i++ {
				for range MergeChannels(done, valStreams) {
					// Read fro channel
				}
			}

		})
	}
}
