package execution

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stackup-wallet/stackup-bundler/pkg/entrypoint"
	"github.com/stackup-wallet/stackup-bundler/pkg/entrypoint/reverts"
	"github.com/stackup-wallet/stackup-bundler/pkg/errors"
	"github.com/stackup-wallet/stackup-bundler/pkg/userop"
)

func SimulateHandleOp(
	rpc *rpc.Client,
	entryPoint common.Address,
	op *userop.UserOperation,
	target common.Address,
	data []byte,
) (*reverts.ExecutionResultRevert, error) {
	ep, err := entrypoint.NewEntrypoint(entryPoint, ethclient.NewClient(rpc))
	if err != nil {
		return nil, err
	}

	// fmt.Println("VerificationGasLimit", op.VerificationGasLimit)
	op.VerificationGasLimit = big.NewInt(400000)

	rawCaller := &entrypoint.EntrypointRaw{Contract: ep}
	err = rawCaller.Call(
		nil,
		nil,
		"simulateHandleOp",
		entrypoint.UserOperation(*op),
		target,
		data,
	)

	// ethClient := ethclient.NewClient(rpc)

	// ethClient.CallContract()

	sim, simErr := reverts.NewExecutionResult(err)
	if simErr != nil {
		fo, foErr := reverts.NewFailedOp(err)
		if foErr != nil {
			return nil, fmt.Errorf("%s, %s", simErr, foErr)
		}
		return nil, errors.NewRPCError(errors.REJECTED_BY_EP_OR_ACCOUNT, fo.Reason, fo)
	}

	return sim, nil
}
