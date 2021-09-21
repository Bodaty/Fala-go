// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build amd64 && !purego && gc
// +build amd64,!purego,gc

package keccak

// this is called when AVX512 is unavailable
func KeccakF1600(a *[25]uint64)