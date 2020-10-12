package tube

import (
	"fmt"
)

var ERR_READ_FROM_WRITE_TUBE = fmt.Errorf("read from a WriterTube")
var ERR_WRITE_TO_READERTUBE = fmt.Errorf("write to a ReaderTube")
var ERR_TUBE_IS_FULL = fmt.Errorf("tube is full")
