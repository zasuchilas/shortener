package shortener

import "context"

// ReadURL _
func (s *service) ReadURL(ctx context.Context, shortURL string) (origURL string, err error) {
	origURL, err = s.shortenerRepo.ReadURL(ctx, shortURL)
	if err != nil {
		return "", err
	}
	return origURL, nil
}
