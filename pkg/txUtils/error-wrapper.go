package txUtils

import (
	"fmt"
)

// Rollbacker is the interface that wraps the basic Rollback method.
type Rollbacker interface {
	Rollback() error
}

// Rollback calls to Rollback and wraps the given error
// with the rollback error if occurred.
func Rollback(tx Rollbacker, err error) error {
	if rerr := tx.Rollback(); rerr != nil {
		err = fmt.Errorf("%w: %v", err, rerr)
	}
	return err
}
