package domain

import "testing"

func TestBudget_Validate(t *testing.T) {
	tests := []struct {
		name    string
		budget  Budget
		wantErr bool
	}{
		{
			name: "valid budget",
			budget: Budget{
				Category: "food",
				Limit:    1000,
			},
			wantErr: false,
		},
		{
			name: "empty category",
			budget: Budget{
				Category: "",
				Limit:    1000,
			},
			wantErr: true,
		},
		{
			name: "negative limit",
			budget: Budget{
				Category: "food",
				Limit:    -100,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.budget.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v, got %v", tt.wantErr, err)
			}
		})
	}
}
