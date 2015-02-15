package subscription

import (
	"database/sql"
	"testing"
	"time"
)

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

func TestFixupSubTime(t *testing.T) {
	data := []*Row{
		{
			Starttimestamp: time.Now().Add(-(time.Hour * 24 * 7 * 2)),
			Endtimestamp:   time.Now().Add(-(time.Hour * 24 * 7 * 2)).AddDate(0, 1, 0),
		},
		{
			Starttimestamp: time.Now(),
			Endtimestamp:   time.Now().AddDate(0, 1, 0),
		},
		{
			Starttimestamp: time.Now().AddDate(0, 1, 0),
			Endtimestamp:   time.Now().AddDate(0, 2, 0),
		},
		{
			Starttimestamp: time.Now().AddDate(0, 2, 0),
			Endtimestamp:   time.Now().AddDate(0, 3, 0),
		},
	}
	expected := []time.Time{
		data[0].Endtimestamp,
		data[0].Endtimestamp.Add(getDurationUpTo(data, 1)),
		data[0].Endtimestamp.Add(getDurationUpTo(data, 2)),
		data[0].Endtimestamp.Add(getDurationUpTo(data, 3)),
	}

	fixupSubTime(data)
	for k, v := range data {
		if !v.Endtimestamp.Equal(expected[k]) {
			t.Errorf("Expected date %+v, got %+v at index %v", expected[k], v.Endtimestamp, k)
		}
	}

	data = []*Row{
		{
			Starttimestamp: time.Now().Add(-(time.Hour * 24 * 7 * 2)),
			Endtimestamp:   time.Now().Add(-(time.Hour * 24 * 7 * 2)).AddDate(0, 1, 0),
		},
		{
			Starttimestamp: time.Now().AddDate(0, 1, 0),
			Endtimestamp:   time.Now().AddDate(0, 2, 0),
		},
		{
			Starttimestamp: time.Now().AddDate(0, 2, 0),
			Endtimestamp:   time.Now().AddDate(0, 3, 0),
		},
		{
			Starttimestamp: time.Now(),
			Endtimestamp:   time.Now().AddDate(0, 1, 0),
		},
	}
	expected = []time.Time{
		data[0].Endtimestamp,
		data[0].Endtimestamp.Add(getDurationUpTo(data, 1)),
		data[0].Endtimestamp.Add(getDurationUpTo(data, 2)),
		data[0].Endtimestamp.Add(getDurationUpTo(data, 3)),
	}

	fixupSubTime(data)
	for k, v := range data {
		if !v.Endtimestamp.Equal(expected[k]) {
			t.Errorf("Expected2 date %+v, got %+v at index %v", expected[k], v.Endtimestamp, k)
		}
	}
}

func TestAssembleSubs(t *testing.T) {
	newrow := &Row{
		ID:   sql.NullInt64{1, true},
		Tier: sql.NullInt64{1, true},
	}
	data := []*Row{
		{
			Tier: sql.NullInt64{1, true},
		},
	}

	data = assembleSubs(data, newrow)
	if len(data) != 2 && data[1].ID.Int64 != 1 {
		t.Errorf("EVEN THE MOST BASIC SHIT FAILED %+v", data)
	}

	data = []*Row{
		{
			Tier: sql.NullInt64{2, true},
		},
		{
			Tier: sql.NullInt64{2, true},
		},
		{
			Tier: sql.NullInt64{1, true},
		},
		{
			Tier: sql.NullInt64{1, true},
		},
	}

	data = assembleSubs(data, newrow)
	if len(data) != 5 && data[4].ID.Int64 != 1 {
		t.Errorf("was expecting the new row to be in position 4")
	}

	newrow.Tier.Int64 = 2
	data = data[0:3]
	data = assembleSubs(data, newrow)
	if len(data) != 5 && data[2].ID.Int64 != 1 {
		t.Errorf("was expecting the new row to be in position 2")
	}

	newrow.Tier.Int64 = 3
	data = data[0:3]
	data = assembleSubs(data, newrow)
	if len(data) != 5 && data[0].ID.Int64 != 1 {
		t.Errorf("was expecting the new row to be in position 0")
	}
}
