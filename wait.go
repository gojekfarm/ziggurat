package ziggurat

import "sync"

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
