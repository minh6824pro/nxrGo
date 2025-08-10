package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/payOSHQ/payos-lib-golang"
	"log"
	"net/http"
)

type WebhookController struct {
}

func NewWebhookController() *WebhookController {
	return &WebhookController{}
}

func (pc *WebhookController) HandleWebhook(c *gin.Context) {
	var webhookBody payos.WebhookType
	if err := c.ShouldBindJSON(&webhookBody); err != nil {
		log.Println("Invalid webhook payload:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	log.Println("Received webhook payload:", webhookBody)

	// TODO: Xác thực chữ ký webhook nếu PayOS cung cấp SDK/hàm verify

	log.Printf("[Webhook] OrderCode: %d | Amount: %.2f | Status: %s\n",
		webhookBody.Code,
		webhookBody.Success,
		webhookBody.Desc,
	)

	c.String(http.StatusOK, "Webhook processed successfully")
}
