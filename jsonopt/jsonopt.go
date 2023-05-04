package jsonopt

import (
	"encoding/json"
)

func Encode(obj any) string {
	marshal, err := json.Marshal(obj)
	if err != nil {
		return ""
	}
	return string(marshal)
}

func EncodeFormat(obj any) string {
	marshal, err := json.MarshalIndent(obj, "", " ")
	if err != nil {
		return ""
	}
	return string(marshal)
}

func Decode[T any, P string | []byte](str P) T {
	var obj T
	_ = json.Unmarshal([]byte(str), &obj)
	return obj
}
