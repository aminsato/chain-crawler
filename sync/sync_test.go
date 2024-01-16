package sync

import (
	"context"
	"testing"
	"time"

	"chain-crawler/db"
	"chain-crawler/model"
	"chain-crawler/node"
	"chain-crawler/utils"
	"github.com/stretchr/testify/require"
)

func TestTotalPaidFee(t *testing.T) {
	testDb := db.NewMemDB[model.Account]()
	log := utils.NewNopZapLogger()
	ethNode := node.NewMockNode()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := testDb.Add(db.LastHeightKey, model.Account{
		Address:      "",
		TotalPaidFee: 0,
		LastHeight:   1,
	})
	if err != nil {
		log.Errorw(err.Error())
		return
	}
	ethNode.Add(1, model.Account{
		Address:      "0x1",
		TotalPaidFee: 300,
		LastHeight:   200,
	})
	ethNode.Add(2, model.Account{
		Address:      "0x2",
		TotalPaidFee: 300,
		LastHeight:   300,
	})
	ethNode.Add(3, model.Account{
		Address:      "0x1",
		TotalPaidFee: 100,
		LastHeight:   400,
	})
	sync, er := New(ctx, ethNode, testDb, log)
	require.NoError(t, er)

	err = sync.Start()
	require.NoError(t, err)
	res, err := testDb.Get("0x1")
	require.NoError(t, err)

	require.Equal(t, uint64(400), res.TotalPaidFee)
	res2, err := testDb.Get("0x2")
	require.NoError(t, err)
	require.Equal(t, uint64(300), res2.TotalPaidFee)

	_, err = testDb.Get("0x3")
	require.NotNil(t, err)
	require.Equal(t, true, testDb.IsNotFoundError(err))
}

func TestLastHeightKey(t *testing.T) {
	testDb := db.NewMemDB[model.Account]()
	log := utils.NewNopZapLogger()
	ethNode := node.NewMockNode()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := testDb.Add(db.LastHeightKey, model.Account{
		Address:      "",
		TotalPaidFee: 0,
		LastHeight:   100,
	})
	if err != nil {
		return
	}
	ethNode.Add(3, model.Account{
		Address:      "0x3",
		TotalPaidFee: 100,
		LastHeight:   4000,
	})
	sync, err := New(ctx, ethNode, testDb, log)
	if err != nil {
		log.Errorw(err.Error())
		return
	}
	err = sync.Start()
	require.NoError(t, err)
	res, err := testDb.Get(db.LastHeightKey)
	require.NoError(t, err)
	require.Equal(t, int64(4000), res.LastHeight)
}

func TestDuplicateEntry(t *testing.T) {
	testDb := db.NewMemDB[model.Account]()
	log := utils.NewNopZapLogger()
	ethNode := node.NewMockNode()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := testDb.Add(db.LastHeightKey, model.Account{
		Address:      "",
		TotalPaidFee: 0,
		LastHeight:   0,
	})
	if err != nil {
		return
	}
	ethNode.Add(3, model.Account{
		Address:      "0x3",
		TotalPaidFee: 4000,
		LastHeight:   100,
	})

	sync, err := New(ctx, ethNode, testDb, log)
	if err != nil {
		log.Errorw(err.Error())
		return
	}
	err = sync.Start()
	require.NoError(t, err)
	account, err := testDb.Get("0x3")
	require.NoError(t, err)
	require.Equal(t, uint64(4000), account.TotalPaidFee)

	ethNode = node.NewMockNode()
	ethNode.Add(3, model.Account{
		Address:      "0x3",
		TotalPaidFee: 4000,
		LastHeight:   100,
	})
	sync, _ = New(ctx, ethNode, testDb, log)
	err = sync.Start()
	require.NoError(t, err)
	account, err = testDb.Get("0x3")
	require.NoError(t, err)
	require.Equal(t, uint64(4000), account.TotalPaidFee)
}

func TestRecords(t *testing.T) {
	testDb := db.NewMemDB[model.Account]()
	log := utils.NewNopZapLogger()
	err := testDb.Add(db.LastHeightKey, model.Account{
		Address:      "",
		TotalPaidFee: 0,
		LastHeight:   0,
	})
	if err != nil {
		log.Errorw(err.Error())
		return
	}
	err = testDb.Add("1", model.Account{
		Address:      "0x3",
		TotalPaidFee: 4000,
		LastHeight:   100,
	})
	if err != nil {
		log.Errorw(err.Error())
		return
	}
	err = testDb.Add("2", model.Account{
		Address:      "0x3",
		TotalPaidFee: 4000,
		LastHeight:   100,
	})
	if err != nil {
		log.Errorw(err.Error())
		return
	}
	err = testDb.Add("3", model.Account{
		Address:      "0x3",
		TotalPaidFee: 4000,
		LastHeight:   100,
	})
	if err != nil {
		log.Errorw(err.Error())
		return
	}
	records := make(chan db.DBItem[model.Account], 1)
	go func() {
		err = testDb.Records(nil, nil, records)
	}()
	count := int(0)
	isOpen := true
	for isOpen {
		_, ok := <-records
		if ok {
			count++
		} else {
			isOpen = false
		}
	}
	require.Equal(t, 4, count)
	require.NoError(t, err)
}
