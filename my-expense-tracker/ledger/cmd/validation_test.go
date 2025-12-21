package main

import "testing"

func TestValidateTransaction(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		category string
		wantErr  bool
	}{
		{
			name:     "Valid transaction",
			amount:   100.0,
			category: "Food",
			wantErr:  false,
		},
		{
			name:     "Zero amount",
			amount:   0,
			category: "Food",
			wantErr:  true,
		},
		{
			name:     "Negative amount",
			amount:   -10.0,
			category: "Taxi",
			wantErr:  true,
		},
		{
			name:     "Empty category",
			amount:   100.0,
			category: "",
			wantErr:  true,
		},
		{
			name:     "Long category",
			amount:   100.0,
			category: "This is a very very very very very very long category name",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTransaction(tt.amount, tt.category)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
