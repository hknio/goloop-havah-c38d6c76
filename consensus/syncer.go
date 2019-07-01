package consensus

import (
	"time"

	"github.com/icon-project/goloop/common"
	"github.com/icon-project/goloop/common/log"
	"github.com/icon-project/goloop/consensus/internal/fastsync"
	"github.com/icon-project/goloop/module"
)

const (
	configSendBPS                   = -1
	configRoundStateMessageInterval = 300 * time.Millisecond
	configFastSyncThreshold         = 4
)

type Engine interface {
	GetCommitBlockParts(h int64) PartSet
	GetCommitPrecommits(h int64) *voteList
	GetPrecommits(r int32) *voteList
	GetVotes(r int32, prevotesMask *bitArray, precommitsMask *bitArray) *voteList
	GetRoundState() *peerRoundState

	Height() int64
	Round() int32
	Step() step

	ReceiveBlockPartMessage(msg *blockPartMessage, unicast bool) (int, error)
	ReceiveVoteMessage(msg *voteMessage, unicast bool) (int, error)
	ReceiveBlock(br fastsync.BlockResult)
}

type Syncer interface {
	Start() error
	Stop()
	OnEngineStepChange()
}

var syncerProtocols = []module.ProtocolInfo{
	protoBlockPart,
	protoRoundState,
	protoVoteList,
}

type peer struct {
	*syncer
	id         module.PeerID
	wakeUpChan chan struct{}
	stopped    chan struct{}
	logger     log.Logger

	running bool
	*peerRoundState
}

func newPeer(syncer *syncer, id module.PeerID) *peer {
	peerLogger := syncer.logger.WithFields(log.Fields{
		"peer": common.HexPre(id.Bytes()),
	})
	return &peer{
		syncer:     syncer,
		id:         id,
		wakeUpChan: make(chan struct{}, 1),
		stopped:    make(chan struct{}),
		logger:     peerLogger,
		running:    true, // TODO better way
	}
}

func (p *peer) setRoundState(prs *peerRoundState) {
	p.peerRoundState = prs
	p.wakeUp()
}

func (p *peer) doSync() (module.ProtocolInfo, message) {
	e := p.engine
	if p.peerRoundState == nil {
		p.logger.Tracef("nil peer round state\n")
		return nil, nil
	}

	if !p.peerRoundState.Sync {
		p.logger.Tracef("peer round state: no sync\n")
		return nil, nil
	}

	if p.Height < e.Height() || (p.Height == e.Height() && e.Step() >= stepCommit) {
		if p.BlockPartsMask == nil {
			vl := e.GetCommitPrecommits(p.Height)
			msg := newVoteListMessage()
			msg.VoteList = vl
			p.BlockPartsMask = newBitArray(e.GetCommitBlockParts(p.Height).Parts())
			p.logger.Tracef("PC for commit %v\n", p.Height)
			return protoVoteList, msg
		}
		partSet := e.GetCommitBlockParts(p.Height)
		mask := p.BlockPartsMask.Copy()
		mask.Flip()
		mask.AssignAnd(partSet.GetMask())
		idx := mask.PickRandom()
		if idx < 0 {
			p.logger.Tracef("no bp to send: %v/%v\n", p.BlockPartsMask, partSet.GetMask())
			return nil, nil
		}
		part := partSet.GetPart(idx)
		msg := newBlockPartMessage()
		msg.Height = p.Height
		msg.Index = uint16(idx)
		msg.BlockPart = part.Bytes()
		p.BlockPartsMask.Set(idx)
		return protoBlockPart, msg
	}
	if p.Height > e.Height() {
		p.logger.Tracef("higher peer height %v > %v\n", p.Height, e.Height())
		if p.Height > e.Height()+configFastSyncThreshold && p.syncer.fetchCanceler == nil {
			blk, err := p.syncer.bm.GetBlockByHeight(e.Height() - 1)
			if err != nil {
				return nil, nil
			}
			p.syncer.fetchCanceler, _ = p.syncer.fsm.FetchBlocks(e.Height(), -1, blk, NewCommitVoteSetFromBytes, p.syncer)
		}
		return nil, nil
	}

	if p.Round < e.Round() && e.Step() >= stepPrecommitWait {
		vl := e.GetPrecommits(e.Round())
		msg := newVoteListMessage()
		msg.VoteList = vl
		p.peerRoundState = nil
		p.logger.Tracef("PC for round %v\n", e.Round())
		return protoVoteList, msg
	} else if p.Round < e.Round() {
		// TODO: check peer step
		vl := e.GetPrecommits(e.Round() - 1)
		msg := newVoteListMessage()
		msg.VoteList = vl
		p.peerRoundState = nil
		p.logger.Tracef("PC for round %v (prev round)\n", e.Round())
		return protoVoteList, msg
	} else if p.Round == e.Round() {
		rs := e.GetRoundState()
		p.logger.Tracef("r=%v pv=%v/%v pc=%v/%v\n", e.Round(), p.PrevotesMask, rs.PrevotesMask, p.PrecommitsMask, rs.PrecommitsMask)
		pv := p.PrevotesMask.Copy()
		pv.Flip()
		pc := p.PrecommitsMask.Copy()
		pc.Flip()
		vl := e.GetVotes(e.Round(), pv, pc)
		if vl.Len() > 0 {
			msg := newVoteListMessage()
			msg.VoteList = vl
			p.peerRoundState = nil
			return protoVoteList, msg
		}
	}

	p.logger.Tracef("nothing to send\n")
	return nil, nil
}

