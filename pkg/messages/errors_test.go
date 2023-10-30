package messages

import (
	"testing"
)

func TestPurchaseError_Error(t *testing.T) {
	tests := []struct {
		name string
		p    *PurchaseError
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.Error(); got != tt.want {
				t.Errorf("PurchaseError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExchangeError_Error(t *testing.T) {
	tests := []struct {
		name string
		f    *ExchangeError
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.Error(); got != tt.want {
				t.Errorf("ExchangeError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
