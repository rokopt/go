//lint:file-ignore U1001 Ignore all unused code, staticcheck doesn't understand testify/suite

package processors

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

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

	sequence uint32
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

func (s *ClaimableBalancesTransactionProcessorTestSuiteLedger) mockTransactionBatchAdd(transactionID, internalID int64, err error) {
	s.mockTransactionBatchInsertBuilder.On("Add", transactionID, internalID).Return(err).Once()
}

func (s *ClaimableBalancesTransactionProcessorTestSuiteLedger) mockOperationBatchAdd(operationID, internalID int64, err error) {
	s.mockOperationBatchInsertBuilder.On("Add", operationID, internalID).Return(err).Once()
}

func (s *ClaimableBalancesTransactionProcessorTestSuiteLedger) TestEmptyClaimableBalances() {
	// What is this expecting? Doesn't seem to assert anything meaningful...
	err := s.processor.Commit()
	s.Assert().NoError(err)
}

func (s *ClaimableBalancesTransactionProcessorTestSuiteLedger) testOperationInserts(balanceID xdr.ClaimableBalanceId, body xdr.OperationBody, change *xdr.LedgerEntryChange) {
	// Setup the transaction
	internalID := int64(1234)
	txn := createTransaction(true, 1)
	txn.Envelope.Operations()[0].Body = body
	if change != nil {
		txn.Meta.V = 2
		txn.Meta.V2 = &xdr.TransactionMetaV2{
			Operations: []xdr.OperationMeta{
				{Changes: xdr.LedgerEntryChanges{*change}},
			},
		}
	}
	txnID := toid.New(int32(s.sequence), int32(txn.Index), 0).ToInt64()
	opID := (&transactionOperationWrapper{
		index:          uint32(0),
		transaction:    txn,
		operation:      txn.Envelope.Operations()[0],
		ledgerSequence: s.sequence,
	}).ID()

	b64ID, _ := xdr.MarshalBase64(balanceID)

	// Setup a q
	s.mockQ.On("CreateHistoryClaimableBalances", mock.AnythingOfType("[]xdr.ClaimableBalanceId"), maxBatchSize).
		Run(func(args mock.Arguments) {
			arg := args.Get(0).([]xdr.ClaimableBalanceId)
			s.Assert().ElementsMatch(
				[]xdr.ClaimableBalanceId{
					balanceID,
				},
				arg,
			)
		}).Return(map[string]int64{
		b64ID: internalID,
	}, nil).Once()

	// Prepare to process transactions successfully
	s.mockQ.On("NewTransactionClaimableBalanceBatchInsertBuilder", maxBatchSize).
		Return(s.mockTransactionBatchInsertBuilder).Once()
	s.mockTransactionBatchAdd(txnID, internalID, nil)
	s.mockTransactionBatchInsertBuilder.On("Exec").Return(nil).Once()

	// Prepare to process operations successfully
	s.mockQ.On("NewOperationClaimableBalanceBatchInsertBuilder", maxBatchSize).
		Return(s.mockOperationBatchInsertBuilder).Once()
	s.mockOperationBatchAdd(opID, internalID, nil)
	s.mockOperationBatchInsertBuilder.On("Exec").Return(nil).Once()

	// Process the transaction
	err := s.processor.ProcessTransaction(txn)
	s.Assert().NoError(err)
	err = s.processor.Commit()
	s.Assert().NoError(err)
}

func (s *ClaimableBalancesTransactionProcessorTestSuiteLedger) TestIngestClaimableBalancesInsertsClaimClaimableBalance() {
	balanceID := xdr.ClaimableBalanceId{
		Type: xdr.ClaimableBalanceIdTypeClaimableBalanceIdTypeV0,
		V0:   &xdr.Hash{1, 2, 3},
	}
	s.testOperationInserts(balanceID, xdr.OperationBody{
		Type: xdr.OperationTypeClaimClaimableBalance,
		ClaimClaimableBalanceOp: &xdr.ClaimClaimableBalanceOp{
			BalanceId: balanceID,
		},
	}, nil)
}

