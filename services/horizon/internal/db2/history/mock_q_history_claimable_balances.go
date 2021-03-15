package history

import (
	"github.com/stretchr/testify/mock"
)

// MockQHistoryClaimableBalances is a mock implementation of the QClaimableBalances interface
type MockQHistoryClaimableBalances struct {
	mock.Mock
}

func (m *MockQClaimableBalances) CreateHistoryClaimableBalances(ids []string, maxBatchSize int) (map[string]int64, error) {
	a := m.Called(ids, maxBatchSize)
	return a.Get(0).(map[string]int64), a.Error(1)
}

func (m *MockQClaimableBalances) NewTransactionClaimableBalanceBatchInsertBuilder(maxBatchSize int) TransactionClaimableBalanceBatchInsertBuilder {
	a := m.Called(maxBatchSize)
	return a.Get(0).(TransactionClaimableBalanceBatchInsertBuilder)
}

// MockTransactionClaimableBalanceBatchInsertBuilder is a mock implementation of the
// TransactionClaimableBalanceBatchInsertBuilder interface
type MockTransactionClaimableBalanceBatchInsertBuilder struct {
	mock.Mock
}

func (m *MockTransactionClaimableBalanceBatchInsertBuilder) Add(transactionID, accountID int64) error {
	a := m.Called(transactionID, accountID)
	return a.Error(0)
}

func (m *MockTransactionClaimableBalanceBatchInsertBuilder) Exec() error {
	a := m.Called()
	return a.Error(0)
}

// NewOperationClaimableBalanceBatchInsertBuilder mock
func (m *MockQClaimableBalances) NewOperationClaimableBalanceBatchInsertBuilder(maxBatchSize int) OperationClaimableBalanceBatchInsertBuilder {
	a := m.Called(maxBatchSize)
	return a.Get(0).(OperationClaimableBalanceBatchInsertBuilder)
}

// MockOperationClaimableBalanceBatchInsertBuilder is a mock implementation of the
// OperationClaimableBalanceBatchInsertBuilder interface
type MockOperationClaimableBalanceBatchInsertBuilder struct {
	mock.Mock
}

func (m *MockOperationClaimableBalanceBatchInsertBuilder) Add(transactionID, accountID int64) error {
	a := m.Called(transactionID, accountID)
	return a.Error(0)
}

func (m *MockOperationClaimableBalanceBatchInsertBuilder) Exec() error {
	a := m.Called()
	return a.Error(0)
}
