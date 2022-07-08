// Copyright (C) 2019-2022 Algorand, Inc.
// This file is part of go-algorand
//
// go-algorand is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// go-algorand is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with go-algorand.  If not, see <https://www.gnu.org/licenses/>.

package internal

import (
	"fmt"

	"github.com/algorand/go-algorand/data/basics"
	"github.com/algorand/go-algorand/data/bookkeeping"
	"github.com/algorand/go-algorand/data/transactions"
	"github.com/algorand/go-algorand/data/transactions/logic"
	"github.com/algorand/go-algorand/ledger/ledgercore"
)

type TransactionGroupDebugger interface {
	AboutToEvalTransaction(groupIndex int, prevTxibs []transactions.SignedTxnInBlock)

	GetLogicDebugger() logic.DebuggerHook
}

// ProcessTransactionGroupForDebugger ..
func (eval *BlockEvaluator) ProcessTransactionGroupForDebugger(group []transactions.SignedTxnWithAD, debugger TransactionGroupDebugger) (ledgercore.StateDelta, []transactions.SignedTxnInBlock, error) {
	err := eval.transactionGroup(group, debugger)
	if err != nil {
		return ledgercore.StateDelta{}, []transactions.SignedTxnInBlock{},
			fmt.Errorf("ProcessTransactionGroupForDebugger() err: %w", err)
	}

	// Finally, process any pending end-of-block state changes.
	err = eval.endOfBlock()
	if err != nil {
		return ledgercore.StateDelta{}, []transactions.SignedTxnInBlock{},
			fmt.Errorf("ProcessTransactionGroupForDebugger() err: %w", err)
	}

	return eval.state.deltas(), eval.block.Payset, nil
}

type AssetParamsWithCreator struct {
	basics.AssetParams
	Creator basics.Address
}

type AppParamsWithCreator struct {
	basics.AppParams
	Creator basics.Address
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

// EvalLevel represents a "level" being evaluated. Each level is a group of transactions
// and its corresponding debugger trace, if one exists.
type EvalLevel struct {
	TxnGroup      []transactions.SignedTxn
	TxnGroupIndex int
	TxnGroupADs   []transactions.ApplyData

	ChildLevels EvalLevels
	Parent      *EvalLevel

	Trace []logic.DebugState
}

type EvalLevels []EvalLevel

type EvalDebugger struct {
	ctx *DebuggerContext

	currentLevel *EvalLevel
}

func (ed EvalDebugger) recordToTrace(state *logic.DebugState) {
	ed.currentLevel.Trace = append(ed.currentLevel.Trace, *state)
}

type logicDebuggerHook struct {
	ed *EvalDebugger
}

func (ldh logicDebuggerHook) Register(state *logic.DebugState) error {
	ldh.ed.recordToTrace(state)
	return nil
}

func (ldh logicDebuggerHook) Update(state *logic.DebugState) error {
	// called before every TEAL op in the program
	ldh.ed.recordToTrace(state)
	return nil
}

func (ldh logicDebuggerHook) Complete(state *logic.DebugState) error {
	return nil
}

func (ldh logicDebuggerHook) EnterInners(ep *logic.EvalParams) error {
	txnGroup := make([]transactions.SignedTxn, len(ep.TxnGroup))
	for i, txn := range ep.TxnGroup {
		txnGroup[i] = txn.SignedTxn
	}

	// descend a level deeper
	childLevel := EvalLevel{
		TxnGroup: txnGroup,
	}
	ldh.ed.currentLevel.ChildLevels = append(ldh.ed.currentLevel.ChildLevels, childLevel)
	ldh.ed.currentLevel = &childLevel

	return nil
}

func (ldh logicDebuggerHook) InnerTxn(groupIndex int, ep *logic.EvalParams) error {
	// include previous apply data
	if groupIndex > 0 {
		prevTxnIndex := groupIndex - 1
		ldh.ed.currentLevel.TxnGroupADs = append(ldh.ed.currentLevel.TxnGroupADs, ep.TxnGroup[prevTxnIndex].ApplyData)
	}

	return nil
}

func (ldh logicDebuggerHook) LeaveInners(ep *logic.EvalParams) error {
	// ascend to parent level
	ldh.ed.currentLevel = ldh.ed.currentLevel.Parent
	return nil
}
