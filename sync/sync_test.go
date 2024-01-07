package sync

import (
	"context"
	"ethereum-crawler/db"
	"ethereum-crawler/model"
	"ethereum-crawler/node"
	"ethereum-crawler/utils"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

// init test with new mock node, db, log
var (
	ctx     = context.Background()
	ethNode = node.NewMockNode()
	testDb  = db.NewMemDB[model.Account]()
	log     = utils.NewNopZapLogger()
)

func init() {
	ctx = context.Background()
	ethNode = node.NewMockNode()
	testDb = db.NewMemDB[model.Account]()
	log = utils.NewNopZapLogger()
}
func TestNew(t *testing.T) {
	_, er := New(ctx, ethNode, testDb, log)
	require.NoError(t, er)
}
func TestTotalPaidFee(t *testing.T) {

	ctx, _ = context.WithTimeout(context.Background(), time.Second)
	testDb.Add(LastHeightKey, model.Account{
		Address:      "",
		TotalPaidFee: 0,
		Height:       1,
	})
	ethNode.Add(1, model.Account{
		Address:      "0x1",
		TotalPaidFee: 300,
		Height:       200,
	})
	ethNode.Add(2, model.Account{
		Address:      "0x2",
		TotalPaidFee: 300,
		Height:       300,
	})
	ethNode.Add(3, model.Account{
		Address:      "0x1",
		TotalPaidFee: 100,
		Height:       400,
	})
	sync, er := New(ctx, ethNode, testDb, log)
	require.NoError(t, er)

	err := sync.Start()
	require.NoError(t, err)
	res, err := testDb.Get("0x1")
	require.Equal(t, uint64(400), res.TotalPaidFee)
	res2, err := testDb.Get("0x2")
	require.Equal(t, uint64(300), res2.TotalPaidFee)

	_, err = testDb.Get("0x3")
	require.NotNil(t, err)
	require.Equal(t, true, testDb.IsNotFoundError(err))
}

func TestLastHeightKey(t *testing.T) {
	ctx, _ = context.WithTimeout(context.Background(), time.Second)
	testDb.Add(LastHeightKey, model.Account{
		Address:      "",
		TotalPaidFee: 0,
		Height:       100,
	})
	ethNode.Add(3, model.Account{
		Address:      "0x3",
		TotalPaidFee: 100,
		Height:       4000,
	})
	sync, err := New(ctx, ethNode, testDb, log)
	err = sync.Start()
	require.NoError(t, err)
	res, err := testDb.Get(LastHeightKey)
	require.NoError(t, err)
	require.Equal(t, int64(4000), res.Height)
}

func TestDuplicateEntry(t *testing.T) {
	ctx, _ = context.WithTimeout(context.Background(), time.Second)

	testDb.Add(LastHeightKey, model.Account{
		Address:      "",
		TotalPaidFee: 0,
		Height:       0,
	})
	ethNode.Add(3, model.Account{
		Address:      "0x3",
		TotalPaidFee: 4000,
		Height:       100,
	})

	sync, err := New(ctx, ethNode, testDb, log)
	err = sync.Start()
	require.NoError(t, err)
	account, err := testDb.Get("0x3")
	require.NoError(t, err)
	require.Equal(t, uint64(4000), account.TotalPaidFee)

	ethNode = node.NewMockNode()
	ethNode.Add(3, model.Account{
		Address:      "0x3",
		TotalPaidFee: 4000,
		Height:       100,
	})
	sync, err = New(ctx, ethNode, testDb, log)
	err = sync.Start()
	require.NoError(t, err)
	account, err = testDb.Get("0x3")
	require.NoError(t, err)
	require.Equal(t, uint64(4000), account.TotalPaidFee)

}
