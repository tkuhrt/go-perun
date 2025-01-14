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

	"github.com/stretchr/testify/assert"

	ethchannel "perun.network/go-perun/backend/ethereum/channel"
	"perun.network/go-perun/backend/ethereum/channel/test"
)

func TestBlockTimeout_IsElapsed(t *testing.T) {
	assert := assert.New(t)
	sb := test.NewSimulatedBackend()
	bt := ethchannel.NewBlockTimeout(sb, 100)

	// We use nil contexts in the following because we're working with a simulated
	// blockchain, which ignores the ctx.
	for i := 0; i < 10; i++ {
		assert.False(bt.IsElapsed(nil))
		sb.Commit() // advances block time by 10 sec
	}
	assert.True(bt.IsElapsed(nil))
}

func TestBlockTimeout_Wait(t *testing.T) {
	sb := test.NewSimulatedBackend()
	bt := ethchannel.NewBlockTimeout(sb, 100)

	t.Run("cancelWait", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		wait := make(chan error)
		go func() {
			wait <- bt.Wait(ctx)
		}()

		cancel()
		select {
		case err := <-wait:
			assert.Error(t, err)
		case <-time.After(100 * time.Millisecond):
			t.Error("expected Wait to return")
		}
	})

	t.Run("normalWait", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		wait := make(chan error)
		go func() {
			wait <- bt.Wait(ctx)
		}()

		for i := 0; i < 10; i++ {
			select {
			case err := <-wait:
				t.Error("Wait returned before timeout with error", err)
			default: // Wait shouldn't return before the timeout is reached
			}
			sb.Commit() // advances block time by 10 sec
		}
		select {
		case err := <-wait:
			assert.NoError(t, err)
		case <-time.After(100 * time.Millisecond):
			t.Error("expected Wait to return after timeout is reached")
		}
	})
}
