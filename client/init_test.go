// Copyright 2019 - See NOTICE file for copyright holders.
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
	"math/rand"

	"github.com/sirupsen/logrus"

	"perun.network/go-perun/apps/payment"
	_ "perun.network/go-perun/backend/sim" // backend init
	plogrus "perun.network/go-perun/log/logrus"
	pkgtest "perun.network/go-perun/pkg/test"
	wallettest "perun.network/go-perun/wallet/test"
)

// This file initializes the blockchain and logging backend for both, whitebox
// and blackbox tests (files *_test.go in packages client and client_test).
func init() {
	plogrus.Set(logrus.WarnLevel, &logrus.TextFormatter{ForceColors: true})

	// Tests of package client use the payment app for now...
	rng := rand.New(rand.NewSource(pkgtest.Seed("test app def")))
	appDef := wallettest.NewRandomAddress(rng)
	payment.SetAppDef(appDef) // payment app address has to be set once at startup
}
