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

package test

import (
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"hash/fnv"
	"math/rand"
	"os"
	"strconv"
	"time"
)

const envTestSeed = "GOTESTSEED"

var rootSeed int64

func init() {
	rootSeed = genRootSeed()
	fmt.Printf("pkg/test: using rootSeed %d\n", rootSeed)
}

func genRootSeed() (rootSeed int64) {
	if val, ok := os.LookupEnv(envTestSeed); ok {
		rootSeed, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			panic("Could not parse GOTESTSEED as int64")
		}
		return rootSeed
	}
	return time.Now().UnixNano()
}

// Prng returns a pseudo-RNG that is seeded with the output of the `Seed`
// function by passing it `t.Name()`.
// Use it in tests with: rng := pkgtest.Prng(t).
func Prng(t interface{ Name() string }, args ...interface{}) *rand.Rand {
	return rand.New(rand.NewSource(Seed(t.Name(), args...)))
}

// Seed generates a seed that is dependent on the rootSeed and the passed
// seed argument.
// To fix this seed, set the GOTESTSEED environment variable.
// Example: GOTESTSEED=123 go test ./...
// Does not work with function pointers or structs without public fields.
func Seed(seed string, args ...interface{}) int64 {
	hasher := fnv.New64a()
	enc := gob.NewEncoder(hasher)
	if err := enc.Encode(seed); err != nil {
		panic("Could not gob-encode seed")
	}
	for _, arg := range args {
		if err := enc.Encode(arg); err != nil {
			panic(fmt.Sprintf("Could not gob-encode value: %v", err))
		}
	}
	if err := binary.Write(hasher, binary.LittleEndian, rootSeed); err != nil {
		panic("Could not hash the root seed")
	}
	return int64(hasher.Sum64())
}
