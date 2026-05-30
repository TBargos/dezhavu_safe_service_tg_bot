package models

// Структура ответа из Бэкенда. Структура obj определяется в функции, но всегда указатель, т.к. может прийти null
type BackendResponse[T any] struct {
	Success bool   `json:"success"`
	Obj     *T     `json:"obj"`
	Msg     string `json:"msg"`
}

type ProfileInfoResponse struct {
	Status      string
	ExpireDate  string
	RemainTime  string
	ConnectLink string
}
