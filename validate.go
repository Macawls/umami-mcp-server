package main

import "fmt"

func validateWebsiteID(id string) error {
	return validateID(id, "website ID")
}

func validateSessionID(id string) error {
	return validateID(id, "session ID")
}

func validateID(id, label string) error {
	if id == "" || len(id) > 36 {
		return fmt.Errorf("invalid %s", label)
	}
	for _, c := range id {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') || c == '-') {
			return fmt.Errorf("invalid %s", label)
		}
	}
	return nil
}
