package brave

import "fmt"

type APIError struct {
	StatusCode int
	Status     string
	Body       string
}

func (e *APIError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("api error: status=%d (%s)", e.StatusCode, e.Status)
	}
	return fmt.Sprintf("api error: status=%d (%s): %s", e.StatusCode, e.Status, e.Body)
}
