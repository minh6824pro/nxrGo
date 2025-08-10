package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/services"
	"github.com/payOSHQ/payos-lib-golang"
	"log"
	"net/http"
)

type WebhookController struct {
	orderService services.OrderService
}

func NewWebhookController(orderService services.OrderService) *WebhookController {
	return &WebhookController{
		orderService: orderService,
	}
}

func (pc *WebhookController) HandleWebhook(c *gin.Context) {
	var webhookBody payos.WebhookType
	if err := c.ShouldBindJSON(&webhookBody); err != nil {
		log.Println("Invalid webhook payload:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	log.Println("Received webhook payload:", webhookBody)

	data, err := payos.VerifyPaymentWebhookData(webhookBody)
	if err != nil {
		log.Println("Invalid webhook payload:", err)
		return
	}
	log.Println(data)
	log.Printf("[Webhook] DraftOrderId: %d | Reference: %s | Status: %s\n",
		data.OrderCode,
		data.Reference,
		data.Desc,
	)
	if data.Reference == "TF230204212323" {
		return
	}
	pc.orderService.PayOSPaymentSuccess(c, uint(data.OrderCode))
	c.String(http.StatusOK, "Webhook processed successfully")
}
