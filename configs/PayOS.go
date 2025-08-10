package configs

import (
	"github.com/payOSHQ/payos-lib-golang"
	"log"
	"os"
)

func InitPayOS() {
	payos.Key(os.Getenv("PAYOS_CLIENT_ID"), os.Getenv("PAYOS_API_KEY"), os.Getenv("PAYOS_CHECKSUM_KEY"))
	data, err := payos.ConfirmWebhook("https://80119fc0b10f.ngrok-free.app/api/payos/webhook")
	if err != nil {
		log.Println(err.Error())
	}
	log.Println(data)
}
