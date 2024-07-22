package model

import (
	"encoding/json"
)

type Phone struct {
	Data struct {
		Brand string `json:"brand"`
		Name  string `json:"name"`
		Image string `json:"image"`
	} `json:"data"`
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func ConvertToPerson(jsonString string) (Phone, error) {
	var phone Phone
	err := json.Unmarshal([]byte(jsonString), &phone)
	if err != nil {
		return Phone{}, err
	}
	return phone, nil
}
