package tube

import "io"

// Page structure
// Full: [0, 1, 2, 3, ...] header = PAGESIZE  tempPageDataIndex = 0
// Partical: [....0, 1, 2, 3] header = 0, tempPageDataIndex = 4. After Flush(), header = 4, tempPageDataIndex = 0
type InternalTube struct {
	role    TubeRole
	address string

	//for reader/writer
	pageIndex int
	pageCnt   int

	//memory alignment
	isEOF       byte
	pageHeaders []byte
	pageData    []byte

	//for writer only
	tempPageDataIndex int
}

func (itube *InternalTube) Type() TubeType {
	return INTERNAL
}

func (itube *InternalTube) Role() TubeRole {
	return itube.role
}

func (itube *InternalTube) Capacity() int {
	return len(itube.pageData)
}

func (itube *InternalTube) Address() string {
	return itube.address
}

func (itube *InternalTube) Read(data []byte) (n int, err error) {
	if itube.role != READER {
		return -1, ERR_READ_FROM_WRITE_TUBE
	}

	i, lt := 0, len(data)
	for i < lt {
		header := itube.pageHeaders[itube.pageIndex]
		if header == 0 {
			break
		}

		ls := int(header)
		pageDataBgn, pageDataEnd := itube.pageIndex*PAGESIZE+PAGESIZE-int(header), itube.pageIndex*PAGESIZE+PAGESIZE
		lc := copy(data[i:], itube.pageData[pageDataBgn:pageDataEnd])

		//reset page header
		nls := ls - lc
		itube.pageHeaders[itube.pageIndex] = byte(nls)
		i += lc
		if nls == 0 {
			itube.incPageIndex()
		}
	}

	if itube.isEOF != 0 {
		err = io.EOF
	}

	return i, err
}

func (itube *InternalTube) Write(data []byte) (n int, err error) {
	header := itube.pageHeaders[itube.pageIndex]
	if header != 0 {
		return 0, ERR_TUBE_IS_FULL
	}

	lb, ld := itube.tempPageDataIndex, len(data)
	if lb + ld < PAGESIZE {
		copy(itube.pageData[itube.pageIndex * PAGESIZE + lb:], data)
		itube.tempPageDataIndex += ld
		return ld, nil
	}

	lc := copy(itube.pageData[itube.pageIndex * PAGESIZE + lb:itube.pageIndex * PAGESIZE + PAGESIZE], data)
	itube.incPageIndex()
	itube.tempPageDataIndex = 0

	i := lc
	header = itube.pageHeaders[itube.pageIndex]

	for i < ld && itube.pageHeaders[itube.pageIndex] == 0 {
		lc = copy(itube.pageData[itube.pageIndex * PAGESIZE:itube.pageIndex * PAGESIZE + PAGESIZE], data[i:])
		i += lc
		if lc == PAGESIZE {
			itube.pageHeaders[itube.pageIndex] = byte(PAGESIZE)
			itube.incPageIndex()

		} else {
			itube.tempPageDataIndex = lc
		}
	}

	return i, nil
}

func (itube *InternalTube) Flush() error {
	if itube.tempPageDataIndex != 0 {
		for i, j := itube.pageIndex * PAGESIZE + itube.tempPageDataIndex - 1, itube.pageIndex * PAGESIZE + PAGESIZE - 1; i >= 0; i, j = i-1, j- 1 {
			ii, jj := itube.pageIndex * PAGESIZE + i, itube.pageIndex * PAGESIZE + j
			itube.pageData[jj] = itube.pageData[ii]
		}
		itube.pageHeaders[itube.pageIndex] = byte(itube.tempPageDataIndex)
		itube.incPageIndex()
		itube.tempPageDataIndex = 0
	}
	return nil
}

func (itube *InternalTube) Close() error {
	itube.Flush()
	itube.isEOF = 1
	return nil
}

//increment pageIndex by 1
func (itube *InternalTube) incPageIndex() {
	itube.pageIndex++
	itube.pageIndex %= itube.pageCnt
}

