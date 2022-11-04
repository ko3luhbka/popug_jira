package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"
	"github.com/golang-jwt/jwt/v4"

	"github.com/ko3luhbka/auth/db"
)

const (
	authSrvURL = ":8080"
)

var jwtKey = []byte("secret!")

func InitOauthServer(repo *db.Repo) *server.Server {
	manager := manage.NewDefaultManager()

	jwtAccessGenerate := newJWTAccessGenerate(repo)
	manager.MapAccessGenerate(jwtAccessGenerate)

	manager.MustTokenStorage(store.NewMemoryTokenStore())

	// client memory store
	clientStore := store.NewClientStore()
	clientStore.Set("000000", &models.Client{
		ID:     "000000",
		Secret: "999999",
		Domain: "http://localhost:8081",
	})
	manager.MapClientStorage(clientStore)

	srv := server.NewDefaultServer(manager)
	srv.SetAllowGetAccessRequest(true)
	srv.SetClientInfoHandler(server.ClientFormHandler)
	srv.SetResponseTokenHandler(tokenResponseHandler)

	srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		log.Println("Internal Error:", err.Error())
		return
	})

	srv.SetResponseErrorHandler(func(re *errors.Response) {
		log.Println("Response Error:", re.Error.Error())
	})

	srv.SetUserAuthorizationHandler(userAuthorizeHandler)

	log.Println("oauth server is initialized")
	return srv
}

func userAuthorizeHandler(w http.ResponseWriter, r *http.Request) (userID string, err error) {
	uid := sessionMgr.GetString(r.Context(), sessionUserID)
	if uid == "" {
		w.Header().Set("Location", "http://localhost:8080/login")
		w.WriteHeader(http.StatusFound)
		return
	}

	userID = uid
	return
}

type jwtAccessGenerate struct {
	repo *db.Repo
}

func newJWTAccessGenerate(repo *db.Repo) jwtAccessGenerate {
	return jwtAccessGenerate{
		repo: repo,
	}
}

type CustomJwtClaims struct {
	jwt.StandardClaims
	UserUUID string `json:"user_uuid"`
	UserRole string `json:"user_role"`
}

func (j jwtAccessGenerate) Token(ctx context.Context, data *oauth2.GenerateBasic, isGenRefresh bool) (string, string, error) {
	user, err := j.repo.GetByID(ctx, data.UserID)
	if err != nil {
		return "", "", err
	}

	claims := &CustomJwtClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(30 * time.Minute).Unix(),
		},
		UserUUID: data.UserID,
		UserRole: user.Role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		log.Println(err)
		return "", "", err
	}

	return tokenString, "", nil
}

func tokenResponseHandler(w http.ResponseWriter, data map[string]interface{}, header http.Header, statusCode ...int) error {
	tokenString, ok := data["access_token"]
	if !ok {
		return fmt.Errorf("empty access_token field")
	}
	tokenExpiration, ok := data["expires_in"]
	if !ok {
		return fmt.Errorf("empty expires_in field")
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString.(string),
		Expires: time.Unix(tokenExpiration.(int64), 0),
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(data)
}
