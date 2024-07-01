package main

import "github.com/trvita/caldav-client-yandex/ui"

func main() {
	// v, _ := uuid.NewV7()
	// //uid := base64.NewEncoding("")
	// bin, _ := v.MarshalBinary()
	// val := base32.StdEncoding.EncodeToString([]byte(bin))
	// fmt.Println(val)
	ui.StartMenu("https://caldav.yandex.ru")
}
