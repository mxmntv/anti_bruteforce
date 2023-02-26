package model

type Request struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	IP       string `json:"ip"`
}
