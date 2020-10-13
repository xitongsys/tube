package tube

import (
	"os"
	"syscall"
	"unsafe"
)

type MmapTube struct {
	InternalTube
	Address string
}

func NewMmapTubeWriter(capacity int, address string) (*MmapTube, error) {
	f, err := os.OpenFile(address, os.O_CREATE | os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	pageCnt := (capacity + PAGESIZE - 1) / PAGESIZE

	rawBuf, err := syscall.Mmap(int(f.Fd()), 0, pageCnt * PAGESIZE, syscall.PROT_READ | syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return nil, err
	}
	f.Truncate(int64(pageCnt * PAGESIZE))

	ptr := uintptr(unsafe.Pointer(&rawBuf[0]))

	var sl = struct {
		addr uintptr
		len int 
		cap int
	}{ptr, pageCnt * PAGESIZE + pageCnt + 1, pageCnt * PAGESIZE + pageCnt + 1}

	data := *(*[]byte)(unsafe.Pointer(&sl))

	mt := & MmapTube{
		InternalTube: *NewInternalTubeWriterFromData(data),
		Address: address,
	}

	return mt, nil
}

func NewMmapTubeReader(capacity int, address string) (*MmapTube, error) {
	f, err := os.OpenFile(address, os.O_CREATE | os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	pageCnt := (capacity + PAGESIZE - 1) / PAGESIZE

	rawBuf, err := syscall.Mmap(int(f.Fd()), 0, pageCnt * PAGESIZE, syscall.PROT_READ | syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return nil, err
	}
	f.Truncate(int64(pageCnt * PAGESIZE))

	ptr := uintptr(unsafe.Pointer(&rawBuf[0]))
	
	var sl = struct {
		addr uintptr
		len int 
		cap int
	}{ptr, pageCnt * PAGESIZE + pageCnt + 1, pageCnt * PAGESIZE + pageCnt + 1}

	data := *(*[]byte)(unsafe.Pointer(&sl))

	mt := & MmapTube{
		InternalTube: *NewInternalTubeReaderFromData(data),
		Address: address,
	}

	return mt, nil
}

