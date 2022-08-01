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

package v2_test

import (
	"encoding/binary"
	"testing"

	v2 "github.com/algorand/go-algorand/daemon/algod/api/server/v2"
	"github.com/algorand/go-algorand/data/basics"
	"github.com/algorand/go-algorand/data/transactions"
	"github.com/algorand/go-algorand/data/transactions/logic"
	simulatortesting "github.com/algorand/go-algorand/ledger/testing/simulator"
	"github.com/algorand/go-algorand/protocol"
	"github.com/algorand/go-algorand/test/partitiontest"
	"github.com/stretchr/testify/require"
)

// ==============================
// > Simulation Test Helpers
// ==============================

func uint64ToBytes(num uint64) []byte {
	ibytes := make([]byte, 8)
	binary.BigEndian.PutUint64(ibytes, num)
	return ibytes
}

// ==============================
// > Simulation Tests
// ==============================

func TestPayTxn(t *testing.T) {
	partitiontest.PartitionTest(t)
	t.Parallel()

	l := simulatortesting.MakeSimulationTestLedger()
	s := v2.MakeSimulator(l)

	accounts := simulatortesting.MakeTestAccounts()
	sender := accounts[0].Address

	txgroup := []transactions.SignedTxn{
		{
			Txn: transactions.Transaction{
				Type:   protocol.PaymentTx,
				Header: simulatortesting.MakeBasicTxnHeader(sender),
				PaymentTxnFields: transactions.PaymentTxnFields{
					Receiver: sender,
					Amount:   basics.MicroAlgos{Raw: 0},
				},
			},
		},
	}

	result, err := s.SimulateSignedTxGroup(txgroup)
	require.NoError(t, err)
	require.Empty(t, result.FailureMessage)
}

func TestOverspendPayTxn(t *testing.T) {
	partitiontest.PartitionTest(t)
	t.Parallel()

	l := simulatortesting.MakeSimulationTestLedger()
	s := v2.MakeSimulator(l)

	accounts := simulatortesting.MakeTestAccounts()
	sender := accounts[0].Address
	balances := simulatortesting.MakeTestBalances()

	txgroup := []transactions.SignedTxn{
		{
			Txn: transactions.Transaction{
				Type:   protocol.PaymentTx,
				Header: simulatortesting.MakeBasicTxnHeader(sender),
				PaymentTxnFields: transactions.PaymentTxnFields{
					Receiver: sender,
					Amount:   basics.MicroAlgos{Raw: balances[sender] + 100}, // overspend
				},
			},
		},
	}

	result, err := s.SimulateSignedTxGroup(txgroup)
	require.NoError(t, err)
	require.Contains(t, *result.FailureMessage, "tried to spend {1000000100}")
}

func TestSimpleGroupTxn(t *testing.T) {
	partitiontest.PartitionTest(t)
	t.Parallel()

	l := simulatortesting.MakeSimulationTestLedger()
	s := v2.MakeSimulator(l)

	accounts := simulatortesting.MakeTestAccounts()
	sender1 := accounts[0].Address
	sender2 := accounts[1].Address

	// Send money back and forth
	txgroup := []transactions.SignedTxn{
		{
			Txn: transactions.Transaction{
				Type:   protocol.PaymentTx,
				Header: simulatortesting.MakeBasicTxnHeader(sender1),
				PaymentTxnFields: transactions.PaymentTxnFields{
					Receiver: sender2,
					Amount:   basics.MicroAlgos{Raw: 1000000},
				},
			},
		},
		{
			Txn: transactions.Transaction{
				Type:   protocol.PaymentTx,
				Header: simulatortesting.MakeBasicTxnHeader(sender2),
				PaymentTxnFields: transactions.PaymentTxnFields{
					Receiver: sender1,
					Amount:   basics.MicroAlgos{Raw: 0},
				},
			},
		},
	}

	// Should fail if there is no group parameter
	result, err := s.SimulateSignedTxGroup(txgroup)
	require.NoError(t, err)
	require.Contains(t, *result.FailureMessage, "had zero Group but was submitted in a group of 2")

	// Add group parameter
	simulatortesting.AttachGroupID(txgroup)

	// Check balances before transaction
	sender1Data, _, err := l.LookupWithoutRewards(l.Latest(), sender1)
	require.NoError(t, err)
	require.Equal(t, basics.MicroAlgos{Raw: 1000000000}, sender1Data.MicroAlgos)

	sender2Data, _, err := l.LookupWithoutRewards(l.Latest(), sender2)
	require.NoError(t, err)
	require.Equal(t, basics.MicroAlgos{Raw: 0}, sender2Data.MicroAlgos)

	// Should now pass
	result, err = s.SimulateSignedTxGroup(txgroup)
	require.NoError(t, err)
	require.Empty(t, result.FailureMessage)

	// Confirm balances have not changed
	sender1Data, _, err = l.LookupWithoutRewards(l.Latest(), sender1)
	require.NoError(t, err)
	require.Equal(t, basics.MicroAlgos{Raw: 1000000000}, sender1Data.MicroAlgos)

	sender2Data, _, err = l.LookupWithoutRewards(l.Latest(), sender2)
	require.NoError(t, err)
	require.Equal(t, basics.MicroAlgos{Raw: 0}, sender2Data.MicroAlgos)
}

const trivialAVMProgram = `#pragma version 2
int 1`

