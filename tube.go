package tube

type TubeType int 
const (
	INTERNAL TubeType = iota
	MMAP
	SOCKET 
)

type TubeRole int
const (
	READER TubeRole = iota
	WRITER
	BOTH
)

const PAGESIZE int = 255
const TUBEEOF byte = 255

type Tube interface {
	Type() TubeType
	Role() TubeRole
	Capacity() int
	Address() string
	Write(data []byte) (int, error)
	Read(data []byte) (int, error)
	Flush() error
	Close() error
}