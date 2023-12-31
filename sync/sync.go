package sync

//
//import (
//	"context"
//	"ethereum-crawler/blockchain"
//	"ethereum-crawler/utils"
//)
//
//type Synchronizer struct {
//	startingBlockNumber *uint64
//	blockchain          *blockchain.Blockchain
//	log                 utils.SimpleLogger
//}
//
//func New() *Synchronizer {
//	s := &Synchronizer{}
//	return s
//}
//
//func (s *Synchronizer) Run(ctx context.Context) error {
//
//	s.syncBlocks(ctx)
//	return nil
//}
//func (s *Synchronizer) syncBlocks(syncCtx context.Context) {
//	defer func() {
//		s.startingBlockNumber = nil
//	}()
//
//	nextHeight := s.nextHeight()
//	startingHeight := nextHeight
//	s.startingBlockNumber = &startingHeight
//
//	latestSem := make(chan struct{}, 1)
//
//	streamCtx, streamCancel := context.WithCancel(syncCtx)
//
//	pendingSem := make(chan struct{}, 1)
//
//	for {
//		select {
//		case <-streamCtx.Done():
//			streamCancel()
//
//			select {
//			case <-syncCtx.Done():
//				pendingSem <- struct{}{}
//				latestSem <- struct{}{}
//				return
//			default:
//				streamCtx, streamCancel = context.WithCancel(syncCtx)
//				nextHeight = s.nextHeight()
//				s.log.Warnw("Restarting sync process", "height", nextHeight)
//			}
//		default:
//			nextHeight++
//		}
//	}
//}
//func (s *Synchronizer) nextHeight() uint64 {
//	nextHeight := uint64(0)
//	if h, err := s.blockchain.Height(); err == nil {
//		nextHeight = h + 1
//	}
//	return nextHeight
//}
