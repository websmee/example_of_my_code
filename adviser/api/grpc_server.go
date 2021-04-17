package api

import (
	"context"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/tracing/zipkin"
	"github.com/go-kit/kit/transport"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
	"github.com/shopspring/decimal"

	"github.com/websmee/example_of_my_code/adviser/api/proto"
	"github.com/websmee/example_of_my_code/adviser/domain/candlestick"
)

type grpcServer struct {
	proto.UnimplementedAdviserServer
	getAdvices grpctransport.Handler
}

func NewGRPCServer(endpoints Adviser, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer, logger log.Logger) proto.AdviserServer {
	options := []grpctransport.ServerOption{
		grpctransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
	}

	if zipkinTracer != nil {
		options = append(options, zipkin.GRPCServerTrace(zipkinTracer))
	}

	return &grpcServer{
		getAdvices: grpctransport.NewServer(
			endpoints.GetAdvicesEndpoint,
			decodeGRPCGetAdvicesRequest,
			encodeGRPCGetAdvicesResponse,
			append(options, grpctransport.ServerBefore(opentracing.GRPCToContext(otTracer, "GetAdvices", logger)))...,
		),
	}
}

func (s *grpcServer) GetAdvices(ctx context.Context, req *proto.GetAdvicesRequest) (*proto.GetAdvicesReply, error) {
	_, rep, err := s.getAdvices.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*proto.GetAdvicesReply), nil
}

func decodeGRPCGetAdvicesRequest(_ context.Context, _ interface{}) (interface{}, error) {
	return GetAdvicesRequest{}, nil
}

func encodeGRPCGetAdvicesResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(GetAdvicesResponse)
	advices := make(map[int64]*proto.Advice, len(resp.Advices))
	for i := range resp.Advices {
		advices[int64(i)] = &proto.Advice{
			Quote: &proto.AdviceQuote{
				Symbol: resp.Advices[i].Quote.Symbol,
				Name:   resp.Advices[i].Quote.Name,
			},
			Candlesticks:     encodeCandlesticks(resp.Advices[i].Candlesticks),
			Price:            decimalToFloat32(resp.Advices[i].Price),
			Amount:           decimalToFloat32(resp.Advices[i].Price),
			TakeProfitPrice:  decimalToFloat32(resp.Advices[i].Price),
			TakeProfitAmount: decimalToFloat32(resp.Advices[i].Price),
			StopLossPrice:    decimalToFloat32(resp.Advices[i].Price),
			StopLossAmount:   decimalToFloat32(resp.Advices[i].Price),
			Leverage:         int64(resp.Advices[i].Leverage),
			ExpiresAt:        resp.Advices[i].ExpiresAt.Unix(),
		}
	}

	return &proto.GetAdvicesReply{Advices: advices, Err: err2str(resp.Err)}, nil
}

func encodeCandlesticks(candlesticks []candlestick.Candlestick) map[int64]*proto.AdviceCandlestick {
	cs := make(map[int64]*proto.AdviceCandlestick, len(candlesticks))
	for i := range candlesticks {
		cs[candlesticks[i].Timestamp.Unix()] = &proto.AdviceCandlestick{
			Open:      decimalToFloat32(candlesticks[i].Open),
			Low:       decimalToFloat32(candlesticks[i].Low),
			High:      decimalToFloat32(candlesticks[i].High),
			Close:     decimalToFloat32(candlesticks[i].Close),
			AdjClose:  decimalToFloat32(candlesticks[i].AdjClose),
			Volume:    int64(candlesticks[i].Volume),
			Timestamp: candlesticks[i].Timestamp.Unix(),
			Interval:  string(candlesticks[i].Interval),
		}
	}

	return cs
}

func err2str(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func decimalToFloat32(v decimal.Decimal) float32 {
	r, _ := v.Float64()
	return float32(r)
}
