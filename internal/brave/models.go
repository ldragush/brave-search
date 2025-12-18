package brave

type SearchResponse struct {
	Type  string `json:"type"`
	Query struct {
		MoreResultsAvailable bool `json:"more_results_available"`
	} `json:"query"`
	Web struct {
		Type    string `json:"type"`
		Results []struct {
			URL string `json:"url"`
		} `json:"results"`
	} `json:"web"`
}

func (r SearchResponse) WebURLs() []string {
	out := make([]string, 0, len(r.Web.Results))
	for _, it := range r.Web.Results {
		if it.URL != "" {
			out = append(out, it.URL)
		}
	}
	return out
}

func (r SearchResponse) MoreResultsAvailable() bool {
	return r.Query.MoreResultsAvailable
}
