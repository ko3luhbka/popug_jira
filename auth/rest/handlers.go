package rest

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/ko3luhbka/auth/mq"
	"github.com/ko3luhbka/auth/rest/model"
)

const sessionUserID = "LoggedInUserID"

func (s Server) pingHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pong")
}

func (s Server) createUser(w http.ResponseWriter, r *http.Request) {
	var u model.User
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.Unmarshal(body, &u); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := u.Validate(); err != nil {
		log.Printf("invalid user: %v\n", err)
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	created, err := s.repo.Create(r.Context(), *u.ToEntity())
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	e := &mq.UserEvent{
		Name: mq.UserCreatedEvent,
		Data: model.EntityToAssignee(created),
	}
	if err := s.mq.Produce(r.Context(), e); err != nil {
		log.Println(err)
	}

	m := model.User{}
	m.FromEntity(created)
	writeJSONResponse(m, w)
}

func (s Server) getUser(w http.ResponseWriter, r *http.Request) {
	uuid := parseUUID(r)

	u, err := s.repo.GetByID(r.Context(), uuid)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	m := model.User{}
	m.FromEntity(u)
	writeJSONResponse(m, w)
}

func (s Server) getAllUsera(w http.ResponseWriter, r *http.Request) {
	entities, err := s.repo.GetAll(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	models := make([]model.User, len(entities))
	for i, u := range entities {
		m := model.User{}
		m.FromEntity(&u)
		models[i] = m
	}
	writeJSONResponse(models, w)
}

func (s Server) updateUser(w http.ResponseWriter, r *http.Request) {
	uuid := parseUUID(r)

	var u model.User
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.Unmarshal(body, &u); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	u.ID = uuid

	updated, err := s.repo.Update(r.Context(), *u.ToEntity())
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	e := &mq.UserEvent{
		Name: mq.UserCreatedEvent,
		Data: model.EntityToAssignee(updated),
	}
	if err := s.mq.Produce(r.Context(), e); err != nil {
		log.Println(err)
	}

	m := model.User{}
	m.FromEntity(updated)
	writeJSONResponse(m, w)
}

func (s Server) deleteUser(w http.ResponseWriter, r *http.Request) {
	uuid := parseUUID(r)

	if err := s.repo.Delete(r.Context(), uuid); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	e := &mq.UserEvent{
		Name: mq.UserDeletedEvent,
		Data: &model.Assignee{
			ID: uuid,
		},
	}
	if err := s.mq.Produce(r.Context(), e); err != nil {
		log.Println(err)
	}

	w.WriteHeader(http.StatusNoContent)
	return
}

func (s Server) loginUser(w http.ResponseWriter, r *http.Request) {
	var queryPart string
	rawQuery := r.URL.RawQuery
	if rawQuery != "" {
		queryPart = fmt.Sprintf("?%s", rawQuery)
	}

	templateData := map[string]any{
		"title":     "Popug Jira login page",
		"login_url": fmt.Sprintf("/login%s", queryPart),
	}
	tmpl, err := template.ParseFiles("views/login.html")

	if r.Method == http.MethodGet {
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl.Execute(w, templateData)
		return

	} else if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		v := r.PostForm
		var ul model.UserLogin
		ul.Username = v.Get("username")
		ul.Password = v.Get("password")
		if err := ul.Validate(); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		user, err := s.repo.GetByName(r.Context(), ul.Username)
		if err != nil {
			templateData["error"] = "invalid username or password"
			tmpl.Execute(w, templateData)
			return
		}

		if user.Password != ul.Password {
			templateData["error"] = "invalid username or password"
			tmpl.Execute(w, templateData)
			return
		}

		sessionMgr.Put(r.Context(), sessionUserID, user.ID)

		oauthURL := fmt.Sprintf("http://localhost:8080/oauth/authorization-grant%s", queryPart)
		w.Header().Set("Location", oauthURL)
		w.WriteHeader(http.StatusFound)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
	return
}

func (s Server) authorizationGrant(w http.ResponseWriter, r *http.Request) {
	queryStr := r.URL.RawQuery
	if userID := sessionMgr.GetString(r.Context(), sessionUserID); userID == "" {
		w.Header().Set("Location", fmt.Sprintf("/login?%s", queryStr))
		w.WriteHeader(http.StatusFound)
		return
	}

	templateData := map[string]any{
		"title":     "Authorization grant",
		"auth_url":  fmt.Sprintf("/oauth/authorize?%s", queryStr),
		"login_url": fmt.Sprintf("/login"),
		"app_id":    r.URL.Query().Get("client_id"),
		"scope":     r.URL.Query().Get("scope"),
	}
	tmpl, err := template.ParseFiles("views/auth_grant.html")
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, templateData)
	return
}

func (s Server) authorize(w http.ResponseWriter, r *http.Request) {
	if err := s.oauth.HandleAuthorizeRequest(w, r); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (s Server) getToken(w http.ResponseWriter, r *http.Request) {
	if err := s.oauth.HandleTokenRequest(w, r); err != nil {
		log.Println(err)
	}
}

func (s Server) validateToken(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	token := string(b)
	if token == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("empty token"))
		return
	}

	token = strings.TrimSpace(token)
	claims := &CustomJwtClaims{}
	t, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			log.Println("invalid jwt token signature")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		log.Printf("failed to parse jwt token: %v\n", err)

		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !t.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	resp, err := json.Marshal(claims)
	if err != nil {
		log.Printf("failed to marshal jwt claims: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func parseUUID(r *http.Request) string {
	splittedPath := strings.Split(r.URL.Path, "/")
	return splittedPath[len(splittedPath)-1]
}

func writeJSONResponse(model any, w http.ResponseWriter) {
	resp, err := json.Marshal(model)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
	return
}
