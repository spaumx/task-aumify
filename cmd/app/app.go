package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"task-aumify/internal/priceChecker"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("Received termination signal. Cancelling context...")
		cancel()
	}()

	checker := priceChecker.NewClient("", "")
	if err := checker.CheckPrices(ctx); err != nil {
		log.Fatal(err)
	}
	log.Println("Program terminated gracefully.")
}
