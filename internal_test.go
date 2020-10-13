package tube

import (
	"testing"
	"sync"
	//"math/rand"
	"time"
)


func Test(t *testing.T) {
	wb := NewInternalTubeWriter(PAGESIZE * 10)
	rb := NewInternalTubeReader(wb)

	bgnTime0 := time.Now().UnixNano()
	data := make([]byte, 1024 * 1024 * 1000)
	for i := 0; i < len(data); i++ {
		//data[i] = byte(rand.Intn(255))
		data[i] = byte(i)
	}

	bgnTime := time.Now().UnixNano()

	t.Log("Time: ", (bgnTime - bgnTime0) / 1e6)

	wg := &sync.WaitGroup{}

	//write
	wg.Add(1)
	go func(){
		BS := 200
		buf := make([]byte, BS)
		for i := 0; i < len(data); i++ {
			j := i % BS
			buf[j] = data[i]
			if j == BS - 1 || i == len(data) - 1 {
				wb.Write(buf[:j + 1])
			}
		}
		wb.Close()
		wg.Done()
	}()

	
	readData := make([]byte, 0)
	wg.Add(1)
	//read
	go func(){
		BS := 100
		buf := make([]byte, BS)
		c := 0
		var err error
		for err == nil {
			c, err = rb.Read(buf)
			if c > 0 {
				readData = append(readData, buf[:c]...)
			} else {
				time.Sleep(time.Microsecond)
			}
		}
		wg.Done()
	}()
	
	wg.Wait()

	endTime := time.Now().UnixNano()

	t.Log("Time: ", (endTime - bgnTime) / 1e6)

	if string(data) != string(readData) {

		t.Fatal("write != read", len(data), len(readData))
	}

}