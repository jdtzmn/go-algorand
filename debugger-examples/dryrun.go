package debuggerexamples

import (
	"github.com/algorand/go-algorand/daemon/algod/api/server/v2/generated"
	"github.com/algorand/go-algorand/data/transactions"
	"github.com/algorand/go-algorand/data/transactions/logic"
)

type dryrunDebuggerHooks struct {
	trace []logic.DebugState
}

func (ddh *dryrunDebuggerHooks) BeforeTxn(groupIndex int, prevTxibs []transactions.SignedTxnInBlock) error {
	return nil
}

func (ddh *dryrunDebuggerHooks) AfterTxn(groupIndex int, prevTxibs []transactions.SignedTxnInBlock) error {
	return nil
}

func (ddh *dryrunDebuggerHooks) BeforeInnerTxn(ep *logic.EvalParams) error {
	return nil
}

func (ddh *dryrunDebuggerHooks) AfterInnerTxn(ep *logic.EvalParams) error {
	return nil
}

func (ddh *dryrunDebuggerHooks) BeforeAppEval(groupIndex int, ep *logic.EvalParams) error {
	return nil
}

func (ddh *dryrunDebuggerHooks) AfterAppEval(groupIndex int, ep *logic.EvalParams) error {
	return nil
}

func (ddh *dryrunDebuggerHooks) OnTealOp(state *logic.DebugState) (action DebuggerAction, err error) {
	// to be implemented
	ddh.trace = append(ddh.trace, *state)
	return Continue, nil
}

type DryrunRequest struct {
	// Txns is transactions to simulate
	Txns []transactions.SignedTxn `codec:"txns"` // not supposed to be serialized

	// Optional, useful for testing Application Call txns.
	Accounts []generated.Account `codec:"accounts"`

	Apps []generated.Application `codec:"apps"`

	// ProtocolVersion specifies a specific version string to operate under, otherwise whatever the current protocol of the network this algod is running in.
	ProtocolVersion string `codec:"protocol-version"`

	// Round is available to some TEAL scripts. Defaults to the current round on the network this algod is attached to.
	Round uint64 `codec:"round"`

	// LatestTimestamp is available to some TEAL scripts. Defaults to the latest confirmed timestamp this algod is attached to.
	LatestTimestamp int64 `codec:"latest-timestamp"`

	Sources []generated.DryrunSource `codec:"sources"`
}

func generateDryrunContext(dr *DryrunRequest) *DebuggerContext {
	// to be implemented
	return &DebuggerContext{}
}

func generateResponseFromTrace(trace []logic.DebugState) *generated.DryrunResponse {
	// to be implemented
	return &generated.DryrunResponse{}
}

func doDryrunRequest(dr *DryrunRequest, response *generated.DryrunResponse) {
	ddh := &dryrunDebuggerHooks{}
	_, _, err := EvalForDebugger(generateDryrunContext(dr), ddh)
	if err != nil {
		// handle error
	}

	response = generateResponseFromTrace(ddh.trace)
}
