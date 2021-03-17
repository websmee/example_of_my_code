package grpc

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/pkg/errors"
	"github.com/websmee/example_of_my_code/quotes/api/proto"

	"github.com/websmee/example_of_my_code/adviser/domain/quote"
)

func (r quotesAppGRPCClient) GetQuotes(ctx context.Context) ([]quote.Quote, error) {
	resp, err := r.getQuotesEndpoint(ctx, GetQuotesRequest{})
	if err != nil {
		return nil, errors.Wrap(err, "GetQuotes failed")
	}

	return resp.(GetQuotesResponse).Quotes, nil
}

var (
	_ endpoint.Failer = GetQuotesResponse{}
)

type GetQuotesRequest struct{}

type GetQuotesResponse struct {
	Quotes []quote.Quote
	Err    error
}

func (r GetQuotesResponse) Failed() error { return r.Err }

func decodeGRPCGetQuotesResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	reply := grpcReply.(*proto.GetQuotesReply)
	qs := make([]quote.Quote, len(reply.Quotes))
	for i := range reply.Quotes {
		qs[i] = quote.Quote{
			ID:     reply.Quotes[i].Id,
			Symbol: reply.Quotes[i].Symbol,
			Name:   reply.Quotes[i].Name,
		}
	}

	return GetQuotesResponse{
		Quotes: qs,
		Err:    nil,
	}, nil
}

func encodeGRPCGetQuotesRequest(_ context.Context, _ interface{}) (interface{}, error) {
	return &proto.GetQuotesRequest{}, nil
}
