// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package wire

import (
	"encoding/binary"
	"io"

	"github.com/pkg/errors"
)

// Int32 is a serializable network 32 bit integer.
type Int32 int32

func (i32 *Int32) Decode(reader io.Reader) error {
	buf := make([]byte, 4)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return errors.Wrap(err, "failed to read int32")
	}
	*i32 = Int32(binary.LittleEndian.Uint32(buf))
	return nil
}

func (i32 Int32) Encode(writer io.Writer) error {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(i32))
	if _, err := writer.Write(buf); err != nil {
		return errors.Wrap(err, "failed to write int32")
	}
	return nil
}