package processors

import (
	"github.com/stellar/go/ingest"
	"github.com/stellar/go/services/horizon/internal/db2/history"
	"github.com/stellar/go/support/errors"
	"github.com/stellar/go/xdr"
)

type claimableBalance struct {
	claimabeBalanceID int64
	transactionSet    map[int64]struct{}
	operationSet      map[int64]struct{}
}

func (b claimableBalance) addOperationID(id int64) {
	if b.operationSet == nil {
		b.operationSet = map[int64]struct{}{}
	}
	b.operationSet[id] = struct{}{}
}

type ClaimableBalancesTransactionProcessor struct {
	sequence            uint32
	claimableBalanceSet map[string]claimableBalance
	qClaimableBalances  history.QClaimableBalances
}

func NewClaimableBalancesTransactionProcessor(Q history.QClaimableBalances, sequence uint32) *ClaimableBalancesTransactionProcessor {
	return &ClaimableBalancesTransactionProcessor{
		qClaimableBalances:  Q,
		sequence:            sequence,
		claimableBalanceSet: map[string]claimableBalance{},
	}
}

func (p *ClaimableBalancesTransactionProcessor) ProcessTransaction(transaction ingest.LedgerTransaction) error {
	err := p.addTransactionClaimableBalances(p.claimableBalanceSet, p.sequence, transaction)
	if err != nil {
		return err
	}

	err = p.addOperationClaimableBalances(p.claimableBalanceSet, p.sequence, transaction)
	if err != nil {
		return err
	}

	return nil
}

func (p *ClaimableBalancesTransactionProcessor) addTransactionClaimableBalances(set map[string]claimableBalance, sequence uint32, transaction ingest.LedgerTransaction) error {
	panic("did not implement")
}

func (p *ClaimableBalancesTransactionProcessor) addOperationClaimableBalances(cbSet map[string]claimableBalance, sequence uint32, transaction ingest.LedgerTransaction) error {
	claimableBalances, err := operationsClaimableBalances(transaction, sequence)
	if err != nil {
		return errors.Wrap(err, "could not determine operation participants")
	}

	for operationID, cbs := range claimableBalances {
		for _, cb := range cbs {
			hexID, err := xdr.MarshalHex(cb)
			if err != nil {
				return errors.New("error parsing BalanceID")
			}
			entry := cbSet[hexID]
			entry.addOperationID(operationID)
			cbSet[hexID] = entry
		}
	}

	return nil
}

func operationsClaimableBalances(transaction ingest.LedgerTransaction, sequence uint32) (map[int64][]xdr.ClaimableBalanceId, error) {
	return nil, errors.New("TODO: Implement operationsClaimableBalances")
}

func (p *ClaimableBalancesTransactionProcessor) Commit() error {
	return errors.New("TODO: Implement ClaimableBalancesTransactionProcessor.Commit")
}
