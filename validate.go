package main

import "fmt"

func validateWebsiteID(id string) error {
	if len(id) == 0 || len(id) > 36 {
		return fmt.Errorf("invalid website ID")
	}
	for _, c := range id {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') || c == '-') {
			return fmt.Errorf("invalid website ID")
		}
	}
	return nil
}
