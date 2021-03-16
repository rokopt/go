//lint:file-ignore U1001 Ignore all unused code, staticcheck doesn't understand testify/suite

package processors

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/stellar/go/ingest"
	"github.com/stellar/go/services/horizon/internal/db2/history"
	"github.com/stellar/go/services/horizon/internal/toid"
	"github.com/stellar/go/xdr"
)

type ClaimableBalancesTransactionProcessorTestSuiteLedger struct {
	suite.Suite
	processor                         *ClaimableBalancesTransactionProcessor
	mockQ                             *history.MockQHistoryClaimableBalances
	mockTransactionBatchInsertBuilder *history.MockTransactionClaimableBalanceBatchInsertBuilder
	mockOperationBatchInsertBuilder   *history.MockOperationClaimableBalanceBatchInsertBuilder

	sequence     uint32
	ids          []string
	toInternalID map[string]int64
	txs          []ingest.LedgerTransaction
}

func TestClaimableBalancesTransactionProcessorTestSuiteLedger(t *testing.T) {
	suite.Run(t, new(ClaimableBalancesTransactionProcessorTestSuiteLedger))
}

func (s *ClaimableBalancesTransactionProcessorTestSuiteLedger) SetupTest() {
	s.mockQ = &history.MockQHistoryClaimableBalances{}
	s.mockTransactionBatchInsertBuilder = &history.MockTransactionClaimableBalanceBatchInsertBuilder{}
	s.mockOperationBatchInsertBuilder = &history.MockOperationClaimableBalanceBatchInsertBuilder{}
	s.sequence = 20

	s.processor = NewClaimableBalancesTransactionProcessor(
		s.mockQ,
		s.sequence,
	)
}

func (s *ClaimableBalancesTransactionProcessorTestSuiteLedger) TearDownTest() {
	s.mockQ.AssertExpectations(s.T())
	s.mockTransactionBatchInsertBuilder.AssertExpectations(s.T())
	s.mockOperationBatchInsertBuilder.AssertExpectations(s.T())
}

func (s *ClaimableBalancesTransactionProcessorTestSuiteLedger) mockTransactionBatchAdd(transactionID, cbID int64, err error) {
	s.mockTransactionBatchInsertBuilder.On(
		"Add", transactionID, cbID,
	).Return(err).Once()
}

func (s *ClaimableBalancesTransactionProcessorTestSuiteLedger) mockOperationBatchAdd(operationID, cbID int64, err error) {
	s.mockOperationBatchInsertBuilder.On(
		"Add", operationID, cbID,
	).Return(err).Once()
}

func (s *ClaimableBalancesTransactionProcessorTestSuiteLedger) TestEmptyClaimableBalances() {
	// What is this expecting? Doesn't seem to assert anything meaningful...
	err := s.processor.Commit()
	s.Assert().NoError(err)
}

func (s *ClaimableBalancesTransactionProcessorTestSuiteLedger) TestIngestClaimableBalancesInsertsTransactions() {
	// Setup a q
	s.mockQ.On("CreateHistoryClaimableBalances", mock.AnythingOfType("[]string"), maxBatchSize).
		Run(func(args mock.Arguments) {
			arg := args.Get(0).([]string)
			s.Assert().ElementsMatch(
				"TODO: Fill this in?",
				arg,
			)
		}).Return(map[xdr.ClaimableBalanceId]int64{
		// TODO: Fill this in
	}, nil).Once()

	// Setup the transaction
	txn := createTransaction(true, 1)
	txn.Index = 2
	txn.Envelope.Operations()[0].Body = xdr.OperationBody{
		Type:                     xdr.OperationTypeCreateClaimableBalance,
		CreateClaimableBalanceOp: &xdr.CreateClaimableBalanceOp{},
	}
	txnID := toid.New(int32(s.sequence), 1, 0).ToInt64()
	opID := int64(0)      // TODO: Figure this out
	balanceID := int64(0) // TODO: Figure this out

	// Prepare to process transactions successfully
	s.mockQ.On("NewTransactionClaimableBalanceBatchInsertBuilder", maxBatchSize).
		Return(s.mockTransactionBatchInsertBuilder).Once()
	s.mockTransactionBatchAdd(txnID, balanceID, nil)
	s.mockTransactionBatchInsertBuilder.On("Exec").Return(nil).Once()

	// Prepare to process operations successfully
	opBatch := &history.MockOperationClaimableBalanceBatchInsertBuilder{}
	s.mockQ.On("NewOperationClaimableBalanceBatchInsertBuilder", maxBatchSize).
		Return(opBatch).Once()
	s.mockOperationBatchAdd(opID, balanceID, nil)
	s.mockOperationBatchInsertBuilder.On("Exec").Return(nil).Once()

	// Process the transaction
	s.Assert().NoError(s.processor.ProcessTransaction(txn))
	s.Assert().NoError(s.processor.Commit())
}

// it extracts mappings from claimable balances -> operations
// it extracts mappings from claimable balances -> transactions
