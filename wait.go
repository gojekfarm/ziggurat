package ziggurat

import "sync"

// WaitAll is a util function which takes in n channels
// and waits on all of them
// every channel must be of type chan struct{}
func WaitAll(chans ...chan struct{}) {
	var wg sync.WaitGroup
	for _, c := range chans {
		wg.Add(1)
		cCopy := c
		go func() {
			<-cCopy
			wg.Done()
		}()
	}
	wg.Wait()
}
