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

type ClaimableBalancesProcessor struct {
	sequence            uint32
	claimableBalanceSet map[string]claimableBalance
	qClaimableBalances  history.QClaimableBalances
	cache               *ingest.ChangeCompactor
}

func NewClaimableBalancesProcessor(Q history.QClaimableBalances, sequence uint32) *ClaimableBalancesProcessor {
	p := &ClaimableBalancesProcessor{
		qClaimableBalances:  Q,
		sequence:            sequence,
		claimableBalanceSet: map[string]claimableBalance{},
	}
	p.reset()
	return p
}

func (p *ClaimableBalancesProcessor) ProcessTransaction(transaction ingest.LedgerTransaction) error {
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

func (p *ClaimableBalancesProcessor) addTransactionClaimableBalances(set map[string]claimableBalance, sequence uint32, transaction ingest.LedgerTransaction) error {
	panic("did not implement")
}

func (p *ClaimableBalancesProcessor) addOperationClaimableBalances(cbSet map[string]claimableBalance, sequence uint32, transaction ingest.LedgerTransaction) error {
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

}

func (p *ClaimableBalancesProcessor) reset() {
	p.cache = ingest.NewChangeCompactor()
}

func (p *ClaimableBalancesProcessor) ProcessChange(change ingest.Change) error {
	if change.Type != xdr.LedgerEntryTypeClaimableBalance {
		return nil
	}

	err := p.cache.AddChange(change)
	if err != nil {
		return errors.Wrap(err, "error adding to ledgerCache")
	}

	if p.cache.Size() > maxBatchSize {
		err = p.Commit()
		if err != nil {
			return errors.Wrap(err, "error in Commit")
		}
		p.reset()
	}

	return nil
}

func (p *ClaimableBalancesProcessor) Commit() error {
	batch := p.qClaimableBalances.NewClaimableBalancesBatchInsertBuilder(maxBatchSize)

	changes := p.cache.GetChanges()
	for _, change := range changes {
		var err error
		var rowsAffected int64
		var action string
		var ledgerKey xdr.LedgerKey

		switch {
		case change.Pre == nil && change.Post != nil:
			// Created
			action = "inserting"
			err = batch.Add(change.Post)
			rowsAffected = 1
		case change.Pre != nil && change.Post == nil:
			// Removed
			action = "removing"
			cBalance := change.Pre.Data.MustClaimableBalance()
			err = ledgerKey.SetClaimableBalance(cBalance.BalanceId)
			if err != nil {
				return errors.Wrap(err, "Error creating ledger key")
			}
			rowsAffected, err = p.qClaimableBalances.RemoveClaimableBalance(cBalance)
		default:
			// Updated
			action = "updating"
			cBalance := change.Post.Data.MustClaimableBalance()
			err = ledgerKey.SetClaimableBalance(cBalance.BalanceId)
			if err != nil {
				return errors.Wrap(err, "Error creating ledger key")
			}
			rowsAffected, err = p.qClaimableBalances.UpdateClaimableBalance(*change.Post)
		}

		if err != nil {
			return err
		}

		if rowsAffected != 1 {
			ledgerKeyString, err := ledgerKey.MarshalBinaryBase64()
			if err != nil {
				return errors.Wrap(err, "Error marshalling ledger key")
			}
			return ingest.NewStateError(errors.Errorf(
				"%d rows affected when %s claimable balance: %s",
				rowsAffected,
				action,
				ledgerKeyString,
			))
		}
	}

	err := batch.Exec()
	if err != nil {
		return errors.Wrap(err, "error executing batch")
	}

	return nil
}
