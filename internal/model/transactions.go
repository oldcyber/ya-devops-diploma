package model

import (
	"database/sql"
	"errors"
	"time"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

type Transactions struct {
	TransactionID     int             `json:"transaction_id"`
	OrderNumber       int             `json:"order_number"`
	OrderDate         string          `json:"order_date"`
	OrderStatus       string          `json:"order_status,omitempty"`
	TransactionType   string          `json:"transaction_type"`
	TransactionAmount decimal.Decimal `json:"transaction_amount,omitempty"`
	UserID            int             `json:"user_id"`
}

type OutTransactions struct {
	OrderNumber       string  `json:"number"`
	OrderStatus       string  `json:"status"`
	TransactionAmount float32 `json:"accrual"`
	OrderDate         string  `json:"uploaded_at"`
}

type Balance struct {
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

type Withdrawal struct {
	Order       string    `json:"order"`
	Sum         float32   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

type OrderNumbers struct {
	OrderNumber string `json:"order_number"`
}

// GetUserBalance returns user balance (GET /api/user/balance)
func (t *Balance) GetUserBalance(db *sql.DB, userID int) (Balance, int, error) {
	var balance Balance
	ub, _ := GetUserBalance(db, userID).Float64()
	balance.Current = float32(ub)
	uw, _ := GetUserWithdraw(db, userID).Float64()
	balance.Withdrawn = float32(uw)
	return balance, Ok, nil
}

func GetUserWithdraw(db *sql.DB, id int) decimal.Decimal {
	var withdraw decimal.Decimal
	_ = db.QueryRow(`SELECT COALESCE(SUM(transaction_amount), 0) AS outcome 
		FROM transactions 
		WHERE user_id = $1 
		AND order_status = 'PROCESSED' 
		AND transaction_type = 'outcome' 
		group by user_id`, id).Scan(&withdraw)
	return withdraw
}

func GetUserBalance(db *sql.DB, id int) decimal.Decimal {
	var balance decimal.Decimal
	_ = db.QueryRow(`SELECT (select COALESCE(SUM(transaction_amount), 0) AS income 
		from transactions 
		WHERE user_id = $1 
		AND order_status = 'PROCESSED' 
		AND transaction_type = 'income') - 
		(select COALESCE(SUM(transaction_amount), 0) AS outcome 
		FROM transactions 
		WHERE user_id = $1 
		AND order_status = 'PROCESSED' 
		AND transaction_type = 'outcome') 
		FROM transactions 
		group by user_id`, id).Scan(&balance)
	return balance
}

// GetAllTransactionsByUserID returns all transactions by user id (GET /api/user/orders)
func (t *OutTransactions) GetAllTransactionsByUserID(db *sql.DB, userID int) ([]OutTransactions, error) {
	var transactions []OutTransactions
	var err error
	if userID == 0 {
		err = errors.New("user id is empty")
		return transactions, err
	}
	res, err := db.Query(`SELECT order_number, order_status, transaction_amount, order_date FROM
		 transactions WHERE user_id = $1 AND transaction_type = 'income' order by order_date desc`, userID)
	if err != nil {
		return nil, err
	}
	defer func(res *sql.Rows) {
		err = res.Close()
		if err != nil {
			log.Error(err)
		}
	}(res)

	var tr OutTransactions
	for res.Next() {
		if err = res.Scan(
			&tr.OrderNumber,
			&tr.OrderStatus,
			&tr.TransactionAmount,
			&tr.OrderDate,
		); err != nil {
			log.Error(err)
			return nil, err
		}
		transactions = append(transactions, tr)
	}
	return transactions, nil
}

// AddTransaction adds new transaction to the database (POST /api/user/orders)
func (t *Transactions) AddTransaction(db *sql.DB, orderNumber, userID int) (int, error) {
	var err error
	res := Valid(orderNumber)
	if !res {
		log.Error("Invalid order number")
		return WrongOrderNum, err
	}
	fRes := findOrderNumber(db, orderNumber, userID)
	switch fRes {
	case OrderAccepted:
		_, err = db.Exec(`INSERT INTO transactions (user_id, order_number, order_date, 
            transaction_type, order_status, transaction_amount)
			VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (transaction_id) DO UPDATE SET 
			(user_id, order_number, order_date, transaction_type, order_status, transaction_amount)
			    = ($1, $2, $3, $4, $5, $6)`, userID, orderNumber, time.Now(), "income", "NEW", 0)
		if err != nil {
			log.Error(err)
			return InternalError, err
		}
		return OrderAccepted, nil
	case Ok:
		log.Info("Order :", Ok)
		return Ok, nil
	case AlienOrderNum:
		log.Info("Order :", AlienOrderNum)
		return AlienOrderNum, nil
	}
	return 0, err
}

// AddWithdraw adds new withdrawal to the database (POST /api/user/balance/withdraw)
func (t *Transactions) AddWithdraw(db *sql.DB, orderNumber, userID int, amount decimal.Decimal) int {
	var err error
	withdrawal := searchForWithdrawal(db, orderNumber, userID)
	if withdrawal {
		res := Valid(orderNumber)
		if !res {
			log.Error("Invalid order number")
			return WrongOrderNum
		}
		balance := GetUserBalance(db, userID)
		if balance.LessThan(amount) {
			return InsufficientFunds
		}
		_, err = db.Exec(`INSERT INTO transactions (user_id, order_number, order_date, transaction_type, 
             transaction_amount, order_status) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (transaction_id) DO UPDATE SET 
            (user_id, order_number, order_date, transaction_type, transaction_amount, order_status) 
                = ($1, $2, $3, $4, $5, $6)`, userID, orderNumber, time.Now(), "outcome", amount, "PROCESSED")
		if err != nil {
			log.Error(err)
			return InternalError
		}
		return Ok
	}
	return 0
}

// GetAllWithdrawsByUserID returns all withdrawals by user id (GET /api/user/balance/withdraw)
func (t *Withdrawal) GetAllWithdrawsByUserID(db *sql.DB, userID int) (with []Withdrawal, status int) {
	var withdrawal []Withdrawal
	res, err := db.Query(`SELECT order_number, transaction_amount, order_date FROM transactions WHERE
		 user_id = $1 AND transaction_type = 'outcome' order by order_date desc`, userID)
	if err != nil {
		log.Error(err)
		return nil, NoData
	}
	defer func(res *sql.Rows) {
		err = res.Close()
		if err != nil {
			log.Error(err)
		}
	}(res)

	var tr Withdrawal
	for res.Next() {
		if err = res.Scan(
			&tr.Order,
			&tr.Sum,
			&tr.ProcessedAt,
		); err != nil {
			log.Error(err)
			return nil, InternalError
		}
		withdrawal = append(withdrawal, tr)
	}
	return withdrawal, Ok
}
