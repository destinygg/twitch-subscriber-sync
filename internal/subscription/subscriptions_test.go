package subscription

import (
	"database/sql"
	"testing"
	"time"
)

type StartEnd struct {
	startt time.Time
	endt   time.Time
}

func getDurationUpTo(rows []*Row, index int) time.Duration {
	var sum time.Duration
	for i, j := 1, len(rows); i < j; i++ {
		v := rows[i]
		sum += v.Endtimestamp.Sub(v.Starttimestamp)

		if i == index {
			break
		}
	}

	return sum
}

func checkStartEndTimes(prefix string, data []*Row, expected []StartEnd, t *testing.T) {
	for k, v := range data {
		if !v.Starttimestamp.Equal(expected[k].startt) {
			t.Errorf(
				"%v Expected startt date %+v, got %+v at index %v",
				prefix, expected[k].startt, v.Starttimestamp, k,
			)
		}

		if !v.Endtimestamp.Equal(expected[k].endt) {
			t.Errorf(
				"%v Expected end date %+v, got %+v at index %v",
				prefix, expected[k].endt, v.Endtimestamp, k,
			)
		}
	}
}

func TestFixupSubTime(t *testing.T) {
	now := time.Now().UTC().Round(time.Second)

	data := []*Row{ // inserting at the second position
		{
			Starttimestamp: now.Add(-(time.Hour * 24 * 7 * 2)),
			Endtimestamp:   now.Add(-(time.Hour * 24 * 7 * 2)).AddDate(0, 1, 0),
		},
		{
			Starttimestamp: now,
			Endtimestamp:   now.AddDate(0, 1, 0),
		},
		{
			Starttimestamp: now.AddDate(0, 1, 0),
			Endtimestamp:   now.AddDate(0, 2, 0),
		},
		{
			Starttimestamp: now.AddDate(0, 2, 0),
			Endtimestamp:   now.AddDate(0, 3, 0),
		},
	}
	expected := []StartEnd{
		{
			startt: now,                  // originally in the past, should start from now
			endt:   data[0].Endtimestamp, // but the end should not change
		},
		{
			startt: data[0].Endtimestamp, // continues from the previous end, no overlap
			endt:   data[0].Endtimestamp.Add(getDurationUpTo(data, 1)),
		},
		{
			startt: data[0].Endtimestamp.Add(getDurationUpTo(data, 1)),
			endt:   data[0].Endtimestamp.Add(getDurationUpTo(data, 2)),
		},
		{
			startt: data[0].Endtimestamp.Add(getDurationUpTo(data, 2)),
			endt:   data[0].Endtimestamp.Add(getDurationUpTo(data, 3)),
		},
	}

	fixupSubTime(data)
	checkStartEndTimes("1", data, expected, t)

	data = []*Row{ // now inserting at the end
		{
			Starttimestamp: now.Add(-(time.Hour * 24 * 7 * 2)),
			Endtimestamp:   now.Add(-(time.Hour * 24 * 7 * 2)).AddDate(0, 1, 0),
		},
		{
			Starttimestamp: now.AddDate(0, 1, 0),
			Endtimestamp:   now.AddDate(0, 2, 0),
		},
		{
			Starttimestamp: now.AddDate(0, 2, 0),
			Endtimestamp:   now.AddDate(0, 3, 0),
		},
		{
			Starttimestamp: now,
			Endtimestamp:   now.AddDate(0, 1, 0),
		},
	}
	expected = []StartEnd{
		{
			startt: now,
			endt:   data[0].Endtimestamp,
		},
		{
			startt: data[0].Endtimestamp,
			endt:   data[0].Endtimestamp.Add(getDurationUpTo(data, 1)),
		},
		{
			startt: data[0].Endtimestamp.Add(getDurationUpTo(data, 1)),
			endt:   data[0].Endtimestamp.Add(getDurationUpTo(data, 2)),
		},
		{
			startt: data[0].Endtimestamp.Add(getDurationUpTo(data, 2)),
			endt:   data[0].Endtimestamp.Add(getDurationUpTo(data, 3)),
		},
	}

	fixupSubTime(data)
	checkStartEndTimes("2", data, expected, t)

}

func checkOrderOfIDs(prefix string, data []*Row, expected []int64, t *testing.T) {
	if len(expected) != len(data) {
		t.Errorf("%v Expected the amount of input data to equal the expected data", prefix)
		return
	}

	for k, v := range data {
		if v.ID.Int64 != expected[k] {
			t.Errorf(
				"%v Expected the id at index %v to equal %v got %+v",
				prefix, k, expected[k], v.ID.Int64,
			)
			return
		}
	}
}

func TestAssembleSubs(t *testing.T) {
	newrow := &Row{
		ID:   sql.NullInt64{1, true},
		Tier: sql.NullInt64{1, true},
	}
	data := []*Row{ // just check if the sub was inserted at the end
		{
			ID:   sql.NullInt64{2, true},
			Tier: sql.NullInt64{1, true},
		},
	}
	expected := []int64{2, 1}

	data = assembleSubs(data, newrow)
	checkOrderOfIDs("1", data, expected, t)

	data = []*Row{ // checking if the sub was inserted at the end again
		{
			ID:   sql.NullInt64{2, true},
			Tier: sql.NullInt64{2, true},
		},
		{
			ID:   sql.NullInt64{3, true},
			Tier: sql.NullInt64{2, true},
		},
		{
			ID:   sql.NullInt64{4, true},
			Tier: sql.NullInt64{1, true},
		},
		{
			ID:   sql.NullInt64{5, true},
			Tier: sql.NullInt64{1, true},
		},
	}
	expected = []int64{2, 3, 4, 5, 1}

	data = assembleSubs(data, newrow)
	checkOrderOfIDs("2", data, expected, t)

	// check if the new sub was inserted after subscriptions of the other tier2s,
	// aka at 3rd position (index 2)
	newrow = &Row{
		ID:   sql.NullInt64{6, true},
		Tier: sql.NullInt64{2, true},
	}
	expected = []int64{2, 3, 6, 4, 5, 1}

	data = assembleSubs(data, newrow)
	checkOrderOfIDs("3", data, expected, t)

	// check if the new sub is inserted at the first place since its the highest
	// tier
	newrow = &Row{
		ID:   sql.NullInt64{7, true},
		Tier: sql.NullInt64{3, true},
	}
	expected = []int64{7, 2, 3, 6, 4, 5, 1}

	data = assembleSubs(data, newrow)
	checkOrderOfIDs("4", data, expected, t)
}
