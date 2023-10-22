package reverts

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
)

type FailedStrRevert struct {
	Reason string
}

func failedStr() abi.Error {
	reason, _ := abi.NewType("string", "string", nil)
	return abi.NewError("Error", abi.Arguments{
		{Name: "reason", Type: reason},
	})
}

func NewFailedStr(err error) (*FailedStrRevert, error) {
	rpcErr, ok := err.(rpc.DataError)
	if !ok {
		return nil, fmt.Errorf(
			"failedStr: cannot assert type: error is not of type rpc.DataError, err: %s",
			err,
		)
	}

	data, ok := rpcErr.ErrorData().(string)
	if !ok {
		return nil, fmt.Errorf(
			"failedOp: cannot assert type: data is not of type string, err: %s, data: %s",
			rpcErr.Error(),
			rpcErr.ErrorData(),
		)
	}

	failedStr := failedStr()
	revert, err := failedStr.Unpack(common.Hex2Bytes(data[2:]))
	if err != nil {
		return nil, fmt.Errorf("failedStr: %s, data: %s", err, data)
	}

	args, ok := revert.([]any)
	if !ok {
		return nil, errors.New("failedStr: cannot assert type: args is not of type []any")
	}
	if len(args) != 1 {
		return nil, fmt.Errorf("failedStr: invalid args length: expected 1, got %d", len(args))
	}

	reason, ok := args[0].(string)
	if !ok {
		return nil, errors.New("failedStr: cannot assert type: reason is not of type string")
	}

	return &FailedStrRevert{
		Reason: reason,
	}, nil
}
