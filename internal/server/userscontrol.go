package server

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/oldcyber/ya-devops-diploma/internal/model"
	log "github.com/sirupsen/logrus"
)

// register register new users POST /api/user/register (200, 400, 409, 500)
func (a *App) register(w http.ResponseWriter, r *http.Request) {
	var item model.User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&item); err != nil {
		log.Error(err)
		w.WriteHeader(model.InternalError)
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(r.Body)
	res := item.AddUser(a.DB, item.Login, item.Password)
	w.Header().Set("Content-Type", "application/json")
	switch res {
	case model.Ok:
		token, status := item.LoginUser(a.DB, item.Login, item.Password)
		w.Header().Set("Authorization", "Bearer "+token)
		w.WriteHeader(status)
	default:
		w.WriteHeader(res)
	}
}

// login login registered users POST /api/user/login (200, 400, 401, 500)
func (a *App) login(w http.ResponseWriter, r *http.Request) {
	var item model.User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&item); err != nil {
		w.WriteHeader(model.InternalError)
		log.Error(err)
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(r.Body)
	if item.Password == "" {
		w.WriteHeader(model.NotLoggedIn)
		log.Error("required Password")
		return
	}
	if item.Login == "" {
		w.WriteHeader(model.NotLoggedIn)
		log.Error("required Email")
		return
	}
	token, res := item.LoginUser(a.DB, item.Login, item.Password)
	w.Header().Set("Content-Type", "application/json")
	switch res {
	case model.Ok:
		w.Header().Set("Authorization", "Bearer "+token)
		w.WriteHeader(res)
	default:
		w.WriteHeader(res)
	}
}
