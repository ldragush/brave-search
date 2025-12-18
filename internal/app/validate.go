package app

import "fmt"

func validateSafeSearch(v string) error {
	switch v {
	case "off", "moderate", "strict":
		return nil
	default:
		return fmt.Errorf("invalid --safesearch value: %q (allowed: off|moderate|strict)", v)
	}
}
