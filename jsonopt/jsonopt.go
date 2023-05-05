package jsonopt

import (
	"encoding/json"
)

func Encode(obj any) (string, error) {
	marshal, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(marshal), nil
}

func EncodeFormat(obj any) string {
	marshal, err := json.MarshalIndent(obj, "", " ")
	if err != nil {
		return ""
	}
	return string(marshal)
}

func Decode[T any, P string | []byte](str P) (T, error) {
	var obj T
	err := json.Unmarshal([]byte(str), &obj)
	return obj, err
}
