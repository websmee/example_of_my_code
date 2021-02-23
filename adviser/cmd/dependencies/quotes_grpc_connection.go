package dependencies

import (
	"strconv"

	consulsd "github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
	"google.golang.org/grpc"
)

func GetQuotesGRPCConnection(addr, consulAddr, consulPort string) (*grpc.ClientConn, error) {
	var discoveredQuotesAddr string
	{
		if addr == "" {
			consulConfig := api.DefaultConfig()

			consulConfig.Address = "http://" + consulAddr + ":" + consulPort
			consulClient, err := api.NewClient(consulConfig)
			if err != nil {
				return nil, err
			}
			client := consulsd.NewClient(consulClient)

			entries, _, err := client.Service("Quotes", "", false, &api.QueryOptions{})
			if err != nil || len(entries) == 0 {
				return nil, err
			}
			discoveredQuotesAddr = entries[0].Service.Address + ":" + strconv.Itoa(entries[0].Service.Port)
		} else {
			discoveredQuotesAddr = addr
		}
	}

	var quotesConn *grpc.ClientConn
	{
		var err error
		quotesConn, err = grpc.Dial(
			discoveredQuotesAddr,
			grpc.WithInsecure(),
		)
		if err != nil {
			return nil, err
		}
	}

	return quotesConn, nil
}
