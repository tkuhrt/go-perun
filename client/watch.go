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

package client

import (
	"context"

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
)

// Watch starts the channel watcher routine. It subscribes to RegisteredEvents
// on the adjudicator. If an event is registered, it is handled by making sure
// the latest state is registered and then all funds withdrawn to the receiver
// specified in the adjudicator that was passed to the channel.
//
// If handling failed, the watcher routine returns the respective error. It is
// the user's job to restart the watcher after the cause of the error got fixed.
func (c *Channel) Watch() error {
	log := c.Log().WithField("proc", "watcher")
	defer log.Info("Watcher returned.")

	ctx := c.Ctx()
	sub, err := c.adjudicator.SubscribeRegistered(ctx, c.Params())
	if err != nil {
		return errors.WithMessage(err, "subscribing to RegisteredEvents")
	}
	// nolint:errcheck
	defer sub.Close()
	// nolint:errcheck,gosec
	c.OnCloseAlways(func() { sub.Close() })

	// Wait for on-chain event
	reg := sub.Next()
	log.Infof("New RegisteredEvent: %v", reg)
	if reg == nil {
		err := sub.Err() // err might be nil if subscription got orderly closed
		log.Debugf("Subscription closed: %v", err)
		return errors.WithMessage(err, "subscription closed")
	}

	return errors.WithMessage(
		c.handleRegisteredEvent(ctx, reg),
		"handling RegisteredEvent")
}

// handleRegisteredEvent stores the passed RegisteredEvent to the machine and
// settles the channel.
func (c *Channel) handleRegisteredEvent(ctx context.Context, reg *channel.RegisteredEvent) error {
	log := c.Log().WithField("proc", "watcher")
	// Lock machine while registering is in progress.
	if !c.machMtx.TryLockCtx(ctx) {
		return errors.Errorf("locking machine mutex in time: %v", ctx.Err())
	}
	defer c.machMtx.Unlock()

	if c.machine.Phase() == channel.Withdrawn {
		// If a Settle call by the user caused this event, the channel will be
		// withdrawn already and we're done.
		log.Debug("Channel already withdrawn.")
		return nil
	}

	if err := c.machine.SetRegistered(ctx, reg); err != nil {
		return errors.WithMessage(err, "setting machine to Registered phase")
	}

	return c.settle(ctx)
}

// Settle settles the channel: it is made sure that the current state is
// registered and the final balance withdrawn. This call blocks until the
// channel has been successfully withdrawn.
func (c *Channel) Settle(ctx context.Context) error {
	if !c.machMtx.TryLockCtx(ctx) {
		return errors.Errorf("locking machine mutex in time: %v", ctx.Err())
	}
	defer c.machMtx.Unlock()
	// Wrap the context to make sure that the settle call stops as soon as the
	// channel controller is closed.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	c.OnClose(cancel)
	return c.settle(ctx)
}

// settle makes sure that the current state is registered and the final balance
// withdrawn. This call blocks until the channel has been successfully
// withdrawn.
//
// The caller is expected to have locked the channel mutex.
func (c *Channel) settle(ctx context.Context) error {
	ver, reg := c.machine.State().Version, c.machine.Registered()
	// If the machine is at least in phase Registered, reg shouldn't be nil. We
	// still catch this case to be future proof.
	if c.machine.Phase() < channel.Registered || reg == nil || reg.Version < ver {
		if reg != nil && reg.Version < ver {
			c.Log().Warnf("Lower version %d (< %d) registered, refuting...", reg.Version, ver)
		}
		if err := c.register(ctx); err != nil {
			return errors.WithMessage(err, "registering")
		}
		c.Log().Info("Channel state registered.")
	}

	if reg = c.machine.Registered(); !reg.Timeout.IsElapsed(ctx) {
		if c.machine.State().IsFinal {
			c.Log().Warnf(
				"Unexpected withdrawal timeout while settling final state. Waiting until %v.",
				reg.Timeout)
		} else {
			c.Log().Infof("Waiting until %v for withdrawal.", reg.Timeout)
		}

		if err := reg.Timeout.Wait(ctx); err != nil {
			return errors.WithMessage(err, "waiting for timeout")
		}
	}

	if err := c.withdraw(ctx); err != nil {
		return errors.WithMessage(err, "withdrawing")
	}
	c.Log().Info("Withdrawal successful.")
	c.wallet.DecrementUsage(c.machine.Account().Address())
	return nil
}

// register calls Regsiter on the adjudicator with the current channel state and
// progresses the machine phases. When successful, the resulting RegisteredEvent
// is saved to the phase machine.
//
// The caller is expected to have locked the channel mutex.
func (c *Channel) register(ctx context.Context) error {
	if err := c.machine.SetRegistering(ctx); err != nil {
		return err
	}

	reg, err := c.adjudicator.Register(ctx, c.machine.AdjudicatorReq())
	if err != nil {
		return errors.WithMessage(err, "calling Register")
	}
	if ver := c.machine.State().Version; reg.Version != ver {
		return errors.Errorf(
			"unexpected version %d registered, expected %d", reg.Version, ver)
	}

	return c.machine.SetRegistered(ctx, reg)
}

// withdraw calls Withdraw on the adjudicator with the current channel state and
// progresses the machine phases.
//
// The caller is expected to have locked the channel mutex.
func (c *Channel) withdraw(ctx context.Context) error {
	if err := c.machine.SetWithdrawing(ctx); err != nil {
		return err
	}

	req := c.machine.AdjudicatorReq()
	if err := c.adjudicator.Withdraw(ctx, req); err != nil {
		return errors.WithMessage(err, "calling Withdraw")
	}

	return c.machine.SetWithdrawn(ctx)
}
