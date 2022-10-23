package database

import (
	"database/sql"
	"fmt"
	"os"
	"text/tabwriter"
)

// Print debug Rows results
func DebugPrintRows(rows *sql.Rows) {
	cols, _ := rows.Columns()
	count := len(cols)
	values := make([]interface{}, count)
	valuesPtrs := make([]interface{}, count)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 0, ' ', tabwriter.Debug)
	for _, col := range cols {
		fmt.Fprintf(w, "%s\t", col)
	}
	fmt.Fprintf(w, "\n")

	for i := 0; i < count; i++ {
		fmt.Fprintf(w, "\t")
	}

	fmt.Fprintf(w, "\n")

	for rows.Next() {
		for i, _ := range cols {
			valuesPtrs[i] = &values[i]
		}
		rows.Scan(valuesPtrs...)

		finalValues := make(map[string]interface{})
		for i, col := range cols {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}

			finalValues[col] = v
		}

		for _, col := range cols {
			fmt.Fprintf(w, "%v\t", finalValues[col])
		}
		fmt.Fprintf(w, "\n")
	}
	w.Flush()
}

// Print debug a single row (does not run rows.next())
func DebugPrintRow(rows *sql.Rows) {
	cols, _ := rows.Columns()
	count := len(cols)
	values := make([]interface{}, count)
	valuesPtrs := make([]interface{}, count)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 0, ' ', tabwriter.Debug)
	for _, col := range cols {
		fmt.Fprintf(w, "%s\t", col)
	}
	fmt.Fprintf(w, "\n")

	for i := 0; i < count; i++ {
		fmt.Fprintf(w, "\t")
	}

	fmt.Fprintf(w, "\n")

	for i, _ := range cols {
		valuesPtrs[i] = &values[i]
	}
	rows.Scan(valuesPtrs...)

	finalValues := make(map[string]interface{})
	for i, col := range cols {
		var v interface{}
		val := values[i]
		b, ok := val.([]byte)
		if ok {
			v = string(b)
		} else {
			v = val
		}

		finalValues[col] = v
	}

	for _, col := range cols {
		fmt.Fprintf(w, "%v\t", finalValues[col])
	}
	fmt.Fprintf(w, "\n")
	w.Flush()
}

func (db *DB) PrintBookmarks() error {

	var url, tags string

	rows, err := db.Handle.Query("select url,tags from bookmarks")

	for rows.Next() {
		err = rows.Scan(&url, &tags)
		if err != nil {
			return err
		}
		log.Debugf("url:%s  tags:%s", url, tags)
	}

	return nil
}
