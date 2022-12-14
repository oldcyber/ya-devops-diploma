package server

import "github.com/oldcyber/ya-devops-diploma/internal/middleware"

func (a *App) appRouter() {
	s := a.Router.PathPrefix("/api/user").Subrouter()
	// register user
	s.HandleFunc("/register", a.register).Methods("POST")
	// login user
	s.HandleFunc("/login", a.login).Methods("POST")
	// get all user orders
	s.HandleFunc("/orders", middleware.SetMiddlewareAuthentication(a.getAllTransactionsByUserID)).Methods("GET")
	// post new order
	s.HandleFunc("/orders", middleware.SetMiddlewareAuthentication(a.addTransaction)).Methods("POST")
	// get current balance
	s.HandleFunc("/balance", middleware.SetMiddlewareAuthentication(a.getUserBalance)).Methods("GET")
	//	 post withdrawal request
	s.HandleFunc("/balance/withdraw", middleware.SetMiddlewareAuthentication(a.addWithdraw)).Methods("POST")
	//	 get all withdrawal requests
	s.HandleFunc("/withdrawals", middleware.SetMiddlewareAuthentication(a.getAllWithdrawsByUserID)).Methods("GET")
}
