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

package ledger

import (
	"errors"
	"fmt"

	"github.com/algorand/go-algorand/config"
	"github.com/algorand/go-algorand/crypto"
	"github.com/algorand/go-algorand/data/basics"
	"github.com/algorand/go-algorand/data/bookkeeping"
	"github.com/algorand/go-algorand/data/transactions"
	"github.com/algorand/go-algorand/ledger/internal"
	"github.com/algorand/go-algorand/ledger/ledgercore"
)

type debuggerLedgerForEval interface {
	LatestBlockHdr() bookkeeping.BlockHeader
	GetAccount(basics.Address) (ledgercore.AccountData, bool)
	GetAsset(basics.AssetIndex) (basics.AssetParams, bool)
	GetAssetCreator(basics.AssetIndex) (basics.Address, bool)
	GetAssetHolding(basics.AssetIndex, basics.Address) (basics.AssetHolding, bool)
	GetApp(basics.AppIndex) (basics.AppParams, bool)
	GetAppCreator(basics.AppIndex) (basics.Address, bool)
	GetAppLocalState(basics.AppIndex, basics.Address) (basics.AppLocalState, bool)
	LatestTotals() ledgercore.AccountTotals
}

// Converter between indexerLedgerForEval and ledgerForEvaluator interfaces.
type debuggerLedgerConnector struct {
	l debuggerLedgerForEval
}

func (dlc debuggerLedgerConnector) checkRound(round basics.Round, caller string) error {
	latestHeader := dlc.l.LatestBlockHdr()
	if round == latestHeader.Round {
		return nil
	}
	return fmt.Errorf("%s() evaluator called this function for the wrong round %d, latest round is %d", caller, round, latestHeader.Round)
}

// BlockHdr is part of LedgerForEvaluator interface.
func (dlc debuggerLedgerConnector) BlockHdr(round basics.Round) (bookkeeping.BlockHeader, error) {
	if err := dlc.checkRound(round, "BlockHdr"); err != nil {
		return bookkeeping.BlockHeader{}, err
	}
	return dlc.l.LatestBlockHdr(), nil
}

// CheckDup is part of LedgerForEvaluator interface.
func (dlc debuggerLedgerConnector) CheckDup(config.ConsensusParams, basics.Round, basics.Round, basics.Round, transactions.Txid, ledgercore.Txlease) error {
	// This function is not used by evaluator.
	return errors.New("CheckDup() not implemented")
}

// LookupWithoutRewards is part of LedgerForEvaluator interface.
func (dlc debuggerLedgerConnector) LookupWithoutRewards(round basics.Round, address basics.Address) (ledgercore.AccountData, basics.Round, error) {
	if err := dlc.checkRound(round, "LookupWithoutRewards"); err != nil {
		return ledgercore.AccountData{}, 0, err
	}

	account, _ := dlc.l.GetAccount(address)

	return account, round, nil
}

func (dlc debuggerLedgerConnector) LookupApplication(round basics.Round, addr basics.Address, aidx basics.AppIndex) (ledgercore.AppResource, error) {
	if err := dlc.checkRound(round, "LookupApplication"); err != nil {
		return ledgercore.AppResource{}, err
	}

	appParams, hasParams := dlc.l.GetApp(aidx)
	appLocalState, hasLocalState := dlc.l.GetAppLocalState(aidx, addr)

	var response ledgercore.AppResource
	if hasParams {
		response.AppParams = &appParams
	}
	if hasLocalState {
		response.AppLocalState = &appLocalState
	}

	return response, nil
}

func (dlc debuggerLedgerConnector) LookupAsset(round basics.Round, addr basics.Address, aidx basics.AssetIndex) (ledgercore.AssetResource, error) {
	if err := dlc.checkRound(round, "LookupAsset"); err != nil {
		return ledgercore.AssetResource{}, err
	}

	assetParams, hasParams := dlc.l.GetAsset(aidx)
	assetHolding, hasHolding := dlc.l.GetAssetHolding(aidx, addr)

	var response ledgercore.AssetResource
	if hasParams {
		response.AssetParams = &assetParams
	}
	if hasHolding {
		response.AssetHolding = &assetHolding
	}

	return response, nil
}

// GetCreatorForRound is part of LedgerForEvaluator interface.
func (dlc debuggerLedgerConnector) GetCreatorForRound(_ basics.Round, cindex basics.CreatableIndex, ctype basics.CreatableType) (basics.Address, bool, error) {
	switch ctype {
	case basics.AssetCreatable:
		assetCreator, assetExists := dlc.l.GetAssetCreator(basics.AssetIndex(cindex))
		return assetCreator, assetExists, nil
	case basics.AppCreatable:
		appCreator, appExists := dlc.l.GetAppCreator(basics.AppIndex(cindex))
		return appCreator, appExists, nil
	default:
		return basics.Address{}, false, fmt.Errorf("unknown creatable type %v", ctype)
	}
}

// GenesisHash is part of LedgerForEvaluator interface.
func (dlc debuggerLedgerConnector) GenesisHash() crypto.Digest {
	return dlc.l.LatestBlockHdr().GenesisHash
}

// GenesisProto is part of LedgerForEvaluator interface.
func (dlc debuggerLedgerConnector) GenesisProto() config.ConsensusParams {
	return config.Consensus[dlc.l.LatestBlockHdr().CurrentProtocol]
}

// Totals is part of LedgerForEvaluator interface.
func (dlc debuggerLedgerConnector) LatestTotals() (basics.Round, ledgercore.AccountTotals, error) {
	return dlc.l.LatestBlockHdr().Round, dlc.l.LatestTotals(), nil
}

// CompactCertVoters is part of LedgerForEvaluator interface.
func (dlc debuggerLedgerConnector) CompactCertVoters(_ basics.Round) (*ledgercore.VotersForRound, error) {
	// This function is not used by evaluator.
	return nil, errors.New("CompactCertVoters() not implemented")
}

// EvalForDebugger ...
func EvalForDebugger(l debuggerLedgerForEval, debugger internal.TransactionGroupDebugger, stxns []transactions.SignedTxn) (ledgercore.StateDelta, []transactions.SignedTxnInBlock, error) {
	dlc := debuggerLedgerConnector{
		l: l,
	}

	nextBlock := bookkeeping.MakeBlock(l.LatestBlockHdr())
	nextBlockProto := config.Consensus[nextBlock.BlockHeader.CurrentProtocol]

	eval, err := internal.StartEvaluator(
		dlc, nextBlock.BlockHeader,
		internal.EvaluatorOptions{
			PaysetHint:  len(stxns),
			ProtoParams: &nextBlockProto,
			Generate:    true,
			Validate:    false,
		})
	if err != nil {
		return ledgercore.StateDelta{}, []transactions.SignedTxnInBlock{},
			fmt.Errorf("EvalForIndexer() err: %w", err)
	}

	group := make([]transactions.SignedTxnWithAD, len(stxns))
	for i, stxn := range stxns {
		group[i] = transactions.SignedTxnWithAD{
			SignedTxn: stxn,
		}
	}

	return eval.ProcessTransactionGroupForDebugger(group, debugger)
}
