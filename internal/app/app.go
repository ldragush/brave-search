package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"
	"time"

	"github.com/haltman-io/brave-search/internal/brave"
	"github.com/haltman-io/brave-search/internal/cli"
	"github.com/haltman-io/brave-search/internal/config"
	"github.com/haltman-io/brave-search/internal/input"
	"github.com/haltman-io/brave-search/internal/output"
	"github.com/haltman-io/brave-search/internal/proxy"
	"github.com/haltman-io/brave-search/internal/ratelimit"
	"github.com/haltman-io/brave-search/internal/store"
	"github.com/haltman-io/brave-search/internal/ui"
)

func Run(codename, version string, argv []string) int {
	opts, parseResult := cli.ParseOptions(argv)

	if parseResult == cli.ParseShowHelp {
		if !opts.Silent {
			ui.PrintBanner(os.Stdout, codename, version)
		}
		PrintUsage(os.Stdout)
		return 0
	}
	if parseResult == cli.ParseShowVersion {
		if !opts.Silent {
			ui.PrintBanner(os.Stdout, codename, version)
		}
		fmt.Fprintln(os.Stdout, version)
		return 0
	}
	if parseResult == cli.ParseError {
		if !opts.Silent {
			ui.PrintBanner(os.Stdout, codename, version)
		}
		ui.NewLogger(opts.Silent, opts.Debug).Errorf("invalid arguments. use --help to see usage.")
		return 2
	}

	log := ui.NewLogger(opts.Silent, opts.Debug)

	if !opts.Silent {
		ui.PrintBanner(os.Stdout, codename, version)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	cfgPath, err := config.EnsureConfigFile()
	if err != nil {
		log.Errorf("failed to ensure config file: %v", err)
		return 1
	}
	if opts.Debug && !opts.Silent {
		log.Debugf("config: %s", cfgPath)
	}

	keyProvider, err := resolveAPIKeyProvider(opts, cfgPath)
	if err != nil {
		log.Errorf("%v", err)
		return 1
	}

	queries, err := input.GatherQueries(opts, os.Stdin)
	if err != nil {
		log.Errorf("failed to gather search queries: %v", err)
		return 1
	}
	if len(queries) == 0 {
		log.Errorf("no search query provided. use --search-query/-sq, --search-query-file/-sqf, or pipe with --stdin.")
		return 1
	}

	if opts.Count <= 0 || opts.Count > 20 {
		log.Errorf("--count must be between 1 and 20.")
		return 1
	}
	if opts.Page < 0 || opts.Page > 9 {
		log.Errorf("--page must be between 0 and 9.")
		return 1
	}
	if err := validateSafeSearch(opts.SafeSearch); err != nil {
		log.Errorf("%v", err)
		return 1
	}
	if opts.Threads <= 0 {
		opts.Threads = 1
	}
	if opts.RateLimit <= 0 {
		opts.RateLimit = 5
	}
	if opts.RetryCount < 0 {
		opts.RetryCount = 0
	}
	if opts.RetryWaitTime <= 0 {
		opts.RetryWaitTime = 3 * time.Second
	}

	transport, err := proxy.BuildTransport(opts)
	if err != nil {
		log.Errorf("failed to build http transport: %v", err)
		return 1
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   40 * time.Second,
	}

	lim := ratelimit.New(opts.RateLimit)
	defer lim.Stop()

	braveClient := brave.NewClient(httpClient)

	writer := output.NewWriter(os.Stdout, opts.Silent)
	results := store.NewResultStore()

	jobs := make(chan string)
	var wg sync.WaitGroup

	for i := 0; i < opts.Threads; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for q := range jobs {
				if ctx.Err() != nil {
					return
				}
				if !opts.Silent {
					log.Infof("query: %s", q)
				}
				if err := runQuery(ctx, log, writer, results, lim, braveClient, keyProvider, opts, q); err != nil {
					// after retries exhausted, this is a hard failure per your requirement
					if !opts.Silent {
						log.Errorf("fatal: query failed after %d attempts (%s): %v", 1+opts.RetryCount, q, err)
					}
					cancel()
					return
				}

			}
		}(i + 1)
	}

	for _, q := range queries {
		select {
		case <-ctx.Done():
			break
		case jobs <- q:
		}
	}
	close(jobs)
	wg.Wait()

	if opts.OutputFile != "" {
		all := results.Values()
		sort.Strings(all)
		if err := output.WriteLinesToFile(opts.OutputFile, all); err != nil {
			log.Errorf("failed to write output file: %v", err)
			return 1
		}
		if !opts.Silent {
			log.Infof("saved: %s (%d unique)", opts.OutputFile, len(all))
		}
	}

	return 0
}

