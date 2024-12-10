package storage

import (
	"github.com/zasuchilas/shortener/internal/app/models"
	"testing"
)

func Test_checkUserURLs(t *testing.T) {
	type args struct {
		userID  int64
		urlRows map[string]*models.URLRow
	}

	rows := make(map[string]*models.URLRow, 1)
	rows["1"] = &models.URLRow{
		ID:       0,
		ShortURL: "1",
		OrigURL:  "1",
		UserID:   1,
		Deleted:  false,
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "nil",
			args: args{
				userID:  0,
				urlRows: rows,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkUserURLs(tt.args.userID, tt.args.urlRows); (err != nil) != tt.wantErr {
				t.Errorf("checkUserURLs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
