package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	pb "github.com/b-harvest/all-in-one-admin-backend/all-in-one-admin-backend"
	"net"
	"strings"
	"time"

	"github.com/tendermint/tendermint/crypto"
	client "github.com/tendermint/tendermint/rpc/client/http"
	"google.golang.org/grpc"
)

type BlockIDFlag byte
type Address = crypto.Address

type StatusInfo struct {
	LatestBlockHeight int64  `json:"latest_block_height"`
	CatchingUp        bool   `json:"catching_up"`
	Moniker           string `json:"moniker"`
}

type SignInfo struct {
	LatestBlockHeight int64 `json:"latest_block_height"`
	SignInfo          bool  `json:"SignInfo"`
}

type server struct {
	pb.UnimplementedMonitoringServer
}

//Future function: indep alarm health check
func alarm_health_check() {
	for {
		time.Sleep(2 * time.Second)
		fmt.Println("Hello world!")
	}
}
func (s *server) GetvalidatorSignInfo(ctx context.Context, in *pb.SignInfoRequest) (*pb.SignInfoResponse, error) {
	httpClient, _ := client.NewWithTimeout(in.GetNodeURI(), "/websocket", 3)
	status, err := httpClient.Status()
	commit_height := int64(status.SyncInfo.LatestBlockHeight)
	commit, err := httpClient.Commit(&commit_height)
	precommit := false
	for _, value := range commit.SignedHeader.Commit.Signatures {
		validatoraddress := hex.EncodeToString(value.ValidatorAddress)
		if strings.ToUpper(validatoraddress) == in.GetValidatorAddress() {
			log.Printf("precommit: %s", value.ValidatorAddress)
			log.Printf("precommit: %s", in.GetValidatorAddress())
			precommit = true
		}
	}

	if err != nil {
		println(err.Error())
		return &pb.SignInfoResponse{Status: "ERROR"}, err
	}
	var u = SignInfo{
		status.SyncInfo.LatestBlockHeight,
		precommit,
	}
	marshal_u, _ := json.Marshal(u)
	log.Printf("Received profile: %v", in.GetNodeURI())
	return &pb.SignInfoResponse{Status: string(marshal_u)}, nil
}

func (s *server) GetnodeStatus(ctx context.Context, in *pb.StatusRequest) (*pb.StatusResponse, error) {
	httpClient, _ := client.NewWithTimeout(in.GetNodeURI(), "/websocket", 3)
	status, err := httpClient.Status()

	if err != nil {
		println(err.Error())
		return &pb.StatusResponse{Status: "ERROR"}, err
	}

	var u = StatusInfo{
		status.SyncInfo.LatestBlockHeight,
		status.SyncInfo.CatchingUp,
		status.NodeInfo.Moniker,
	}
	marshal_u, _ := json.Marshal(u)
	log.Printf("Received profile: %v", in.GetNodeURI())
	return &pb.StatusResponse{Status: string(marshal_u)}, nil
}

func main() {
	go alarm_health_check()
	lis, err := net.Listen("tcp", ":8088")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterMonitoringServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
