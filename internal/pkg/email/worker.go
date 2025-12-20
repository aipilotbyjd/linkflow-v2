package email

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

type Worker struct {
	service *Service
}

func NewWorker(service *Service) *Worker {
	return &Worker{service: service}
}

func (w *Worker) HandleSendEmail(ctx context.Context, task *asynq.Task) error {
	var email Email
	if err := json.Unmarshal(task.Payload(), &email); err != nil {
		return fmt.Errorf("failed to unmarshal email: %w", err)
	}

	return w.service.sendDirect(&email)
}

func (w *Worker) RegisterHandlers(mux *asynq.ServeMux) {
	mux.HandleFunc("email:send", w.HandleSendEmail)
}
