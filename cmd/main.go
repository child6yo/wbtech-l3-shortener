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
	"github.com/child6yo/wbtech-l3-shortener/internal/repository/clickhouse"
	"github.com/child6yo/wbtech-l3-shortener/internal/repository/postgres"
	"github.com/child6yo/wbtech-l3-shortener/internal/repository/redis"
	"github.com/child6yo/wbtech-l3-shortener/internal/usecase"
	"github.com/wb-go/wbf/config"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

const (
	minimizeLinkRoute = "/shorten"
	getFullLinkRoute  = "/s/:short_url"
	analyticsRoute    = "/analytics/:short_url"
)

type appConfig struct {
	address string

	pgHost     string
	pgPort     string
	pgUsername string
	pgDBName   string
	pgPassword string
	pgSSLMode  string

	chHost     string
	chPort     string
	chUsername string
	chDBName   string
	chPassword string

	redisAddr     string
	redisPassword string
	redisDB       int
}

func initConfig(configFilePath, envFilePath, envPrefix string) (*appConfig, error) {
	appConfig := &appConfig{}

	cfg := config.New()

	err := cfg.Load(configFilePath, envFilePath, envPrefix)
	if err != nil {
		return appConfig, fmt.Errorf("failed to load config: %w", err)
	}

	appConfig.address = cfg.GetString("app_address")

	// PostgreSQL
	appConfig.pgHost = cfg.GetString("PG_HOST")
	appConfig.pgPort = cfg.GetString("PG_PORT")
	appConfig.pgPassword = cfg.GetString("PG_PASSWORD")
	appConfig.pgUsername = cfg.GetString("PG_USER")
	appConfig.pgDBName = cfg.GetString("PG_DB")
	appConfig.pgSSLMode = cfg.GetString("PG_SSLMODE")

	// ClickHouse
	appConfig.chHost = cfg.GetString("CH_HOST")
	appConfig.chPort = cfg.GetString("CH_PORT")
	appConfig.chUsername = cfg.GetString("CH_USER")
	appConfig.chPassword = cfg.GetString("CH_PASSWORD")
	appConfig.chDBName = cfg.GetString("CH_DB")

	// Redis
	appConfig.redisAddr = cfg.GetString("REDIS_ADDR")
	appConfig.redisPassword = cfg.GetString("REDIS_PASSWORD")
	appConfig.redisDB = cfg.GetInt("REDIS_DB")

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
		fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
			cfg.pgHost, cfg.pgPort, cfg.pgUsername, cfg.pgDBName, cfg.pgPassword, cfg.pgSSLMode),
	)
	if err != nil {
		lgr.Fatal().Err(err).Send()
	}

	adb, err := clickhouse.NewClickhouseDB(
		fmt.Sprintf("%s:%s", cfg.chHost, cfg.chPort),
		cfg.chUsername, cfg.chPassword, cfg.chDBName)
	if err != nil {
		lgr.Fatal().Err(err).Send()
	}

	rds, err := redis.NewRedis(cfg.redisAddr, cfg.redisPassword, cfg.redisDB)
	if err != nil {
		lgr.Fatal().Err(err).Send()
	}

	lr := postgres.NewLinksRepository(db)
	tr := clickhouse.NewTransitsRepository(adb)

	sh := usecase.NewLinksShortener(lr, rds)
	ans := usecase.NewAnalyticsManager(tr)

	sc := httpctrl.NewShortenerController(sh, ans)
	mdlw := httpctrl.NewMiddleware(logger.NewLoggerAdapter(lgr))

	srv := ginext.New("")
	srv.Use(ginext.Logger(), ginext.Recovery(), mdlw.ErrHandlingMiddleware())
	srv.POST(minimizeLinkRoute, sc.Shorten)
	srv.GET(getFullLinkRoute, sc.Redirect)
	srv.GET(analyticsRoute, sc.GetAnalytics)

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

	if err := db.Master.Close(); err != nil {
		lgr.Err(err).Send()
	}

	if err := adb.Close(); err != nil {
		lgr.Err(err).Send()
	}

	if err := rds.Client.Close(); err != nil {
		lgr.Err(err).Send()
	}

	wg.Wait()

	lgr.Info().Msg("app exited")
}
