package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/luoxiaojun1992/miniprogram/internal/app"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	// configPath may be empty; configuration is then loaded from APP_* environment variables.

	cfg, err := app.InitConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	p, err := app.NewProvider(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "init provider: %v\n", err)
		os.Exit(1)
	}

	router := app.InitRouter(p)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		p.Log.Infof("server starting on :%d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			p.Log.WithError(err).Fatal("server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	p.Log.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		p.Log.WithError(err).Error("server shutdown error")
	}
	p.Log.Info("server exited")
}
