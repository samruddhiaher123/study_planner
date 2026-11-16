package services

type DayProgress struct {
	Day       string
	Date      string
	Completed int
	Total     int
	Progress  int
}

type DashboardData struct {
	TodayDate             string
	TodayProgress         int
	TodayPlanned          int
	TodayCompleted        int
	TodayRemaining        int
	TotalTasks            int
	Completed             int
	Pending               int
	NotStarted            int
	InProgress            int
	OnHold                int
	Week                  []DayProgress
	WeeklyProgress        int
	PreviousWeekProgress  int
	PreviousWeekAvailable bool
	WeekChange            int
	CurrentStreak         int
	LongestStreak         int
	LastStudyDate         string
}
