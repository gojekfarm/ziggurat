package ziggurat

import "sync"

func Join(chans ...chan error) []error {
	errors := make([]error, 0, len(chans))
	receiver := make(chan error)
	var wg sync.WaitGroup
	for _, c := range chans {
		ch := c
		wg.Add(1)
		go func() {
			receiver <- <-ch
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(receiver)
	}()
	for e := range receiver {
		errors = append(errors, e)
	}
	return errors
}
