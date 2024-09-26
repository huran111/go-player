package rtmp

import (
	"context"
	"fmt"
	"io"
	"net"
	"player/internal/config"

	"github.com/yutopp/go-rtmp"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	addChan  chan *Channel
	Channels map[string]*Channel
	*net.TCPListener
	*rtmp.Server
}

func (svr *Server) Run(ctx context.Context) error {

	errGrp, _ := errgroup.WithContext(ctx)

	errGrp.Go(func() error {

		for c := range svr.addChan {
			svr.Channels[c.live] = c
		}

		return nil
	})

	errGrp.Go(func() error {
		return svr.Serve(svr.TCPListener)
	})

	if err := errGrp.Wait(); err != nil {
		return err
	}

	return nil

}

func NewServer(cfg *config.Config) (*Server, error) {

	tcpAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", "0.0.0.0", cfg.RtmpServerPort))
	if err != nil {
		return nil, err
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil, err
	}

	Chan := make(chan *Channel, 1024)

	srv := rtmp.NewServer(&rtmp.ServerConfig{

		OnConnect: func(conn net.Conn) (io.ReadWriteCloser, *rtmp.ConnConfig) {
			return conn, &rtmp.ConnConfig{
				Handler: &Stream{
					Chan: Chan,
				},
			}
		},
	})

	return &Server{
		Chan,
		map[string]*Channel{},
		listener,
		srv,
	}, nil

}