func TestSimpleAppCall(t *testing.T) {
	partitiontest.PartitionTest(t)
	t.Parallel()

	l := simulatortesting.MakeSimulationTestLedger()
	s := v2.MakeSimulator(l)

	accounts := simulatortesting.MakeTestAccounts()
	sender := accounts[0].Address

	// Compile AVM program
	ops, err := logic.AssembleString(trivialAVMProgram)
	require.NoError(t, err, ops.Errors)
	prog := ops.Program

	// Create program and call it
	futureAppID := 1
	txgroup := []transactions.SignedTxn{
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
		{
			Txn: transactions.Transaction{
				Type:   protocol.ApplicationCallTx,
				Header: simulatortesting.MakeBasicTxnHeader(sender),
				ApplicationCallTxnFields: transactions.ApplicationCallTxnFields{
					ApplicationID:     basics.AppIndex(futureAppID),
					ApprovalProgram:   prog,
					ClearStateProgram: prog,
				},
			},
		},
	}

	simulatortesting.AttachGroupID(txgroup)
	result, err := s.SimulateSignedTxGroup(txgroup)
	require.NoError(t, err)
	require.Empty(t, result.FailureMessage)
}

func TestSignatureCheck(t *testing.T) {
	partitiontest.PartitionTest(t)
	t.Parallel()

	l := simulatortesting.MakeSimulationTestLedger()
	s := v2.MakeSimulator(l)

	accounts := simulatortesting.MakeTestAccounts()
	sender := accounts[0].Address

	txgroup := []transactions.SignedTxn{
		{
			Txn: transactions.Transaction{
				Type:   protocol.PaymentTx,
				Header: simulatortesting.MakeBasicTxnHeader(sender),
				PaymentTxnFields: transactions.PaymentTxnFields{
					Receiver: sender,
					Amount:   basics.MicroAlgos{Raw: 0},
				},
			},
		},
	}

	// should error without a signature
	result, err := s.SimulateSignedTxGroup(txgroup)
	require.NoError(t, err)
	require.Empty(t, result.FailureMessage)
	require.Contains(t, *result.SignatureFailureMessage, "signedtxn has no sig")

	// add signature
	signatureSecrets, err := simulatortesting.SignatureSecretsFromPrivateKey(accounts[0].PrivateKey)
	require.NoError(t, err)
	txgroup[0] = txgroup[0].Txn.Sign(signatureSecrets)

	// should not error now that we have a signature
	result, err = s.SimulateSignedTxGroup(txgroup)
	require.NoError(t, err)
	require.Empty(t, result.FailureMessage)
	require.Empty(t, result.SignatureFailureMessage)
}

const accountBalanceCheckProgram = `#pragma version 4
  txn ApplicationID      // [appId]
	bz end                 // []
  int 1                  // [1]
  balance                // [bal[1]]
  itob                   // [itob(bal[1])]
  txn ApplicationArgs 0  // [itob(bal[1]), args[0]]
  ==                     // [itob(bal[1])=?=args[0]]
	assert
	b end
end:
  int 1                  // [1]
`

func TestBalanceChangesWithApp(t *testing.T) {
	// Send a payment transaction to a new account and confirm its balance within an app call
	partitiontest.PartitionTest(t)
	t.Parallel()

	l := simulatortesting.MakeSimulationTestLedger()
	s := v2.MakeSimulator(l)

	accounts := simulatortesting.MakeTestAccounts()
	sender := accounts[0].Address
	receiver := accounts[1].Address
	sendAmount := uint64(100000000)

	// Compile approval program
	ops, err := logic.AssembleString(accountBalanceCheckProgram)
	require.NoError(t, err, ops.Errors)
	approvalProg := ops.Program

	// Compile clear program
	ops, err = logic.AssembleString(trivialAVMProgram)
	require.NoError(t, err, ops.Errors)
	clearStateProg := ops.Program

	futureAppID := 1
	txgroup := []transactions.SignedTxn{
		// create app
		{
			Txn: transactions.Transaction{
				Type:   protocol.ApplicationCallTx,
				Header: simulatortesting.MakeBasicTxnHeader(sender),
				ApplicationCallTxnFields: transactions.ApplicationCallTxnFields{
					ApplicationID:     0,
					ApprovalProgram:   approvalProg,
					ClearStateProgram: clearStateProg,
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
		// check balance
		{
			Txn: transactions.Transaction{
				Type:   protocol.ApplicationCallTx,
				Header: simulatortesting.MakeBasicTxnHeader(sender),
				ApplicationCallTxnFields: transactions.ApplicationCallTxnFields{
					ApplicationID:     basics.AppIndex(futureAppID),
					ApprovalProgram:   approvalProg,
					ClearStateProgram: clearStateProg,
					Accounts:          []basics.Address{receiver},
					ApplicationArgs:   [][]byte{uint64ToBytes(0)},
				},
			},
		},
		// send payment
		{
			Txn: transactions.Transaction{
				Type:   protocol.PaymentTx,
				Header: simulatortesting.MakeBasicTxnHeader(sender),
				PaymentTxnFields: transactions.PaymentTxnFields{
					Receiver: receiver,
					Amount:   basics.MicroAlgos{Raw: sendAmount},
				},
			},
		},
		// check balance changed
		{
			Txn: transactions.Transaction{
				Type:   protocol.ApplicationCallTx,
				Header: simulatortesting.MakeBasicTxnHeader(sender),
				ApplicationCallTxnFields: transactions.ApplicationCallTxnFields{
					ApplicationID:     basics.AppIndex(futureAppID),
					ApprovalProgram:   approvalProg,
					ClearStateProgram: clearStateProg,
					Accounts:          []basics.Address{receiver},
					ApplicationArgs:   [][]byte{uint64ToBytes(sendAmount)},
				},
			},
		},
	}

	simulatortesting.AttachGroupID(txgroup)
	result, err := s.SimulateSignedTxGroup(txgroup)
	require.NoError(t, err)
	require.Empty(t, result.FailureMessage)
}
