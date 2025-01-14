// Copyright 2020 - See NOTICE file for copyright holders.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package channel_test

import (
	"context"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ethchannel "perun.network/go-perun/backend/ethereum/channel"
	"perun.network/go-perun/backend/ethereum/channel/test"
	ethwallet "perun.network/go-perun/backend/ethereum/wallet"
	ethwallettest "perun.network/go-perun/backend/ethereum/wallet/test"
	"perun.network/go-perun/channel"
	channeltest "perun.network/go-perun/channel/test"
	pkgtest "perun.network/go-perun/pkg/test"
)

const defaultTxTimeout = 2 * time.Second

func signState(t *testing.T, accounts []*ethwallet.Account, params *channel.Params, state *channel.State) channel.Transaction {
	// Sign valid state.
	sigs := make([][]byte, len(accounts))
	for i := range accounts {
		sig, err := channel.Sign(accounts[i], params, state)
		assert.NoError(t, err, "Sign should not return error")
		sigs[i] = sig
	}
	return channel.Transaction{
		State: state,
		Sigs:  sigs,
	}
}

func TestSubscribeRegistered(t *testing.T) {
	rng := pkgtest.Prng(t)
	// create test setup
	s := test.NewSetup(t, rng, 1)
	// create valid state and params
	params, state := channeltest.NewRandomParamsAndState(rng, channeltest.WithChallengeDuration(uint64(100*time.Second)), channeltest.WithParts(s.Parts...), channeltest.WithAssets((*ethchannel.Asset)(&s.Asset)), channeltest.WithIsFinal(false))
	// Set up subscription
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	registered, err := s.Adjs[0].SubscribeRegistered(ctx, params)
	require.NoError(t, err, "Subscribing to valid params should not error")
	// we need to properly fund the channel
	txCtx, txCancel := context.WithTimeout(context.Background(), defaultTxTimeout)
	defer txCancel()
	// fund the contract
	reqFund := channel.FundingReq{
		Params: params,
		State:  state,
		Idx:    channel.Index(0),
	}
	require.NoError(t, s.Funders[0].Fund(txCtx, reqFund), "funding should succeed")
	// Now test the register function
	tx := signState(t, s.Accs, params, state)
	req := channel.AdjudicatorReq{
		Params: params,
		Acc:    s.Accs[0],
		Idx:    channel.Index(0),
		Tx:     tx,
	}
	event, err := s.Adjs[0].Register(txCtx, req)
	assert.NoError(t, err, "Registering state should succeed")
	assert.Equal(t, event, registered.Next(), "Events should be equal")
	assert.NoError(t, registered.Close(), "Closing event channel should not error")
	assert.Nil(t, registered.Next(), "Next on closed channel should produce nil")
	assert.NoError(t, registered.Err(), "Closing should produce no error")
	// Setup a new subscription
	registered2, err := s.Adjs[0].SubscribeRegistered(ctx, params)
	assert.NoError(t, err, "registering two subscriptions should not fail")
	assert.Equal(t, event, registered2.Next(), "Events should be equal")
	assert.NoError(t, registered2.Close(), "Closing event channel should not error")
	assert.Nil(t, registered2.Next(), "Next on closed channel should produce nil")
	assert.NoError(t, registered2.Err(), "Closing should produce no error")
}

func TestValidateAdjudicator(t *testing.T) {
	// Test setup
	rng := pkgtest.Prng(t)
	s := test.NewSimSetup(rng)

	t.Run("no_adj_code", func(t *testing.T) {
		randomAddr := (common.Address)(ethwallettest.NewRandomAddress(rng))
		ctx, cancel := context.WithTimeout(context.Background(), defaultTxTimeout)
		defer cancel()
		require.True(t, ethchannel.IsContractBytecodeError(ethchannel.ValidateAdjudicator(ctx, *s.CB, randomAddr)))
	})
	t.Run("incorrect_adj_code", func(t *testing.T) {
		randomAddr := (common.Address)(ethwallettest.NewRandomAddress(rng))
		ctx, cancel := context.WithTimeout(context.Background(), defaultTxTimeout)
		defer cancel()
		incorrectCodeAddr, err := ethchannel.DeployETHAssetholder(ctx, *s.CB, randomAddr)
		require.NoError(t, err)
		require.True(t, ethchannel.IsContractBytecodeError(ethchannel.ValidateAdjudicator(ctx, *s.CB, incorrectCodeAddr)))
	})
	t.Run("correct_adj_code", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), defaultTxTimeout)
		defer cancel()
		adjudicatorAddr, err := ethchannel.DeployAdjudicator(ctx, *s.CB)
		require.NoError(t, err)
		t.Logf("adjudicator address is %v", adjudicatorAddr)
		require.NoError(t, ethchannel.ValidateAdjudicator(ctx, *s.CB, adjudicatorAddr))
	})
}
