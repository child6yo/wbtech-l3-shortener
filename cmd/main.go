package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	httpctrl "github.com/child6yo/wbtech-l3-shortener/internal/controller/http"
	"github.com/child6yo/wbtech-l3-shortener/internal/logger"
	"github.com/child6yo/wbtech-l3-shortener/internal/repository/postgres"
	"github.com/child6yo/wbtech-l3-shortener/internal/usecase"
	"github.com/wb-go/wbf/config"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

const (
	minimizeLinkRoute = "/shorten"
)

type appConfig struct {
	address string

	pgHost     string
	pgPort     string
	pgUsername string
	pgDBName   string
	pgPassword string
	pgSSLMode  string
}

func initConfig(configFilePath, envFilePath, envPrefix string) (*appConfig, error) {
	appConfig := &appConfig{}

	cfg := config.New()

	err := cfg.Load(configFilePath, envFilePath, envPrefix)
	if err != nil {
		return appConfig, fmt.Errorf("failed to load config: %w", err)
	}

	appConfig.address = cfg.GetString("app_address")

	appConfig.pgHost = cfg.GetString("pg_host")
	appConfig.pgPort = cfg.GetString("pg_port")
	appConfig.pgPassword = cfg.GetString("PG_PASSWORD")
	appConfig.pgUsername = cfg.GetString("pg_username")
	appConfig.pgDBName = cfg.GetString("pg_db_name")
	appConfig.pgSSLMode = cfg.GetString("pg_ssl_mode")

	return appConfig, nil
}

func main() {
	var wg sync.WaitGroup

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	zlog.InitConsole()
	lgr := zlog.Logger

	cfg, err := initConfig("config/config.yml", ".env", "")
	if err != nil {
		lgr.Fatal().Err(err).Send()
	}

	db, err := postgres.NewMSPostgresDB(
		cfg.pgHost, cfg.pgPort, cfg.pgUsername,
		cfg.pgDBName, cfg.pgPassword, cfg.pgSSLMode,
	)

	lr := postgres.NewLinksRepository(db)
	sh := usecase.NewLinksShortener(lr)
	sc := httpctrl.NewShortenerController(sh)
	mdlw := httpctrl.NewMiddleware(logger.NewLoggerAdapter(lgr))

	srv := ginext.New("")
	srv.Use(ginext.Logger(), ginext.Recovery(), mdlw.ErrHandlingMiddleware())
	srv.POST(minimizeLinkRoute, sc.Shorten)

	httpServer := &http.Server{
		Addr:    cfg.address,
		Handler: srv,
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			lgr.Err(err).Send()
		}
	}()

	<-ctx.Done()
	lgr.Info().Msg("shutting down gracefully...")

	if err := httpServer.Shutdown(context.Background()); err != nil {
		lgr.Err(err).Send()
	}

	wg.Wait()

	lgr.Info().Msg("app exited")
}
