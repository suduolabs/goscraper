package helpers

import (
	"testing"
	"time"

	"goscraper/src/types"
)

func TestPlannerPage(t *testing.T) {
	cases := []struct {
		date time.Time
		want string
	}{
		// June 2026 is the tail of the 2025-26 EVEN semester, the page that
		// used to be hardcoded.
		{time.Date(2026, time.June, 30, 0, 0, 0, 0, time.UTC), "Academic_Planner_2025_26_EVEN"},
		{time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC), "Academic_Planner_2025_26_EVEN"},
		// July flips to the next academic year's ODD semester.
		{time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC), "Academic_Planner_2026_27_ODD"},
		{time.Date(2026, time.December, 31, 0, 0, 0, 0, time.UTC), "Academic_Planner_2026_27_ODD"},
		// Year boundaries keep the two-digit suffix padded.
		{time.Date(2099, time.August, 1, 0, 0, 0, 0, time.UTC), "Academic_Planner_2099_00_ODD"},
		{time.Date(2100, time.March, 1, 0, 0, 0, 0, time.UTC), "Academic_Planner_2099_00_EVEN"},
	}

	for _, c := range cases {
		if got := PlannerPage(c.date); got != c.want {
			t.Errorf("PlannerPage(%s) = %q, want %q", c.date.Format("2006-01-02"), got, c.want)
		}
	}
}

func TestSelectToday(t *testing.T) {
	// August starts on the 8th, as a semester beginning mid-month does. Day
	// positions and day numbers therefore disagree, which is what the old
	// index-based lookup got wrong.
	calendar := []types.CalendarMonth{
		{Month: "Aug '26", Days: []types.Day{
			{Date: "8", Day: "Sat"},
			{Date: "9", Day: "Sun"},
			{Date: "10", Day: "Mon"},
		}},
		{Month: "Sep '26", Days: []types.Day{
			{Date: "1", Day: "Tue"},
		}},
	}

	today, tomorrow, idx := SelectToday(calendar, time.Date(2026, time.August, 9, 0, 0, 0, 0, time.UTC))
	if today == nil || today.Date != "9" {
		t.Fatalf("today = %v, want day 9", today)
	}
	if tomorrow == nil || tomorrow.Date != "10" {
		t.Fatalf("tomorrow = %v, want day 10", tomorrow)
	}
	if idx != 0 {
		t.Errorf("monthIndex = %d, want 0", idx)
	}

	// Last day of a month rolls into the first day of the next one.
	_, tomorrow, _ = SelectToday(calendar, time.Date(2026, time.August, 10, 0, 0, 0, 0, time.UTC))
	if tomorrow == nil || tomorrow.Date != "1" {
		t.Fatalf("tomorrow across month boundary = %v, want day 1", tomorrow)
	}

	// A month the planner does not cover must report nothing rather than
	// falling back to the first month, which reported January as "today".
	today, tomorrow, _ = SelectToday(calendar, time.Date(2026, time.November, 3, 0, 0, 0, 0, time.UTC))
	if today != nil || tomorrow != nil {
		t.Errorf("out-of-range month returned today=%v tomorrow=%v, want nil, nil", today, tomorrow)
	}

	// A day inside a covered month but missing from it is also nothing.
	today, _, _ = SelectToday(calendar, time.Date(2026, time.August, 3, 0, 0, 0, 0, time.UTC))
	if today != nil {
		t.Errorf("missing day returned %v, want nil", today)
	}

	// Empty calendar must not panic.
	if today, tomorrow, idx = SelectToday(nil, time.Now()); today != nil || tomorrow != nil || idx != 0 {
		t.Errorf("empty calendar returned %v, %v, %d", today, tomorrow, idx)
	}
}

// The signed-out page carries neither the planner table nor the zmlvalue
// marker. That used to come back as a 200 with an empty Calendar, which the
// frontend rendered as a crash instead of sending the user to log in again.
func TestParseCalendarSignedOut(t *testing.T) {
	c := NewCalendarFetcher(time.Date(2026, time.August, 19, 0, 0, 0, 0, time.UTC), "")

	resp, err := c.parseCalendar("<html><body>Please sign in to continue</body></html>")
	if err == nil {
		t.Fatalf("signed-out page parsed without error, got %+v", resp)
	}
	if err.Error() != "invalid response format" {
		t.Errorf("err = %q, want %q so HandleError maps it to tokenInvalid", err.Error(), "invalid response format")
	}
	if resp != nil {
		t.Errorf("response = %+v, want nil", resp)
	}
}

func TestParseCalendarReadsPlannerTable(t *testing.T) {
	html := `<table bgcolor="#FFFFFF">
<tr><th>Aug '26</th><th>Day</th><th>Event</th><th>DO</th><th></th></tr>
<tr><td>18</td><td>Tue</td><td></td><td>1</td><td></td></tr>
<tr><td>19</td><td>Wed</td><td></td><td>2</td><td></td></tr>
</table>`

	c := NewCalendarFetcher(time.Date(2026, time.August, 19, 0, 0, 0, 0, time.UTC), "")
	resp, err := c.parseCalendar(html)
	if err != nil {
		t.Fatalf("parseCalendar returned %v", err)
	}
	if len(resp.Calendar) != 1 || resp.Calendar[0].Month != "Aug '26" {
		t.Fatalf("calendar = %+v, want one Aug '26 month", resp.Calendar)
	}
	if len(resp.Calendar[0].Days) != 2 {
		t.Fatalf("days = %+v, want 2", resp.Calendar[0].Days)
	}
	if resp.Today == nil || resp.Today.Date != "19" {
		t.Errorf("today = %+v, want day 19", resp.Today)
	}
}
