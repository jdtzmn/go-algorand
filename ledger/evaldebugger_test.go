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

package ledger_test

import (
	"testing"

	"github.com/algorand/go-algorand/data/basics"
	"github.com/algorand/go-algorand/data/transactions"
	"github.com/algorand/go-algorand/data/transactions/logic"
	"github.com/algorand/go-algorand/ledger"
	simulatortesting "github.com/algorand/go-algorand/ledger/testing/simulator"
	"github.com/algorand/go-algorand/protocol"
	"github.com/algorand/go-algorand/test/partitiontest"
	"github.com/stretchr/testify/require"
)

// byte 0x068101 is `#pragma version 6; int 1;`
var innerTxnTestProgram string = `itxn_begin
int appl
itxn_field TypeEnum
int NoOp
itxn_field OnCompletion
byte 0x068101
dup
itxn_field ApprovalProgram
itxn_field ClearStateProgram
itxn_submit
int 1
`

type testDbgHook struct {
	log []string
}

func (d *testDbgHook) BeforeAppEval(state *logic.DebugState) error {
	d.log = append(d.log, "beforeAppEval")
	return nil
}

func (d *testDbgHook) BeforeTealOp(state *logic.DebugState) error {
	d.log = append(d.log, "beforeTealOp")
	return nil
}

func (d *testDbgHook) BeforeInnerTxn(ep *logic.EvalParams) error {
	d.log = append(d.log, "beforeInnerTxn")
	return nil
}

func (d *testDbgHook) AfterInnerTxn(ep *logic.EvalParams) error {
	d.log = append(d.log, "afterInnerTxn")
	return nil
}

func (d *testDbgHook) AfterTealOp(state *logic.DebugState) error {
	d.log = append(d.log, "afterTealOp")
	return nil
}

func (d *testDbgHook) AfterAppEval(state *logic.DebugState) error {
	d.log = append(d.log, "afterAppEval")
	return nil
}

func tealOpLogs(count int) []string {
	var log []string

	for i := 0; i < count; i++ {
		log = append(log, "beforeTealOp", "afterTealOp")
	}

	return log
}

func flatten(rows [][]string) []string {
	var out []string
	for _, row := range rows {
		out = append(out, row...)
	}
	return out
}

func TestEvalForDebuggerHooks(t *testing.T) {
	partitiontest.PartitionTest(t)
	t.Parallel()

	l := simulatortesting.MakeSimulationTestLedger()

	accounts := simulatortesting.MakeTestAccounts()
	sender := accounts[0].Address

	// Compile AVM program
	ops, err := logic.AssembleStringWithVersion(innerTxnTestProgram, uint64(6))
	require.NoError(t, err, ops.Errors)
	prog := ops.Program

	// Fund and call an inner transaction app
	futureAppID := 2
	txgroup := []transactions.SignedTxn{
		{
			Txn: transactions.Transaction{
				Type:   protocol.PaymentTx,
				Header: simulatortesting.MakeBasicTxnHeader(sender),
				PaymentTxnFields: transactions.PaymentTxnFields{
					Receiver: basics.AppIndex(futureAppID).Address(),
					Amount:   basics.MicroAlgos{Raw: 1000000},
				},
			},
		},
		{
			Txn: transactions.Transaction{
				Type:   protocol.ApplicationCallTx,
				Header: simulatortesting.MakeBasicTxnHeader(sender),
				ApplicationCallTxnFields: transactions.ApplicationCallTxnFields{
					ApplicationID:     0,
					ApprovalProgram:   prog,
					ClearStateProgram: prog,
					LocalStateSchema: basics.StateSchema{
						NumUint:      0,
						NumByteSlice: 0,
					},
					GlobalStateSchema: basics.StateSchema{
						NumUint:      0,
						NumByteSlice: 0,
					},
				},
			},
		},
	}

	simulatortesting.AttachGroupID(txgroup)

	testDbg := &testDbgHook{}
	_, _, err = ledger.EvalForDebugger(l, txgroup, testDbg)
	require.NoError(t, err)

	expectedLog := flatten([][]string{
		{"beforeAppEval"},
		tealOpLogs(9),
		{"beforeTealOp",
			"beforeInnerTxn",
			"beforeAppEval"},
		tealOpLogs(1),
		{"afterAppEval",
			"afterInnerTxn",
			"afterTealOp"},
		tealOpLogs(1),
		{"afterAppEval"},
	})
	require.Equal(t, expectedLog, testDbg.log)
}
