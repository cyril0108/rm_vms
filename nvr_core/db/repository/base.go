package repository

import (
	"nvr_core/logger"
	"strings"
	"time"
)

var LOG = logger.NewLogger("[nvr_core][db][repository]")

type PartialUpdateInterfaces map[string]interface{}

// Format:
// JoinSetFieldsClause("UPDATE table SET", updates)
func JoinSetFieldsClause(queryPrefix string, updates PartialUpdateInterfaces) string {

	var args []interface{}
	var setClauses []string

	// Iterate through the safely constructed map
	for col, val := range updates {
		setClauses = append(setClauses, col+" = ?")
		args = append(args, val)
	}

	// Always enforce the updated_at timestamp
	setClauses = append(setClauses, "updated_at = ?")
	args = append(args, time.Now().Unix())

	// Stitch the query together safely
	return strings.Join(setClauses, ", ")

}