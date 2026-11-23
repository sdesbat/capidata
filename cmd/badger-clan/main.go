package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/example/capidata/badger/internal/adapter/journal"
	"github.com/example/capidata/badger/internal/adapter/postgresprotocol"
	"github.com/example/capidata/badger/internal/domain"
	"github.com/example/capidata/badger/internal/usecase"
)

func main() {
	root := flag.String("journal", "/var/lib/badger", "node journal path")
	tenant := flag.String("tenant", "default", "tenant namespace used when metadata is unavailable")
	flag.Parse()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	j, err := journal.OpenJSONL(*root)
	if err != nil {
		logger.Error("journal", "error", err)
		os.Exit(1)
	}
	defer j.Close()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	correlator := usecase.NewCorrelator(*tenant)
	service := usecase.NewStatusService(j)
	// In production the protocol messages are fed by the Linux eBPF source. stdin is a small
	// framing harness used by tests and by local debugging; each line is hex protocol payload.
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		var hexPayload string
		if _, err := fmt.Fscan(os.Stdin, &hexPayload); err != nil {
			return
		}
		payload, err := hex.DecodeString(hexPayload)
		if err != nil {
			continue
		}
		messages, err := (&postgresprotocol.Parser{}).Feed(payload)
		if err != nil {
			continue
		}
		for _, message := range messages {
			for _, event := range correlator.Consume(message) {
				if err := service.Record(ctx, event); err != nil {
					logger.Error("record", "error", err)
				}
			}
		}
	}
}

var _ = domain.Pending
