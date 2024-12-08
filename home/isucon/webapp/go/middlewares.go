package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/catatsuy/cache"
)

// func appAuthMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		ctx := r.Context()
// 		c, err := r.Cookie("app_session")
// 		if errors.Is(err, http.ErrNoCookie) || c.Value == "" {
// 			writeError(w, http.StatusUnauthorized, errors.New("app_session cookie is required"))
// 			return
// 		}
// 		accessToken := c.Value
// 		user := &User{}
// 		err =   db.GetContext(ctx, user, "SELECT * FROM users WHERE access_token = ?", accessToken)
// 		if err != nil {
// 			if errors.Is(err, sql.ErrNoRows) {
// 				writeError(w, http.StatusUnauthorized, errors.New("invalid access token"))
// 				return
// 			}
// 			writeError(w, http.StatusInternalServerError, err)
// 			return
// 		}

// 		ctx = context.WithValue(ctx, "user", user)
// 		next.ServeHTTP(w, r.WithContext(ctx))
// 	})
// }

// appAuthMiddlewareをcacheを使って高速化
// グローバルキャッシュの初期化
var appUserCache = cache.NewWriteHeavyCache[string, *User]() // キャッシュをグローバルで共有

// ミドルウェア関数の修正
func appAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		c, err := r.Cookie("app_session")

		if errors.Is(err, http.ErrNoCookie) || c.Value == "" {
			writeError(w, http.StatusUnauthorized, errors.New("app_session cookie is required"))
			return
		}

		// Cookieの値を取得
		accessToken := c.Value

		// キャッシュからユーザー情報を取得
		user, found := appUserCache.Get(accessToken)
		if !found {
			// キャッシュにデータがなければデータベースから取得
			user = &User{}
			err = db.GetContext(ctx, user, "SELECT * FROM users WHERE access_token = ?", accessToken)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					writeError(w, http.StatusUnauthorized, errors.New("invalid access token"))
					return
				}
				writeError(w, http.StatusInternalServerError, err)
				return
			}

			// キャッシュに保存（TTLは設定しない）
			appUserCache.Set(accessToken, user)
		}

		// コンテキストにユーザー情報を設定
		ctx = context.WithValue(ctx, "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}


func ownerAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		c, err := r.Cookie("owner_session")
		if errors.Is(err, http.ErrNoCookie) || c.Value == "" {
			writeError(w, http.StatusUnauthorized, errors.New("owner_session cookie is required"))
			return
		}
		accessToken := c.Value
		owner := &Owner{}
		if err := db.GetContext(ctx, owner, "SELECT * FROM owners WHERE access_token = ?", accessToken); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeError(w, http.StatusUnauthorized, errors.New("invalid access token"))
				return
			}
			writeError(w, http.StatusInternalServerError, err)
			return
		}

		ctx = context.WithValue(ctx, "owner", owner)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// func chairAuthMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		ctx := r.Context()
// 		c, err := r.Cookie("chair_session")
// 		if errors.Is(err, http.ErrNoCookie) || c.Value == "" {
// 			writeError(w, http.StatusUnauthorized, errors.New("chair_session cookie is required"))
// 			return
// 		}
// 		accessToken := c.Value
// 		chair := &Chair{}
// 		err = db.GetContext(ctx, chair, "SELECT * FROM chairs WHERE access_token = ?", accessToken)
// 		if err != nil {
// 			if errors.Is(err, sql.ErrNoRows) {
// 				writeError(w, http.StatusUnauthorized, errors.New("invalid access token"))
// 				return
// 			}
// 			writeError(w, http.StatusInternalServerError, err)
// 			return
// 		}

// 		ctx = context.WithValue(ctx, "chair", chair)
// 		next.ServeHTTP(w, r.WithContext(ctx))
// 	})
// }

// chairAuthMiddlewareをcacheを使って高速化
var chairCache = cache.NewWriteHeavyCache[string, *Chair]() // グローバルキャッシュ

func chairAuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()
        c, err := r.Cookie("chair_session")
        if errors.Is(err, http.ErrNoCookie) || c.Value == "" {
            writeError(w, http.StatusUnauthorized, errors.New("chair_session cookie is required"))
            return
        }

        accessToken := c.Value
        chair, found := chairCache.Get(accessToken)
        if !found {
            chair = &Chair{}
            err = db.GetContext(ctx, chair, "SELECT * FROM chairs WHERE access_token = ?", accessToken)
            if err != nil {
                if errors.Is(err, sql.ErrNoRows) {
                    writeError(w, http.StatusUnauthorized, errors.New("invalid access token"))
                    return
                }
                writeError(w, http.StatusInternalServerError, err)
                return
            }

            // キャッシュに保存（TTLは設定しない）
			chairCache.Set(accessToken, chair)
        }

        ctx = context.WithValue(ctx, "chair", chair)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
