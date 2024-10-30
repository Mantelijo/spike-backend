package main

import (
	"log/slog"
	"os"

	"github.com/Mantelijo/spike-backend/internal/api"
	"github.com/Mantelijo/spike-backend/internal/data"
	"github.com/Mantelijo/spike-backend/internal/data/cache"
	"github.com/Mantelijo/spike-backend/internal/data/database"
	"github.com/Mantelijo/spike-backend/internal/svc"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

func main() {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	})
	slog.SetDefault(slog.New(handler))

	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		slog.Error("reading .env", slog.Any("error", err))
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: viper.GetString("REDIS_CONN_URL"),
	})
	c := cache.NewRedisCache(rdb)

	dbSvc, err := database.NewDbService(viper.GetString("DB_DSN"))
	if err != nil {
		slog.Error("connecting to db", slog.Any("error", err))
	}
	ds := &data.DataStore{
		RedisCache: c,
		DBService:  dbSvc,
	}

	go svc.RunDataReconciler(
		ds.RedisCache,
		ds.DBService,
	)

	api.NewHttpApi(viper.GetString("API_ADDR"), viper.GetString("API_PORT"), ds).Start()
}
