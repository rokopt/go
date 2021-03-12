//lint:file-ignore U1001 Ignore all unused code, staticcheck doesn't understand testify/suite

package processors

import (
	"testing"

	"github.com/stellar/go/ingest"
	"github.com/stellar/go/services/horizon/internal/db2/history"
	"github.com/stellar/go/xdr"
	"github.com/stretchr/testify/suite"
)

func TestClaimableBalancesTransactionProcessorTestSuiteState(t *testing.T) {
	suite.Run(t, new(ClaimableBalancesTransactionProcessorTestSuiteState))
}

type ClaimableBalancesTransactionProcessorTestSuiteState struct {
	suite.Suite
	processor              *ClaimableBalancesTransactionProcessor
	mockQ                  *history.MockQClaimableBalances
	mockBatchInsertBuilder *history.MockClaimableBalancesBatchInsertBuilder
}

func (s *ClaimableBalancesTransactionProcessorTestSuiteState) SetupTest() {
	s.mockQ = &history.MockQClaimableBalances{}
	s.mockBatchInsertBuilder = &history.MockClaimableBalancesBatchInsertBuilder{}

	s.mockQ.
		On("NewClaimableBalancesBatchInsertBuilder", maxBatchSize).
		Return(s.mockBatchInsertBuilder)

	s.processor = NewClaimableBalancesTransactionProcessor(s.mockQ, 0)
}

func (s *ClaimableBalancesTransactionProcessorTestSuiteState) TearDownTest() {
	s.mockBatchInsertBuilder.On("Exec").Return(nil).Once()
	s.Assert().NoError(s.processor.Commit())
	s.mockQ.AssertExpectations(s.T())
	s.mockBatchInsertBuilder.AssertExpectations(s.T())
}

func (s *ClaimableBalancesTransactionProcessorTestSuiteState) TestNoEntries() {
	// Nothing processed, assertions in TearDownTest.
}

func (s *ClaimableBalancesTransactionProcessorTestSuiteState) TestCreatesClaimableBalances() {
	lastModifiedLedgerSeq := xdr.Uint32(123)
	cBalance := xdr.ClaimableBalanceEntry{
		BalanceId: xdr.ClaimableBalanceId{
			Type: xdr.ClaimableBalanceIdTypeClaimableBalanceIdTypeV0,
			V0:   &xdr.Hash{1, 2, 3},
		},
		Claimants: []xdr.Claimant{},
		Asset:     xdr.MustNewCreditAsset("USD", "GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML"),
		Amount:    10,
	}

	s.mockBatchInsertBuilder.On("Add", &xdr.LedgerEntry{
		LastModifiedLedgerSeq: lastModifiedLedgerSeq,
		Data: xdr.LedgerEntryData{
			Type:             xdr.LedgerEntryTypeClaimableBalance,
			ClaimableBalance: &cBalance,
		},
	}).Return(nil).Once()

	err := s.processor.ProcessChange(ingest.Change{
		Type: xdr.LedgerEntryTypeClaimableBalance,
		Pre:  nil,
		Post: &xdr.LedgerEntry{
			Data: xdr.LedgerEntryData{
				Type:             xdr.LedgerEntryTypeClaimableBalance,
				ClaimableBalance: &cBalance,
			},
			LastModifiedLedgerSeq: lastModifiedLedgerSeq,
		},
	})
	s.Assert().NoError(err)
}

func TestClaimableBalancesTransactionProcessorTestSuiteLedger(t *testing.T) {
	suite.Run(t, new(ClaimableBalancesTransactionProcessorTestSuiteLedger))
}

type ClaimableBalancesTransactionProcessorTestSuiteLedger struct {
	suite.Suite
	processor              *ClaimableBalancesTransactionProcessor
	mockQ                  *history.MockQClaimableBalances
	mockBatchInsertBuilder *history.MockClaimableBalancesBatchInsertBuilder
}

func (s *ClaimableBalancesTransactionProcessorTestSuiteLedger) SetupTest() {
	s.mockQ = &history.MockQClaimableBalances{}
	s.mockBatchInsertBuilder = &history.MockClaimableBalancesBatchInsertBuilder{}

	s.mockQ.
		On("NewClaimableBalancesBatchInsertBuilder", maxBatchSize).
		Return(s.mockBatchInsertBuilder)

	s.processor = NewClaimableBalancesTransactionProcessor(s.mockQ, 0)
}

func (s *ClaimableBalancesTransactionProcessorTestSuiteLedger) TearDownTest() {
	s.mockBatchInsertBuilder.On("Exec").Return(nil).Once()
	s.Assert().NoError(s.processor.Commit())
	s.mockQ.AssertExpectations(s.T())
}

func (s *ClaimableBalancesTransactionProcessorTestSuiteLedger) TestNoTransactions() {
	// Nothing processed, assertions in TearDownTest.
}