func resolveAPIKeyProvider(opts cli.Options, cfgPath string) (config.APIKeyProvider, error) {
	if opts.APIKey != "" {
		return config.NewStaticKey(opts.APIKey), nil
	}

	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}
	if len(cfg.APIKeys) == 0 {
		return nil, fmt.Errorf("no API key configured. set --api-key or add one or more keys to %s under api_keys[].", cfgPath)
	}
	return config.NewKeyRing(cfg.APIKeys), nil
}

func runQuery(
	ctx context.Context,
	log *ui.Logger,
	writer *output.Writer,
	results *store.ResultStore,
	lim *ratelimit.Limiter,
	c *brave.Client,
	keyProvider config.APIKeyProvider,
	opts cli.Options,
	searchQuery string,
) error {
	page := opts.Page

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		req := brave.SearchRequest{
			Q:          searchQuery,
			Count:      opts.Count,
			Offset:     page,
			SafeSearch: opts.SafeSearch,
			Freshness:  opts.Freshness,
		}

		resp, headers, err := searchWithRetry(ctx, log, lim, c, keyProvider, req, opts)
		if err != nil {
			return err
		}

		if opts.Debug && !opts.Silent {
			log.Debugf("ratelimit: limit=%q policy=%q remaining=%q reset=%q",
				headers.Get("X-Ratelimit-Limit"),
				headers.Get("X-Ratelimit-Policy"),
				headers.Get("X-Ratelimit-Remaining"),
				headers.Get("X-Ratelimit-Reset"),
			)
		}

		for _, u := range resp.WebURLs() {
			if results.Add(u) {
				writer.Println(u)
			}
		}

		if !opts.All {
			return nil
		}

		if !resp.MoreResultsAvailable() {
			return nil
		}
		if page >= 9 {
			if !opts.Silent {
				log.Warnf("pagination stopped at page=9 (API offset max is 9).")
			}
			return nil
		}

		page++
		if opts.Debug && !opts.Silent {
			log.Debugf("auto-scroll: next page=%d", page)
		}
	}
}

func searchWithRetry(
	ctx context.Context,
	log *ui.Logger,
	lim *ratelimit.Limiter,
	c *brave.Client,
	apiKeyProvider config.APIKeyProvider,
	req brave.SearchRequest,
	opts cli.Options,
) (brave.SearchResponse, http.Header, error) {

	attempts := 1 + opts.RetryCount
	var lastErr error

	for attempt := 1; attempt <= attempts; attempt++ {
		if ctx.Err() != nil {
			return brave.SearchResponse{}, nil, ctx.Err()
		}

		if err := lim.Wait(ctx); err != nil {
			return brave.SearchResponse{}, nil, err
		}

		apiKey := apiKeyProvider.Next()

		resp, headers, err := c.Search(ctx, apiKey, req, opts.Debug)
		if err == nil {
			return resp, headers, nil
		}

		lastErr = err

		// Decide if retryable
		retryable := false

		// API status-based retry
		if apiErr, ok := err.(*brave.APIError); ok {
			switch {
			case apiErr.StatusCode == 429:
				retryable = true
			case apiErr.StatusCode >= 500 && apiErr.StatusCode <= 599:
				retryable = true
			default:
				retryable = false
			}
		} else {
			// Network / decode / transient errors: retryable
			retryable = true
		}

		if !retryable || attempt == attempts {
			return brave.SearchResponse{}, nil, lastErr
		}

		if opts.Debug && !opts.Silent {
			log.Debugf("retry: attempt=%d/%d wait=%s err=%v", attempt, attempts, opts.RetryWaitTime, err)
		}

		// Wait before next attempt
		select {
		case <-ctx.Done():
			return brave.SearchResponse{}, nil, ctx.Err()
		case <-time.After(opts.RetryWaitTime):
		}
	}

	return brave.SearchResponse{}, nil, lastErr
}
