package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/oldcyber/ya-devops-diploma/internal/model"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

type ToOrder struct {
	Order string `json:"order"`
}

type FromOrder struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual"`
}

//func (o *ToOrder) postOrder(order int, url string) int {
//	// 202 - accepted,  409 - conflict, 400 - bad request
//	// Оборачиваем order в json
//	var ord ToOrder
//	ord.Order = strconv.Itoa(order)
//	orderJSON, err := json.Marshal(ord)
//	if err != nil {
//		log.Error(err)
//		return model.InternalError
//	}
//	ctx := context.Background()
//	client := &http.Client{}
//	urlSuffix := "/api/orders"
//	req, err := http.NewRequestWithContext(ctx, "POST", url+urlSuffix, bytes.NewBuffer(orderJSON))
//	req.Header.Set("Content-Type", "application/json")
//	if err != nil {
//		log.Error("Ошибка запроса: ", err)
//	}
//	resp, err := client.Do(req)
//	if err != nil {
//		log.Error("Ошибка при отправке данных в blackbox: ", err)
//		return resp.StatusCode
//	}
//	defer resp.Body.Close()
//	return resp.StatusCode
//}

func (a *App) GetOrders(url string) {
	// 202 - accept, 204 - no content
	ctx := context.Background()
	var fO FromOrder
	var resp *http.Response
	client := &http.Client{}
	log.Info("queue: ", a.Queue)
	for s := range a.Queue {
		log.Info("GET Order: ", s)
		urlSuffix := "/api/orders/"
		id := strconv.Itoa(s)
		body := bytes.NewReader([]byte{})
		req, err := http.NewRequestWithContext(ctx, "GET", url+urlSuffix+id, body)
		if err != nil {
			log.Error("Ошибка запроса: ", err)
			return
		}
		resp, err = client.Do(req)
		if err != nil {
			log.Error("Ошибка при отправке данных в сервис метрик: ", err)
			return
		}
		log.Info("Blackbox GET status code: ", resp.StatusCode)
		switch resp.StatusCode {
		case model.Ok:
			decoder := json.NewDecoder(resp.Body)
			err = decoder.Decode(&fO)
			if err != nil {
				log.Error("Ошибка декодирования: ", err)
				return
			}
			order, _ := strconv.Atoi(fO.Order)
			var accrual decimal.Decimal
			accrual = decimal.NewFromFloat32(fO.Accrual)
			_, err = a.DB.Exec("UPDATE transactions SET order_status = $1, transaction_amount = $2 WHERE order_number = $3 ",
				fO.Status, accrual, order)
			if err != nil {
				log.Error(err)
				return
			}
			switch fO.Status {
			case "PROCESSED":
				log.Info(fO.Status)
				return
			default:
				log.Info("Add again to queue: ", fO.Status)
				a.Queue <- s
				return
			}
		case model.OrderAccepted:
			decoder := json.NewDecoder(resp.Body)
			err = decoder.Decode(&fO)
			if err != nil {
				log.Error("Ошибка декодирования: ", err)
				return
			}
			order, _ := strconv.Atoi(fO.Order)
			_, err = a.DB.Exec("UPDATE transactions SET order_status = $1, transaction_amount = $2 WHERE order_number = $3 ",
				fO.Status, fO.Accrual, order)
			if err != nil {
				log.Error(err)
				return
			}
			switch fO.Status {
			case "PROCESSED":
				log.Info(fO.Status)
				return
			default:
				log.Info("Add again to queue: ", fO.Status)
				a.Queue <- s
				return
			}
		case model.WrongQueryFormat:
			log.Error("Неверный формат запроса")
			return
		case model.NoData:
			log.Error("Заказ не найден:", s)
			return
		default:
			log.Error("Неизвестная ошибка")
			return
		}
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()
	log.Info("fO: ", fO)
}
