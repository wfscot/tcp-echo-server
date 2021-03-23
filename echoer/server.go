package echoer

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"net"
)

// defines a generic echo server object
// a server represents a single server instance that will accept any number of incoming connections on a given port.
// for each connection, it will echo whatever data is received until the client closes the connection.

type Server interface {
	Run(context.Context) error
}

type tcpEchoServer struct {
	listenPort    int
	announceAlive bool
}

func NewTcpEchoServer(listenPort int, announceAlive bool) Server {
	return &tcpEchoServer{
		listenPort:    listenPort,
		announceAlive: announceAlive,
	}
}

func (s *tcpEchoServer) Run(ctx context.Context) error {
	// use the log object from the context with additional fields
	log := log.Ctx(ctx).With().Str("func", "tcpEchoServer.Run").Logger()

	// use a ListenConfig so it can be torn down via context
	lc := net.ListenConfig{}

	// establish the listener on all interfaces
	ln, err := lc.Listen(ctx, "tcp", fmt.Sprintf(":%d", s.listenPort))
	if err != nil {
		log.Error().Err(err).Msg("error while establishing listener")
		return err
	}
	defer ln.Close()
	log.Info().Stringer("listenAddr", ln.Addr()).Msg("listener established")

	// for some reason, the listener is staying open even after the context is cancelled. force it closed.
	go func() {
		<-ctx.Done()
		ln.Close()
	}()

	i := 0
	for {
		i++
		log := log.With().Int("connNum", i).Logger()
		log.Debug().Msg("waiting for connection")
		conn, err := ln.Accept()
		if err != nil {
			// if the context has been canceled, ignore the final error and return
			if ctx.Err() != nil {
				return nil
			}
			// otherwise log the error and continue
			log.Error().Int("connNum", i).Err(err).Msg("error while accepting client connection")
			continue
		}
		log = log.With().Stringer("clientAddr", conn.RemoteAddr()).Logger()
		log.Info().Msg("accepted client connection")

		// store logger in the context
		ctx = log.WithContext(ctx)

		// invoke echo logic in a routine
		go func(ctx context.Context) {
			e := NewEchoer(conn, s.announceAlive)
			err := e.Run(ctx)
			if err != nil {
				log.Error().Err(err).Msg("echoer exited with error")
			}
		}(ctx)
	}
}
