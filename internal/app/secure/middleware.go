package secure

import (
	"context"
	"errors"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type ContextKey string

const (
	TokenCookieName             = "token"
	ContextUserIDKey ContextKey = "userID"
)

func (s *Secure) SecureMiddleware(h http.Handler) http.Handler {
	sec := func(w http.ResponseWriter, r *http.Request) {

		userID, err := s.GetTokenUserID(r)
		if err != nil {
			logger.Log.Debug("create and set token with new userID")
			var e error
			userID, e = s.SetTokenWithUserID(r.Context(), w)
			if e != nil {
				logger.Log.Error("setting token with new userID", zap.Error(e))
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
		logger.Log.Debug("get userID from token cookie", zap.Int64("userID", userID))
		ctx := context.WithValue(r.Context(), ContextUserIDKey, userID)

		h.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(sec)
}

func (s *Secure) GuardMiddleware(h http.Handler) http.Handler {
	sec := func(w http.ResponseWriter, r *http.Request) {

		userID, err := s.GetTokenUserID(r)
		if err != nil {
			logger.Log.Debug("unauthorized request (hasn't contain valid token cookie)", zap.String("error", err.Error()))
			w.WriteHeader(http.StatusUnauthorized)
			//h.ServeHTTP(w, r)
		} else {
			logger.Log.Debug("get userID from token cookie", zap.Int64("userID", userID))
			ctx := context.WithValue(r.Context(), ContextUserIDKey, userID)
			h.ServeHTTP(w, r.WithContext(ctx))
		}
	}

	return http.HandlerFunc(sec)
}

func (s *Secure) GetTokenUserID(r *http.Request) (userID int64, err error) {

	token, err := r.Cookie(TokenCookieName)
	if err != nil {
		logger.Log.Debug("getting token cookie", zap.Error(err))
		return 0, err
	}

	// checking token cookie params
	err = checkTokenCookie(token)
	if err != nil {
		logger.Log.Debug("checking token cookie params", zap.Error(err))
		return 0, err
	}

	userID, userHash, err := s.unpackTokenCookieData(token.Value)
	if err != nil {
		logger.Log.Debug("unpack token cookie", zap.Error(err))
		return 0, err
	}

	found, err := s.CheckUser(r.Context(), userID, userHash)
	if !found {
		return 0, errors.New("userID not found in secure file")
	}
	if err != nil {
		return 0, err
	}

	return
}

func (s *Secure) SetTokenWithUserID(ctx context.Context, w http.ResponseWriter) (userID int64, err error) {
	userID, err = s.NewUser(ctx)
	if err != nil {
		logger.Log.Error("getting new user id", zap.Error(err))
		return 0, err
	}

	// creating nonce before encryption
	nonce, err := generateRandom(s.aesgcm.NonceSize())
	if err != nil {
		logger.Log.Error("creating nonce", zap.Error(err))
		return 0, err
	}

	token := s.packTokenCookieData(userID, nonce)
	logger.Log.Debug("creating hexadecimal token", zap.String("token", token))

	// setting token cookie
	cookie := &http.Cookie{
		Name:  TokenCookieName,
		Value: token,
		//Path:       "/",
		//Domain:     "",
		Expires: time.Now().Add(time.Hour * 1000),
		//RawExpires: "",
		//MaxAge:     0,
		//Secure:     true,
		//HttpOnly:   true,
		//SameSite:   0,
		//Raw:        "",
		//Unparsed:   nil,
	}
	http.SetCookie(w, cookie)

	return userID, nil
}

func checkTokenCookie(token *http.Cookie) error {
	//if !token.Secure {
	//	return errors.New("token cookie is not secure")
	//}

	//if !token.HttpOnly {
	//	return errors.New("token cookie is not HttpOnly")
	//}

	//if token.Expires.Before(time.Now()) {
	//	return errors.New("token cookie is expired")
	//}

	if token.Value == "" {
		return errors.New("token cookie has empty value")
	}

	return token.Valid()
}
