/*
 * Copyright 2020 ICON Foundation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// -build base

package hvh

import (
	"bytes"
	"encoding/json"
	"math/big"

	"github.com/icon-project/goloop/common"
	"github.com/icon-project/goloop/common/codec"
	"github.com/icon-project/goloop/common/crypto"
	"github.com/icon-project/goloop/common/errors"
	"github.com/icon-project/goloop/havah/hvhmodule"
	"github.com/icon-project/goloop/module"
	"github.com/icon-project/goloop/service/contract"
	"github.com/icon-project/goloop/service/state"
	"github.com/icon-project/goloop/service/transaction"
	"github.com/icon-project/goloop/service/txresult"
)

type baseDataJSON struct {
	issueAmount *common.HexInt `json:"issueAmount"`

	//rewardTotal  *common.HexInt `json:"rewardTotal"`
	//rewardRemain *common.HexInt `json:"rewardRemain"`
}

func parseBaseData(data []byte) (*baseDataJSON, error) {
	if data == nil {
		return nil, nil
	}
	jso := new(baseDataJSON)
	jd := json.NewDecoder(bytes.NewBuffer(data))
	jd.DisallowUnknownFields()
	if err := jd.Decode(jso); err != nil {
		return nil, err
	}
	return jso, nil
}

type baseV3Data struct {
	Version   common.HexUint16 `json:"version"`
	From      *common.Address  `json:"from,omitempty"` // it should be nil
	TimeStamp common.HexInt64  `json:"timestamp"`
	DataType  string           `json:"dataType,omitempty"`
	Data      json.RawMessage  `json:"data,omitempty"`
}

func (tx *baseV3Data) calcHash() ([]byte, error) {
	sha := bytes.NewBuffer(nil)
	sha.Write([]byte("icx_sendTransaction"))

	// data
	if tx.Data != nil {
		sha.Write([]byte(".data."))
		if len(tx.Data) > 0 {
			var obj interface{}
			if err := json.Unmarshal(tx.Data, &obj); err != nil {
				return nil, err
			}
			if bs, err := transaction.SerializeValue(obj); err != nil {
				return nil, err
			} else {
				sha.Write(bs)
			}
		}
	}

	// dataType
	sha.Write([]byte(".dataType."))
	sha.Write([]byte(tx.DataType))

	// timestamp
	sha.Write([]byte(".timestamp."))
	sha.Write([]byte(tx.TimeStamp.String()))

	// version
	sha.Write([]byte(".version."))
	sha.Write([]byte(tx.Version.String()))

	return crypto.SHA3Sum256(sha.Bytes()), nil
}

type baseV3 struct {
	baseV3Data

	id    []byte
	hash  []byte
	bytes []byte
}

func (tx *baseV3) Version() int {
	return module.TransactionVersion3
}

func (tx *baseV3) Prepare(ctx contract.Context) (state.WorldContext, error) {
	lq := []state.LockRequest{
		{state.WorldIDStr, state.AccountWriteLock},
	}
	wc := ctx.GetFuture(lq)
	wc.WorldVirtualState().Ensure()

	return wc, nil
}

func (tx *baseV3) Execute(ctx contract.Context, estimate bool) (txresult.Receipt, error) {
	if estimate {
		return nil, errors.InvalidStateError.New("EstimationNotAllowed")
	}
	info := ctx.TransactionInfo()
	if info == nil {
		return nil, errors.InvalidStateError.New("TransactionInfoUnavailable")
	}
	if info.Index != 0 {
		return nil, errors.CriticalFormatError.New("BaseMustBeTheFirst")
	}

	cc := contract.NewCallContext(ctx, ctx.GetStepLimit(state.StepLimitTypeInvoke), false)
	defer cc.Dispose()

	icc := NewCallContext(cc, tx.From())
	es := cc.GetExtensionState().(*ExtensionStateImpl)
	if err := es.OnBaseTx(icc, tx.Data); err != nil {
		return nil, err
	}

	// Make a receipt
	r := txresult.NewReceipt(ctx.Database(), ctx.Revision(), cc.Treasury())
	cc.GetEventLogs(r)
	r.SetResult(module.StatusSuccess, new(big.Int), new(big.Int), nil)
	es.ClearCache()
	return r, nil
}

func (tx *baseV3) Dispose() {
	//panic("implement me")
}

func (tx *baseV3) Group() module.TransactionGroup {
	return module.TransactionGroupNormal
}

func (tx *baseV3) ID() []byte {
	if tx.id == nil {
		if bs, err := tx.baseV3Data.calcHash(); err != nil {
			panic(err)
		} else {
			tx.id = bs
		}
	}
	return tx.id
}

func (tx *baseV3) From() module.Address {
	return state.SystemAddress
}

func (tx *baseV3) Bytes() []byte {
	if tx.bytes == nil {
		if bs, err := codec.BC.MarshalToBytes(&tx.baseV3Data); err != nil {
			panic(err)
		} else {
			tx.bytes = bs
		}
	}
	return tx.bytes
}

func (tx *baseV3) Hash() []byte {
	if tx.hash == nil {
		tx.hash = crypto.SHA3Sum256(tx.Bytes())
	}
	return tx.hash
}

func (tx *baseV3) Verify() error {
	return nil
}

func (tx *baseV3) ToJSON(version module.JSONVersion) (interface{}, error) {
	jso := map[string]interface{}{
		"version":   &tx.baseV3Data.Version,
		"timestamp": &tx.baseV3Data.TimeStamp,
		"dataType":  tx.baseV3Data.DataType,
		"data":      tx.baseV3Data.Data,
	}
	jso["txHash"] = common.HexBytes(tx.ID())
	return jso, nil
}

func (tx *baseV3) ValidateNetwork(nid int) bool {
	return true
}

func (tx *baseV3) PreValidate(wc state.WorldContext, update bool) error {
	return nil
}

func (tx *baseV3) GetHandler(cm contract.ContractManager) (transaction.Handler, error) {
	return tx, nil
}

func (tx *baseV3) Timestamp() int64 {
	return tx.baseV3Data.TimeStamp.Value
}

func (tx *baseV3) Nonce() *big.Int {
	return nil
}

func (tx *baseV3) To() module.Address {
	return state.SystemAddress
}

func (tx *baseV3) IsSkippable() bool {
	return false
}

func checkBaseV3JSON(jso map[string]interface{}) bool {
	if d, ok := jso["dataType"]; !ok || d != "base" {
		return false
	}
	if v, ok := jso["version"]; !ok || v != "0x3" {
		return false
	}
	return true
}

func parseBaseV3JSON(bs []byte, raw bool) (transaction.Transaction, error) {
	tx := new(baseV3)
	if err := json.Unmarshal(bs, &tx.baseV3Data); err != nil {
		return nil, transaction.InvalidFormat.Wrap(err, "InvalidJSON")
	}
	if tx.baseV3Data.From != nil {
		return nil, transaction.InvalidFormat.New("InvalidFromValue(NonNil)")
	}
	return tx, nil
}

type baseV3Header struct {
	Version common.HexUint16 `json:"version"`
	From    *common.Address  `json:"from"` // it should be nil
}

func checkBaseV3Bytes(bs []byte) bool {
	var vh baseV3Header
	if _, err := codec.BC.UnmarshalFromBytes(bs, &vh); err != nil {
		return false
	}
	return vh.From == nil
}

func parseBaseV3Bytes(bs []byte) (transaction.Transaction, error) {
	tx := new(baseV3)
	if _, err := codec.BC.UnmarshalFromBytes(bs, &tx.baseV3Data); err != nil {
		return nil, err
	}
	return tx, nil
}

func RegisterBaseTx() {
	transaction.RegisterFactory(&transaction.Factory{
		Priority:    15,
		CheckJSON:   checkBaseV3JSON,
		ParseJSON:   parseBaseV3JSON,
		CheckBinary: checkBaseV3Bytes,
		ParseBinary: parseBaseV3Bytes,
	})
}

func (es *ExtensionStateImpl) OnBaseTx(cc hvhmodule.CallContext, data []byte) error {
	height := cc.BlockHeight()
	issueStart := es.state.GetIssueStart()

	if !(issueStart > 0 && height >= issueStart) {
		panic("This method MUST NOT be called in this case")
	}

	baseData, err := parseBaseData(data)
	if err != nil {
		return transaction.InvalidFormat.Wrap(err, "Failed to parse baseData")
	}

	termPeriod := es.state.GetTermPeriod()
	issueAmount := es.state.GetIssueAmount()
	baseTxCount := height - issueStart
	termSeq := baseTxCount / termPeriod

	if baseData.issueAmount.Value().Cmp(issueAmount) != 0 {
		return transaction.InvalidTxValue.Errorf(
			"IssueAmount mismatch: actual(%s) != expected(%s)", issueAmount, baseData.issueAmount)
	}

	if err = es.onTermEnd(cc, termSeq-1); err != nil {
		return err
	}
	if err = es.onTermStart(cc, termSeq); err != nil {
		return err
	}
	return nil
}

func (es *ExtensionStateImpl) onTermEnd(cc hvhmodule.CallContext, termSeq int64) error {
	var err error
	if termSeq >= 0 {
		// TxFee Distribution
		if err = distributeFee(cc, cc.Treasury(), hvhmodule.BigRatEcoSystemProportion); err != nil {
			return err
		}
		// ServiceFee Distribution
		if err = distributeFee(cc, hvhmodule.ServiceTreasury, hvhmodule.BigRatEcoSystemProportion); err != nil {
			return err
		}
	}
	return nil
}

func (es *ExtensionStateImpl) onTermStart(cc hvhmodule.CallContext, termSeq int64) error {
	var err error
	issueAmount := es.state.GetIssueAmount()
	reductionCycle := es.state.GetIssueReductionCycle()

	// Reduce the amount of coin to issue by 30% every reduction cycle
	if termSeq > 0 && termSeq%reductionCycle == 0 {
		reductionRate := es.state.GetIssueReductionRate()
		newIssueAmount := calcIssueAmount(issueAmount, reductionRate)

		if issueAmount.Cmp(newIssueAmount) != 0 {
			if err = es.state.SetBigInt(hvhmodule.VarIssueAmount, newIssueAmount); err != nil {
				return err
			}
			es.Logger().Infof(
				"IssueAmount is reduced: rate=%v before=%s after=%s",
				reductionRate, issueAmount, newIssueAmount)
			issueAmount = newIssueAmount
		}
	}

	if err = issueCoin(cc, termSeq, issueAmount); err != nil {
		return err
	}

	// Reset reward-related states
	if err = es.state.OnTermStart(termSeq, issueAmount); err != nil {
		return err
	}

	return nil
}

func issueCoin(cc hvhmodule.CallContext, termSeq int64, amount *big.Int) error {
	if amount != nil && amount.Sign() > 0 {
		newTotalSupply, err := cc.AddTotalSupply(amount)
		if err != nil {
			return err
		}
		if err = cc.Deposit(hvhmodule.PublicTreasury, amount); err != nil {
			return err
		}
		onICXIssuedEvent(cc, termSeq, amount, newTotalSupply)
	}
	return nil
}

func distributeFee(cc hvhmodule.CallContext, from module.Address, proportion *big.Rat) error {
	var err error
	balance := cc.GetBalance(from)
	if balance.Sign() > 0 {
		ecoAmount := new(big.Int).Mul(balance, proportion.Num())
		ecoAmount.Div(ecoAmount, proportion.Denom())
		susAmount := new(big.Int).Sub(balance, ecoAmount)

		if err = cc.Transfer(from, hvhmodule.SustainableFund, susAmount); err != nil {
			return err
		}
		if err = cc.Transfer(from, hvhmodule.EcoSystem, ecoAmount); err != nil {
			return err
		}
	}
	return nil
}

func calcIssueAmount(curIssueAmount *big.Int, reductionRate *big.Rat) *big.Int {
	amount := new(big.Int).Set(curIssueAmount)
	numerator := new(big.Int).Sub(reductionRate.Denom(), reductionRate.Num())
	amount.Mul(amount, numerator)
	amount.Div(amount, reductionRate.Denom())
	return amount
}
