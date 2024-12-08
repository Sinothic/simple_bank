package util

import "testing"

func TestIsSupportedCurrency(t *testing.T) {
	type args struct {
		currency string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "supported currency",
			args: args{currency: "USD"},
			want: true,
		},
		{
			name: "unsupported currency",
			args: args{currency: "JPY"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSupportedCurrency(tt.args.currency); got != tt.want {
				t.Errorf("IsSupportedCurrency() = %v, want %v", got, tt.want)
			}
		})
	}
}
