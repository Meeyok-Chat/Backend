package models

import "errors"

var ErrChatNotInCache = errors.New("chat not in cache")

type HTTPError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}
