package repository

import (
	"context"
	"database/sql"
	"fmt"

	"nvr_core/db/models"
)

type LayoutRepository interface {
	GetByUser(ctx context.Context, userID int64) ([]*models.Layout, error)
	GetLayout(ctx context.Context, userID int64, layoutID int64) (*models.Layout, error)

	Create(ctx context.Context, userID int64, layout *models.Layout) (int64, error)
	Update(ctx context.Context, userID int64, layout *models.Layout) error
	UpdatePartial(ctx context.Context, userID int64, layout *models.Layout, updateItems bool) error
	Delete(ctx context.Context, userID int64, id int64) error
}

type layoutRepo struct {
	db *sql.DB
}

func NewLayoutRepository(db *sql.DB) LayoutRepository {
	return &layoutRepo{db: db}
}

// Create inserts a new layout and its items within a single transaction.
func (r *layoutRepo) Create(ctx context.Context, userID int64, layout *models.Layout) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	// Defer a rollback in case anything fails. If tx.Commit() is called, this does nothing.
	defer tx.Rollback()

	// Insert the parent layout
	res, err := tx.ExecContext(ctx, 
		`INSERT INTO layouts (user_id, name, mode, payload) VALUES (?, ?, ?, ?)`,
		userID, layout.Name, layout.Mode, string(layout.Payload),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert layout: %w", err)
	}

	layoutID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Insert all child items
	if len(layout.Items) > 0 {
		stmt, err := tx.PrepareContext(ctx, `INSERT INTO layout_items (layout_id, type, payload) VALUES (?, ?, ?)`)
		if err != nil {
			return 0, fmt.Errorf("failed to prepare item statement: %w", err)
		}
		defer stmt.Close()

		for _, item := range layout.Items {
			_, err = stmt.ExecContext(ctx, layoutID, item.Type, string(item.Payload))
			if err != nil {
				return 0, fmt.Errorf("failed to insert layout item: %w", err)
			}
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return layoutID, nil
}

func (l *layoutRepo) GetLayout(ctx context.Context, userID int64, layoutID int64) (*models.Layout, error) {

	var lo models.Layout
	var payload string

	query := `
		SELECT id, user_id, name, mode, payload, created_at
		FROM layouts WHERE user_id = ? AND id = ?
	`

	err := l.db.QueryRowContext(ctx, query, userID, layoutID).Scan(
		&lo.ID, &lo.UserID, &lo.Name, &lo.Mode, &payload, &lo.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Cleanly return nil if the layout doesn't exist or doesn't belong to this user
			return nil, fmt.Errorf("layout not found") 
		}
		return nil, fmt.Errorf("failed to fetch layout: %w", err)
	}

	lo.Payload = []byte(payload)
	lo.Items = make([]models.LayoutItem, 0)

	itemRows, err := l.db.QueryContext(ctx,
		`SELECT id, layout_id, type, payload 
		 FROM layout_items 
		 WHERE layout_id = ? 
		 ORDER BY id ASC`, 
		layoutID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch layout items: %w", err)
	}
	defer itemRows.Close()


	for itemRows.Next() {
		var item models.LayoutItem
		var itemPayload string
		if err := itemRows.Scan(&item.ID, &item.LayoutID, &item.Type, &itemPayload); err != nil {
			return nil, err
		}
		item.Payload = []byte(itemPayload)

		// Append the item to the correct parent layout using our map
		lo.Items = append(lo.Items, item)
	}

	if err := itemRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over layout items: %w", err)
	}

	return &lo, nil

}

