package model

import (
	"database/sql"

	log "github.com/sirupsen/logrus"
)

const (
	WrongOrderNum     = 422
	Ok                = 200
	OrderAccepted     = 202
	NoData            = 204
	WrongQueryFormat  = 400
	NotLoggedIn       = 401
	AlienOrderNum     = 409
	InsufficientFunds = 402
	InternalError     = 500
	ten               = 10
	nine              = 9
	two               = 2
)

func Valid(number int) bool {
	return (number%10+checksum(number/ten))%10 == 0
}

func checksum(number int) int {
	var luhn int

	for i := 0; number > 0; i++ {
		cur := number % ten

		if i%2 == 0 { // even
			cur *= two
			if cur > nine {
				cur = cur%ten + cur/ten
			}
		}

		luhn += cur
		number /= ten
	}
	return luhn % ten
}

// findOrderNumber returns order number by order_number
func findOrderNumber(db *sql.DB, orderNumber, userID int) int {
	var (
		on  int
		uid int
	)
	err := db.QueryRow("SELECT order_number, user_id FROM transactions WHERE order_number = $1",
		orderNumber).Scan(&on, &uid)
	// номер не найден - значит будет новый
	log.Info("Order number: ", on, " User ID: ", uid)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return OrderAccepted
		default:
			return InternalError
		}
	}
	if on != orderNumber {
		return OrderAccepted
	}
	if uid != userID || uid != 0 {
		return AlienOrderNum
	} else {
		return Ok
	}
}

func findUser(db *sql.DB, login string) int {
	var user User
	err := db.QueryRow("SELECT login FROM users WHERE login = $1", login).Scan(&user.Login)
	switch err {
	case sql.ErrNoRows:
		return Ok
	case nil:
		return InternalError
	default:
		return AlienOrderNum
	}
}

// searchForWithdrawal searches for withdrawal in the database
func searchForWithdrawal(db *sql.DB, orderNumber, userID int) bool {
	var orderStatus string
	err := db.QueryRow("SELECT transaction_id FROM transactions WHERE order_number = $1 AND user_id = $2 AND"+
		" transaction_type = 'outcome'", orderNumber, userID).Scan(&orderStatus)
	return err != nil
}
