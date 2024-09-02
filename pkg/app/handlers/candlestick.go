package handlers

import (
	"context"
	"fmt"

	"github.com/ramasbeinaty/trading-chart-service/pkg/domain/candlestick"
	"github.com/ramasbeinaty/trading-chart-service/pkg/domain/subscription"
	"github.com/ramasbeinaty/trading-chart-service/pkg/domain/uids"
	candlestickpb "github.com/ramasbeinaty/trading-chart-service/proto/candlestick/contracts"
)

type CandlestickHandler struct {
	candlestickpb.UnimplementedCandlestickServiceServer
	candlestickService  *candlestick.CandlestickService
	subscriptionService *subscription.SubscriptionService
	uidService          *uids.UIDService
}

var _ candlestickpb.CandlestickServiceServer = &CandlestickHandler{}

func NewCandlestickHandler(
	candlestickService *candlestick.CandlestickService,
	subscriptionService *subscription.SubscriptionService,
	uidService *uids.UIDService,
) *CandlestickHandler {
	return &CandlestickHandler{
		candlestickService:  candlestickService,
		subscriptionService: subscriptionService,
		uidService:          uidService,
	}
}

func (h *CandlestickHandler) SubscribeToCandlesticks(
	req *candlestickpb.SubscribeToStreamRequest,
	srv candlestickpb.CandlestickService_SubscribeToCandlesticksServer,
) error {
	if len(req.Symbols) == 0 {
		return fmt.Errorf("Failed to validate request - symbol must not be empty")
	}

	id, err := h.uidService.GenerateUID()
	if err != nil {
		return fmt.Errorf("Failed to generate an id for subscriber")
	}

	err = h.subscriptionService.AddUpdateSubscriber(
		srv.Context(),
		id,
		req.Symbols,
		srv,
	)
	if err != nil {
		return fmt.Errorf(
			"Failed to add symbols %v to subscriber %d",
			req.Symbols,
			id,
		)
	}

	// cleanup when client disconnects
	defer h.subscriptionService.RemoveSubscriber(srv.Context(), id, nil)

	// block until context is done or client disconnects
	<-srv.Context().Done()
	return srv.Context().Err()
}

// if not symbols are provided, will unsubscribe from all symbols
func (h *CandlestickHandler) UnsubscribeFromCandlesticks(
	ctx context.Context,
	req *candlestickpb.UnsubscribeFromStreamRequest,
) (*candlestickpb.GenericResponse, error) {
	if req.SubscriberId == 0 {
		return nil, fmt.Errorf("Failed to validate request - a valid subscriber id must be provided")
	}

	err := h.subscriptionService.RemoveSubscriber(ctx, req.SubscriberId, req.Symbols)
	if err != nil {
		return nil, fmt.Errorf("Unsubscribing failed - %w", err)
	}

	var message string
	if len(req.Symbols) != 0 {
		message = fmt.Sprintf("Successfully unsubscribed from %s", req.Symbols)
	} else {
		message = fmt.Sprintf("Successfully unsubscribed from all")
	}

	return &candlestickpb.GenericResponse{
		Message: message,
	}, nil
}
