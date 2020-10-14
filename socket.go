package tube

import (
	"io"
	"net"
)

type SocketTube struct {
	InternalTube
	Address	string

	role TubeRole

	//for writer
	closed	bool
	tcpListener *net.TCPListener

	//for reader
	tcpConn net.Conn
	
}

func NewSocketTubeWriter(capacity int, address string) (*SocketTube, error) {
	if _, err := net.ResolveTCPAddr("tcp", address); err != nil {
		return nil, err
	}

	st := & SocketTube {
		InternalTube: *NewInternalTubeWriter(capacity),
		Address: address,
		closed: false,
		role: WRITER,

	}

	return st, nil
}

func NewSocketTubeReader(capacity int, address string) (*SocketTube, error) {
	st := & SocketTube {
		InternalTube: *NewInternalTubeWriter(capacity),
		Address: address,
		closed: false,
		role: READER,
	}

	return st, nil
}

func (st *SocketTube) Start() error {
	if st.role == READER {
		return st.startReader()
	}

	return st.startWriter()
}

func (st *SocketTube) startReader() (err error) {
	st.tcpConn, err = net.Dial("tcp", st.Address)
	go func(){
		for {
			if _, err := io.Copy(st, st.tcpConn); err != nil {
				return
			}
		}
	}()
	return err
}

func (st *SocketTube) startWriter() error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", st.Address)
	if err != nil {
		return err
	}

	st.tcpListener, err = net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil
	}

	go func() {
		for {
			conn, err := st.tcpListener.AcceptTCP()
			if err != nil && st.closed {
				return
			}

			go func(){
				defer func(){
					conn.Close()
				}()

				if _, err := io.Copy(conn, st); err != nil {
					return
				}
			}()

		}
	}()

	return nil
}

func (st *SocketTube) Close() error {
	st.InternalTube.Close()

	if st.role == READER {
		return st.tcpConn.Close()
	} else {
		return st.tcpListener.Close()
	}
}