func (p *peer) sync() {
	var nextSendTime *time.Time

	p.logger.Debugf("peer start sync\n")
	for {
		<-p.wakeUpChan

		p.logger.Tracef("peer.wakeUp\n")
		p.mutex.Lock()
		if !p.running {
			p.mutex.Unlock()
			p.logger.Tracef("peer is not running\n")
			p.stopped <- struct{}{}
			break
		}
		now := time.Now()
		if nextSendTime != nil && now.Before(*nextSendTime) {
			p.mutex.Unlock()
			p.logger.Tracef("peer.now=%v nextSendTime=%v\n", now.Format(time.StampMicro), nextSendTime.Format(time.StampMicro))
			continue
		}
		proto, msg := p.doSync()
		p.mutex.Unlock()

		if msg == nil {
			nextSendTime = nil
			continue
		}

		msgBS, err := msgCodec.MarshalToBytes(msg)
		if err != nil {
			p.logger.Panicf("peer.sync: %v\n", err)
		}
		p.logger.Debugf("sendMessage %v\n", msg)
		if err = p.ph.Unicast(proto, msgBS, p.id); err != nil {
			p.logger.Warnf("peer.sync: %v\n", err)
		}
		if configSendBPS < 0 {
			p.wakeUp()
			continue
		}
		if nextSendTime == nil {
			nextSendTime = &now
		}
		delta := time.Second * time.Duration(len(msgBS)) / configSendBPS
		next := nextSendTime.Add(delta)
		nextSendTime = &next
		waitTime := nextSendTime.Sub(now)
		p.logger.Tracef("msg size=%v delta=%v waitTime=%v\n", len(msgBS), delta, waitTime)
		if waitTime > time.Duration(0) {
			time.AfterFunc(waitTime, func() {
				p.wakeUp()
			})
		} else {
			p.wakeUp()
		}
	}
}

func (p *peer) stop() {
	p.running = false
	p.wakeUp()
	p.syncer.mutex.CallAfterUnlock(func() {
		<-p.stopped
	})
}

func (p *peer) wakeUp() {
	select {
	case p.wakeUpChan <- struct{}{}:
	default:
	}
}

type syncer struct {
	engine Engine
	logger log.Logger
	nm     module.NetworkManager
	bm     module.BlockManager
	mutex  *common.Mutex
	addr   module.Address
	fsm    fastsync.Manager

	ph            module.ProtocolHandler
	peers         []*peer
	timer         *time.Timer
	lastSendTime  time.Time
	running       bool
	fetchCanceler func() bool
}

func newSyncer(e Engine, logger log.Logger, nm module.NetworkManager, bm module.BlockManager, mutex *common.Mutex, addr module.Address) Syncer {
	fsm, err := fastsync.NewManager(nm, bm)
	if err != nil {
		return nil
	}
	fsm.StartServer()
	return &syncer{
		engine: e,
		logger: logger,
		nm:     nm,
		bm:     bm,
		mutex:  mutex,
		addr:   addr,
		fsm:    fsm,
	}
}

func (s *syncer) Start() error {
	var err error
	s.ph, err = s.nm.RegisterReactor("consensus.sync", s, syncerProtocols, configSyncerPriority)
	if err != nil {
		return err
	}

	peerIDs := s.nm.GetPeers()
	s.peers = make([]*peer, len(peerIDs))
	for i, peerID := range peerIDs {
		s.logger.Debugf("Start: starting peer list %v\n", common.HexPre(peerID.Bytes()))
		s.peers[i] = newPeer(s, peerID)
		go s.peers[i].sync()
	}

	s.sendRoundStateMessage()
	s.running = true
	return nil
}

