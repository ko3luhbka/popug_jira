package rest

import (
	"log"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-oauth2/oauth2/v4/server"

	"github.com/ko3luhbka/auth/db"
	"github.com/ko3luhbka/auth/mq"
)

const (
	listenAddr = "0.0.0.0:8080"
)

var sessionMgr = scs.New()

type Server struct {
	mq    *mq.Client
	repo  *db.Repo
	oauth *server.Server
	mux   *http.ServeMux
}

func NewServer(repo *db.Repo, mq *mq.Client) (*Server, error) {
	sessionMgr.Lifetime = 24 * time.Hour
	oauthSrv := InitOauthServer(repo)

	srv := &Server{
		repo:  repo,
		mq:    mq,
		oauth: oauthSrv,
		mux:   http.NewServeMux(),
	}

	srv.initRoutes()
	return srv, nil
}

func (s Server) Run() error {
	log.Printf("server is running on %s\n", listenAddr)
	return http.ListenAndServe(listenAddr, sessionMgr.LoadAndSave(s.mux))
}

func (s Server) initRoutes() {
	s.mux.HandleFunc("/ping", s.pingHandler)
	s.mux.HandleFunc("/create-user/", s.createUser)
	s.mux.HandleFunc("/get-user/", s.getUser)
	s.mux.HandleFunc("/get-users/", s.getAllUsera)
	s.mux.HandleFunc("/update-user/", s.updateUser)
	s.mux.HandleFunc("/delete-user/", s.deleteUser)

	s.mux.HandleFunc("/login", s.loginUser)

	s.mux.HandleFunc("/oauth/authorization-grant", s.authorizationGrant)
	s.mux.HandleFunc("/oauth/authorize", s.authorize)
	s.mux.HandleFunc("/oauth/get-token", s.getToken)
	s.mux.HandleFunc("/oauth/validate-token", s.validateToken)
}
