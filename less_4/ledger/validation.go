package main

import (
	"errors"
	"fmt"
)

type Validatable interface {
	Validate() error
}

func CheckValid(v Validatable) error {
	return v.Validate()
}

func (t Transaction) Validate() error {
	if t.Category == "" {
		return errors.New("category cannot be empty")
	}
	if t.Amount <= 0 {
		return errors.New("amount must be greater than zero")
	}
	return nil
}

func (b Budget) Validate() error {
	if b.Category == "" {
		return errors.New("budget category cannot be empty")
	}
	if b.Limit <= 0 {
		return errors.New("budget limit must be greater than zero")
	}
	return nil
}

func PrintValidation(v Validatable) {
	err := CheckValid(v)
	switch e := err.(type) {
	case nil:
		fmt.Printf("%T: validation passed\n", v)
	default:
		fmt.Printf("%T: validation failed → %v\n", v, e)
	}
}
