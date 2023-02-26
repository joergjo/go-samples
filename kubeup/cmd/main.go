package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/joergjo/go-samples/kubeup"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Adjust zerlog's configuration so it mirrors the CloudEvents SDK log output
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.MessageFieldName = "msg"
	zerolog.TimestampFieldName = "ts"
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	port := flag.Int("port", 8000, "HTTP listen port")
	path := flag.String("path", "/webhook", "WebHook path")
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Level(zerolog.DebugLevel)
	}
	var opts []kubeup.Options = []kubeup.Options{
		kubeup.WithLogging(),
	}
	if envVars := getEnvVars("KU_SENDGRID_APIKEY",
		"KU_SENDGRID_FROM",
		"KU_SENDGRID_TO",
		"KU_SENDGRID_SUBJECT"); envVars != nil {
		email := kubeup.EmailTemplate{
			From:    envVars["KU_SENDGRID_FROM"],
			To:      envVars["KU_SENDGRID_TO"],
			Subject: envVars["KU_SENDGRID_SUBJECT"],
			Templ:   nil,
		}
		opts = append(
			opts,
			kubeup.WithSendgrid(envVars["KU_SENDGRID_APIKEY"], email))
	}
	if envVars := getEnvVars("KU_SMTP_HOST",
		"KU_SMTP_PORT",
		"KU_SMTP_USERNAME",
		"KU_SMTP_PASSWORD",
		"KU_SMTP_FROM",
		"KU_SMTP_TO",
		"KU_SMTP_SUBJECT"); envVars != nil {
		email := kubeup.EmailTemplate{
			From:    envVars["KU_SMTP_FROM"],
			To:      envVars["KU_SMTP_TO"],
			Subject: envVars["KU_SMTP_SUBJECT"],
			Templ:   nil,
		}
		port, err := strconv.Atoi(envVars["KU_SMTP_PORT"])
		if err != nil {
			log.Fatal().Err(err).Msg("Fatal error parsing SMTP port")
		}
		opts = append(
			opts,
			kubeup.WithSMTP(
				envVars["KU_SMTP_HOST"],
				port,
				envVars["KU_SMTP_USERNAME"],
				envVars["KU_SMTP_PASSWORD"],
				email))
	}

	p, err := kubeup.NewPublisher(opts...)
	if err != nil {
		log.Fatal().Err(err).Msg("Invalid configuration")
	}

	h, err := kubeup.NewCloudEventHandler(context.Background(), p)
	if err != nil {
		log.Fatal().Err(err).Msg("Fatal error creating CloudEvent receiver")
	}

	s := newServer(*port, *path, h)
	done := make(chan struct{})
	go shutdown(s, done, 10*time.Second)

	log.Info().Msgf("Starting server on port %d", *port)
	err = s.ListenAndServe()
	log.Info().Msgf("Waiting for server to shut down")
	<-done
	log.Info().Err(err).Msg("Server has shut down")
}

func getEnvVars(vars ...string) map[string]string {
	envVars := make(map[string]string, 4)
	for _, k := range vars {
		v, ok := os.LookupEnv(k)
		if !ok {
			return nil
		}
		log.Debug().Msgf("%s=%q", k, v)
		envVars[k] = v
	}
	return envVars
}

func newServer(port int, path string, h http.Handler) *http.Server {
	mux := http.NewServeMux()
	mux.Handle(path, h)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	s := http.Server{Addr: fmt.Sprintf(":%d", port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      mux}
	return &s
}

func shutdown(srv *http.Server, srvClosed chan<- struct{}, timeout time.Duration) {
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigch
	log.Warn().Msgf("Received signal %v, shutting down", sig)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Error shutting down server")
	}
	close(srvClosed)
}
