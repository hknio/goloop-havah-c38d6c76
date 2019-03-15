package v3

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"strconv"

	"github.com/icon-project/goloop/common"
	"github.com/icon-project/goloop/common/db"
	"github.com/icon-project/goloop/module"
	"github.com/icon-project/goloop/server/jsonrpc"
)

func MethodRepository() *jsonrpc.MethodRepository {
	mr := jsonrpc.NewMethodRepository()

	mr.RegisterMethod("icx_getLastBlock", getLastBlock)
	mr.RegisterMethod("icx_getBlockByHeight", getBlockByHeight)
	mr.RegisterMethod("icx_getBlockByHash", getBlockByHash)
	mr.RegisterMethod("icx_call", call)
	mr.RegisterMethod("icx_getBalance", getBalance)
	mr.RegisterMethod("icx_getScoreApi", getScoreApi)
	mr.RegisterMethod("icx_getTotalSupply", getTotalSupply)
	mr.RegisterMethod("icx_getTransactionResult", getTransactionResult)
	mr.RegisterMethod("icx_getTransactionByHash", getTransactionByHash)
	mr.RegisterMethod("icx_sendTransaction", sendTransaction)

	mr.RegisterMethod("icx_getDataByHash", getDataByHash)
	mr.RegisterMethod("icx_getBlockHeaderByHeight", getBlockHeaderByHeight)
	mr.RegisterMethod("icx_getVotesByHeight", getVotesByHeight)
	mr.RegisterMethod("icx_getProofForResult", getProofForResult)

	return mr
}

func getLastBlock(ctx *jsonrpc.Context, params *jsonrpc.Params) (interface{}, error) {
	if !params.IsEmpty() {
		return nil, jsonrpc.ErrInvalidParams()
	}

	chain, _ := ctx.Chain()
	bm := chain.BlockManager()

	block, _ := bm.GetLastBlock()
	result, _ := block.ToJSON(3)

	return result, nil
}

func getBlockByHeight(ctx *jsonrpc.Context, params *jsonrpc.Params) (interface{}, error) {
	var param BlockHeightParam
	if err := params.Convert(&param); err != nil {
		return nil, err
	}

	chain, _ := ctx.Chain()
	bm := chain.BlockManager()

	block, _ := bm.GetBlockByHeight(param.Height.Value())
	result, _ := block.ToJSON(3)

	return result, nil
}

func getBlockByHash(ctx *jsonrpc.Context, params *jsonrpc.Params) (interface{}, error) {
	var param BlockHashParam
	if err := params.Convert(&param); err != nil {
		return nil, err
	}

	chain, _ := ctx.Chain()
	bm := chain.BlockManager()

	block, _ := bm.GetBlock(param.Hash.Bytes())
	result, _ := block.ToJSON(3)

	return result, nil
}

func call(ctx *jsonrpc.Context, params *jsonrpc.Params) (interface{}, error) {
	return nil, nil
}

func getBalance(ctx *jsonrpc.Context, params *jsonrpc.Params) (interface{}, error) {
	var param AddressParam
	if err := params.Convert(&param); err != nil {
		return nil, err
	}

	chain, _ := ctx.Chain()
	bm := chain.BlockManager()
	sm := chain.ServiceManager()

	var balance common.HexInt
	block, _ := bm.GetLastBlock()
	balance.Set(sm.GetBalance(block.Result(), param.Address.Address()))

	return balance, nil
}

func getScoreApi(ctx *jsonrpc.Context, params *jsonrpc.Params) (interface{}, error) {
	var param ScoreAddressParam
	if err := params.Convert(&param); err != nil {
		return nil, err
	}
	// TODO : service interface required
	return nil, nil
}

func getTotalSupply(ctx *jsonrpc.Context, params *jsonrpc.Params) (interface{}, error) {
	if !params.IsEmpty() {
		return nil, jsonrpc.ErrInvalidParams()
	}
	// TODO : service interface required
	return nil, nil
}

