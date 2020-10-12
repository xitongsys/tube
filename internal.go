package tube

import (
	"io"
)

type InternalTube struct {
	role TubeRole
	address string
	pageIndex int
	pageCnt int
	pageData []byte
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

func (itube *InternalTube) Size() int {
	i0 := itube.pageIndex
	l, r := 1, itube.pageCnt
	for l <= r {
		m := l + (r - l) / 2
		i := (i0 + m - 1) % itube.pageCnt
		header := itube.pageData[i * PAGESIZE]
		if itube.role == READER {
			if header != 0 && header != TUBEEOF {
				l = m + 1
			} else {
				r = m - 1
			}

		} else { // WRITER
			if header == 0 {
				l = m + 1
			} else {
				r = m - 1
			}
		}
	}

	if r > 0 {
		return r
	}

	if itube.pageData[i0 * PAGESIZE] == TUBEEOF {
		return -1
	}

	return 0
}

func (itube *InternalTube) Address() string {
	return itube.address
}

func (itube *InternalTube) Read(data []byte) (n int, err error) {
	if itube.role != READER {
		return -1, ERR_READ_FROM_WRITE_TUBE
	}

	i, ld = 0, len(data)
	for i < ld {
		header := itube.pageData[itube.pageIndex * PAGESIZE]
		if header == TUBEEOF {
			err = io.EOF
			break
		} else if header == 0 {
			break
		}
		
		
	}
}




