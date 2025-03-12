package main

import (
	"encoding/base64"
	"fmt"
)

func main() {
	data := "asldfasdfalsho012untuku39os-seller019di01fstaging1lnaslkdfnalsdfkjasdf"
	encoded := base64.StdEncoding.EncodeToString([]byte(data))
	encoded2 := base64.StdEncoding.EncodeToString([]byte(encoded))
	fmt.Println("Base64 Encoded:", encoded2)
}
