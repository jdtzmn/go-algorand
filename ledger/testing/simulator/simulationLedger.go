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

package testing

import (
	"crypto/ed25519"
	"errors"
	"fmt"

	"github.com/algorand/go-algorand/config"
	"github.com/algorand/go-algorand/crypto"
	"github.com/algorand/go-algorand/crypto/passphrase"
	"github.com/algorand/go-algorand/data/basics"
	"github.com/algorand/go-algorand/data/bookkeeping"
	"github.com/algorand/go-algorand/data/transactions"
	"github.com/algorand/go-algorand/ledger/ledgercore"
	ledgertesting "github.com/algorand/go-algorand/ledger/testing"
	"github.com/algorand/go-algorand/libgoal"
	"github.com/algorand/go-algorand/protocol"
)

// ==============================
// > Simulation Test Ledger
// ==============================

type simulationTestLedger struct {
	*ledgertesting.Ledger

	hdr bookkeeping.BlockHeader
}

func (sl *simulationTestLedger) Latest() basics.Round {
	return sl.hdr.Round
}

func (sl *simulationTestLedger) BlockHdr(rnd basics.Round) (blk bookkeeping.BlockHeader, err error) {
	if rnd != sl.Latest() {
		err = fmt.Errorf(
			"BlockHdr() evaluator called this function for the wrong round %d, "+
				"latest round is %d",
			rnd, sl.Latest())
		return
	}

	return sl.hdr, nil
}

// override the test ledger's BlockHdrCached method to return the same header
func (sl *simulationTestLedger) BlockHdrCached(rnd basics.Round) (bookkeeping.BlockHeader, error) {
	return sl.BlockHdr(rnd)
}

func (sl *simulationTestLedger) CheckDup(currentProto config.ConsensusParams, current basics.Round, firstValid basics.Round, lastValid basics.Round, txid transactions.Txid, txl ledgercore.Txlease) error {
	// Never throw an error during these tests since it's a simulation ledger.
	// In production, the actual ledger method is used.
	return nil
}

func (sl *simulationTestLedger) CompactCertVoters(rnd basics.Round) (*ledgercore.VotersForRound, error) {
	panic("CompactCertVoters() should not be called in a simulation ledger")
}

func (sl *simulationTestLedger) GenesisHash() crypto.Digest {
	return sl.hdr.GenesisHash
}

func (sl *simulationTestLedger) GenesisProto() config.ConsensusParams {
	return config.Consensus[sl.hdr.CurrentProtocol]
}

func (sl *simulationTestLedger) GetCreatorForRound(round basics.Round, cidx basics.CreatableIndex, ctype basics.CreatableType) (creator basics.Address, ok bool, err error) {
	if round != sl.Latest() {
		err = fmt.Errorf(
			"GetCreatorForRound() evaluator called this function for the wrong round %d, "+
				"latest round is %d",
			round, sl.Latest())
		return
	}

	return sl.GetCreator(cidx, ctype)
}

func (sl *simulationTestLedger) LookupAsset(rnd basics.Round, addr basics.Address, aidx basics.AssetIndex) (ledgercore.AssetResource, error) {
	assetParams, addr, err := sl.AssetParams(aidx)
	if err != nil {
		return ledgercore.AssetResource{}, err
	}

	assetHolding, err := sl.AssetHolding(addr, aidx)
	if err != nil {
		return ledgercore.AssetResource{}, err
	}

	return ledgercore.AssetResource{
		AssetParams:  &assetParams,
		AssetHolding: &assetHolding,
	}, nil
}

func (sl *simulationTestLedger) LookupWithoutRewards(rnd basics.Round, addr basics.Address) (ledgercore.AccountData, basics.Round, error) {
	if rnd != sl.Latest() {
		return ledgercore.AccountData{}, basics.Round(0), fmt.Errorf(
			"LookupWithoutRewards() evaluator called this function for the wrong round %d, "+
				"latest round is %d",
			rnd, sl.Latest())
	}

	acctData, err := sl.AccountData(addr)
	if err != nil {
		return ledgercore.AccountData{}, basics.Round(0), err
	}

	return acctData, sl.Latest(), nil
}

// ============================================
// > Simulation Test Ledger Helper Methods
// ============================================

func makeTestClient() libgoal.Client {
	c, err := libgoal.MakeClientFromConfig(libgoal.ClientConfig{
		AlgodDataDir: "NO_DIR",
	}, libgoal.DynamicClient)
	if err != nil {
		panic(err)
	}

	return c
}

func SignatureSecretsFromPrivateKey(privateKey ed25519.PrivateKey) (*crypto.SignatureSecrets, error) {
	var sk crypto.PrivateKey
	copy(sk[:], privateKey)
	return crypto.SecretKeyToSignatureSecrets(sk)
}

