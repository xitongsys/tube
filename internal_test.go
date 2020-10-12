package tube

import (
	"testing"
	"sync"
)


func Test(t *testing.T) {
	tb := NewInternalTube(1024)
	n := 1024

	wg := &sync.WaitGroup{}

	//write
	wg.Add(1)
	go func(){
		buf := make([]byte, 1024)
		for i := 0; i<len(buf); i++ {
			buf[i] = byte(i % PAGESIZE)
		}

		for i := 0; i < n; i++ {
			tb.Write(buf)
		}

		tb.Flush()
		tb.Close()

		wg.Done()
	}()

	wg.Add(1)
	//read
	go func(){
		buf := make([]byte, 1024)
		c := 0
		var err error

		for err == nil {
			c, err = tb.Read(buf)
			if c > 0 {
				t.Log(c)
			}
		}

		wg.Done()
	}()

	wg.Wait()
}