package main

import (
	"context"
	"fmt"
	"github.com/pborman/getopt/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/wfscot/tcp-echo-server/echoer"
	"os"
	"os/signal"
	"strconv"
)

func main() {
	// configure the zerolog for pretty commmand line feedback
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// add fields to logger
	log := log.With().Str("func", "main").Logger()

	// default log level is warn
	zerolog.SetGlobalLevel(zerolog.WarnLevel)

	// use getopt to process command line flags. this is used instead of flag pkg due to the strong historic precedent.
	getopt.SetParameters("listenPort")
	verbosity := getopt.Counter('v', "verbosity. can be used multiple times to further increase.")
	quiet := getopt.Bool('q', "quiet. do not print any log info. overrides verbosity flag.")
	announceAlive := getopt.BoolLong("announcealive", 'a', "announce \"alive\" every 5 seconds.")

	// use ParseV2 simply to make sure that we have the v2 version of getopt
	getopt.ParseV2()

	// after flags we should have exactly 1 args
	args := getopt.Args()
	if len(args) != 1 {
		fmt.Printf("error: wrong number of arguments (%d)\n", len(args))
		getopt.Usage()
		os.Exit(1)
	}

	// parse listenPort
	listenPort64, err := strconv.ParseInt(args[0], 10, 16)
	if err != nil {
		fmt.Printf("error: invalid integer for listenPort (got %s)\n", args[0])
		getopt.Usage()
		os.Exit(1)
	}
	listenPort := int(listenPort64)

	// set verbosity. quiet overrides verbosity flag.
	if *quiet {
		zerolog.SetGlobalLevel(zerolog.Disabled)
	} else {
		switch *verbosity {
		case 0:
			zerolog.SetGlobalLevel(zerolog.WarnLevel)
		case 1:
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		case 2:
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		case 3:
			zerolog.SetGlobalLevel(zerolog.TraceLevel)
		default:
			log.Warn().Int("verbosity", *verbosity).Msg("got invalid verbosity count. using max (trace level)")
			zerolog.SetGlobalLevel(zerolog.TraceLevel)
		}
	}

	// establish the context with a cancel function and embed the logger
	ctx, cancel := context.WithCancel(context.Background())
	ctx = log.WithContext(ctx)

	// handle SIGINT (control+c)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		log.Info().Msg("interrupt received. exiting.")
		cancel()
		// don't exit yet. let context cancellation do its magic.
	}()

	// create the server and run it
	srv := echoer.NewTcpEchoServer(listenPort, *announceAlive)
	err = srv.Run(ctx)
	if err != nil {
		log.Error().Err(err).Msg("server exited with error")
		os.Exit(1)
	}

	os.Exit(0)
}
