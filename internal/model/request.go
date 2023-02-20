package model

type Request struct {
	Login    string `json:"login" validate:"nonzero"`
	Password string `json:"password" validate:"nonzero"`
	IP       string `json:"ip" validate:"nonzero"`
}
