package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ko4etov/gophkeeper/internal/client/config"
	"github.com/Ko4etov/gophkeeper/internal/client/crypto"
	grpcclient "github.com/Ko4etov/gophkeeper/internal/client/grpc_client"
	"github.com/Ko4etov/gophkeeper/internal/client/service/auth"
	"github.com/Ko4etov/gophkeeper/internal/client/service/data"
	"github.com/Ko4etov/gophkeeper/internal/client/service/storage"
	"github.com/Ko4etov/gophkeeper/internal/client/service/sync"
	"github.com/Ko4etov/gophkeeper/internal/client/shell"
)

var (
    // Эти переменные заполняются при сборке:
    // go build -ldflags="-X 'main.version=1.0.0' -X 'main.buildDate=$(date)'"
    version   = "dev"
    buildDate = "unknown"
)

func main() {
    mainCtx, cancel := context.WithCancel(context.Background())
    defer cancel()

    config, err := config.New()
    if err != nil {
        fmt.Fprintf(os.Stderr, "❌ Failed to load config: %v\n", err)
        os.Exit(1)
    }
    
    buildInfo := &grpcclient.BuildInfo{
        Version:   version,
        BuildDate: buildDate,
    }
    
    grpcclient, err := grpcclient.New(mainCtx, config, buildInfo)
    if err != nil {
        fmt.Fprintf(os.Stderr, "❌ Failed to create client: %v\n", err)
        os.Exit(1)
    }

    session := crypto.NewSession(15 * time.Minute)

    storage, err := storage.NewStorage(config, session)
    if err != nil {
        fmt.Fprintf(os.Stderr, "❌ Failed to create shell: %v\n", err)
        os.Exit(1)
    }

    authService, err := auth.NewAuthService(mainCtx, grpcclient, storage)
    if err != nil {
        fmt.Fprintf(os.Stderr, "❌ Failed to create shell: %v\n", err)
        os.Exit(1)
    }

    dataService, err := data.NewDataService(mainCtx, grpcclient, storage)
    if err != nil {
        fmt.Fprintf(os.Stderr, "❌ Failed to create shell: %v\n", err)
        os.Exit(1)
    }

    syncManager := sync.NewSyncManager(grpcclient, dataService, authService)
    
    sh, err := shell.New(authService, config, buildInfo, dataService, syncManager, session)
    if err != nil {
        fmt.Fprintf(os.Stderr, "❌ Failed to create shell: %v\n", err)
        os.Exit(1)
    }
    
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
    go func() {
        <-sigCh
        cancel()
        os.Exit(0)
    }()
    
    // Запускаем shell
    if err := sh.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "❌ Shell error: %v\n", err)
        os.Exit(1)
    }
}