package node

import (
	"ethereum-crawler/model"
	"sort"
)

type MockNode struct {
	mockedData []model.Account
}

func NewMockNode() *MockNode {
	return &MockNode{
		mockedData: make([]model.Account, 0),
	}
}

func (m *MockNode) Add(_ int64, account model.Account) {
	m.mockedData = append(m.mockedData, account)
}

func (m *MockNode) Sync(start int64, x chan model.Account) {
	sort.Slice(m.mockedData, func(i, j int) bool {
		return m.mockedData[i].Height < m.mockedData[j].Height
	})
	for _, account := range m.mockedData {
		if account.Height >= start {
			x <- account
		}
	}
}
