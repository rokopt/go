//lint:file-ignore U1001 Ignore all unused code, staticcheck doesn't understand testify/suite

package processors

import (
	"testing"

	"github.com/stellar/go/ingest"
	"github.com/stellar/go/services/horizon/internal/db2/history"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
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
}

func (s *ClaimableBalancesTransactionProcessorTestSuiteLedger) TearDownTest() {
	s.mockQ.AssertExpectations(s.T())
	s.mockTransactionBatchInsertBuilder.AssertExpectations(s.T())
	s.mockOperationBatchInsertBuilder.AssertExpectations(s.T())
}

func (s *ClaimableBalancesTransactionProcessorTestSuiteLedger) TestEmptyClaimableBalances() {
	// What is this expecting? Doesn't seem to assert anything...
	err := s.processor.Commit()
	s.Assert().NoError(err)
}

func (s *ParticipantsProcessorTestSuiteLedger) TestIngestClaimableBalancesInsertsOperationsAndTransactions() {
	s.mockQ.On("CreateHistoryClaimableBalances", mock.AnythingOfType("[]string"), maxBatchSize).
		Run(func(args mock.Arguments) {
			arg := args.Get(0).([]string)
			s.Assert().ElementsMatch(
				s.addresses,
				arg,
			)
		}).Return(s.addressToID, nil).Once()
	s.mockQ.On("NewTransactionParticipantsBatchInsertBuilder", maxBatchSize).
		Return(s.mockBatchInsertBuilder).Once()
	s.mockQ.On("NewOperationParticipantBatchInsertBuilder", maxBatchSize).
		Return(s.mockOperationsBatchInsertBuilder).Once()

	s.mockSuccessfulTransactionBatchAdds()
	s.mockSuccessfulOperationBatchAdds()

	s.mockBatchInsertBuilder.On("Exec").Return(nil).Once()
	s.mockOperationsBatchInsertBuilder.On("Exec").Return(nil).Once()

	for _, tx := range s.txs {
		err := s.processor.ProcessTransaction(tx)
		s.Assert().NoError(err)
	}
	err := s.processor.Commit()
	s.Assert().NoError(err)
}

// it extracts mappings from claimable balances -> operations
// it extracts mappings from claimable balances -> transactions
