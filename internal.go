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
	role 		TubeRole
	pageIndex	int
	pageCnt		int

	//memory alignment
	data 		[]byte
	isEOF       *byte
	pageHeaders []byte
	pageData    []byte

	//for reader only
	eofFlag bool

	//for writer only
	tempPageDataIndex int
}

func NewInternalTubeWriter(capacity int) *InternalTube {
	pageCnt := (capacity + PAGESIZE - 1) / PAGESIZE
	data := make([]byte, pageCnt + pageCnt * PAGESIZE + 1)
	t := & InternalTube{
		role:			WRITER,
		pageCnt:		pageCnt,
		data: 			data,
		isEOF:			&data[0],
		pageHeaders:	data[1:1 + pageCnt],
		pageData: 		data[pageCnt+1:],
	}

	return t
}

func NewInternalTubeWriterFromData(data []byte) *InternalTube {
	pageCnt := (len(data) - 1) / (PAGESIZE + 1)
	t := & InternalTube{
		role:			WRITER,
		pageCnt:		pageCnt,
		data: 			data,
		isEOF:			&data[0],
		pageHeaders:	data[1:1 + pageCnt],
		pageData: 		data[pageCnt+1:],
	}

	return t
}

func NewInternalTubeReader(wt *InternalTube) *InternalTube {
	t := & InternalTube{
		role:			READER,
		pageCnt:		wt.pageCnt,
		data: 			wt.data,
		isEOF:			&wt.data[0],
		pageHeaders:	wt.data[1:1 + wt.pageCnt],
		pageData: 		wt.data[wt.pageCnt+1:],
	}

	return t
}

func NewInternalTubeReaderFromData(data []byte) *InternalTube {
	pageCnt := (len(data) - 1) / (PAGESIZE + 1)
	t := & InternalTube{
		role:			READER,
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
	return itube.role
}

func (itube *InternalTube) Address() string {
	return ""
}

func (itube *InternalTube) Capacity() int {
	return len(itube.pageData)
}

func (itube *InternalTube) Read(data []byte) (n int, err error) {
	if *itube.isEOF != 0 {
		itube.eofFlag = true
	}

	i, lt := 0, len(data)
	for i < lt && itube.pageHeaders[itube.pageIndex] != 0 {
		header := itube.pageHeaders[itube.pageIndex]
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

	if itube.eofFlag && i == 0 {
		err = io.EOF
	}

	return i, err
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

	return i, nil
}

func (itube *InternalTube) write(data []byte) (n int, err error) {
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
	itube.pageHeaders[itube.pageIndex] = byte(PAGESIZE)
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
		for i, j := itube.pageIndex * PAGESIZE + itube.tempPageDataIndex - 1, itube.pageIndex * PAGESIZE + PAGESIZE - 1; i >= itube.pageIndex * PAGESIZE; i, j = i - 1, j - 1 {
			itube.pageData[j] = itube.pageData[i]
		}

		itube.pageHeaders[itube.pageIndex] = byte(itube.tempPageDataIndex)
		itube.incPageIndex()
		itube.tempPageDataIndex = 0
	}

	return nil
}

func (itube *InternalTube) Close() error {
	itube.Flush()
	*itube.isEOF = 1
	return nil
}

//increment pageIndex by 1
func (itube *InternalTube) incPageIndex() {
	itube.pageIndex++
	itube.pageIndex %= itube.pageCnt
}

