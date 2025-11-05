package api

import (
	"errors"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	WEEKDAY = "w"
	DAY     = "d"
	MONTH   = "m"
	YEAR    = "y"
)

var repeatTypes = []string{WEEKDAY, DAY, MONTH, YEAR}

// NextDate return format - YYYYMMDD
// NextDate(now, "20240229", "y") = 20250301
// NextDate(now, "20240113", "d 7") = 20240120
// NextDate(now, "20240116", "m 16,5") = 20240205
// NextDate(now, "20240201", "m -1,18") = 20240218
func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	parsedDate, err := time.Parse(DATEFORMAT, dstart)
	if err != nil {
		return "", err
	}

	parsedRepeat, err := parseRepeat(repeat)
	if err != nil {
		return "", err
	}

	if parsedRepeat.rType == DAY {
		for {
			parsedDate = parsedDate.AddDate(0, 0, parsedRepeat.days[0])
			if afterNow(parsedDate, now) {
				break
			}
		}
	}

	if parsedRepeat.rType == YEAR {
		for {
			parsedDate = parsedDate.AddDate(1, 0, 0)
			if afterNow(parsedDate, now) {
				break
			}
		}
	}

	if parsedRepeat.rType == WEEKDAY {
		var dayOfWeek [8]bool
		for _, day := range parsedRepeat.days {
			dayOfWeek[day] = true
		}

		for {
			parsedDate = parsedDate.AddDate(0, 0, 1)
			if afterNow(parsedDate, now) && dayOfWeek[parsedDate.Weekday()] {
				break
			}
		}
	}

	if parsedRepeat.rType == MONTH {
		var dayOfMonth [32]bool
		var months [13]bool
		for _, day := range parsedRepeat.days {
			dayOfMonth[day] = true
		}
		for _, month := range parsedRepeat.months {
			months[month] = true
		}

		for {
			parsedDate = parsedDate.AddDate(0, 0, 1)
			if afterNow(parsedDate, now) {
				day := parsedDate.Day()
				month := parsedDate.Month()
				if dayOfMonth[day] && months[month] {
					break
				}
			}
		}
	}

	return parsedDate.Format(DATEFORMAT), nil
}

type Parsed struct {
	rType  string
	days   []int
	months []int
}

func parseRepeat(repeat string) (*Parsed, error) {
	result := new(Parsed)

	if repeat == "" {
		return result, errors.New("repeat is empty")
	}

	params := strings.Split(repeat, " ")
	rType := params[0]
	rParams := params[1:]

	if rType == "" || !slices.Contains(repeatTypes, rType) {
		return result, errors.New("invalid repeat type")
	}

	// d - day - max 400
	// example - d 1, d 7, d 60
	if rType == DAY {
		if len(rParams) == 0 || len(rParams) > 1 {
			return result, errors.New("invalid day repeat params")
		}

		days, err := strconv.Atoi(rParams[0])
		if err != nil {
			return result, err
		}

		if days < 0 || days > 400 {
			return result, errors.New("invalid day repeat params")
		}
		result.days = []int{days}
	}

	if rType == YEAR {
		if len(rParams) > 0 {
			return result, errors.New("invalid year repeat params")
		}
	}

	// w - weak
	// example - w 7; w 1,4,5; w 2,3;
	if rType == WEEKDAY {
		if len(rParams) == 0 {
			return result, errors.New("invalid weekday repeat params")
		}

		daysArr := strings.Split(rParams[0], ",")

		if len(daysArr) == 0 || len(daysArr) > 7 {
			return result, errors.New("invalid weekday repeat params")
		}

		parsedDaysArr := []int{}

		for _, day := range daysArr {
			count, err := strconv.Atoi(day)
			if err != nil {
				return result, err
			}

			if count < 1 || count > 7 {
				return result, errors.New("invalid weekday repeat params")
			}

			parsedDaysArr = append(parsedDaysArr, count)
		}

		result.days = parsedDaysArr
	}

	// m - month date
	// m D,D M,M
	// example - m 4; m 1,15,25; m -1; m -2; m 3 1,3,6; m 1,-1 2,8
	if rType == MONTH {
		if len(rParams) == 0 || len(rParams) > 2 {
			return result, errors.New("invalid month repeat params")
		}
		daysArr := strings.Split(rParams[0], ",")
		monthsArr := strings.Split(rParams[1], ",")

		parsedDaysArr := []int{}
		for _, day := range daysArr {
			count, err := strconv.Atoi(day)
			if err != nil {
				return result, err
			}

			if count < -31 || count > 31 {
				return result, errors.New("invalid month repeat params")
			}

			parsedDaysArr = append(parsedDaysArr, count)
		}

		parsedMonthsArr := []int{}
		for _, month := range monthsArr {
			count, err := strconv.Atoi(month)
			if err != nil {
				return result, err
			}

			if count < 1 || count > 12 {
				return result, errors.New("invalid month repeat params")
			}

			parsedMonthsArr = append(parsedMonthsArr, count)
		}

		result.months = parsedMonthsArr
		result.days = parsedDaysArr
	}

	result.rType = rType

	return result, nil
}

func afterNow(date, now time.Time) bool {
	return date.After(now)
}

func nextDateHandler(w http.ResponseWriter, req *http.Request) {
	nowStr := req.FormValue("now")
	date := req.FormValue("date")
	repeat := req.FormValue("repeat")

	var parsedNow time.Time
	if nowStr != "" {
		var err error
		parsedNow, err = time.Parse(DATEFORMAT, nowStr)
		if err != nil {
			http.Error(w, "incorrect now", http.StatusBadRequest)
			return
		}
	} else {
		parsedNow = time.Now() // Используем текущее время, если now не указано
	}

	nextDate, err := NextDate(parsedNow, date, repeat)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	io.WriteString(w, nextDate)
}
