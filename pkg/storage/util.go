package storage

import "fmt"

func IsOk(res interface{}) error {
	if res.(string) == "OK" {
		return nil
	} else {
		return fmt.Errorf("Not Ok. Got [%T]%+v", res, res)
	}
}
