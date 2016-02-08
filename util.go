package adt

import "time"

var (
	julianBase          = time.Date(-4713, time.November, 24, 12, 0, 0, 0, time.UTC)
	julianMidnightLocal = time.Date(-4713, time.November, 24, 0, 0, 0, 0, time.Local)
)

func adtDatetimeToTime(date, ms int32) time.Time {
	d := adtDateToTime(date)
	d = d.Add(time.Duration(ms) * time.Millisecond)
	return d
}

func adtDateToTime(i int32) time.Time {
	return julianMidnightLocal.AddDate(0, 0, int(i))
}
