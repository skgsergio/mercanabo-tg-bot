// Copyright (c) 2020 Sergio Conde skgsergio@gmail.com
//
// This program is free software: you can redistribute it and/or modify it under
// the terms of the GNU General Public License as published by the Free Software
// Foundation, version 3.
//
// This program is distributed in the hope that it will be useful, but WITHOUT ANY
// WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A
// PARTICULAR PURPOSE. See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along with
// this program. If not, see <https://www.gnu.org/licenses/>.
//
// SPDX-License-Identifier: GPL-3.0-only

package main

import (
	"strconv"
)

// parseUint8 parses a string and converts it to uint8
func parseUint8(s string) (uint8, error) {
	u, err := strconv.ParseUint(s, 10, 8)

	if err != nil {
		return 0, err
	}

	return uint8(u), nil
}

// parseUint32 parses a string and converts it to uint32
func parseUint32(s string) (uint32, error) {
	u, err := strconv.ParseUint(s, 10, 32)

	if err != nil {
		return 0, err
	}

	return uint32(u), nil
}

// parseInt64 parses a string and converts it to int64
func parseInt64(s string) (int64, error) {
	i, err := strconv.ParseInt(s, 10, 32)

	if err != nil {
		return 0, err
	}

	return int64(i), nil
}

// maxUint32 returns the maximum value from two uint32 values
func maxUint32(x, y uint32) uint32 {
	if x < y {
		return y
	}

	return x
}

// minUint32 returns the minimum value from two uint32 values
func minUint32(x, y uint32) uint32 {
	if x > y {
		return y
	}

	return x
}
