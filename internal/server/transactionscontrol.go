package server

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/oldcyber/ya-devops-diploma/internal/model"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

type addWithdraw struct {
	Order string          `json:"order"`
	Sum   decimal.Decimal `json:"sum"`
}

func getUID(r *http.Request) (int, error) {
	uid, err := strconv.Atoi(r.Header.Get("uid"))
	if err != nil {
		log.Error(err)
		return 0, err
	}
	return uid, nil
}

// getAllTransactionsByUserID GET /api/user/orders (200, 204, 401, 500)
func (a *App) getAllTransactionsByUserID(w http.ResponseWriter, r *http.Request) {
	var item model.OutTransactions
	uid, err := getUID(r)
	log.Info("get orders uid: ", uid)
	if err != nil {
		MyError("error get uid", err)
		w.WriteHeader(model.InternalError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	log.Info("GET all order transactions by user id: ", uid)
	res, err := item.GetAllTransactionsByUserID(a.DB, uid)
	log.Info("order transactions result: ", res)
	if err != nil {
		MyError("get order transactions error", err)
		w.WriteHeader(model.NoData)
		_, _ = w.Write([]byte("No data"))
		return
	}
	rRes, _ := json.Marshal(res)
	log.Info("get order transactions result: ", string(rRes))
	if string(rRes) != "null" {
		w.WriteHeader(model.Ok)
		_, _ = w.Write(rRes)
	} else {
		w.WriteHeader(model.NoData)
	}
}

// addTransaction POST /api/user/orders (200, 202, 400, 401, 409, 422, 500)
func (a *App) addTransaction(w http.ResponseWriter, r *http.Request) {
	var item model.Transactions
	//var order ToOrder
	var res int
	// user id
	uid, err := getUID(r)
	log.Info("add order uid: ", uid)
	if err != nil {
		MyError("error get uid", err)
		w.WriteHeader(model.InternalError)
		return
	}
	responseData, err := io.ReadAll(r.Body)
	if err != nil {
		MyError("error read body", err)
		w.WriteHeader(model.InternalError)
		return
	}
	orderNumber, _ := strconv.Atoi(string(responseData))
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(r.Body)
	log.Info("POST order number: ", orderNumber)
	//url := &a.Cfg.AccrualSystemAddress
	//bbStatus := order.postOrder(orderNumber, *url)
	//log.Info("POST order status: ", bbStatus)
	//switch bbStatus {
	//case model.OrderAccepted:
	res, err = item.AddTransaction(a.DB, orderNumber, uid)
	if err != nil {
		MyError("error add transaction", err)
		w.WriteHeader(model.InternalError)
		return
	}
	log.Info("Add transaction status: ", res)
	switch res {
	case model.OrderAccepted:
		log.Info("Prepare to queue")
		a.Queue <- orderNumber
		w.WriteHeader(model.OrderAccepted)
		log.Info("Print to Header: ", model.OrderAccepted)
		return
	default:
		w.WriteHeader(res)
		log.Info("Print to Header: ", res)
		return
		//}
		//case model.AlienOrderNum:
		//	w.WriteHeader(model.OrderAccepted)
		//	log.Info("from BlackBox Print to Header: ", model.OrderAccepted)
		//	return
		//case model.WrongQueryFormat:
		//	w.WriteHeader(model.WrongOrderNum)
		//	log.Info("from BlackBox Print to Header: ", model.WrongOrderNum)
	}
}

// getUserBalance GET /api/user/balance (200, 401, 500))
func (a *App) getUserBalance(w http.ResponseWriter, r *http.Request) {
	var item model.Balance
	w.Header().Set("Content-Type", "application/json")
	uid, err := getUID(r)
	if err != nil {
		MyError("error get uid", err)
		w.WriteHeader(model.InternalError)
		return
	}
	res, status, err := item.GetUserBalance(a.DB, uid)
	if err != nil {
		MyError("error get user balance", err)
		w.WriteHeader(model.InternalError)
		return
	}

	w.WriteHeader(status)
	rRes, _ := json.Marshal(res)
	_, _ = w.Write(rRes)
}

// addWithdraw POST /api/user/balance/withdraw (200, 401, 402, 422, 500)
func (a *App) addWithdraw(w http.ResponseWriter, r *http.Request) {
	var item model.Transactions
	var ad addWithdraw
	uid, err := getUID(r)
	if err != nil {
		w.WriteHeader(model.InternalError)
		MyError("error get uid", err)
		return
	}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&ad)
	if err != nil {
		w.WriteHeader(model.InternalError)
		MyError("error decode body", err)
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(r.Body)

	order, _ := strconv.Atoi(ad.Order)
	sum := ad.Sum
	res := item.AddWithdraw(a.DB, order, uid, sum)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(res)
}

// getAllWithdrawsByUserID GET /api/user/withdraw (200, 204, 401, 500)
func (a *App) getAllWithdrawsByUserID(w http.ResponseWriter, r *http.Request) {
	var item model.Withdrawal
	uid, err := getUID(r)
	if err != nil {
		w.WriteHeader(model.InternalError)
		MyError("error get uid", err)
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(r.Body)

	res, status := item.GetAllWithdrawsByUserID(a.DB, uid)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	switch status {
	case model.Ok:
		rRes, _ := json.Marshal(res)
		if string(rRes) != "null" {
			_, _ = w.Write(rRes)
		}
	case model.NoData:
		_, _ = w.Write([]byte("No data"))
	}
}
