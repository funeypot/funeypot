// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_truncateString(t *testing.T) {
	type args struct {
		s   string
		max int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "regular",
			args: args{
				s:   "hello world",
				max: 10,
			},
			want: "hello w...",
		},
		{
			name: "too short",
			args: args{
				s:   "hello world",
				max: 1,
			},
			want: "h",
		},
		{
			name: "empty",
			args: args{
				s:   "",
				max: 10,
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, truncateString(tt.args.s, tt.args.max))
		})
	}
}
