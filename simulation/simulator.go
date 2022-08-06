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

package simulation

import (
	"errors"

	"github.com/algorand/go-algorand/daemon/algod/api/server/v2/generated"
	"github.com/algorand/go-algorand/data"
	"github.com/algorand/go-algorand/data/basics"
	"github.com/algorand/go-algorand/data/bookkeeping"
	"github.com/algorand/go-algorand/data/transactions"
	"github.com/algorand/go-algorand/data/transactions/verify"
	"github.com/algorand/go-algorand/ledger/ledgercore"
)

// Ledger is wrapper around a real ledger that keeps track of what's going on
type Ledger struct {
	start basics.Round
	*data.Ledger
}

func NewLedger(underlying *data.Ledger) Ledger {
	return Ledger{underlying.Latest(), underlying}
}

// Latest returns the "locked in" start, regardless of whether the underlying
// Ledger has moved on.
func (sl Ledger) Latest() basics.Round {
	return sl.start
}

// checkWellFormed checks that the transaction is well-formed.
func (sl Ledger) checkWellFormed(txgroup []transactions.SignedTxn) error {
	hdr, err := sl.BlockHdr(sl.start)
	if err != nil {
		return err
	}

	_, err = verify.TxnGroup(txgroup, hdr, nil)
	if err != nil {
		return err
	}

	return nil
}

// evaluate creates a new block with txgroup given, and returns all the details
func (sl Ledger) evaluate(stxns []transactions.SignedTxn) (ledgercore.StateDelta, []transactions.SignedTxnInBlock, error) {
	prevBlockHdr, err := sl.BlockHdr(sl.start)
	if err != nil {
		return ledgercore.StateDelta{}, []transactions.SignedTxnInBlock{}, err
	}
	nextBlock := bookkeeping.MakeBlock(prevBlockHdr)

	// sl has 'StartEvaluator' because *data.Ledger is embedded (and that, in
	// turn, embeds *ledger.Ledger)
	eval, err := sl.StartEvaluator(nextBlock.BlockHeader, 0, 0)
	if err != nil {
		return ledgercore.StateDelta{}, []transactions.SignedTxnInBlock{}, err
	}

	group := transactions.WrapSignedTxnsWithAD(stxns)

	err = eval.TransactionGroup(group)
	if err != nil {
		return ledgercore.StateDelta{}, []transactions.SignedTxnInBlock{}, err
	}

	// Finally, process any pending end-of-block state changes.
	vb, err := eval.GenerateBlock()
	if err != nil {
		return ledgercore.StateDelta{}, []transactions.SignedTxnInBlock{}, err
	}

	return vb.Delta(), vb.Block().Payset, nil
}

// Simulate simulates a transaction group using the simulator.
func (sl Ledger) Simulate(txgroup []transactions.SignedTxn) (generated.SimulationResult, error) {
	var result generated.SimulationResult

	// check that the transaction is well-formed. Signatures are checked after evaluation
	err := sl.checkWellFormed(txgroup)
	if err != nil {
		if errors.Is(err, errors.New("signature missing")) { // ahh! not right!
			// note the error for reporting later
			msg := err.Error()
			result.SignatureFailureMessage = &msg
		} else {
			return result, err
		}
	}

	sl.evaluate(txgroup)

	return result, nil
}
