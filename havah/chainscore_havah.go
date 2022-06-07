/*
 * Copyright 2020 ICON Foundation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package havah

import (
	"math/big"

	"github.com/icon-project/goloop/common"
	"github.com/icon-project/goloop/common/errors"
	"github.com/icon-project/goloop/havah/hvh"
	"github.com/icon-project/goloop/havah/hvhmodule"
	"github.com/icon-project/goloop/havah/hvhutils"
	"github.com/icon-project/goloop/module"
	"github.com/icon-project/goloop/service/scoreresult"
)

func (s *chainScore) getExtensionState() (*hvh.ExtensionStateImpl, error) {
	es := s.cc.GetExtensionState()
	if es == nil {
		err := errors.Errorf("ExtensionState is nil")
		return nil, s.toScoreResultError(scoreresult.UnknownFailureError, err)
	}
	esi := es.(*hvh.ExtensionStateImpl)
	esi.SetLogger(hvhutils.NewLogger(s.cc.Logger()))
	return esi, nil
}

func (s *chainScore) toScoreResultError(code errors.Code, err error) error {
	msg := err.Error()
	if logger := s.cc.Logger(); logger != nil {
		logger = hvhutils.NewLogger(logger)
		logger.Errorf(msg)
	}
	return code.Wrap(err, msg)
}

func (s *chainScore) newCallContext() hvhmodule.CallContext {
	return hvh.NewCallContext(s.cc, s.from)
}

func (s *chainScore) Ex_getUSDTPrice() (*big.Int, error) {
	es, err := s.getExtensionState()
	if err != nil {
		return nil, err
	}
	return es.GetUSDTPrice()
}

func (s *chainScore) Ex_setUSDTPrice(price *common.HexInt) error {
	es, err := s.getExtensionState()
	if err != nil {
		return err
	}
	return es.SetUSDTPrice(price.Value())
}

func (s *chainScore) Ex_getIssueInfo() (map[string]interface{}, error) {
	es, err := s.getExtensionState()
	if err != nil {
		return nil, err
	}
	return es.GetIssueInfo(s.newCallContext())
}

func (s *chainScore) Ex_startRewardIssue(height *common.HexInt) error {
	es, err := s.getExtensionState()
	if err != nil {
		return err
	}
	startBH := height.Int64()
	if startBH <= s.cc.BlockHeight() {
		return scoreresult.RevertedError.New("Invalid height")
	}
	return es.StartRewardIssue(s.newCallContext(), startBH)
}

func (s *chainScore) Ex_addPlanetManager(address module.Address) error {
	es, err := s.getExtensionState()
	if err != nil {
		return err
	}
	return es.AddPlanetManager(address)
}

func (s *chainScore) Ex_removePlanetManager(address module.Address) error {
	es, err := s.getExtensionState()
	if err != nil {
		return err
	}
	return es.RemovePlanetManager(address)
}

func (s *chainScore) Ex_isPlanetManager(address module.Address) (bool, error) {
	es, err := s.getExtensionState()
	if err != nil {
		return false, err
	}
	return es.IsPlanetManager(address)
}

func (s *chainScore) Ex_registerPlanet(
	id int64,
	isPrivate, isCompany bool, owner module.Address, usdt, price *common.HexInt) error {
	es, err := s.getExtensionState()
	if err != nil {
		return err
	}
	return es.RegisterPlanet(s.newCallContext(), id, isPrivate, isCompany, owner, usdt.Value(), price.Value())
}

func (s *chainScore) Ex_unregisterPlanet(id int64) error {
	es, err := s.getExtensionState()
	if err != nil {
		return err
	}
	return es.UnregisterPlanet(s.newCallContext(), id)
}

func (s *chainScore) Ex_setPlanetOwner(id int64, owner module.Address) error {
	es, err := s.getExtensionState()
	if err != nil {
		return err
	}
	return es.SetPlanetOwner(s.newCallContext(), id, owner)
}

func (s *chainScore) Ex_getPlanetInfo(id int64) (map[string]interface{}, error) {
	es, err := s.getExtensionState()
	if err != nil {
		return nil, err
	}
	return es.GetPlanetInfo(s.newCallContext(), id)
}

func (s *chainScore) Ex_reportPlanetWork(id int64) error {
	es, err := s.getExtensionState()
	if err != nil {
		return err
	}
	return es.ReportPlanetWork(s.newCallContext(), id)
}

func (s *chainScore) Ex_claimPlanetReward(ids []int64) error {
	es, err := s.getExtensionState()
	if err != nil {
		return err
	}
	return es.ClaimPlanetReward(s.newCallContext(), ids)
}