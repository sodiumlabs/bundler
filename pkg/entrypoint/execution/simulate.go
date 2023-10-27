package execution

import (
	"context"
	"fmt"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stackup-wallet/stackup-bundler/pkg/entrypoint"
	"github.com/stackup-wallet/stackup-bundler/pkg/entrypoint/reverts"
	"github.com/stackup-wallet/stackup-bundler/pkg/errors"
	"github.com/stackup-wallet/stackup-bundler/pkg/signer"
	"github.com/stackup-wallet/stackup-bundler/pkg/userop"
)

func SimulateHandleOp(
	signer *signer.EOA,
	chainId *big.Int,
	rpc *rpc.Client,
	entryPoint common.Address,
	op *userop.UserOperation,
	target common.Address,
	data []byte,
) (*reverts.ExecutionResultRevert, error) {
	ethClient := ethclient.NewClient(rpc)
	ep, err := entrypoint.NewEntrypoint(entryPoint, ethClient)
	if err != nil {
		return nil, err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(signer.PrivateKey, chainId)
	if err != nil {
		return nil, err
	}
	auth.GasLimit = math.MaxUint64
	auth.NoSend = true

	op.MaxFeePerGas = big.NewInt(1)
	op.MaxPriorityFeePerGas = big.NewInt(1)

	tx, err := ep.SimulateHandleOp(auth, entrypoint.UserOperation(*op), target, data)
	if err != nil {
		return nil, err
	}

	_, err = ethClient.EstimateGas(context.Background(), ethereum.CallMsg{
		From:       signer.Address,
		To:         tx.To(),
		Gas:        tx.Gas(),
		GasPrice:   tx.GasPrice(),
		GasFeeCap:  tx.GasFeeCap(),
		GasTipCap:  tx.GasTipCap(),
		Value:      tx.Value(),
		Data:       tx.Data(),
		AccessList: tx.AccessList(),
	})

	sim, simErr := reverts.NewExecutionResult(err)
	if simErr != nil {
		fo, foErr := reverts.NewFailedOp(err)
		if foErr != nil {
			fs, fsErr := reverts.NewFailedStr(err)
			if fsErr != nil {
				return nil, fmt.Errorf("%s, %s, %s", simErr, foErr, fsErr)
			}
			return nil, errors.NewRPCError(errors.REJECTED_BY_EP_OR_ACCOUNT, fs.Reason, fs)
		}
		return nil, errors.NewRPCError(errors.REJECTED_BY_EP_OR_ACCOUNT, fo.Reason, fo)
	}

	return sim, nil
}

func EstimateCreationGas(
	signer *signer.EOA,
	rpc *rpc.Client,
	op *userop.UserOperation,
) (uint64, error) {
	// if wallet inited
	if len(op.InitCode) == 0 {
		return 0, nil
	}

	factoryAddress := common.BytesToAddress(op.InitCode[:20])
	walletCreateCallData := op.InitCode[20:]

	ethClient := ethclient.NewClient(rpc)

	return ethClient.EstimateGas(context.Background(), ethereum.CallMsg{
		From:       signer.Address,
		To:         &factoryAddress,
		Gas:        math.MaxUint64,
		GasPrice:   nil,
		GasFeeCap:  nil,
		GasTipCap:  nil,
		Value:      big.NewInt(0),
		Data:       walletCreateCallData,
		AccessList: nil,
	})
}
