// Copyright (c) 2025 Chakib Ben Ziane <contact@blob42.xyz>  and [`gosuki` contributors](https://github.com/blob42/gosuki/graphs/contributors).
// All rights reserved.
//
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This file is part of GoSuki.
//
// GoSuki is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// GoSuki is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with gosuki.  If not, see <http://www.gnu.org/licenses/>.
package config

import (
	"fmt"
	"reflect"
	"time"

	"github.com/hako/durafmt"
	"github.com/mitchellh/mapstructure"
)

var (
	mapDecoderConfig = &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			StringToDurationFunc(),
		),
	}
)

// StringToDurationHookFunc returns a mapstructure.DecodeHookFunc that converts
// strings to time.Duration using https://github.com/hako/durafmt
func StringToDurationFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data any) (any, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(time.Duration(0)) {
			return data, nil
		}

		// Convert it by parsing
		duration, err := durafmt.ParseString(data.(string))
		if err != nil {
			return time.Duration(0), fmt.Errorf("failed parsing duration %v", data)
		}

		return duration.Duration(), nil
	}
}
