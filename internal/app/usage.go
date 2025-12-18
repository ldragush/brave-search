package app

import (
	"fmt"
	"io"
)

func PrintUsage(w io.Writer) {
	fmt.Fprintln(w, `
USAGE:
  brave-search [options]

INPUT:
  -sq,  --search-query <q>            Search query (repeatable). Comma-separated supported.
  -sqf, --search-query-file <file>    File with queries (one per line). Repeatable. Comma-separated supported.
       --stdin                        Enable stdin/pipeline input explicitly.

API:
  -ak, --api-key <key>                Brave API key (overrides config rotation).
  -c,  --count <n>                    Results per request (1..20). Default: 20
  -p,  --page <n>                     Page/offset (0..9). Default: 0
       --all                          Auto-scroll pages (until exhaustion or page=9)
       --safesearch <off|moderate|strict>  Default: off
       --freshness <pd|pw|pm|py|YYYY-MM-DDtoYYYY-MM-DD>

NETWORK:
       --proxy <url>                  http://host:port or socks5://host:port
       --proxy-auth <user:pass>       Proxy credentials (HTTP or SOCKS5)
       --no-proxy                     Ignore proxy env vars
  -k,  --insecure                     TLS bypass (curl-style)

PERF / OUTPUT:
       --threads <n>                  Worker threads. Default: 1
  -rl, --rate-limit <rps>             Max requests/sec (global). Default: 5
  -o,  --output <file>                Save deduped+sorted results to file
       --retry-count <n>              Max retries on API errors. Default: 3
       --retry-wait-time <dur>        Wait between retries (e.g. 3s, 500ms). Default: 3s

MISC:
  -s,  --silent                       Only print extracted results
       --debug                        Debug logs (incl. rate limit headers)
       --version                      Print version and exit
  -h,  --help                         Show this help

EXAMPLES:
  brave-search -sq "site:thc.org"
  brave-search -sq "site:thc.org,site:example.com" --threads 5 --rate-limit 10
  brave-search -sqf queries.txt --all
  cat queries.txt | brave-search --stdin --all --silent
`)
}