func (s *ClaimableBalancesTransactionProcessorTestSuiteLedger) TestIngestClaimableBalancesInsertsClawbackClaimableBalance() {
	balanceID := xdr.ClaimableBalanceId{
		Type: xdr.ClaimableBalanceIdTypeClaimableBalanceIdTypeV0,
		V0:   &xdr.Hash{1, 2, 3},
	}
	s.testOperationInserts(balanceID, xdr.OperationBody{
		Type: xdr.OperationTypeClawbackClaimableBalance,
		ClawbackClaimableBalanceOp: &xdr.ClawbackClaimableBalanceOp{
			BalanceId: balanceID,
		},
	}, nil)
}

func (s *ClaimableBalancesTransactionProcessorTestSuiteLedger) TestIngestClaimableBalancesInsertsCreateClaimableBalance() {
	balanceID := xdr.ClaimableBalanceId{
		Type: xdr.ClaimableBalanceIdTypeClaimableBalanceIdTypeV0,
		V0:   &xdr.Hash{1, 2, 3},
	}
	s.testOperationInserts(balanceID, xdr.OperationBody{
		Type:                     xdr.OperationTypeCreateClaimableBalance,
		CreateClaimableBalanceOp: &xdr.CreateClaimableBalanceOp{},
	}, &xdr.LedgerEntryChange{
		Type: xdr.LedgerEntryChangeTypeLedgerEntryCreated,
		Created: &xdr.LedgerEntry{
			Data: xdr.LedgerEntryData{
				Type: xdr.LedgerEntryTypeClaimableBalance,
				ClaimableBalance: &xdr.ClaimableBalanceEntry{
					BalanceId: balanceID,
				},
			},
		},
	})
}

func (s *ClaimableBalancesTransactionProcessorTestSuiteLedger) TestIngestClaimableBalancesInsertsTransactionsFromMeta() {
	balanceID := xdr.ClaimableBalanceId{
		Type: xdr.ClaimableBalanceIdTypeClaimableBalanceIdTypeV0,
		V0:   &xdr.Hash{1, 2, 3},
	}

	// Setup the transaction
	internalID := int64(1234)
	txn := createTransaction(true, 1)
	// Whatever, just an operation with no claimable balance so we can be sure it is coming from the
	txn.Envelope.Operations()[0].Body = xdr.OperationBody{
		Type: xdr.OperationTypeAllowTrust,
	}
	// Put the transaction id we want to insert in the meta
	txn.Meta.V = 0
	txn.Meta.Operations = &[]xdr.OperationMeta{
		{Changes: xdr.LedgerEntryChanges{
			{
				Type: xdr.LedgerEntryChangeTypeLedgerEntryCreated,
				Created: &xdr.LedgerEntry{
					Data: xdr.LedgerEntryData{
						Type: xdr.LedgerEntryTypeClaimableBalance,
						ClaimableBalance: &xdr.ClaimableBalanceEntry{
							BalanceId: balanceID,
						},
					},
				},
			}},
		},
	}
	txnID := toid.New(int32(s.sequence), int32(txn.Index), 0).ToInt64()

	// Setup a q
	b64ID, _ := xdr.MarshalBase64(balanceID)
	s.mockQ.On("CreateHistoryClaimableBalances", mock.AnythingOfType("[]xdr.ClaimableBalanceId"), maxBatchSize).
		Run(func(args mock.Arguments) {
			arg := args.Get(0).([]xdr.ClaimableBalanceId)
			s.Assert().ElementsMatch(
				[]xdr.ClaimableBalanceId{
					balanceID,
				},
				arg,
			)
		}).Return(map[string]int64{
		b64ID: internalID,
	}, nil).Once()

	// Prepare to process transactions successfully
	s.mockQ.On("NewTransactionClaimableBalanceBatchInsertBuilder", maxBatchSize).
		Return(s.mockTransactionBatchInsertBuilder).Once()
	s.mockTransactionBatchAdd(txnID, internalID, nil)
	s.mockTransactionBatchInsertBuilder.On("Exec").Return(nil).Once()

	s.mockQ.On("NewOperationClaimableBalanceBatchInsertBuilder", maxBatchSize).
		Return(s.mockOperationBatchInsertBuilder).Once()
	s.mockOperationBatchInsertBuilder.On("Exec").Return(nil).Once()

	// Should not insert any operations.

	// Process the transaction
	err := s.processor.ProcessTransaction(txn)
	s.Assert().NoError(err)
	err = s.processor.Commit()
	s.Assert().NoError(err)
}
