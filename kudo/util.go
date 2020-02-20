package main

import (
	"fmt"
	"strings"
)

func idParts(id string) (string, string, error) {
	parts := strings.Split(id, "_")
	if len(parts) != 2 {
		err := fmt.Errorf("Unexpected ID format (%q), expected %q", id, "namespace/name")
		return "", "", err
	}

	return parts[0], parts[1], nil
}

func id(name, namespace string) string {
	return fmt.Sprintf("%v_%v", name, namespace)
}
