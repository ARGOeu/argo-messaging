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
	suite.Equal("Success: ok message", grpcStatus.Result())

	// error status
	grpcStatus2 := GrpcClientStatus{
		err:     status.Error(codes.InvalidArgument, "invalid argument"),
		message: "",
	}

	suite.Equal("Error: invalid argument", grpcStatus2.Result())
}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
