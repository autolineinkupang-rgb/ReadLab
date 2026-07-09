package ticket

import "time"

const MakassarTimezone = "Asia/Makassar"

var makassarLoc = func() *time.Location {
	loc, err := time.LoadLocation(MakassarTimezone)
	if err != nil {
		return time.UTC
	}
	return loc
}()

func MakassarNow() time.Time {
	return time.Now().In(makassarLoc)
}

func TodayMakassarBoundary() time.Time {
	now := MakassarNow()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, makassarLoc)
}
