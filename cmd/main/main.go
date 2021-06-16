package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/amorey/elbx/pkg/worker"
	"github.com/amorey/elbx/pkg/models"
	"github.com/amorey/elbx/pkg/sqsmonitor"
)

// Get env var or default
func getEnv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		if value != "" {
			return value
		}
	}
	return fallback
}

// Parse env var to boolean if key exists
func getBoolEnv(key string, fallback bool) bool {
	envStrValue := getEnv(key, "")
	if envStrValue == "" {
		return fallback
	}
	envBoolValue, err := strconv.ParseBool(envStrValue)
	if err != nil {
		panic("Env Var " + key + " must be either true or false")
	}
	return envBoolValue
}

func main() {
	var wg sync.WaitGroup

	// init context and listen for termination signals
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	
	// parse flags
	debug := flag.Bool("debug", getBoolEnv("DEBUG", false), "Set log level to debug")
	queueUrl := flag.String("queue-url", getEnv("QUEUE_URL", ""), "SQS queue URL")
	flag.Parse()

	// configure logger
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
	
	zerolog.TimestampFunc = func() time.Time {
		return time.Now().UTC()
	}
	
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out: os.Stderr,
		TimeFormat: "2006-01-02T15:04:05Z",
		NoColor: true,
	})

	// init comms channel
	commsChan := make(chan models.EventBridgeEvent)
	defer close(commsChan)

	// init sqs monitor
	m, err := sqsmonitor.New(queueUrl)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	
	// init elbx worker
	w, err := worker.New()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	
	// start goroutines
	wg.Add(2)
	go m.WatchForSQSMessages(ctx, commsChan, &wg)
	go w.WatchForEventBridgeEvents(ctx, commsChan, &wg)
	
	log.Info().Msg("NTH-elbx has started successfully!")
	
	// wait for context
	<- ctx.Done()
	stop() // stop receiving signals as soon as possible

	// wait for goroutines to exit
	log.Info().Msg("NTH-elbx is shutting down")
	wg.Wait()
}
