package validate

import (
	"testing"
)

func TestToken(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		want       string
		wantErr    bool
		errMessage string
	}{
		{
			name:    "Valid token",
			token:   "0123456789abcdef0123456789abcdef01234567",
			want:    "0123456789abcdef0123456789abcdef01234567",
			wantErr: false,
		},
		{
			name:       "Invalid token with special characters",
			token:      "0123456789abcdef0123456789abcdef0123456g",
			want:       "",
			wantErr:    true,
			errMessage: "invalid token format",
		},
		{
			name:       "Invalid token with length less than 40",
			token:      "0123456789abcdef0123456789abcdef0123456",
			want:       "",
			wantErr:    true,
			errMessage: "invalid token format",
		},
		{
			name:       "Invalid token with length more than 40",
			token:      "0123456789abcdef0123456789abcdef012345678",
			want:       "",
			wantErr:    true,
			errMessage: "invalid token format",
		},
		{
			name:       "Invalid token with uppercase letters",
			token:      "0123456789ABCDEF0123456789ABCDEF01234567",
			want:       "",
			wantErr:    true,
			errMessage: "invalid token format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Token(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("Token() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && err.Error() != tt.errMessage {
				t.Errorf("Token() error = %v, wantErrMessage %v", err.Error(), tt.errMessage)
			}
			if got != tt.want {
				t.Errorf("Token() = %v, want %v", got, tt.want)
			}
		})
	}
}
