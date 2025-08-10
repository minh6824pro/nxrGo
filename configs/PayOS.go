package configs

import (
	"fmt"
	"github.com/payOSHQ/payos-lib-golang"
	"os"
)

func InitPayOS() {
	payos.Key(os.Getenv("PAYOS_CLIENT_ID"), os.Getenv("PAYOS_API_KEY"), os.Getenv("PAYOS_CHECKSUM_KEY"))
	data, err := payos.ConfirmWebhook("https://0090610f0832.ngrok-free.app/api/payos/webhook")
	if err != nil {
		fmt.Printf(err.Error())
	}
	fmt.Println(data, "?")
}
