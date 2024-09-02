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
	req *candlestickpb.StreamRequest,
	srv candlestickpb.CandlestickService_SubscribeToCandlesticksServer,
) error {
	if req.Symbol == "" {
		return fmt.Errorf("Failed to validate request - symbol must not be empty")
	}

	var id int64
	if req.SubscriberId == 0 {
		var err error

		id, err = h.uidService.GenerateUID()
		if err != nil {
			return fmt.Errorf("Failed to generate an id for subscriber")
		}
	} else {
		id = req.SubscriberId
	}

	err := h.subscriptionService.AddUpdateSubscriber(
		srv.Context(),
		id,
		req.Symbol,
		srv,
	)
	if err != nil {
		return fmt.Errorf(
			"Failed to add symbol %s to subscriber %d",
			req.Symbol,
			id,
		)
	}

	// cleanup when client disconnects
	defer h.subscriptionService.RemoveSubscriber(srv.Context(), id, nil)

	// block until context is done or client disconnects
	<-srv.Context().Done()
	return srv.Context().Err()
}

func (h *CandlestickHandler) UnsubscribeFromCandlesticks(
	ctx context.Context,
	req *candlestickpb.StreamRequest,
) (*candlestickpb.GenericResponse, error) {
	if req.Symbol == "" {
		return nil, fmt.Errorf("Failed to validate request - symbol must not be empty")
	}
	if req.SubscriberId == 0 {
		return nil, fmt.Errorf("Failed to validate request - a valid subscriber id must be provided")
	}

	err := h.subscriptionService.RemoveSubscriber(ctx, req.SubscriberId, &req.Symbol)
	if err != nil {
		return nil, fmt.Errorf("Unsubscribing failed - %w", err)
	}

	return &candlestickpb.GenericResponse{
		Message: "Successfully unsubscribed from " + req.Symbol,
	}, nil
}
