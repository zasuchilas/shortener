package secure

import (
	"net/http"
	"testing"
	"time"
)

func Test_checkTokenCookie(t *testing.T) {
	type args struct {
		token *http.Cookie
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "nil",
			args: args{
				token: &http.Cookie{
					Name:       "",
					Value:      "",
					Path:       "",
					Domain:     "",
					Expires:    time.Time{},
					RawExpires: "",
					MaxAge:     0,
					Secure:     false,
					HttpOnly:   false,
					SameSite:   0,
					Raw:        "",
					Unparsed:   nil,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkTokenCookie(tt.args.token); (err != nil) != tt.wantErr {
				t.Errorf("checkTokenCookie() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
