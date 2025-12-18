package cli

import (
	"flag"
	"io"
	"strings"
	"time"
)

type ParseResult int

const (
	ParseOK ParseResult = iota
	ParseShowHelp
	ParseShowVersion
	ParseError
)

type multiString []string

func (m *multiString) String() string { return strings.Join(*m, ",") }
func (m *multiString) Set(v string) error {
	*m = append(*m, v)
	return nil
}

type Options struct {
	SearchQueries    []string
	SearchQueryFiles []string

	Count       int
	Page        int
	APIKey      string
	All         bool
	UseStdin    bool
	Silent      bool
	Debug       bool
	ProxyURL    string
	ProxyAuth   string
	NoProxy     bool
	InsecureTLS bool

	Threads    int
	OutputFile string
	RateLimit  int

	SafeSearch string
	Freshness  string

	// NEW: retry behavior
	RetryCount    int           // number of retries (attempts = 1 + RetryCount)
	RetryWaitTime time.Duration // wait between retries

	VersionOnly bool
}

func ParseOptions(argv []string) (Options, ParseResult) {
	var opts Options

	fs := flag.NewFlagSet("brave-search", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var sq multiString
	var sqf multiString

	// Query inputs
	fs.Var(&sq, "search-query", "Search query (alias: -sq). Can be repeated. Comma-separated supported.")
	fs.Var(&sq, "sq", "Alias of --search-query.")
	fs.Var(&sqf, "search-query-file", "File with one search query per line (alias: -sqf). Can be repeated. Comma-separated supported.")
	fs.Var(&sqf, "sqf", "Alias of --search-query-file.")

	// API params
	fs.IntVar(&opts.Count, "count", 20, "Max results per API request (1..20). Alias: -c")
	fs.IntVar(&opts.Count, "c", 20, "Alias of --count.")
	fs.IntVar(&opts.Page, "page", 0, "API page/offset (0..9). Alias: -p")
	fs.IntVar(&opts.Page, "p", 0, "Alias of --page.")
	fs.StringVar(&opts.SafeSearch, "safesearch", "off", "SafeSearch mode: off|moderate|strict")
	fs.StringVar(&opts.Freshness, "freshness", "", "Freshness filter: pd|pw|pm|py|YYYY-MM-DDtoYYYY-MM-DD")

	// API key
	fs.StringVar(&opts.APIKey, "api-key", "", "Brave API key (alias: -ak). Overrides config rotation.")
	fs.StringVar(&opts.APIKey, "ak", "", "Alias of --api-key.")

	// Behavior
	fs.BoolVar(&opts.All, "all", false, "Auto-scroll pages until exhaustion (or page=9).")
	fs.BoolVar(&opts.UseStdin, "stdin", false, "Enable stdin (pipeline) input explicitly.")
	fs.BoolVar(&opts.Silent, "silent", false, "Silent mode (only prints extracted results). Alias: -s")
	fs.BoolVar(&opts.Silent, "s", false, "Alias of --silent.")
	fs.BoolVar(&opts.Debug, "debug", false, "Debug mode (prints debugging information).")
	fs.BoolVar(&opts.InsecureTLS, "insecure", false, "TLS bypass (disables certificate verification). Alias: -k")
	fs.BoolVar(&opts.InsecureTLS, "k", false, "Alias of --insecure.")

	// Proxy (HTTP or SOCKS5)
	fs.StringVar(&opts.ProxyURL, "proxy", "", "Proxy URL (http://host:port or socks5://host:port).")
	fs.StringVar(&opts.ProxyAuth, "proxy-auth", "", "Proxy auth in user:pass format (HTTP or SOCKS5).")
	fs.BoolVar(&opts.NoProxy, "no-proxy", false, "Ignore proxy environment variables.")

	// Concurrency / output
	fs.IntVar(&opts.Threads, "threads", 1, "Number of worker threads. Default: 1")
	fs.StringVar(&opts.OutputFile, "output", "", "Save deduped+sorted results to file. Alias: -o")
	fs.StringVar(&opts.OutputFile, "o", "", "Alias of --output.")
	fs.IntVar(&opts.RateLimit, "rate-limit", 5, "Max requests per second (global). Alias: -rl")
	fs.IntVar(&opts.RateLimit, "rl", 5, "Alias of --rate-limit.")

	// NEW: Retry controls
	fs.IntVar(&opts.RetryCount, "retry-count", 3, "Max retries on API error (attempts = 1 + retries). Default: 3")
	fs.DurationVar(&opts.RetryWaitTime, "retry-wait-time", 3*time.Second, "Wait time between retries (e.g. 3s, 1500ms). Default: 3s")

	// Meta
	var help bool
	var versionOnly bool
	fs.BoolVar(&help, "help", false, "Show help.")
	fs.BoolVar(&help, "h", false, "Alias of --help.")
	fs.BoolVar(&versionOnly, "version", false, "Print version and exit.")

	if err := fs.Parse(argv); err != nil {
		return Options{}, ParseError
	}

	if help {
		return Options{Silent: opts.Silent, Debug: opts.Debug}, ParseShowHelp
	}
	if versionOnly {
		opts.VersionOnly = true
		return opts, ParseShowVersion
	}

	opts.SearchQueries = []string(sq)
	opts.SearchQueryFiles = []string(sqf)

	return opts, ParseOK
}
