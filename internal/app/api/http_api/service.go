package http_api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"go.uber.org/zap"

	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/internal/app/secure"
	"github.com/zasuchilas/shortener/internal/app/service"
	"github.com/zasuchilas/shortener/pkg/shortener_http_api_v1"
)

var _ shortener_http_api_v1.ShortenerHTTPApiV1 = (*Implementation)(nil)

type Implementation struct {
	shortenerService service.ShortenerService
}

func NewImplementation(shortenerService service.ShortenerService) *Implementation {
	return &Implementation{shortenerService: shortenerService}
}

// GetUserID gets the userID from the context.
//
// All errors in this method are considered internal (500 InternalServerError)
// because error 401 Unauthorized is returned earlier from middleware.
func GetUserID(r *http.Request) (userID int64, err error) {

	// getting userID from context of request (after SecureMiddleware)
	uid := r.Context().Value(secure.ContextUserIDKey)

	// cast userID from any to int64
	userID, err = strconv.ParseInt(fmt.Sprintf("%d", uid), 10, 64)
	if err != nil {
		logger.Log.Debug("there are problems with userID", zap.Error(err))
		return 0, err
	}
	if userID == 0 {
		logger.Log.Debug("something went wrong: empty userID")
		return 0, errors.New("something went wrong: empty userID")
	}

	return userID, nil
}
