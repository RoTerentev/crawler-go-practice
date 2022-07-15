package internal

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sync"
	"time"
)

const RFC_LINK_SEARCH_REGEXP = `\[\s?RFC\s+(\d+)\s?\]`
const RFC_BASE_URL = "https://www.ietf.org/rfc/"
const IDLE_TIMEOUT = time.Second * 3
const CRAWLING_TIMEOUT = time.Minute * 5

type Crawler struct {
	ctx          context.Context
	config       Config
	rateLimiter  *RateLimiter
	pages        sync.Map
	pagesCrawlCh chan string
}

func NewCrawler(ctx context.Context, cfg Config) *Crawler {
	crawlerCtx := context.Context(ctx)
	return &Crawler{
		ctx:          crawlerCtx,
		config:       cfg,
		rateLimiter:  NewLimiter(time.Second*1, int64(cfg.PagesPerSecond)),
		pages:        sync.Map{},
		pagesCrawlCh: make(chan string, cfg.WorkersAmount*cfg.PagesPerSecond),
	}
}

func (c *Crawler) Run(pageUrl string) {
	wg := sync.WaitGroup{}
	for i := 0; i < c.config.WorkersAmount; i++ {
		wg.Add(1)
		go c.worker(&wg)
	}
	c.addPage(pageUrl)
	wg.Wait()
}

func (c *Crawler) addPage(url string) {
	if _, pageExists := c.pages.Load(url); !pageExists {
		c.pages.Store(url, struct{}{})
		c.pagesCrawlCh <- url
	}
}

func (c *Crawler) worker(wg *sync.WaitGroup) {
	idleTimer := time.NewTimer(IDLE_TIMEOUT)

	defer func() {
		idleTimer.Stop()
		wg.Done()
	}()

	// NOTE: how write it clearly?
	for {
		select {
		case <-idleTimer.C:
			return
		case <-c.ctx.Done():
			return
		case urlToCrawl := <-c.pagesCrawlCh:
			select {
			case <-c.ctx.Done():
				return
			default:
				// TODO: crawl execution timeout should be in err log
				idleTimer.Reset(CRAWLING_TIMEOUT)
				if err := c.crawl(urlToCrawl); err != nil {
					log.Println("crwaling error: ", err)
				}
				idleTimer.Reset(IDLE_TIMEOUT)
			}
		}
	}
}

func (c *Crawler) crawl(url string) error {
	// NOTE: is it require to close channel inside limiter?
	c.rateLimiter.Add() <- struct{}{}
	log.Println("crawling: ", url)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("get page err: %w", err)
	}
	defer resp.Body.Close()

	// TODO: check HTTP status

	// NOTE: crawl page implementation
	s := bufio.NewScanner(resp.Body)
	linkRegex := regexp.MustCompile(RFC_LINK_SEARCH_REGEXP)
	pageEndRegex := regexp.MustCompile(`\[Page \d+\]$`)
	contentPage := ""

	for s.Scan() {
		contentPage += s.Text()
		if pageEndRegex.Match(s.Bytes()) {
			for _, match := range linkRegex.FindAllStringSubmatch(contentPage, -1) {
				// RFC number is match[1]
				go c.addPage(fmt.Sprintf("%srfc%s.txt", RFC_BASE_URL, match[1]))
			}
			contentPage = ""
		}
	}

	return nil
}
