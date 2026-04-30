package main

import (
	"context"
	"crypto-notifier/gen/alert/v1/alertv1connect"
	"crypto-notifier/internal/alert"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	pricev1 "crypto-notifier/gen/price/v1"
	"crypto-notifier/gen/price/v1/pricev1connect"
	"crypto-notifier/internal/pricefeed"

	"connectrpc.com/connect"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type priceServer struct {
	pricev1connect.UnimplementedPriceServiceHandler
	hub *pricefeed.Hub
}

type config struct {
	DatabaseURL    string
	Port           string
	BinanceSymbols []string
}

func loadConfig() config {
	symbols := getEnv("BINANCE_SYMBOLS", "btcusdt")
	return config{
		DatabaseURL:    getEnv("DATABASE_URL", "postgresql://postgres:postgres@localhost:5433/postgres?sslmode=disable"),
		Port:           getEnv("PORT", "8081"),
		BinanceSymbols: strings.Split(symbols, ","),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func (s *priceServer) GetPrice(
	ctx context.Context,
	req *connect.Request[pricev1.GetPriceRequest],
) (*connect.Response[pricev1.GetPriceResponse], error) {

	return connect.NewResponse(&pricev1.GetPriceResponse{
		Price: 100,
	}), nil
}

func (s *priceServer) Subscribe(
	ctx context.Context,
	req *connect.Request[pricev1.SubscribeRequest],
	stream *connect.ServerStream[pricev1.SubscribeResponse],
) error {
	ch := s.hub.Subscribe()
	defer s.hub.Unsubscribe(ch)

	for {
		select {
		case update := <-ch:

			if !contains(req.Msg.Symbols, update.Symbol) {
				continue

			}

			err := stream.Send(&pricev1.SubscribeResponse{
				Symbol:    update.Symbol,
				Price:     update.Price,
				Timestamp: uint64(update.Time),
			})
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func contains(symbols []string, symbol string) bool {
	for _, s := range symbols {
		if strings.EqualFold(s, symbol) {
			return true
		}
	}
	return false

}

func connectDB(dbUrl string) *sql.DB {
	db, err := sql.Open("pgx", dbUrl)
	if err != nil {
		slog.Error("Error connecting to database:", "error", err)
		os.Exit(1)
	}

	return db
}

func main() {

	cfg := loadConfig()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	db := connectDB(cfg.DatabaseURL)
	defer db.Close()

	hub := pricefeed.NewHub()
	binanceClient := pricefeed.NewBinanceClient(hub, cfg.BinanceSymbols)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go hub.Run(ctx)
	go func() {
		err := binanceClient.Run(ctx)
		if err != nil {
			slog.Error("Error running binance client:", "error", err)
		}
	}()

	queries := alert.New(db)
	evaluator := alert.NewEvaluator(hub, queries)
	go evaluator.Run(ctx)

	alertService := alert.NewAlertServer(queries, evaluator.NotifCh, evaluator.RefreshCh)

	mux := http.NewServeMux()

	path, handler := pricev1connect.NewPriceServiceHandler(&priceServer{
		hub: hub,
	})
	mux.Handle(path, handler)

	path, handler = alertv1connect.NewAlertServiceHandler(alertService)
	mux.Handle(path, handler)

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: mux,
	}
	//err := http.ListenAndServe(":8081", mux)
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Error starting server:", "error", err)
			os.Exit(1)
		}
	}()

	slog.Info("Server started", "port", fmt.Sprintf(":%s", cfg.Port))

	<-ctx.Done()
	slog.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := server.Shutdown(shutdownCtx)
	if err != nil {
		slog.Error("Error shutting down server:", "error", err)
		os.Exit(1)
		return
	}

}
