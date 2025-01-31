package repository

import (
	"testing"

	"github.com/zasuchilas/shortener/internal/app/model"
)

func Test_checkUserURLs(t *testing.T) {
	type args struct {
		userID  int64
		urlRows map[string]*model.URLRow
	}

	rows := make(map[string]*model.URLRow, 1)
	rows["1"] = &model.URLRow{
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
