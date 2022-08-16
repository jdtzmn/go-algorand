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
	Stack   map[uint]Mutation
	Scratch map[uint8]Mutation
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
	Effects *Effects
}

type OpCodeTraceElement struct {
	TraceElement
	OpCodeWithArgs string
	PC             uint64
	InnerIndex     uint64
}

type TxnTraceElement struct {
	TraceElement
	Txn       transactions.SignedTxnWithAD // ApplyData.EvalDelta.InnerTxns is not populated, in favor of InnerTxns below
	InnerTxns []TxnTraceElement
	LogicSig  []OpCodeTraceElement // included iff trace is requested
	Trace     []OpCodeTraceElement // included iff trace is requested
}

// ==============================
// > Transaction Results
// ==============================

type FailureLocator struct {
	In string // InnerTxns, LogicSig, or Trace
	At uint64 // index into InnerTxns, LogicSig, or Trace
}

type TxnResult struct {
	TxnTraceElement
	MissingSignature bool
	FailureMessage   string

	// If FailureMessage is not-empty:
	// - if FailedAt is empty, then this transaction failed
	// - if FailedAt is not empty, then this transaction failed at the given location
	//
	// FailedAt is used to avoid repeating the FailureMessage multiple times down the tree.
	FailedAt []FailureLocator
}

type SimulationResult struct {
	Version      uint64
	TxnGroups    [][]TxnResult // txngroups is a list so that supporting multiple in the future is not breaking
	WouldSucceed bool          // false iff failure message or missing sig or budget exceeded
	InitialState Effects
}
