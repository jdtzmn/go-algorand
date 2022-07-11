package debuggerexamples

import (
	"testing"

	"github.com/algorand/go-algorand/data/transactions"
	"github.com/algorand/go-algorand/data/transactions/logic"
	"github.com/stretchr/testify/assert"
)

type stackValue struct {
	Uint  uint64
	Bytes []byte
}

type expectation struct {
	pc        int
	shouldAdd stackValue
}

type testingHooks struct {
	t              *testing.T
	asserts        map[int][]expectation // groupIndex -> expectations
	currentAsserts []expectation
}

func (ddh *testingHooks) BeforeTxn(groupIndex int, prevTxibs []transactions.SignedTxnInBlock) error {
	return nil
}

func (ddh *testingHooks) AfterTxn(groupIndex int, prevTxibs []transactions.SignedTxnInBlock) error {
	return nil
}

func (ddh *testingHooks) BeforeInnerTxn(ep *logic.EvalParams) error {
	return nil
}

func (ddh *testingHooks) AfterInnerTxn(ep *logic.EvalParams) error {
	return nil
}

func (ddh *testingHooks) BeforeAppEval(groupIndex int, ep *logic.EvalParams) error {
	ddh.currentAsserts = ddh.asserts[groupIndex]
	return nil
}

func (ddh *testingHooks) AfterAppEval(groupIndex int, ep *logic.EvalParams) error {
	return nil
}

func (ddh *testingHooks) OnTealOp(state *logic.DebugState) (action DebuggerAction, err error) {
	for _, expect := range ddh.currentAsserts {
		if state.PC == expect.pc {
			stackVal := state.Stack[len(state.Stack)-1]
			assert.Equal(ddh.t, stackVal, expect.shouldAdd)
		}
	}

	return Continue, nil
}

func testGroupTxn(t *testing.T, context *DebuggerContext, asserts map[int][]expectation) {
	th := &testingHooks{t: t, asserts: asserts}

	_, _, err := EvalForDebugger(context, th)
	if err != nil {
		// handle error
	}
}
