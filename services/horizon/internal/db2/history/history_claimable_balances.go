package history

import (
	"sort"

	sq "github.com/Masterminds/squirrel"
	"github.com/stellar/go/support/db"
	"github.com/stellar/go/support/errors"
	"github.com/stellar/go/xdr"
)

// QHistoryClaimableBalances defines account related queries.
type QHistoryClaimableBalances interface {
	CreateHistoryClaimableBalances(ids []string, batchSize int) (map[string]int64, error)
	NewOperationClaimableBalanceBatchInsertBuilder(maxBatchSize int) OperationClaimableBalanceBatchInsertBuilder
	NewTransactionClaimableBalanceBatchInsertBuilder(maxBatchSize int) TransactionClaimableBalanceBatchInsertBuilder
}

// CreateHistoryClaimableBalances creates rows in the history_claimable_balances table for a given list of ids.
// CreateHistoryClaimableBalances returns a mapping of id to its corresponding internal id in the history_claimable_balances table
func (q *Q) CreateHistoryClaimableBalances(ids []string, batchSize int) (map[string]int64, error) {
	builder := &db.BatchInsertBuilder{
		Table:        q.GetTable("history_claimable_balances"),
		MaxBatchSize: batchSize,
		Suffix:       "ON CONFLICT (claimable_balance_id) DO NOTHING",
	}

	// sort before inserting to prevent deadlocks on acquiring a ShareLock
	// https://github.com/stellar/go/issues/2370
	sort.Strings(ids)
	for _, id := range ids {
		err := builder.Row(map[string]interface{}{
			"claimable_balance_id": id,
		})
		if err != nil {
			return nil, errors.Wrap(err, "could not insert history_claimable_balances row")
		}
	}

	err := builder.Exec()
	if err != nil {
		return nil, errors.Wrap(err, "could not exec claimable balance insert builder")
	}

	var cbs []HistoryClaimableBalance
	toInternalID := map[string]int64{}
	const selectBatchSize = 10000

	for i := 0; i < len(ids); i += selectBatchSize {
		end := i + selectBatchSize
		if end > len(ids) {
			end = len(ids)
		}
		subset := ids[i:end]

		if err := q.ClaimableBalancesByIDs(&cbs, subset); err != nil {
			return nil, errors.Wrap(err, "could not select claimable balances")
		}

		for _, cb := range cbs {
			hexID, err := xdr.MarshalHex(cb.BalanceID)
			if err != nil {
				return nil, errors.New("error parsing BalanceID")
			}
			toInternalID[hexID] = cb.InternalID
		}
	}

	return toInternalID, nil
}

// HistoryClaimableBalance is a row of data from the `history_claimable_balances` table
type HistoryClaimableBalance struct {
	BalanceID  string `db:"claimable_balance_id"`
	InternalID int64  `db:"id"`
}

var selectHistoryClaimableBalance = sq.Select("hcb.*").From("history_claimable_balances hcb")

// ClaimableBalancesByIDs loads a rows from `history_accounts`, by claimable_balance_id
func (q *Q) ClaimableBalancesByIDs(dest interface{}, ids []string) error {
	sql := selectHistoryClaimableBalance.Where(map[string]interface{}{
		"hcb.claimable_balance_id": ids, // hcb.claimable_balance_id IN (...)
	})
	return q.Select(dest, sql)
}

type OperationClaimableBalanceBatchInsertBuilder interface {
	Add(operationID, internalID int64) error
	Exec() error
}

type operationClaimableBalanceBatchInsertBuilder struct {
	builder db.BatchInsertBuilder
}

func (q *Q) NewOperationClaimableBalanceBatchInsertBuilder(maxBatchSize int) OperationClaimableBalanceBatchInsertBuilder {
	return &operationClaimableBalanceBatchInsertBuilder{
		builder: db.BatchInsertBuilder{
			Table:        q.GetTable("history_operation_claimable_balances"),
			MaxBatchSize: maxBatchSize,
		},
	}
}

// Add adds a new operation claimable balance to the batch
func (i *operationClaimableBalanceBatchInsertBuilder) Add(operationID, internalID int64) error {
	return i.builder.Row(map[string]interface{}{
		"history_operation_id":         operationID,
		"history_claimable_balance_id": internalID,
	})
}

// Exec flushes all pending operation claimable balances to the db
func (i *operationClaimableBalanceBatchInsertBuilder) Exec() error {
	return i.builder.Exec()
}

type TransactionClaimableBalanceBatchInsertBuilder interface {
	Add(transactionID, internalID int64) error
	Exec() error
}

type transactionClaimableBalanceBatchInsertBuilder struct {
	builder db.BatchInsertBuilder
}

func (q *Q) NewTransactionClaimableBalanceBatchInsertBuilder(maxBatchSize int) TransactionClaimableBalanceBatchInsertBuilder {
	return &transactionClaimableBalanceBatchInsertBuilder{
		builder: db.BatchInsertBuilder{
			Table:        q.GetTable("history_transaction_claimable_balances"),
			MaxBatchSize: maxBatchSize,
		},
	}
}

// Add adds a new transaction claimable balance to the batch
func (i *transactionClaimableBalanceBatchInsertBuilder) Add(transactionID, internalID int64) error {
	return i.builder.Row(map[string]interface{}{
		"history_transaction_id":       transactionID,
		"history_claimable_balance_id": internalID,
	})
}

// Exec flushes all pending transaction claimable balances to the db
func (i *transactionClaimableBalanceBatchInsertBuilder) Exec() error {
	return i.builder.Exec()
}
