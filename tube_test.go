package tube

import (
	"sync"
	"time"
	"fmt"
	"testing"
)

func testTube(wb, rb Tube, size int) error {
	wg := &sync.WaitGroup{} 
	writeData := make([]byte, 0)
	readData := make([]byte, 0)

	wg.Add(1)
	go func() {
		BS := 200
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
		BS := 150
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
	tb := NewInternalTube(PAGESIZE * 10)
	if testTube(tb, tb, 1024 * 10) != nil {
		t.Fatal("write != read")
	}
}

func TestMmapTube(t *testing.T) {
	capacity, address := PAGESIZE * 10, "/tmp/a"
	wb, _ := NewMmapTube(capacity, address)
	rb, _ := NewMmapTube(capacity, address)

	if testTube(wb, rb, 1024 * 10) != nil {
		t.Fatal("write != read")
	}
}

func TestSocketTube(t *testing.T) {
	var err error
	wb, _ := NewSocketTube(PAGESIZE * 10, "127.0.0.1:33333")
	rb, _ := NewSocketTube(PAGESIZE * 10, "127.0.0.1:33333")
	if err = wb.Start(WRITER); err != nil {
		t.Fatal(err)
	}
	if err = rb.Start(READER); err != nil {
		t.Fatal(err)
	}

	if testTube(wb, rb, 1024 * 10) != nil {
		t.Fatal("write != read")
	}
}