package models

type DayProgress struct {
	Day       string `json:"day"`
	Date      string `json:"date"`
	Completed int    `json:"completed"`
	Total     int    `json:"total"`
	Progress  int    `json:"progress"`
}

type DashboardResponse struct {
	TodayDate             string        `json:"todayDate"`
	TodayProgress         int           `json:"todayProgress"`
	TodayPlanned          int           `json:"todayPlanned"`
	TodayCompleted        int           `json:"todayCompleted"`
	TodayRemaining        int           `json:"todayRemaining"`
	TotalTasks            int           `json:"totalTasks"`
	Completed             int           `json:"completed"`
	Pending               int           `json:"pending"`
	NotStarted            int           `json:"notStarted"`
	InProgress            int           `json:"inProgress"`
	OnHold                int           `json:"onHold"`
	Week                  []DayProgress `json:"week"`
	WeeklyProgress        int           `json:"weeklyProgress"`
	PreviousWeekProgress  int           `json:"previousWeekProgress"`
	PreviousWeekAvailable bool          `json:"previousWeekAvailable"`
	WeekChange            int           `json:"weekChange"`
	CurrentStreak         int           `json:"currentStreak"`
	LongestStreak         int           `json:"longestStreak"`
	LastStudyDate         string        `json:"lastStudyDate"`
}
