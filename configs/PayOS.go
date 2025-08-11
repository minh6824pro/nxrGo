package configs

import (
	"github.com/payOSHQ/payos-lib-golang"
	"log"
	"os"
)

func InitPayOS() {
	payos.Key(os.Getenv("PAYOS_CLIENT_ID"), os.Getenv("PAYOS_API_KEY"), os.Getenv("PAYOS_CHECKSUM_KEY"))
	webHookUrl := os.Getenv("BE_URL") + "/api/payos/webhook"
	data, err := payos.ConfirmWebhook(webHookUrl)
	log.Println(webHookUrl)
	if err != nil {
		log.Println(err.Error())
	}
	log.Println(data)
}
