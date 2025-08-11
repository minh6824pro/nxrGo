package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/minh6824pro/nxrGO/services"
	payos "github.com/payOSHQ/payos-lib-golang"
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

// HandleWebhook godoc
// @Summary      Handle incoming webhook from PayOS
// @Description  Receive and verify webhook payload with checksum key, then process payment success.
// @Tags         webhooks
// @Accept       json
// @Produce      plain
// @Param        webhookBody  body  dto.WebhookType  true  "Webhook payload"
// @Success      200  {string}  string  "Webhook processed successfully"
// @Router       /webhook [post]
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
