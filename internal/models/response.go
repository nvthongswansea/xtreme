package models

type SuccessFManResponse struct {
	EntityUUID string `json:"entity_uuid"`
	Message    string `json:"message"`
}

type SuccessRegisterResponse struct {
	Message string `json:"message"`
}

type SuccessAuthenResponse struct {
	Token   string `json:"token"`
	Message string `json:"message"`
}
