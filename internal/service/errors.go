package service

import "errors"

var (
	ErrNotFound  = errors.New("资源不存在")
	ErrForbidden = errors.New("无权操作")
)
