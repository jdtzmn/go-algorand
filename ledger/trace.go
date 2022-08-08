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
	"github.com/algorand/go-algorand/data/basics"
	"github.com/algorand/go-algorand/data/transactions"
)

// ==============================
// > Mutation
// ==============================

type MutationType uint64

const (
	Create MutationType = iota
	Update
	Delete
)

// this should be made stricter
type MutationValue = interface{}

// this would greatly benefit from golang generics
type Mutation struct {
	Type  MutationType
	Value *MutationValue
}

// ==============================
// > Effects
// ==============================

type AccountMutations struct {
	Balance       *Mutation
	AssetsCreated *Mutation
	// ...
}

type AssetMutations struct {
	Creator *Mutation
	Total   *Mutation
	// ...
}

type AppMutations struct {
	Creator     *Mutation
	GlobalState map[string]Mutation
	Boxes       map[string]Mutation
	// ...
}

type OptedInAssetMutations struct {
	Balance *Mutation
	Frozen  *Mutation
	// ...
}

type OptedInAppMutations struct {
	LocalState map[string]Mutation
	// ...
}

type EvalContextEffects struct {
	Stack   map[uint8]Mutation
	Scratch map[string]Mutation
}

type Effects struct {
	Accounts      map[basics.Address]AccountMutations
	Assets        map[basics.AssetIndex]AssetMutations
	Apps          map[basics.AppIndex]AppMutations
	OptedInAssets map[basics.Address]map[basics.AssetIndex]OptedInAssetMutations
	OptedInApps   map[basics.Address]map[basics.AppIndex]OptedInAppMutations

	// opcode only
	EvalContext EvalContextEffects
}

// ==============================
// > Trace
// ==============================

// Trace contains a list of transactions or opcodes

type TraceElementType uint64

const (
	Txn TraceElementType = iota
	OpCode
)

type TraceElement struct {
	// common fields
	Type    TraceElementType
	Events  []TraceElement
	Effects *Effects

	// txn only

	// This is a "transaction path": e.g. [0, 0, 1] means the second inner txn of the first inner txn of the first txn.
	// You can use this transaction path to find the txn data in the `TxnResults` list.
	GroupIndex *[]uint64

	// opcode only
	OpCodeWithArgs string
	PC             uint64
}

type Trace = []TraceElement

// ==============================
// > Initial State
// ==============================

type InitialState struct {
	Effects
	TxGroup []transactions.SignedTxn // raw txn group submitted
}

// ==============================
// > Transaction Results
// ==============================

type TxnResult struct {
	ApplyData        transactions.ApplyData
	MissingSignature bool
	FailureMessage   string // Question: is this still needed if we have the trace?
}

type TxnGroup struct {
	Trace      Trace
	TxnResults []TxnResult
}

type SimulationResult struct {
	Version           uint64
	TxnGroups         []TxnGroup // txngroups is a list so that supporting multiple in the future is not breaking
	MissingSignatures bool
	FailureMessage    string
	InitialState      InitialState
}
