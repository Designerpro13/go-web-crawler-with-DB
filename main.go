package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github/yeshu2004/go-epics/api"
	"github/yeshu2004/go-epics/crawler"
	db "github/yeshu2004/go-epics/db"
	"github/yeshu2004/go-epics/indexer"
)

const (
	expected  = 10000000
	fp_rate   = 0.001
	bfKey     = "filter_bloom"
	batchSize = 100
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Println("Starting web crawler API server")

	postgresConnStr := "host=localhost port=5432 user=postgres password=crawler123 dbname=crawler sslmode=disable"
	postgresDB, err := db.PostgresInit(ctx, postgresConnStr)
	if err != nil {
		log.Fatal("PostgreSQL connection failed:", err)
	}
	defer postgresDB.Close()

	if err := postgresDB.InitSchema(ctx); err != nil {
		log.Fatal("Schema initialization failed:", err)
	}

	rdb, err := db.RedisInit(ctx)
	if err != nil {
		log.Fatal("Redis connection failed:", err)
	}
	defer rdb.Close()

	if err := db.InitializeBloomFilter(ctx, rdb, bfKey, fp_rate, int64(expected)); err != nil {
		log.Fatal("Bloom filter init failed:", err)
	}

	invertedIdx := indexer.NewInvertedIndex(rdb)
	processor := indexer.NewProcessor(invertedIdx, postgresDB, batchSize)

	crawlerInstance := crawler.New(postgresDB, rdb, processor)
	redisQueries := db.NewRedisQueries(rdb)

	server := api.NewServer(postgresDB, redisQueries, crawlerInstance)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		log.Println("Shutting down gracefully...")
		crawlerInstance.Stop()
		cancel()
		os.Exit(0)
	}()

	if err := server.Start("8080"); err != nil {
		log.Fatal("Server failed:", err)
	}
}
