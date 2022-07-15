package main

import (
	"context"
	"crawler/internal"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	wAmnt := flag.Int("n", 1, "amount of workers to crawl a page")
	rate := flag.Int("r", 15, "pages per second crawling limit")
	startUrl := flag.String("s", "https://www.ietf.org/rfc/rfc1912.txt", "url to start crawling")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-sigCh
		log.Println("received stop signal")
		cancel()
	}()

	c := internal.NewCrawler(ctx, internal.Config{
		WorkersAmount:  *wAmnt,
		PagesPerSecond: *rate,
	})

	c.Run(*startUrl)

	log.Println("crawling complete")
}
