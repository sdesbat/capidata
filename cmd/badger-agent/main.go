package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/example/capidata/badger/internal/adapter/journal"
)

func main() {
	path := flag.String("journal", "/var/lib/badger", "shared node journal path")
	flag.Parse()
	j, err := journal.OpenJSONL(*path)
	if err != nil {
		slog.Error("open journal", "error", err)
		os.Exit(1)
	}
	defer j.Close()
	slog.Info("badger agent ready", "journal", *path)
}
