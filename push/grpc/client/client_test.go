package push

import (
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

type ClientTestSuite struct {
	suite.Suite
}

func (suite *ClientTestSuite) TestResult() {

	// ok status
	grpcStatus := GrpcClientStatus{
		err:     nil,
		message: "ok message",
	}
	suite.Equal("ok message", grpcStatus.Result(false))

	// error status
	grpcStatus2 := GrpcClientStatus{
		err:     status.Error(codes.InvalidArgument, "invalid argument"),
		message: "",
	}

	suite.Equal("Error: invalid argument", grpcStatus2.Result(false))

	// unavailable error status
	grpcStatus3 := GrpcClientStatus{
		err:     status.Error(codes.Unavailable, "connection refused"),
		message: "",
	}

	suite.Equal("Push server is currently unavailable", grpcStatus3.Result(false))

	// unavailable detailed status
	grpcStatus4 := GrpcClientStatus{
		err:     status.Error(codes.Unavailable, "connection refused"),
		message: "",
	}

	suite.Equal("connection refused", grpcStatus4.Result(true))
}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
