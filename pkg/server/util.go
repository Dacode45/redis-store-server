package server

import "fmt"

func MultiError(errs ...error) error {
	var finalErr error
	for i, err := range errs {
		if err != nil {
			finalErr = fmt.Errorf("[%d] %w: %w", i, err, finalErr)
		}
	}
	return finalErr
}
