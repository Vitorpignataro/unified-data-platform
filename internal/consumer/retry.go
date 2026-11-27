package consumer

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func ProcessWithRetry(
	ctx context.Context,
	handler Handler,
	msg Message,
	maxAttempts int,
	delay time.Duration,
) (Result, error) {
	if maxAttempts < 1 || delay < 0 {
		return Result{}, fmt.Errorf(
			"configuração de retry inválida",
		)
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return Result{}, err
		}

		result, err := handler.Handle(ctx, msg)
		if err == nil {
			return result, err
		}

		if !errors.Is(err, ErrTransient) ||
			attempt == maxAttempts {
			return Result{}, fmt.Errorf(
				"tentativa %d/%d: %w",
				attempt,
				maxAttempts,
				err,
			)
		}

		timer := time.NewTimer(delay)

		select {
		case <-ctx.Done():
			timer.Stop()
			return Result{}, ctx.Err()

		case <-timer.C:
		}
	}

	return Result{}, fmt.Errorf(
		"processamento não concluído",
	)
}

type flakyHandler struct {
	attempts int
}

func (h *flakyHandler) Handle(
	ctx context.Context,
	msg Message,
) (Result, error) {
	h.attempts++

	if h.attempts < 3 {
		return Result{}, ErrTransient
	}

	return Result{Operation: "c"}, nil
}

func TestProcessWithRetry(t *testing.T) {
	handler := &flakyHandler{}

	result, err := ProcessWithRetry(
		context.Background(),
		handler,
		Message{},
		3,
		0,
	)

	if err != nil {
		t.Fatal(err)
	}

	if result.Operation != "c" {
		t.Fatal("processamento não concluído")
	}

	if handler.attempts != 3 {
		t.Errorf(
			"esperadas 3 tentativas, recebidas %d",
			handler.attempts,
		)
	}
}