func (s *ClaimableBalancesTransactionProcessorTestSuiteLedger) TestNewClaimableBalance() {
	lastModifiedLedgerSeq := xdr.Uint32(123)
	cBalance := xdr.ClaimableBalanceEntry{
		BalanceId: xdr.ClaimableBalanceId{
			Type: xdr.ClaimableBalanceIdTypeClaimableBalanceIdTypeV0,
			V0:   &xdr.Hash{1, 2, 3},
		},
		Claimants: []xdr.Claimant{},
		Asset:     xdr.MustNewCreditAsset("USD", "GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML"),
		Amount:    10,
	}
	entry := xdr.LedgerEntry{
		Data: xdr.LedgerEntryData{
			Type:             xdr.LedgerEntryTypeClaimableBalance,
			ClaimableBalance: &cBalance,
		},
		LastModifiedLedgerSeq: lastModifiedLedgerSeq,
		Ext: xdr.LedgerEntryExt{
			V: 1,
			V1: &xdr.LedgerEntryExtensionV1{
				SponsoringId: nil,
			},
		},
	}
	err := s.processor.ProcessChange(ingest.Change{
		Type: xdr.LedgerEntryTypeClaimableBalance,
		Pre:  nil,
		Post: &entry,
	})
	s.Assert().NoError(err)

	// add sponsor
	updated := xdr.LedgerEntry{
		Data: xdr.LedgerEntryData{
			Type:             xdr.LedgerEntryTypeClaimableBalance,
			ClaimableBalance: &cBalance,
		},
		LastModifiedLedgerSeq: lastModifiedLedgerSeq,
		Ext: xdr.LedgerEntryExt{
			V: 1,
			V1: &xdr.LedgerEntryExtensionV1{
				SponsoringId: xdr.MustAddressPtr("GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML"),
			},
		},
	}

	entry.LastModifiedLedgerSeq = entry.LastModifiedLedgerSeq - 1
	err = s.processor.ProcessChange(ingest.Change{
		Type: xdr.LedgerEntryTypeClaimableBalance,
		Pre:  &entry,
		Post: &updated,
	})
	s.Assert().NoError(err)

	// We use LedgerEntryChangesCache so all changes are squashed
	s.mockBatchInsertBuilder.On(
		"Add",
		&updated,
	).Return(nil).Once()
}

func (s *ClaimableBalancesTransactionProcessorTestSuiteLedger) TestUpdateClaimableBalance() {
	cBalance := xdr.ClaimableBalanceEntry{
		BalanceId: xdr.ClaimableBalanceId{
			Type: xdr.ClaimableBalanceIdTypeClaimableBalanceIdTypeV0,
			V0:   &xdr.Hash{1, 2, 3},
		},
		Claimants: []xdr.Claimant{},
		Asset:     xdr.MustNewCreditAsset("USD", "GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML"),
		Amount:    10,
	}
	lastModifiedLedgerSeq := xdr.Uint32(123)

	pre := xdr.LedgerEntry{
		Data: xdr.LedgerEntryData{
			Type:             xdr.LedgerEntryTypeClaimableBalance,
			ClaimableBalance: &cBalance,
		},
		LastModifiedLedgerSeq: lastModifiedLedgerSeq - 1,
		Ext: xdr.LedgerEntryExt{
			V: 1,
			V1: &xdr.LedgerEntryExtensionV1{
				SponsoringId: nil,
			},
		},
	}

	// add sponsor
	updated := xdr.LedgerEntry{
		Data: xdr.LedgerEntryData{
			Type:             xdr.LedgerEntryTypeClaimableBalance,
			ClaimableBalance: &cBalance,
		},
		LastModifiedLedgerSeq: lastModifiedLedgerSeq,
		Ext: xdr.LedgerEntryExt{
			V: 1,
			V1: &xdr.LedgerEntryExtensionV1{
				SponsoringId: xdr.MustAddressPtr("GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML"),
			},
		},
	}

	err := s.processor.ProcessChange(ingest.Change{
		Type: xdr.LedgerEntryTypeClaimableBalance,
		Pre:  &pre,
		Post: &updated,
	})
	s.Assert().NoError(err)

	s.mockQ.On(
		"UpdateClaimableBalance",
		updated,
	).Return(int64(1), nil).Once()
}

func (s *ClaimableBalancesTransactionProcessorTestSuiteLedger) TestRemoveClaimableBalance() {
	cBalance := xdr.ClaimableBalanceEntry{
		BalanceId: xdr.ClaimableBalanceId{
			Type: xdr.ClaimableBalanceIdTypeClaimableBalanceIdTypeV0,
			V0:   &xdr.Hash{1, 2, 3},
		},
		Claimants: []xdr.Claimant{},
		Asset:     xdr.MustNewCreditAsset("USD", "GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML"),
		Amount:    10,
	}
	lastModifiedLedgerSeq := xdr.Uint32(123)
	pre := xdr.LedgerEntry{
		Data: xdr.LedgerEntryData{
			Type:             xdr.LedgerEntryTypeClaimableBalance,
			ClaimableBalance: &cBalance,
		},
		LastModifiedLedgerSeq: lastModifiedLedgerSeq - 1,
		Ext: xdr.LedgerEntryExt{
			V: 1,
			V1: &xdr.LedgerEntryExtensionV1{
				SponsoringId: nil,
			},
		},
	}
	err := s.processor.ProcessChange(ingest.Change{
		Type: xdr.LedgerEntryTypeClaimableBalance,
		Pre:  &pre,
		Post: nil,
	})
	s.Assert().NoError(err)

	s.mockQ.On(
		"RemoveClaimableBalance",
		cBalance,
	).Return(int64(1), nil).Once()
}
