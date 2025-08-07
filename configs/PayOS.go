package configs

import (
	"github.com/payOSHQ/payos-lib-golang"
	"os"
)

func InitPayOS() {
	payos.Key(os.Getenv("PAYOS_CLIENT_ID"), os.Getenv("PAYOS_API_KEY"), os.Getenv("PAYOS_CHECKSUM_KEY"))

}
