package tube

import (
	"os"
	"syscall"
	"unsafe"
)

type MmapTube struct {
	InternalTube
	address string
	fp *os.File
}

func NewMmapTube(capacity int, address string) (*MmapTube, error) {
	f, err := os.OpenFile(address, os.O_CREATE | os.O_RDWR | os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}

	pageCnt := (capacity + PAGESIZE - 1) / PAGESIZE
	totalSize := pageCnt * PAGESIZE + pageCnt + 1

	rawBuf, err := syscall.Mmap(int(f.Fd()), 0, totalSize, syscall.PROT_READ | syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return nil, err
	}
	f.Truncate(int64(totalSize))

	ptr := uintptr(unsafe.Pointer(&rawBuf[0]))

	var sl = struct {
		addr uintptr
		len int 
		cap int
	}{ptr, totalSize, totalSize}

	data := *(*[]byte)(unsafe.Pointer(&sl))

	mt := & MmapTube{
		InternalTube: *NewInternalTubeFromData(data),
		address: address,
		fp: f,
	}

	return mt, nil
}

func (mt *MmapTube) Close() error {
	mt.InternalTube.Close()
	return mt.fp.Close()
}

func (mt *MmapTube) Address() string {
	return mt.address
}
