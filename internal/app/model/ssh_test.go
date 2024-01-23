package model

import (
	"testing"
)

func TestRecord_MaskedPassword(t *testing.T) {
	tests := []struct {
		password string
		want     string
	}{
		{
			password: "1",
			want:     "*",
		},
		{
			password: "12",
			want:     "**",
		},
		{
			password: "123",
			want:     "1*3",
		},
		{
			password: "1234",
			want:     "1**4",
		},
		{
			password: "12345",
			want:     "1***5",
		},
		{
			password: "123456",
			want:     "12**56",
		},
		{
			password: "123456789012",
			want:     "1234****9012",
		},
		{
			password: "12345678901234567890",
			want:     "1234************7890",
		},
	}
	for _, tt := range tests {
		t.Run(tt.password, func(t *testing.T) {
			r := &Record{
				Password: tt.password,
			}
			if got := r.MaskedPassword(); got != tt.want {
				t.Errorf("MaskedPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}
