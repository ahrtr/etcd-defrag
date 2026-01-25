package agent

import (
	"errors"
	"testing"
)

func TestIsNospaceError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "NOSPACE error uppercase",
			err:  errors.New("etcdserver: mvcc: database space exceeded NOSPACE"),
			want: true,
		},
		{
			name: "NOSPACE error lowercase",
			err:  errors.New("error: nospace alarm triggered"),
			want: false, // Our implementation checks for uppercase NOSPACE
		},
		{
			name: "NOSPACE in message",
			err:  errors.New("alarm NOSPACE is active"),
			want: true,
		},
		{
			name: "different error",
			err:  errors.New("connection refused"),
			want: false,
		},
		{
			name: "timeout error",
			err:  errors.New("context deadline exceeded"),
			want: false,
		},
		{
			name: "permission denied error",
			err:  errors.New("permission denied"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isNospaceError(tt.err)
			if got != tt.want {
				t.Errorf("isNospaceError() = %v, want %v", got, tt.want)
			}
		})
	}
}
