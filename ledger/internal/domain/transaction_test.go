package domain

import (
	"testing"
	"time"
)

func TestTransaction_Validate(t *testing.T) {
	tests := []struct {
		name    string
		tx      Transaction
		wantErr bool
	}{
		{
			name: "valid transaction",
			tx: Transaction{
				Amount:   100,
				Category: "food",
				Date:     time.Now(),
			},
			wantErr: false,
		},
		{
			name: "zero amount",
			tx: Transaction{
				Amount:   0,
				Category: "food",
			},
			wantErr: true,
		},
		{
			name: "empty category",
			tx: Transaction{
				Amount: 100,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tx.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v, got %v", tt.wantErr, err)
			}
		})
	}
}
