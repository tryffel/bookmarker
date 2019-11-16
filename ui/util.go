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
	"fmt"
	"time"
)

const (
	timeHourSeconds  = 3600
	timeDaySeconds   = timeHourSeconds * 24
	timeWeekSeconds  = timeDaySeconds * 7
	timeMonthSeconds = timeDaySeconds * 31
	timeYearSeconds  = timeDaySeconds * 365
)

//PrettyTime shows time in relative / pretty mode.
func TimeSince(t time.Time) string {
	diff := time.Since(t)

	unit := ""
	var count int64 = 0

	if diff < time.Minute {
		unit = "second"
		count = int64(diff.Seconds())
	} else if diff < time.Hour {
		unit = "minute"
		count = int64(diff.Seconds() / 60)
	} else if diff < time.Hour*24 {
		unit = "hour"
		count = int64(diff.Seconds() / timeHourSeconds)
	} else if diff < time.Hour*24*7 {
		unit = "day"
		count = int64(diff.Seconds() / (timeDaySeconds))
	} else if diff < time.Hour*24*31 {
		unit = "week"
		count = int64(diff.Seconds() / (timeWeekSeconds))
	} else if diff < (time.Hour * 24 * 365) {
		unit = "month"
		count = int64(diff.Seconds() / (timeMonthSeconds))
	} else {
		unit = "year"
		count = int64(diff.Seconds() / (timeYearSeconds))
	}

	text := fmt.Sprintf("%d %s", count, unit)
	if count > 1 {
		text += "s"
	}
	return text
}

// ShortTimeSince returns human readable format for time.
func ShortTimeSince(t time.Time) string {
	today := time.Now().YearDay()
	day := t.YearDay()

	diff := time.Since(t)
	if diff < time.Second*60 {
		return "now"
	}
	if diff < 6*time.Hour {
		return TimeSince(t) + " ago"
	}
	if today == day {
		return "today"
	}
	if today-1 == day {
		return "yesterday"
	} else {
		return TimeSince(t) + " ago"
	}
}
