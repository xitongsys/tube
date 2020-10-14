package tube

import (
	"io"
	"time"
)

// Page structure -- right alignment
// Full: [0, 1, 2, 3, ...] header = PAGESIZE  tempPageDataIndex = 0
// Partical: [....0, 1, 2, 3] header = 0, tempPageDataIndex = 4. After Flush(), header = 4, tempPageDataIndex = 0
type InternalTube struct {
	//for reader/writer
	pageCnt		int
	err			error

	//memory alignment
	data 		[]byte
	isEOF       *byte
	pageHeaders []byte
	pageData    []byte

	//for reader only
	eofFlag bool
	readerPageIndex int

	//for writer only
	tempPageDataIndex int
	writerPageIndex int
}

func NewInternalTube(capacity int) *InternalTube {
	pageCnt := (capacity + PAGESIZE - 1) / PAGESIZE
	data := make([]byte, pageCnt + pageCnt * PAGESIZE + 1)
	t := & InternalTube{
		pageCnt:		pageCnt,
		data: 			data,
		isEOF:			&data[0],
		pageHeaders:	data[1:1 + pageCnt],
		pageData: 		data[pageCnt+1:],
	}

	return t
}

func NewInternalTubeFromData(data []byte) *InternalTube {
	pageCnt := (len(data) - 1) / (PAGESIZE + 1)
	t := & InternalTube{
		pageCnt:		pageCnt,
		data: 			data,
		isEOF:			&data[0],
		pageHeaders:	data[1:1 + pageCnt],
		pageData: 		data[pageCnt+1:],
	}

	return t
}

func (itube *InternalTube) Type() TubeType {
	return INTERNAL
}

func (itube *InternalTube) Role() TubeRole {
	return BOTH
}

func (itube *InternalTube) Address() string {
	return ""
}

func (itube *InternalTube) Capacity() int {
	return len(itube.pageData)
}

func (itube *InternalTube) SetError(err error) {
	itube.err = err
}

func (itube *InternalTube) Read(data []byte) (n int, err error) {
	if *itube.isEOF != 0 {
		itube.eofFlag = true
	}

	pageIndex := &itube.readerPageIndex

	i, lt := 0, len(data)
	for i < lt && itube.pageHeaders[*pageIndex] != 0 {
		header := itube.pageHeaders[*pageIndex]
		ls := int(header)
		pageDataBgn, pageDataEnd := (*pageIndex)*PAGESIZE+PAGESIZE-int(header), (*pageIndex)*PAGESIZE+PAGESIZE
		lc := copy(data[i:], itube.pageData[pageDataBgn:pageDataEnd])

		//reset page header
		nls := ls - lc
		itube.pageHeaders[*pageIndex] = byte(nls)
		i += lc
		if nls == 0 {
			itube.incPageIndex(pageIndex)
		}
	}

	if itube.eofFlag && i == 0 {
		itube.err = io.EOF
	}

	return i, itube.err
}

//block write
func (itube *InternalTube) Write(data []byte) (n int, err error){
	i := 0
	for i < len(data) {
		if c, err := itube.write(data[i:]); err != nil && err != ERR_TUBE_IS_FULL {
			return i + c, err

		} else {
			i += c
			if c == 0 {
				time.Sleep(time.Microsecond)
			}
		}
	}

	return i, itube.err
}

func (itube *InternalTube) write(data []byte) (n int, err error) {
	pageIndex := &itube.writerPageIndex
	header := itube.pageHeaders[*pageIndex]
	if header != 0 {
		return 0, ERR_TUBE_IS_FULL
	}

	lb, ld := itube.tempPageDataIndex, len(data)
	if lb + ld < PAGESIZE {
		copy(itube.pageData[(*pageIndex) * PAGESIZE + lb:], data)
		itube.tempPageDataIndex += ld
		return ld, itube.err
	}

	lc := copy(itube.pageData[(*pageIndex) * PAGESIZE + lb: (*pageIndex) * PAGESIZE + PAGESIZE], data)
	itube.pageHeaders[*pageIndex] = byte(PAGESIZE)
	itube.incPageIndex(&itube.writerPageIndex)
	itube.tempPageDataIndex = 0

	i := lc
	header = itube.pageHeaders[*pageIndex]

	for i < ld && itube.pageHeaders[*pageIndex] == 0 {
		lc = copy(itube.pageData[(*pageIndex) * PAGESIZE: (*pageIndex) * PAGESIZE + PAGESIZE], data[i:])
		i += lc
		if lc == PAGESIZE {
			itube.pageHeaders[*pageIndex] = byte(PAGESIZE)
			itube.incPageIndex(&itube.writerPageIndex)

		} else {
			itube.tempPageDataIndex = lc
		}
	}

	return i, itube.err
}

func (itube *InternalTube) Flush() error {
	pageIndex := &itube.writerPageIndex
	if itube.tempPageDataIndex != 0 {
		for i, j := (*pageIndex) * PAGESIZE + itube.tempPageDataIndex - 1, (*pageIndex) * PAGESIZE + PAGESIZE - 1; i >= (*pageIndex) * PAGESIZE; i, j = i - 1, j - 1 {
			itube.pageData[j] = itube.pageData[i]
		}

		itube.pageHeaders[*pageIndex] = byte(itube.tempPageDataIndex)
		itube.incPageIndex(pageIndex)
		itube.tempPageDataIndex = 0
	}

	return itube.err
}

func (itube *InternalTube) Close() error {
	itube.Flush()
	*itube.isEOF = 1
	return itube.err
}

//increment pageIndex by 1
func (itube *InternalTube) incPageIndex(pageIndex *int) {
	(*pageIndex)++
	(*pageIndex) %= itube.pageCnt
}

