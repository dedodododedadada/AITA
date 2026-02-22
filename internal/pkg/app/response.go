package app

import "aita/internal/errcode"

type Response struct {
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	Code    string `json:"code,omitempty"`
	Meta    any    `json:"meta,omitempty"`
}

func Fail(err error) Response {
	return Response{
		Error: err.Error(),
		Code:  errcode.GetBusinessCode(err),
	}
}

func Success(data any) Response {
	return Response{
		Data: data,
		Code: "SUCCESS",
	}
}

func SuccessMsg(msg string) Response {
	return Response{
		Message: msg,
		Code:    "SUCCESS",
	}
}

func SuccessWithMeta(data any, meta any) Response {
	return Response{
		Data: data,
		Meta: meta,
		Code: "SUCCESS",
	}
}
