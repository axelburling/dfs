package grpc

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewPreClient(target string, customOptions []grpc.DialOption) (*grpc.ClientConn, error) {
	var options []grpc.DialOption
	options = append(options, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if len(customOptions) > 0 {
		for _, opt := range customOptions {
			if opt != nil {
				options = append(options, opt)
			}
		}
	}

	conn, err := grpc.NewClient(target, options...)

	if err != nil {
		return nil, err
	}

	return conn, nil
}