func makeSpecialAccounts() (sink, rewards basics.Address) {
	// irrelevant, but deterministic
	sink, err := basics.UnmarshalChecksumAddress("YTPRLJ2KK2JRFSZZNAF57F3K5Y2KCG36FZ5OSYLW776JJGAUW5JXJBBD7Q")
	if err != nil {
		panic(err)
	}
	rewards, err = basics.UnmarshalChecksumAddress("242H5OXHUEBYCGGWB3CQ6AZAMQB5TMCWJGHCGQOZPEIVQJKOO7NZXUXDQA")
	if err != nil {
		panic(err)
	}
	return
}

func makeTestBlockHeader() bookkeeping.BlockHeader {
	// arbitrary genesis information
	genesisID := "simulation-test-v1"
	genesisHash, err := crypto.DigestFromString("3QF7SU53VLAQV6YIWENHUVANS4OFG5PHCTXPPX4EH7FEI3WIMJOQ")
	if err != nil {
		panic(err)
	}

	feeSink, rewardsPool := makeSpecialAccounts()

	// convert test balances to AccountData balances
	testBalances := MakeTestBalances()
	acctDataBalances := make(map[basics.Address]basics.AccountData)
	for addr, balance := range testBalances {
		acctDataBalances[addr] = basics.AccountData{
			MicroAlgos: basics.MicroAlgos{Raw: balance},
		}
	}

	genesisBalances := bookkeeping.MakeGenesisBalances(acctDataBalances, feeSink, rewardsPool)
	genesisBlock, err := bookkeeping.MakeGenesisBlock(protocol.ConsensusCurrentVersion, genesisBalances, genesisID, genesisHash)
	if err != nil {
		panic(err)
	}

	return genesisBlock.BlockHeader
}

type account struct {
	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey
	Address    basics.Address
}

func accountFromMnemonic(mnemonic string) (account, error) {
	key, err := passphrase.MnemonicToKey(mnemonic)
	if err != nil {
		return account{}, err
	}

	decoded := ed25519.NewKeyFromSeed(key)

	pk := decoded.Public().(ed25519.PublicKey)
	sk := decoded

	// Convert the public key to an address
	var addr basics.Address
	n := copy(addr[:], pk)
	if n != ed25519.PublicKeySize {
		return account{}, errors.New("generated public key is the wrong size")
	}

	return account{
		PublicKey:  pk,
		PrivateKey: sk,
		Address:    addr,
	}, nil
}

func MakeTestAccounts() []account {
	// funded
	account1, err := accountFromMnemonic("enforce voyage media inform embody borrow truck flat brave goose edit glide poet describe oxygen teach home choice motion engine wolf iron bachelor ability view")
	if err != nil {
		panic(err)
	}

	// unfunded
	account2, err := accountFromMnemonic("husband around three crystal canvas arrive beach dumb pill sock inflict drink salmon modify gas monkey jungle chronic senior flavor ability witness resist abandon just")
	if err != nil {
		panic(err)
	}

	return []account{account1, account2}
}

func MakeTestBalances() map[basics.Address]uint64 {
	accounts := MakeTestAccounts()

	return map[basics.Address]uint64{
		accounts[0].Address: 1000000000,
	}
}

func MakeSimulationTestLedger() *simulationTestLedger {
	hdr := makeTestBlockHeader()
	balances := MakeTestBalances()
	balances[hdr.RewardsPool] = 1000000 // pool is always 1000000
	round := uint64(1)
	logicLedger := ledgertesting.MakeLedgerForRound(balances, round)
	hdr.Round = basics.Round(round)
	l := simulationTestLedger{logicLedger, hdr}
	return &l
}

// ============================================
// > Simulation Test Ledger Namespace Methods
// ============================================

func MakeBasicTxnHeader(sender basics.Address) transactions.Header {
	hdr := makeTestBlockHeader()

	return transactions.Header{
		Fee:         basics.MicroAlgos{Raw: 1000},
		FirstValid:  basics.Round(1),
		GenesisID:   hdr.GenesisID,
		GenesisHash: hdr.GenesisHash,
		LastValid:   basics.Round(1001),
		Note:        []byte{240, 134, 38, 55, 197, 14, 142, 132},
		Sender:      sender,
	}
}

// Attach group ID to a transaction group. Mutates the group directly.
func AttachGroupID(txgroup []transactions.SignedTxn) error {
	txnArray := make([]transactions.Transaction, len(txgroup))
	for i, txn := range txgroup {
		txnArray[i] = txn.Txn
	}

	client := makeTestClient()
	groupID, err := client.GroupID(txnArray)
	if err != nil {
		return err
	}

	for i := range txgroup {
		txgroup[i].Txn.Header.Group = groupID
	}

	return nil
}
