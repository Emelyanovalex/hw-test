package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	if len(stages) == 0 {
		return in
	}

	current := in
	for _, stage := range stages {
		current = stage(current)
		if done != nil {
			current = forwardUntilDone(done, current)
		}
	}

	return current
}

func forwardUntilDone(done In, source Out) Out {
	result := make(Bi)

	go func() {
		defer close(result)

		for {
			select {
			case <-done:
				go drainChannel(source)
				return
			case value, ok := <-source:
				if !ok {
					return
				}

				if !sendUntilDone(done, result, value) {
					go drainChannel(source)
					return
				}
			}
		}
	}()

	return result
}

func sendUntilDone(done In, destination Bi, value interface{}) bool {
	select {
	case destination <- value:
		return true
	case <-done:
		return false
	}
}

func drainChannel(source Out) {
	for range source {
		// drain source until it is closed
	}
}
