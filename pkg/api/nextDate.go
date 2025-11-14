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
		var dayOfWeek [7]bool
		for _, day := range parsedRepeat.days {
			idx := day % 7 // 7 -> 0 (воскресенье), 1..6 -> 1..6
			dayOfWeek[idx] = true
		}

		for {
			if afterNow(parsedDate, now) {
				weekdayIdx := int(parsedDate.Weekday())
				if dayOfWeek[weekdayIdx] {
					break
				}
			}
			parsedDate = parsedDate.AddDate(0, 0, 1)
		}
	}

	if parsedRepeat.rType == MONTH {
		var dayOfMonth [32]bool
		lastDay := false
		secondLastDay := false
		var months [13]bool
		allMonths := len(parsedRepeat.months) == 0

		for _, day := range parsedRepeat.days {
			switch {
			case day > 0:
				dayOfMonth[day] = true
			case day == -1:
				lastDay = true
			case day == -2:
				secondLastDay = true
			}
		}

		if allMonths {
			for i := 1; i <= 12; i++ {
				months[i] = true
			}
		} else {
			for _, month := range parsedRepeat.months {
				months[month] = true
			}
		}

		for {
			if afterNow(parsedDate, now) {
				monthIdx := parsedDate.Month()
				if months[monthIdx] && validMonthlyDate(parsedDate, &dayOfMonth, lastDay, secondLastDay) {
					break
				}
			}
			parsedDate = parsedDate.AddDate(0, 0, 1)
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

	params := strings.Fields(repeat)
	if len(params) == 0 {
		return result, errors.New("repeat is empty")
	}
	rType := params[0]
	rParams := params[1:]

	if rType == "" || !slices.Contains(repeatTypes, rType) {
		return result, errors.New("invalid repeat type")
	}

	// d - day - max 400
	// example - d 1, d 7, d 60
	if rType == DAY {
		if len(rParams) != 1 {
			return result, errors.New("invalid day repeat params")
		}

		days, err := strconv.Atoi(rParams[0])
		if err != nil {
			return result, err
		}

		if days <= 0 || days > 400 {
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
		if len(rParams) != 1 {
			return result, errors.New("invalid weekday repeat params")
		}

		daysArr := strings.Split(rParams[0], ",")

		if len(daysArr) == 0 || len(daysArr) > 7 {
			return result, errors.New("invalid weekday repeat params")
		}

		parsedDaysArr := make([]int, 0, len(daysArr))

		for _, day := range daysArr {
			day = strings.TrimSpace(day)
			if day == "" {
				return result, errors.New("invalid weekday repeat params")
			}
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
		if len(daysArr) == 0 {
			return result, errors.New("invalid month repeat params")
		}
		monthsArr := make([]string, 0)

		if len(rParams) == 2 {
			monthsArr = strings.Split(rParams[1], ",")
		}

		parsedDaysArr := make([]int, 0, len(daysArr))
		for _, day := range daysArr {
			day = strings.TrimSpace(day)
			if day == "" {
				return result, errors.New("invalid month repeat params")
			}
			count, err := strconv.Atoi(day)
			if err != nil {
				return result, err
			}

			if count == 0 || count > 31 {
				return result, errors.New("invalid month repeat params")
			}

			if count < 0 && count != -1 && count != -2 {
				return result, errors.New("invalid month repeat params")
			}

			parsedDaysArr = append(parsedDaysArr, count)
		}

		parsedMonthsArr := make([]int, 0, len(monthsArr))
		if len(monthsArr) != 0 {
			for _, month := range monthsArr {
				month = strings.TrimSpace(month)
				if month == "" {
					return result, errors.New("invalid month repeat params")
				}
				count, err := strconv.Atoi(month)
				if err != nil {
					return result, err
				}

				if count < 1 || count > 12 {
					return result, errors.New("invalid month repeat params")
				}

				parsedMonthsArr = append(parsedMonthsArr, count)
			}
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

func isLastDayOfMonth(date time.Time) bool {
	return date.AddDate(0, 0, 1).Month() != date.Month()
}

func isSecondLastDayOfMonth(date time.Time) bool {
	return date.AddDate(0, 0, 2).Month() != date.Month()
}

func validMonthlyDate(date time.Time, dayOfMonth *[32]bool, lastDay, secondLastDay bool) bool {
	day := date.Day()
	if dayOfMonth[day] {
		return true
	}
	if lastDay && isLastDayOfMonth(date) {
		return true
	}
	if secondLastDay && isSecondLastDayOfMonth(date) {
		return true
	}
	return false
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
