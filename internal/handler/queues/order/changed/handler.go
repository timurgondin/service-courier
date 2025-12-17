package changed

import (
	"encoding/json"
	"log"
	"service-courier/internal/dto/queues/order/changed"
	"service-courier/internal/model/order"

	"github.com/IBM/sarama"
)

type Handler struct {
	usecase usecase
}

func NewHandler(u usecase) *Handler {
	return &Handler{
		usecase: u,
	}
}

func (h *Handler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for dtoMsg := range claim.Messages() {
		ctx := sess.Context()
		log.Printf("order.changed handler: received message: key=%s, value=%s, partition=%d, offset=%d\n", string(dtoMsg.Key), string(dtoMsg.Value), dtoMsg.Partition, dtoMsg.Offset)

		var msg changed.Message
		err := json.Unmarshal(dtoMsg.Value, &msg)
		if err != nil {
			log.Printf("order.changed handler: received bad message: %v", err)
			sess.MarkMessage(dtoMsg, "")
			continue
		}

		// мы можем отфильтровать сообщения ненужные сразу в хендлере,
		// чтобы не передавать их на слой ниже

		err = h.usecase.Process(ctx, order.Order{
			ID:     msg.OrderID,
			Status: msg.Status,
		})
		if err != nil {
			log.Printf("order.changed handler: failed procces order: %v", err)
		}
		sess.MarkMessage(dtoMsg, "")
	}

	return nil
}
