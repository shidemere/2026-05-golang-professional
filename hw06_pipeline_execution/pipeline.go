// Package hw06pipelineexecution homework solution for lesson 12 from otus.
package hw06pipelineexecution

type (
	// In -> represent input data for single stage.
	In = <-chan interface{}
	// Out -> represent output data for single stage.
	Out = In
	// Bi represent channel that can read and write.
	Bi = chan interface{}
)

// Stage -> represent functon that consumed, do some work with data and produce data.
type Stage func(in In) (out Out)

// ExecutePipeline let us to start out pipeline with several stages.
func ExecutePipeline(in In, done In, stages ...Stage) Out {
	for _, stage := range stages {
		in = wrapper(stage(in), done)
	}
	return in
}

func wrapper(in In, done In) Out {
	out := make(Bi)
	go func() {
		for {
			select {
			case value, ok := <-in:
				if !ok {
					close(out)
					readAndDoNothing(in)
					return
				}
				select {
				case out <- value:
				case <-done:
					close(out)
					return
				}
			case <-done:
				close(out)
				readAndDoNothing(in)
				return
			}
		}
	}()
	return out
}

func readAndDoNothing(in In) {
	for val := range in {
		_ = val
	}
}
