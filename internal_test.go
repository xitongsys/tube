package tube

import (
	"testing"
	"sync"
)


func Test(t *testing.T) {
	wb := NewInternalTubeWriter(PAGESIZE * 10)
	rb := NewInternalTubeReader(wb)
	n := 20

	wg := &sync.WaitGroup{}

	//write
	wg.Add(1)
	go func(){
		buf := make([]byte, 1024)
		for i := 0; i<len(buf); i++ {
			buf[i] = byte(i % PAGESIZE)
		}

		for i := 0; i < n; i++ {
			c, err := wb.Write(buf)
			t.Log("write", c, wb.pageIndex, err)
		}

		wb.Close()

		wg.Done()
	}()

	wg.Add(1)
	//read
	go func(){
		buf := make([]byte, 1024)
		c := 0
		var err error

		for err == nil {
			c, err = rb.Read(buf)
			if c > 0 {
				t.Log(c)
			}

			if err != nil {
				t.Log(err)
			}
		}

		wg.Done()
	}()

	wg.Wait()
}