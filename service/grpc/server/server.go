package service

import (
	"context"
	"fmt"
	"log"
	"net"

	"chain-crawler/db"
	"chain-crawler/model"
	"chain-crawler/service"
	"chain-crawler/utils"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

import pb "chain-crawler/service/grpc/protobuf"

type server struct {
	pb.UnimplementedTransactionServer
	db   db.DB[model.Account]
	log  *utils.ZapLogger
	port uint16
}

func (s *server) Run() (err error) {
	grpcServer := grpc.NewServer()
	pb.RegisterTransactionServer(grpcServer, s)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	err = grpcServer.Serve(lis)
	return err
}

func NewGrpc(db db.DB[model.Account], log *utils.ZapLogger, port uint16) service.Service {
	return &server{
		db:   db,
		port: port,
		log:  log,
	}
}

func (s *server) GetTotalPaidFee(ctx context.Context, in *pb.Address) (*pb.Account, error) {
	res, err := s.db.Get(in.GetAddress())
	if err != nil && !s.db.IsNotFoundError(err) {
		s.log.Errorw(err.Error())
	} else if err != nil && s.db.IsNotFoundError(err) {
		return nil, status.Errorf(codes.NotFound, "address not found: %v", in.GetAddress())
	}
	return &pb.Account{
		Address:      res.Address,
		TotalPaidFee: res.TotalPaidFee,
		LastHeight:   res.LastHeight,
		TxIndex:      int32(res.TxIndex),
		FirstHeight:  res.FirstHeight,
		IsContract:   res.IsContract,
	}, nil
}

func (s *server) GetStatus(ctx context.Context, in *pb.Empty) (*pb.Account, error) {
	res, err := s.db.Get(db.LastHeightKey)
	if err != nil && !s.db.IsNotFoundError(err) {
		s.log.Errorw(err.Error())
	} else if err != nil && s.db.IsNotFoundError(err) {
		return nil, status.Errorf(codes.NotFound, "height not found")
	}
	return &pb.Account{
		LastHeight:   res.LastHeight,
		Address:      res.Address,
		TotalPaidFee: res.TotalPaidFee,
		TxIndex:      int32(res.TxIndex),
		FirstHeight:  res.FirstHeight,
		IsContract:   res.IsContract,
	}, nil
}

func (s *server) GetFirstHeight(ctx context.Context, in *pb.Empty) (*pb.Account, error) {
	records := make(chan db.DBItem[model.Account], 10)
	err := error(nil)
	go func() {
		err = s.db.Records(nil, nil, records)
	}()

	minTransactionHeight := int64(1e10)
	var firstAccount model.Account
	for {
		item, ok := <-records
		if ok {
			if item.Value.FirstHeight < minTransactionHeight && item.Value.FirstHeight != 0 {
				minTransactionHeight = item.Value.FirstHeight
				firstAccount = item.Value
			}
		} else {
			break
		}
	}
	if err != nil && !s.db.IsNotFoundError(err) {
		s.log.Errorw(err.Error())
	}
	return &pb.Account{
		Address:      firstAccount.Address,
		TotalPaidFee: firstAccount.TotalPaidFee,
		LastHeight:   firstAccount.LastHeight,
		TxIndex:      int32(firstAccount.TxIndex),
		FirstHeight:  firstAccount.FirstHeight,
		IsContract:   firstAccount.IsContract,
	}, nil
}
