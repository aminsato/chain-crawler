package node

import (
	"sort"

	"chain-crawler/model"
)

type MockNode struct {
	mockedData []model.Account
}

func (m *MockNode) FirstBlock() int64 {
	return 1
}

func NewMockNode() *MockNode {
	return &MockNode{
		mockedData: make([]model.Account, 0),
	}
}

func (m *MockNode) Add(_ int64, account model.Account) {
	m.mockedData = append(m.mockedData, account)
}

func (m *MockNode) Sync(start int64, x chan model.Account) error {
	sort.Slice(m.mockedData, func(i, j int) bool {
		return m.mockedData[i].LastHeight < m.mockedData[j].LastHeight
	})
	for _, account := range m.mockedData {
		if account.LastHeight >= start {
			x <- account
		}
	}
	return nil
}
