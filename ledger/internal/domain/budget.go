package domain

import "errors"

type Budget struct {
	Category string 
	Limit    float64 
	Period   string 
}

func (b Budget) Validate() error {
	if b.Category == "" {
		return errors.New("budget category cannot be empty")
	}
	if b.Limit <= 0 {
		return errors.New("budget limit should be > 0")
	}

	return nil
}