package tube

import (
	"sync"
	"time"
	"fmt"
	"testing"
)

var testSize = 1024 * 1024 * 100
var capacity = PAGESIZE * 4 * 1024

var readBufferSize = 1024 * 100
var writeBufferSize = 1024 * 100

func testTube(wb, rb Tube, size int) error {
	wg := &sync.WaitGroup{} 
	writeData := make([]byte, 0)
	readData := make([]byte, 0)

	wg.Add(1)
	go func() {
		BS := writeBufferSize
		buf := make([]byte, BS)
		for i := 0; i<size; i++ {
			j := i % BS
			buf[j] = byte(i % 256)
			writeData = append(writeData, buf[j])
			if j == BS - 1 || i == size - 1 {
				wb.Write(buf[:j + 1])
			}
		}

		wb.Close()
		wg.Done()
	}()

	wg.Add(1)
	go func(){
		BS := readBufferSize
		buf := make([]byte, BS)
		c := 0
		var err error 
		for err == nil {
			c, err = rb.Read(buf)
			if c > 0 {
				readData = append(readData, buf[:c]...)
			}else{
				time.Sleep(time.Microsecond)
			}
		}
		wg.Done()
	}()

	wg.Wait()

	fmt.Printf("Write %d Read %d\n", len(writeData), len(readData))
	if string(writeData) != string(readData) {
		return fmt.Errorf("not match")
	}
	return nil
}

func TestInternalTube(t *testing.T) {
	tb := NewInternalTube(capacity)
	if testTube(tb, tb, testSize) != nil {
		t.Fatal("write != read")
	}
}

func TestMmapTube(t *testing.T) {
	address := "/tmp/a"
	wb, _ := NewMmapTube(capacity, address)
	rb, _ := NewMmapTube(capacity, address)

	if testTube(wb, rb, testSize) != nil {
		t.Fatal("write != read")
	}
}

func TestSocketTube(t *testing.T) {
	var err error
	wb, _ := NewSocketTube(capacity, "127.0.0.1:33333")
	rb, _ := NewSocketTube(capacity, "127.0.0.1:33333")
	if err = wb.Start(WRITER); err != nil {
		t.Fatal(err)
	}
	if err = rb.Start(READER); err != nil {
		t.Fatal(err)
	}

	if testTube(wb, rb, testSize) != nil {
		t.Fatal("write != read")
	}
}