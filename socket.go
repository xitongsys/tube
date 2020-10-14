package tube

import (
	"io"
	"net"
)

type SocketTube struct {
	InternalTube
	address	string

	role TubeRole

	//for writer
	closed	bool
	tcpListener *net.TCPListener

	//for reader
	tcpConn net.Conn
	
}

func NewSocketTube(capacity int, address string) (*SocketTube, error) {
	if _, err := net.ResolveTCPAddr("tcp", address); err != nil {
		return nil, err
	}

	st := & SocketTube {
		InternalTube: *NewInternalTube(capacity),
		address: address,
		closed: false,

	}

	return st, nil
}

func (st *SocketTube) Start(role TubeRole) error {
	st.role = role
	if role == READER {
		return st.startReader()
	} else if role == WRITER {
		return st.startWriter()
	}

	return ERR_OPERATION_NOT_SUPPORT
}

func (st *SocketTube) startReader() (err error) {
	st.tcpConn, err = net.Dial("tcp", st.address)
	go func(){
		defer func(){
			st.Close()
		}()

		if _, err := io.Copy(st, st.tcpConn); err != nil {
			st.SetError(err)
			return
		}
	}()

	return err
}

func (st *SocketTube) startWriter() error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", st.address)
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
			} else if err != nil {
				continue
			}

			go func(){
				defer func(){
					conn.Close()
				}()

				if _, err := io.Copy(conn, st); err != nil {
					st.SetError(err)
					return
				}
			}()

		}
	}()

	return nil
}

func (st *SocketTube) Close() error {
	st.InternalTube.Close()
	st.closed = true

	if st.role == READER {
		return st.tcpConn.Close()
	} else {
		return st.tcpListener.Close()
	}
}

func (st *SocketTube) Role() TubeRole {
	return st.role
}

func (st *SocketTube) Address() string {
	return st.address
}