// GetByUser fetches all layouts for a specific user, including all their nested items.
func (l *layoutRepo) GetByUser(ctx context.Context, userID int64) ([]*models.Layout, error) {
	// Fetch the parent layouts
	rows, err := l.db.QueryContext(ctx, 
		`SELECT id, user_id, name, mode, payload, created_at FROM layouts WHERE user_id = ? ORDER BY created_at ASC`, 
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	layoutMap := make(map[int64]*models.Layout)
	var layouts []*models.Layout

	for rows.Next() {
		l := &models.Layout{}
		var payload string // Temporary string to read the JSON text
		if err := rows.Scan(&l.ID, &l.UserID, &l.Name, &l.Mode, &payload, &l.CreatedAt); err != nil {
			return nil, err
		}
		l.Payload = []byte(payload) // Convert back to json.RawMessage
		l.Items = make([]models.LayoutItem, 0) // Initialize empty array for frontend friendliness
		
		layouts = append(layouts, l)
		layoutMap[l.ID] = layouts[len(layouts)-1] // Keep a pointer map to easily append items later
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(layouts) == 0 {
		return layouts, nil // Return empty early if no layouts found
	}

	// Fetch the items for this user's layouts using a JOIN
	// This avoids the N+1 problem (doing a separate query for every single layout)
	itemRows, err := l.db.QueryContext(ctx, `
		SELECT li.id, li.layout_id, li.type, li.payload 
		FROM layout_items li
		JOIN layouts l ON li.layout_id = l.id
		WHERE l.user_id = ?
		ORDER BY li.id ASC`, 
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer itemRows.Close()

	for itemRows.Next() {
		var item models.LayoutItem
		var payload string
		if err := itemRows.Scan(&item.ID, &item.LayoutID, &item.Type, &payload); err != nil {
			return nil, err
		}
		item.Payload = []byte(payload)

		// Append the item to the correct parent layout using our map
		if parentLayout, exists := layoutMap[item.LayoutID]; exists {
			parentLayout.Items = append(parentLayout.Items, item)
		}
	}

	if err := itemRows.Err(); err != nil {
		return nil, err
	}

	return layouts, nil
}

// Update completely replaces a layout's data and items. 
// It requires the userID to ensure a user cannot overwrite another user's layout.
func (r *layoutRepo) Update(ctx context.Context, userID int64, layout *models.Layout) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update the parent layout (Only if it belongs to this user!)
	res, err := tx.ExecContext(ctx, 
		`UPDATE layouts SET name = ?, mode = ?, payload = ? WHERE id = ? AND user_id = ?`,
		layout.Name, layout.Mode, string(layout.Payload), layout.ID, userID,
	)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil || affected == 0 {
		return fmt.Errorf("layout not found or unauthorized")
	}


	// Insert the new items
	if len(layout.Items) > 0 {

		// Wipe existing items for this layout
		_, err = tx.ExecContext(ctx, `DELETE FROM layout_items WHERE layout_id = ?`, layout.ID)
		if err != nil {
			return err
		}

		stmt, err := tx.PrepareContext(ctx, `INSERT INTO layout_items (layout_id, type, payload) VALUES (?, ?, ?)`)
		if err != nil {
			return err
		}
		defer stmt.Close()

		for _, item := range layout.Items {
			_, err = stmt.ExecContext(ctx, layout.ID, item.Type, string(item.Payload))
			if err != nil {
				return err
			}
		}

	}

	return tx.Commit()
}


// Update completely replaces a layout's parent data.
// If updateItems is true, it will also completely rewrite the layout's child items.
func (r *layoutRepo) UpdatePartial(ctx context.Context, userID int64, layout *models.Layout, updateItems bool) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Unconditionally update the parent layout (Only if it belongs to this user!)
	res, err := tx.ExecContext(ctx, 
		`UPDATE layouts SET name = ?, mode = ?, payload = ? WHERE id = ? AND user_id = ?`,
		layout.Name, layout.Mode, string(layout.Payload), layout.ID, userID,
	)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil || affected == 0 {
		return fmt.Errorf("layout not found or unauthorized")
	}

	// Only touch the items if the frontend explicitly included them in the JSON
	if updateItems {
		// Wipe existing items for this layout
		_, err = tx.ExecContext(ctx, `DELETE FROM layout_items WHERE layout_id = ?`, layout.ID)
		if err != nil {
			return err
		}

		// Insert the new items (if the array isn't empty)
		if len(layout.Items) > 0 {
			stmt, err := tx.PrepareContext(ctx, `INSERT INTO layout_items (layout_id, type, payload) VALUES (?, ?, ?)`)
			if err != nil {
				return err
			}
			defer stmt.Close()

			for _, item := range layout.Items {
				_, err = stmt.ExecContext(ctx, layout.ID, item.Type, string(item.Payload))
				if err != nil {
					return err
				}
			}
		}
	}

	return tx.Commit()
}


// Delete removes a layout. The SQLite `ON DELETE CASCADE` handles wiping the layout_items automatically.
func (r *layoutRepo) Delete(ctx context.Context, userID int64, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM layouts WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil || affected == 0 {
		return fmt.Errorf("layout not found or unauthorized")
	}

	return nil
}