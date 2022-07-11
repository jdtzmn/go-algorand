package debuggerexamples

import (
	"github.com/algorand/go-algorand/data/basics"
	"github.com/algorand/go-algorand/data/bookkeeping"
	"github.com/algorand/go-algorand/data/transactions"
	"github.com/algorand/go-algorand/data/transactions/logic"
	"github.com/algorand/go-algorand/ledger/ledgercore"
)

type DebuggerAction int64

const (
	Continue DebuggerAction = iota
	StepInto
	StepOver
	StepOut
)

type DebuggerHooks interface {
	BeforeTxn(groupIndex int, prevTxibs []transactions.SignedTxnInBlock) error
	AfterTxn(groupIndex int, prevTxibs []transactions.SignedTxnInBlock) error
	BeforeInnerTxn(ep *logic.EvalParams) error
	AfterInnerTxn(ep *logic.EvalParams) error
	BeforeAppEval(gropuIndex int, ep *logic.EvalParams) error
	AfterAppEval(groupIndex int, ep *logic.EvalParams) error
	OnTealOp(state *logic.DebugState) (action DebuggerAction, err error)
}

type DebuggerOnChainState struct {
	AccountTotals   ledgercore.AccountTotals
	LastBlockHeader bookkeeping.BlockHeader
	Accounts        map[basics.Address]ledgercore.AccountData
	Assets          map[basics.AssetIndex]AssetParamsWithCreator
	AssetHoldings   map[ledgercore.AccountAsset]basics.AssetHolding
	Apps            map[basics.AppIndex]AppParamsWithCreator
	AppLocalStates  map[ledgercore.AccountApp]basics.AppLocalState
}

type DebuggerParams struct {
	InputTxns []transactions.SignedTxn
}

type DebuggerContext struct {
	ChainState *DebuggerOnChainState
	Params     *DebuggerParams
}

func EvalForDebugger(context *DebuggerContext, hooks DebuggerHooks) (ledgercore.StateDelta, []transactions.SignedTxnInBlock, error) {
	// To be implemented
	return ledgercore.StateDelta{}, []transactions.SignedTxnInBlock{}, nil
}

// ------------------------------------------------------------------------------------------------

type AssetParamsWithCreator struct {
	basics.AssetParams
	Creator basics.Address
}

type AppParamsWithCreator struct {
	basics.AppParams
	Creator basics.Address
}
