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

// LookupLatest would implicitly use the latest round in the _underlying_
// Ledger, it would give wrong results if that ledger has moved forward. But it
// should never be called, as the REST API is the only code using this function,
// and the REST API will never get its grubby paws on a SimulationLedger.
func (l *Ledger) LookupLatest(addr basics.Address) (basics.AccountData, basics.Round, basics.MicroAlgos, error) {
	panic("unexpected call to LookupLatest")
}

// check checks signatures (if present) and well-formedness
func (sl Ledger) check(txgroup []transactions.SignedTxn) error {
	hdr, err := sl.BlockHdr(sl.start)
	if err != nil {
		return err
	}

	_, err = verify.TxnGroup(txgroup, hdr, nil)
	if err != nil && !errors.Is(err, verify.MissingSignatureError) {
		return err
	}

	return nil
}

// evaluate simulates a transaction group as if it were the only transaction
// added as a block to the current blockchain.
func (sl Ledger) evaluate(stxns []transactions.SignedTxn) (*ledgercore.ValidatedBlock, error) {
	prevBlockHdr, err := sl.BlockHdr(sl.start)
	if err != nil {
		return nil, err
	}
	nextBlock := bookkeeping.MakeBlock(prevBlockHdr)

	// sl has 'StartEvaluator' because *data.Ledger is embedded (and that, in
	// turn, embeds *ledger.Ledger)
	eval, err := sl.StartEvaluator(nextBlock.BlockHeader, 0, 0)
	if err != nil {
		return nil, err
	}

	group := transactions.WrapSignedTxnsWithAD(stxns)

	err = eval.TransactionGroup(group)
	if err != nil {
		return nil, err
	}

	// Finally, process any pending end-of-block state changes.
	vb, err := eval.GenerateBlock()
	if err != nil {
		return nil, err
	}

	return vb, nil
}

// Simulate simulates a transaction group using the simulator.
func (sl Ledger) Simulate(txgroup []transactions.SignedTxn) (*ledgercore.ValidatedBlock, error) {
	// check that the transaction is well-formed. Missing signatures are ignored.
	err := sl.check(txgroup)
	if err != nil {
		return nil, err
	}

	return sl.evaluate(txgroup)
}
