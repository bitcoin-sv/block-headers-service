package database

import "fmt"

func wrapIfNeeded(outer, inner error, msg string) error {
	if outer != nil {
		return fmt.Errorf("%w. %s: %w", outer, msg, inner)
	}

	return inner
}
