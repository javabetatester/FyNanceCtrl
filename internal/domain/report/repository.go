package report

import (
	"time"

	"github.com/oklog/ulid/v2"
)

type ReportRepository interface {
	GetMonthlyReport(userID ulid.ULID, month, year int) (*MonthlyReport, error)
	GetYearlyReport(userID ulid.ULID, year int) (*YearlyReport, error)
	GetCategoryReport(userID ulid.ULID, categoryID ulid.ULID, startDate, endDate time.Time) (*CategoryReport, error)
}
