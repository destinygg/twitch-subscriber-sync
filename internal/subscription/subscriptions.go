package subscription

import (
	"database/sql"
	"time"

	"github.com/destinygg/website2/internal/debug"
	"golang.org/x/net/context"
)

// Row mirrors the backing database schema
type Row struct {
	ID             sql.NullInt64
	Donationid     sql.NullInt64
	Fromuserid     sql.NullInt64
	Targetuserid   sql.NullInt64
	Tier           sql.NullInt64
	Starttimestamp time.Time
	Endtimestamp   time.Time
	Timestamp      time.Time
}

func getDBFromContext(ctx context.Context) *sql.DB {
	db, ok := ctx.Value("db").(*sql.DB)
	if !ok {
		panic("Database not found in the context")
	}

	return db
}

// Init expects to be called with the main context
func Init(ctx context.Context) {
	/*
		TODO set up array of durations for subscription end times, sort them
		from lowest to highest, and wait on them, if one passes, signal the users pkg
		to refresh the users flairs, if a new subscription is added, just refresh
		the whole array
		signaling the chat should be done in the users package where we can correctly
		regenerate the whole thing
	*/
}

// Subscribed checks if the given userid is a subscriber and returns the tier
// of the subscription, to differentiate
func Subscribed(ctx context.Context, userid int64) (int, error) {
	db := getDBFromContext(ctx)
	stmt, err := db.Prepare(`
		SELECT tier
		FROM subscriptions
		WHERE
			userid          = ? AND
			starttimestamp <= NOW() AND
			endtimestamp   >= NOW()
		ORDER BY starttimestamp
		LIMIT 1
	`)

	if err != nil {
		d.F("err: %v userid: %v", err, userid)
	}

	var tier int
	err = stmt.QueryRow().Scan(&tier)
	if err != nil {
		return 0, err
	}

	return tier, nil
}

// Add inserts the subscription, modifying all other subscriptions start and end
// times as appropriate
// the logic is the following:
// always use up time from the highest tier, oldest subscription first
// insert the new subscription before the next lower tier sub (or append it
// to the end)
// fix up the times of the subscriptions so that no time is lost and the
// time interval is continous
func Add(ctx context.Context, row *Row) {
	db := getDBFromContext(ctx)
	tx, err := db.Begin()
	if err != nil {
		panic(err.Error())
	}

	// expectations: the intervals will be continous, no gaps and no overlaps
	// the new row will be inserted with start = current time, and
	// end = current time + 1 month
	// if its the first sub for the user, everything will be basically a noop
	rows := getActiveSubs(tx, row.Targetuserid.Int64)

	// insert row into subs based on tier, this makes sure they are in the correct
	// order, but not yet with the correct timestamps
	rows = assembleSubs(rows, row)

	// now fix up the timestamps so that they are continous and start from
	// rows[0], do this by taking the durations in order and adjusting
	// all the start and end timestamps
	fixupSubTime(rows)

	// now update records that have a non-null ID aka where records already exist
	updateSubs(tx, rows)
	// and insert the new record
	insert(tx, row)

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		d.F("err: %v row: %+v", err, row)
	}
}

func getActiveSubs(tx *sql.Tx, userid int64) []*Row {
	cursor, err := tx.Query(`
		SELECT
			id,
			donationid,
			fromuserid,
			targetuserid,
			tier,
			starttimestamp,
			endtimestamp,
			timestamp
		FROM subscriptions
		WHERE
			userid          = ? AND
			starttimestamp <= NOW() AND
			endtimestamp   >= NOW()
		ORDER BY starttimestamp
	`, userid)
	if err != nil {
		tx.Rollback()
		d.F("err: %v targetuserid: %+v", err, userid)
	}
	defer cursor.Close()

	var rows []*Row
	for cursor.Next() {
		var row *Row
		err := cursor.Scan(
			&row.ID,
			&row.Donationid,
			&row.Fromuserid,
			&row.Targetuserid,
			&row.Tier,
			&row.Starttimestamp,
			&row.Endtimestamp,
			&row.Timestamp,
		)

		if err != nil {
			tx.Rollback()
			panic(err.Error())
		}

		rows = append(rows, row)
	}

	return rows
}

func assembleSubs(rows []*Row, row *Row) []*Row {
	var inserted bool
	for k, v := range rows {
		if row.Tier.Int64 > v.Tier.Int64 {
			// insert into slice
			inserted = true
			rows = append(rows[:k], append([]*Row{row}, rows[k:]...)...)
			break
		}
	}

	// did not need to insert into the array, so just append it to the end
	if !inserted {
		rows = append(rows, row)
	}

	return rows
}

func fixupSubTime(rows []*Row) {
	durations := make([]time.Duration, 0, len(rows))
	for _, v := range rows {
		durations = append(durations, v.Endtimestamp.Sub(v.Starttimestamp))
	}

	pe := rows[0].Endtimestamp
	for i, l := 1, len(rows); i < l; i++ {
		rows[i].Starttimestamp = pe
		rows[i].Endtimestamp = pe.Add(durations[i])
		pe = rows[i].Endtimestamp
	}
}

func updateSubs(tx *sql.Tx, rows []*Row) {
	stmt, err := tx.Prepare(`
		UPDATE subscriptions
		SET
			starttimestamp = ?,
			endtimestamp   = ?
		WHERE id = ?
		LIMIT 1
	`)
	if err != nil {
		tx.Rollback()
		panic(err.Error())
	}
	defer stmt.Close()

	for _, v := range rows {
		if v.ID.Valid { // if the id is not valid, its the to-be inserted record
			_, err = stmt.Exec(
				v.Starttimestamp,
				v.Endtimestamp,
				v.ID,
			)

			if err != nil {
				tx.Rollback()
				d.F("err: %v row: %+v", err, v)
			}
		}
	}
}

func insert(tx *sql.Tx, row *Row) sql.NullInt64 {
	res, err := tx.Exec(`
		INSERT INTO subscriptions
		(
			donationid,
			fromuserid,
			targetuserid,
			tier,
			starttimestamp,
			endtimestamp,
			timestamp
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		row.Donationid,
		row.Fromuserid,
		row.Targetuserid,
		row.Tier,
		row.Starttimestamp,
		row.Endtimestamp,
		row.Timestamp,
	)
	if err != nil {
		tx.Rollback()
		panic(err.Error())
	}

	id, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		panic(err.Error())
	}

	return sql.NullInt64{
		Int64: id,
		Valid: true,
	}
}
