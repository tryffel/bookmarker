/*
 *   Copyright 2019 Tero Vierimaa
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package ui

import (
	"testing"
	"time"
)

func TestTimeSince(t *testing.T) {
	now := time.Now().Unix()

	tests := []struct {
		name string
		ts   time.Time
		want string
	}{
		{
			ts:   time.Unix(now-30, 0),
			want: "30 seconds",
		},
		{
			ts:   time.Unix(now-200, 0),
			want: "3 minutes",
		},
		{
			ts:   time.Unix(now-timeHourSeconds-10, 0),
			want: "1 hour",
		},
		{
			ts:   time.Unix(now-timeHourSeconds*2-10, 0),
			want: "2 hours",
		},
		{
			ts:   time.Unix(now-timeDaySeconds*1-10, 0),
			want: "1 day",
		},
		{
			ts:   time.Unix(now-timeDaySeconds*2-10, 0),
			want: "2 days",
		},
		{
			ts:   time.Unix(now-timeWeekSeconds*2-10, 0),
			want: "2 weeks",
		},
		{
			ts:   time.Unix(now-timeMonthSeconds*2-10, 0),
			want: "2 months",
		},
		{
			ts:   time.Unix(now-timeYearSeconds*2-10, 0),
			want: "2 years",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TimeSince(tt.ts); got != tt.want {
				t.Errorf("TimeSince() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShortTimeSince(t *testing.T) {
	now := time.Now().Unix()

	tests := []struct {
		name string
		ts   time.Time
		want string
	}{
		{
			ts:   time.Unix(now-30, 0),
			want: "now",
		},
		{
			ts:   time.Unix(now-200, 0),
			want: "3 minutes ago",
		},
		{
			ts:   time.Unix(now-timeHourSeconds*5, 0),
			want: "5 hours ago",
		},
		{
			ts:   time.Unix(now-timeHourSeconds*8, 0),
			want: "today",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShortTimeSince(tt.ts); got != tt.want {
				t.Errorf("ShortTimeSince() = %v, want %v", got, tt.want)
			}
		})
	}
}
