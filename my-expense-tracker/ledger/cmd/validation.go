package main

import "errors"

func ValidateTransaction(amount float64, category string) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}
	if category == "" {
		return errors.New("category cannot be empty")
	}
	if len(category) > 50 {
		return errors.New("category name too long")
	}
	return nil
}
