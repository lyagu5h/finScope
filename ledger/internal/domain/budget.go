package domain

import "errors"

type Budget struct {
	Category string 
	Limit    float64 
}

func (b Budget) Validate() error {
	if b.Category == "" {
		return errors.New("validation failed: budget category cannot be empty")
	}
	if b.Limit <= 0 {
		return errors.New("validation failed: budget limit should be > 0")
	}

	return nil
}