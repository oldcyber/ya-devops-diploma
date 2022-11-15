package server

import log "github.com/sirupsen/logrus"

func MyError(er string, err error) {
	log.Error(er, er+": ", err)
}