func (s *syncer) OnReceive(sp module.ProtocolInfo, bs []byte,
	id module.PeerID) (bool, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.running {
		return false, nil
	}

	msg, err := unmarshalMessage(sp.Uint16(), bs)
	if err != nil {
		s.logger.Warnf("OnReceive: error=%+v\n", err)
		return false, err
	}
	s.logger.Debugf("OnReceive %v From:%v\n", msg, common.HexPre(id.Bytes()))
	if err := msg.verify(); err != nil {
		return false, err
	}
	var idx int
	switch m := msg.(type) {
	case *blockPartMessage:
		idx, err = s.engine.ReceiveBlockPartMessage(m, true)
		if idx < 0 && err != nil {
			return false, err
		}
		for _, p := range s.peers {
			// TODO check mask for optimization
			if p.peerRoundState != nil && p.Height == m.Height &&
				p.Height == p.engine.Height() && p.BlockPartsMask != nil {
				p.wakeUp()
			}
		}
	case *roundStateMessage:
		for _, p := range s.peers {
			if p.id.Equal(id) {
				p.setRoundState(&m.peerRoundState)
			}
		}
	case *voteListMessage:
		for i := 0; i < m.VoteList.Len(); i++ {
			s.engine.ReceiveVoteMessage(m.VoteList.Get(i), true)
		}
		rs := s.engine.GetRoundState()
		s.logger.Tracef("roundState=%+v\n", *rs)
	default:
		s.logger.Warnf("received unknown message %v\n", msg)
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *syncer) OnFailure(
	err error,
	pi module.ProtocolInfo,
	b []byte,
) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.logger.Debugf("OnFailure: subprotocol:%v err:%+v\n", pi, err)

	if !s.running {
		return
	}
}

func (s *syncer) OnJoin(id module.PeerID) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.logger.Debugf("OnJoin: %v\n", common.HexPre(id.Bytes()))

	if !s.running {
		return
	}

	for _, p := range s.peers {
		if p.id.Equal(id) {
			return
		}
	}
	p := newPeer(s, id)
	s.peers = append(s.peers, p)
	go p.sync()
	s.doSendRoundStateMessage(id)
}

func (s *syncer) OnLeave(id module.PeerID) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.logger.Debugf("OnLeave: %v\n", common.HexPre(id.Bytes()))

	if !s.running {
		return
	}

	for i, p := range s.peers {
		if p.id.Equal(id) {
			last := len(s.peers) - 1
			s.peers[i] = s.peers[last]
			s.peers[last] = nil
			s.peers = s.peers[:last]
			p.stop()
			return
		}
	}
}

func (s *syncer) OnEngineStepChange() {
	if !s.running {
		return
	}
	e := s.engine
	if e.Step() == stepPrecommitWait || e.Step() == stepCommit {
		for _, p := range s.peers {
			if p.peerRoundState != nil {
				p.wakeUp()
			}
		}
	}
	if e.Step() == stepPropose || e.Step() == stepCommit {
		s.sendRoundStateMessage()
	}
}

func (s *syncer) doSendRoundStateMessage(id module.PeerID) {
	e := s.engine
	msg := newRoundStateMessage()
	msg.peerRoundState = *e.GetRoundState()
	bs, err := msgCodec.MarshalToBytes(msg)
	if err != nil {
		s.logger.Panicf("doSendRoundStateMessage: %+v\n", err)
	}
	if id == nil {
		if len(s.peers) > 0 {
			s.logger.Debugf("neighborcastRoundState %v\n", msg)
			err = s.ph.Broadcast(protoRoundState, bs, module.BROADCAST_NEIGHBOR)
		}
	} else {
		s.logger.Debugf("sendRoundState %v To:%v\n", msg, common.HexPre(id.Bytes()))
		err = s.ph.Unicast(protoRoundState, bs, id)
	}
	if err != nil {
		s.logger.Warnf("doSendRoundStateMessage: %+v\n", err)
	}
}

func (s *syncer) sendRoundStateMessage() {
	s.doSendRoundStateMessage(nil)
	s.lastSendTime = time.Now()
	if s.timer != nil {
		s.timer.Stop()
	}

	var timer *time.Timer
	timer = time.AfterFunc(configRoundStateMessageInterval, func() {
		s.mutex.Lock()
		defer s.mutex.Unlock()

		if s.timer != timer {
			return
		}

		s.sendRoundStateMessage()
	})
	s.timer = timer
}

func (s *syncer) Stop() {
	for _, p := range s.peers {
		p.stop()
	}

	s.running = false

	if s.timer != nil {
		s.timer.Stop()
		s.timer = nil
	}
	s.fsm.StopServer()
	if s.fetchCanceler != nil {
		s.fetchCanceler()
		s.fetchCanceler = nil
	}
}

func (s *syncer) OnBlock(br fastsync.BlockResult) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.running {
		return
	}

	s.logger.Debugf("syncer.OnBlock %d\n", br.Block().Height())
	s.engine.ReceiveBlock(br)
}

func (s *syncer) OnEnd(err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.running {
		return
	}

	s.logger.Debugf("syncer.OnEnd %+v\n", err)
	s.fetchCanceler = nil
}
