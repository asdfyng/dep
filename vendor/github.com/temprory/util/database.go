package util

import (
	"database/sql"
	"fmt"
	"github.com/temprory/log"
)

// clear transaction
func ClearTx(tx *sql.Tx) {
	err := tx.Rollback()
	if err != nil && err != sql.ErrTxDone {
		fmt.Fprintf(log.DefaultLogger.Writer, log.LogWithFormater(log.LEVEL_ERROR, log.DefaultLogDepth, log.DefaultLogTimeLayout, "ClearTx failed: %v\n", nil))
	}
}