func getTransactionResult(ctx *jsonrpc.Context, params *jsonrpc.Params) (interface{}, error) {
	var param TransactionHashParam
	if err := params.Convert(&param); err != nil {
		return nil, err
	}

	chain, _ := ctx.Chain()
	bm := chain.BlockManager()

	ti, _ := bm.GetTransactionInfo(param.Hash.Bytes())
	block := ti.Block()
	receipt := ti.GetReceipt()
	res, _ := receipt.ToJSON(3)

	result := res.(map[string]interface{})
	result["blockHash"] = "0x" + hex.EncodeToString(block.ID())
	result["blockHeight"] = "0x" + strconv.FormatInt(int64(block.Height()), 16)
	result["txIndex"] = "0x" + strconv.FormatInt(int64(ti.Index()), 16)

	return result, nil
}

func getTransactionByHash(ctx *jsonrpc.Context, params *jsonrpc.Params) (interface{}, error) {
	var param TransactionHashParam
	if err := params.Convert(&param); err != nil {
		return nil, err
	}

	chain, _ := ctx.Chain()
	bm := chain.BlockManager()

	ti, _ := bm.GetTransactionInfo(param.Hash.Bytes())
	tx := ti.Transaction()
	result, _ := tx.ToJSON(3)

	return result, nil
}

func sendTransaction(ctx *jsonrpc.Context, params *jsonrpc.Params) (interface{}, error) {
	var param TransactionParam
	if err := params.Convert(&param); err != nil {
		return nil, err
	}

	chain, _ := ctx.Chain()
	sm := chain.ServiceManager()
	tx, _ := json.Marshal(param)
	hash, _ := sm.SendTransaction(tx)
	result := "0x" + hex.EncodeToString(hash)

	return result, nil
}

func getDataByHash(ctx *jsonrpc.Context, params *jsonrpc.Params) (interface{}, error) {
	var param DataHashParam
	if err := params.Convert(&param); err != nil {
		return nil, err
	}

	chain, _ := ctx.Chain()
	dbm := chain.Database()

	bucket, err := dbm.GetBucket(db.BytesByHash)
	if err != nil {

	}
	value, err := bucket.Get(param.Hash.Bytes())
	if err != nil {

	}
	if value == nil {
		return nil, jsonrpc.ErrInvalidParams()
	}

	return value, nil
}

func getBlockHeaderByHeight(ctx *jsonrpc.Context, params *jsonrpc.Params) (interface{}, error) {
	var param BlockHeightParam
	if err := params.Convert(&param); err != nil {
		return nil, err
	}

	chain, _ := ctx.Chain()
	bm := chain.BlockManager()

	block, _ := bm.GetBlockByHeight(param.Height.Value())
	buf := bytes.NewBuffer(nil)
	if err := block.MarshalHeader(buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func getVotesByHeight(ctx *jsonrpc.Context, params *jsonrpc.Params) (interface{}, error) {
	var param BlockHeightParam
	if err := params.Convert(&param); err != nil {
		return nil, err
	}

	chain, _ := ctx.Chain()
	cs := chain.Consensus()
	votes, _ := cs.GetVotesByHeight(param.Height.Value())

	return votes.Bytes(), nil
}

func getProofForResult(ctx *jsonrpc.Context, params *jsonrpc.Params) (interface{}, error) {
	var param ProofResultParam
	if err := params.Convert(&param); err != nil {
		return nil, err
	}

	chain, _ := ctx.Chain()
	bm := chain.BlockManager()
	sm := chain.ServiceManager()

	block, _ := bm.GetBlock(param.BlockHash.Bytes())
	blockResult := block.Result()
	receipts := sm.ReceiptListFromResult(blockResult, module.TransactionGroupNormal)
	proofs, _ := receipts.GetProof(int(param.Index.Value()))

	return proofs, nil
}