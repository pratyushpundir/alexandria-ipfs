package service

import (
	"context"
	"fmt"
	"strconv"

	pb "github.com/pratyushpundir/alexandria-api/gen/ipfs/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCServer implements the IPFS gRPC service
type GRPCServer struct {
	pb.UnimplementedIPFSServiceServer
	client IPFSClient
}

// NewGRPCServer creates a new IPFS gRPC server
func NewGRPCServer(client IPFSClient) *GRPCServer {
	return &GRPCServer{
		client: client,
	}
}

// UploadContent uploads raw content to IPFS
func (s *GRPCServer) UploadContent(ctx context.Context, req *pb.UploadContentRequest) (*pb.UploadContentResponse, error) {
	if len(req.Data) == 0 {
		return nil, status.Error(codes.InvalidArgument, "data is required")
	}

	filename := req.Filename
	if filename == "" {
		filename = "content"
	}

	result, err := s.client.Upload(ctx, req.Data, filename)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to upload content: %v", err)
	}

	size, _ := strconv.ParseInt(result.Size, 10, 64)

	return &pb.UploadContentResponse{
		Cid:       result.IPFSHash,
		SizeBytes: size,
	}, nil
}

// UploadProto uploads serialized protobuf data to IPFS
func (s *GRPCServer) UploadProto(ctx context.Context, req *pb.UploadProtoRequest) (*pb.UploadProtoResponse, error) {
	if len(req.ProtoData) == 0 {
		return nil, status.Error(codes.InvalidArgument, "proto_data is required")
	}

	filename := fmt.Sprintf("%s.pb", req.ProtoType)
	if req.ProtoType == "" {
		filename = "data.pb"
	}

	result, err := s.client.Upload(ctx, req.ProtoData, filename)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to upload proto: %v", err)
	}

	size, _ := strconv.ParseInt(result.Size, 10, 64)

	return &pb.UploadProtoResponse{
		Cid:       result.IPFSHash,
		SizeBytes: size,
	}, nil
}

// GetContent retrieves content from IPFS by CID
func (s *GRPCServer) GetContent(ctx context.Context, req *pb.GetContentRequest) (*pb.GetContentResponse, error) {
	if req.Cid == "" {
		return nil, status.Error(codes.InvalidArgument, "cid is required")
	}

	data, err := s.client.Get(ctx, req.Cid)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "failed to get content: %v", err)
	}

	return &pb.GetContentResponse{
		Data:      data,
		SizeBytes: int64(len(data)),
	}, nil
}

// GetProto retrieves protobuf data from IPFS by CID
func (s *GRPCServer) GetProto(ctx context.Context, req *pb.GetProtoRequest) (*pb.GetProtoResponse, error) {
	if req.Cid == "" {
		return nil, status.Error(codes.InvalidArgument, "cid is required")
	}

	data, err := s.client.Get(ctx, req.Cid)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "failed to get proto: %v", err)
	}

	return &pb.GetProtoResponse{
		ProtoData: data,
	}, nil
}

// PinContent pins content on IPFS
func (s *GRPCServer) PinContent(ctx context.Context, req *pb.PinContentRequest) (*pb.PinContentResponse, error) {
	if req.Cid == "" {
		return nil, status.Error(codes.InvalidArgument, "cid is required")
	}

	if err := s.client.Pin(ctx, req.Cid); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to pin content: %v", err)
	}

	return &pb.PinContentResponse{
		Success: true,
	}, nil
}

// UnpinContent unpins content from IPFS
func (s *GRPCServer) UnpinContent(ctx context.Context, req *pb.UnpinContentRequest) (*pb.UnpinContentResponse, error) {
	if req.Cid == "" {
		return nil, status.Error(codes.InvalidArgument, "cid is required")
	}

	if err := s.client.Unpin(ctx, req.Cid); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to unpin content: %v", err)
	}

	return &pb.UnpinContentResponse{
		Success: true,
	}, nil
}

// GetGatewayURL returns the public gateway URL for a CID
func (s *GRPCServer) GetGatewayURL(ctx context.Context, req *pb.GetGatewayURLRequest) (*pb.GetGatewayURLResponse, error) {
	if req.Cid == "" {
		return nil, status.Error(codes.InvalidArgument, "cid is required")
	}

	url := s.client.GetGatewayURL(req.Cid)

	return &pb.GetGatewayURLResponse{
		Url: url,
	}, nil
}
