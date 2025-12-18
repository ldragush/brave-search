package brave

type SearchRequest struct {
	Q          string
	Count      int
	Offset     int
	SafeSearch string
	Freshness  string
}
