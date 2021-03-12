package processors

import (
	"github.com/stellar/go/ingest"
	"github.com/stellar/go/services/horizon/internal/db2/history"
	"github.com/stellar/go/support/errors"
	"github.com/stellar/go/xdr"
)

type claimableBalance struct {
	claimabeBalanceID int64
	internalId        int64 // TODO: Dunno if this is right at all...
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
		return errors.Wrap(err, "could not determine operation claimable balances")
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
	cbs := map[int64][]xdr.ClaimableBalanceId{}

	for opi, op := range transaction.Envelope.Operations() {
		operation := transactionOperationWrapper{
			index:          uint32(opi),
			transaction:    transaction,
			operation:      op,
			ledgerSequence: sequence,
		}

		cb, err := operation.ClaimableBalances()
		if err != nil {
			return cbs, errors.Wrapf(err, "reading operation %v claimable balances", operation.ID())
		}
		cbs[operation.ID()] = cb
	}

	return cbs, nil
}

func (p *ClaimableBalancesTransactionProcessor) Commit() error {
	if len(p.claimableBalanceSet) > 0 {
		if err := p.loadClaimableBalanceIDs(p.claimableBalanceSet); err != nil {
			return err
		}

		if err := p.insertDBTransactionClaimableBalances(p.claimableBalanceSet); err != nil {
			return err
		}

		if err := p.insertDBOperationsClaimableBalances(p.claimableBalanceSet); err != nil {
			return err
		}
	}

	return nil
}

func (p *ClaimableBalancesTransactionProcessor) loadClaimableBalanceIDs(claimableBalanceSet map[string]claimableBalance) error {
	ids := make([]string, 0, len(claimableBalanceSet))
	for id := range claimableBalanceSet {
		ids = append(ids, id)
	}

	toInternalId, err := p.qClaimableBalances.CreateClaimableBalances(ids, maxBatchSize)
	if err != nil {
		return errors.Wrap(err, "Could not create claimable balance ids")
	}

	for _, id := range ids {
		internalId, ok := toInternalId[id]
		if !ok {
			return errors.Errorf("no internal id found for account address %s", internalId)
		}

		cb := claimableBalanceSet[id]
		cb.internalId = internalId
		claimableBalanceSet[id] = cb
	}

	return nil
}

func (p ClaimableBalancesTransactionProcessor) insertDBTransactionParticipants(claimableBalanceSet map[string]claimableBalance) error {
	panic("did not implement")
}

func (p ClaimableBalancesTransactionProcessor) insertDBOperationParticipants(claimableBalanceSet map[string]claimableBalance) error {
	batch := p.qClaimableBalances.NewOperationClaimableBalanceBatchInsertBuilder(maxBatchSize)

	for _, entry := range claimableBalanceSet {
		for operationID := range entry.operationSet {
			if err := batch.Add(operationID, entry.internalId); err != nil {
				return errors.Wrap(err, "could not insert operation claimable balance in db")
			}
		}
	}

	if err := batch.Exec(); err != nil {
		return errors.Wrap(err, "could not flush operation claimable balances to db")
	}
	return nil
}
