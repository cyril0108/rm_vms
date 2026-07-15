package repository

import (
	"nvr_core/db/models"
	"nvr_core/logger"
	"strings"
)

var LOG = logger.NewLogger("[nvr_core][db][repository]")

// Format:
// JoinSetFieldsClause("UPDATE table SET", updates, true)
func JoinSetFieldsClause(queryPrefix string, updates models.PartialUpdateInterfaces, withUpdateAt bool) (string, []interface{}) {

	var args []interface{}
	var setClauses []string

	if withUpdateAt {
		updates.ApplyUpdatedAt()
	}

	// Iterate through the safely constructed map
	for col, val := range updates {
		setClauses = append(setClauses, col+" = ?")
		args = append(args, val)
	}

	// Stitch the query together safely
	return queryPrefix + strings.Join(setClauses, ", "), args

}