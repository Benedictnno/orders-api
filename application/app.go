package application

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type App struct {
	router http.Handler
	rdb    *redis.Client
}

func New() *App {
	app := &App{
		router: loadRoutes(),
		rdb: redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		}),
	}
	return app
}

func (a *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    ":3000",
		Handler: a.router,
	}

	err := a.rdb.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("Error connecting to Redis: %w", err)
	}

	defer func() {
		err := a.rdb.Close()
		if err != nil {
			fmt.Println("Error closing Redis connection:", err)
		}
	}()

	fmt.Println("Connected to Redis successfully")
	ch := make(chan error, 1)
	go func() {

		err = server.ListenAndServe()
		if err != nil {
			ch <- fmt.Errorf("Error starting server: %w", err)
		}
		close(ch)

	}()

	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		timeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err := server.Shutdown(timeout)
		if err != nil {
			return fmt.Errorf("Error shutting down server: %w", err)
		}
	}
	return nil

}
