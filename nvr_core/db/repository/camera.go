package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"nvr_core/db/models"
	"nvr_core/utils"
)

var (
	ErrCameraNotFound = errors.New("camera not found")
	ErrCameraExists   = errors.New("camera already exists")
	ErrCameraHardwareConflict = errors.New("a camera with this serial number and IP address already exists")
)

type CameraRepository interface {
	// Get
	GetByID(ctx context.Context, id int64) (*models.Camera, error)
	GetAll(ctx context.Context) ([]*models.Camera, error)
	GetAllForInSystemCheck(ctx context.Context) ([]*models.Camera, error)
	// Create and Manage
	Create(ctx context.Context, cam *models.Camera) (int64, error)
	// Recreate(ctx context.Context, cam *models.Camera) error
	UpdatePartial(ctx context.Context, id int64, updates models.PartialUpdateInterfaces) error
	SetActivate(ctx context.Context, id int64, active int) error
	Delete(ctx context.Context, id int64) error
	Activate(ctx context.Context, id int64) error
	Deactivate(ctx context.Context, id int64) error
}

type cameraRepo struct {
	db *sql.DB
}

func NewCameraRepository(db *sql.DB) CameraRepository {
	return &cameraRepo{db: db}
}

// prepareForInsert initializes defaults, timestamps, and UUIDs.
func (r *cameraRepo) prepareForInsert(cam *models.Camera) (supportsPTZ int, isActive int) {
	now := time.Now().Unix()
	cam.CreatedAt = now
	cam.UpdatedAt = now

	if cam.Name == "" {
		cam.Name = cam.DefaultName()
	}
	cam.UUID = utils.GenerateCameraUUID(cam.MACAddress, cam.StreamURL)

	if cam.SupportsPTZ { supportsPTZ = 1 }
	if cam.IsActive { isActive = 1 }

	return supportsPTZ, isActive
}

// checkHardwareConflict looks for an existing camera by Serial or IP.
func (r *cameraRepo) checkHardwareConflict(ctx context.Context, tx *sql.Tx, serial, ip string) (id int64, isDeleted int, err error) {
	if serial != "" {
		err = tx.QueryRowContext(ctx, `SELECT id, deleted FROM cameras WHERE serial_number = ?`, serial).Scan(&id, &isDeleted)
	} else {
		err = tx.QueryRowContext(ctx, `SELECT id, deleted FROM cameras WHERE ip_address = ?`, ip).Scan(&id, &isDeleted)
	}
	return id, isDeleted, err
}

// resurrectCamera overwrites a soft-deleted camera's fields and marks it active.
func (r *cameraRepo) resurrectCamera(ctx context.Context, tx *sql.Tx, id int64, cam *models.Camera, supportsPTZ, isActive int) error {
	query := `
		UPDATE cameras SET 
			uuid = ?, name = ?, manufacturer = ?, model = ?, serial_number = ?, 
			ip_address = ?, mac_address = ?, http_port = ?, type = ?, 
			username = ?, password_enc = ?, stream_url = ?, sub_stream_url = ?, 
			onvif_profile_token = ?, sub_stream_profile_token = ?, supports_ptz = ?, 
			retention_gb_limit = ?, is_active = ?, updated_at = ?, deleted = 0
		WHERE id = ?
	`
	_, err := tx.ExecContext(ctx, query,
		cam.UUID, cam.Name, cam.Manufacturer, cam.Model, cam.SerialNumber,
		cam.IPAddress, cam.MACAddress, cam.HTTPPort, cam.Type, cam.Username, cam.PasswordEnc,
		cam.StreamURL, cam.SubStreamURL, cam.OnvifProfileToken, cam.SubStreamProfileToken,
		supportsPTZ, cam.RetentionGBLimit, isActive, cam.UpdatedAt, id,
	)
	return err
}


// insertNewCamera handles a fresh insert into the database.
func (r *cameraRepo) insertNewCamera(ctx context.Context, tx *sql.Tx, cam *models.Camera, supportsPTZ, isActive int) (int64, error) {
	query := `
		INSERT INTO cameras (
			uuid, name, manufacturer, model, serial_number, 
			ip_address, mac_address, http_port, type, 
			username, password_enc, stream_url, sub_stream_url, 
			onvif_profile_token, sub_stream_profile_token, supports_ptz, 
			retention_gb_limit, is_active, created_at, updated_at, deleted
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)
	`
	result, err := tx.ExecContext(ctx, query,
		cam.UUID, cam.Name, cam.Manufacturer, cam.Model, cam.SerialNumber,
		cam.IPAddress, cam.MACAddress, cam.HTTPPort, cam.Type, cam.Username, cam.PasswordEnc,
		cam.StreamURL, cam.SubStreamURL, cam.OnvifProfileToken, cam.SubStreamProfileToken,
		supportsPTZ, cam.RetentionGBLimit, isActive, cam.CreatedAt, cam.UpdatedAt,
	)

	if err != nil {
		if isUniqueConstraintViolation(err) {
			errStr := err.Error()
			if strings.Contains(errStr, "serial_number") || strings.Contains(errStr, "ip_address") || strings.Contains(errStr, "idx_cam_dedup") {
				return 0, ErrCameraHardwareConflict
			}
			return 0, ErrCameraExists
		}
		return 0, err
	}

	return result.LastInsertId()
}

// Create inserts a new camera. The 'cam.ID' must be pre-generated (e.g., UUID).
func (r *cameraRepo) Create(ctx context.Context, cam *models.Camera) (int64, error){

	supportsPTZ, isActive := r.prepareForInsert(cam)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	existingID, isDeleted, err := r.checkHardwareConflict(ctx, tx, cam.SerialNumber, cam.IPAddress)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, err // Real DB error during check
	}

	var generatedID int64

	if err == nil {

		// Hardware conflict detected
		if isDeleted == 0 {
			return 0, ErrCameraHardwareConflict
		}

		// Resurrect soft-deleted camera
		if err := r.resurrectCamera(ctx, tx, existingID, cam, supportsPTZ, isActive); err != nil {
			return 0, err
		}
		generatedID = existingID

	} else {
		// No conflict, safe to insert
		generatedID, err = r.insertNewCamera(ctx, tx, cam, supportsPTZ, isActive)
		if err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	cam.ID = generatedID
	return generatedID, nil

}

// GetByID fetches a specific camera and safely handles SQLite NULLs.
func (r *cameraRepo) GetByID(ctx context.Context, id int64) (*models.Camera, error) {
	query := `
		SELECT id, uuid, name, manufacturer, model, serial_number, 
		       ip_address, mac_address, http_port, type, 
		       username, password_enc, stream_url, sub_stream_url, 
		       onvif_profile_token, sub_stream_profile_token, supports_ptz, 
		       retention_gb_limit, is_active, created_at, updated_at 
		FROM cameras WHERE deleted=0 AND id = ?
	`

	var c models.Camera

	// Temporary variables to hold potentially NULL database columns
	var manufacturer, model, username, passwordEnc sql.NullString
	var subStream, onvifToken, subStreamToken sql.NullString
	var macAddress sql.NullString
	var retentionLimit sql.NullInt64
	var supportsPTZ, isActive int

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID, &c.UUID, &c.Name, &manufacturer, &model, &c.SerialNumber,
		&c.IPAddress, &macAddress, &c.HTTPPort, &c.Type, &username, &passwordEnc,
		&c.StreamURL, &subStream, &onvifToken, &subStreamToken,
		&supportsPTZ, &retentionLimit, &isActive, &c.CreatedAt, &c.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCameraNotFound
		}
		return nil, err
	}

	// Map the safe temporary variables back into the struct pointers
	if manufacturer.Valid { c.Manufacturer = manufacturer.String }
	if model.Valid { c.Model = model.String }
	if username.Valid { c.Username = username.String }
	if macAddress.Valid { c.MACAddress = macAddress.String }
	if passwordEnc.Valid { c.PasswordEnc = passwordEnc.String }
	if subStream.Valid { c.SubStreamURL = subStream.String }
	if onvifToken.Valid { c.OnvifProfileToken = onvifToken.String }
	if subStreamToken.Valid { c.SubStreamProfileToken = subStreamToken.String }
	if retentionLimit.Valid { 
		limit := int(retentionLimit.Int64)
		c.RetentionGBLimit = limit
	}

	c.SupportsPTZ = supportsPTZ == 1
	c.IsActive = isActive == 1

	return &c, nil
}

// GetAll fetches all cameras to initialize the NVR ingestion workers on startup.
func (r *cameraRepo) GetAll(ctx context.Context) ([]*models.Camera, error) {
	query := `
		SELECT id, uuid, name, manufacturer, model, serial_number, 
		       ip_address, mac_address, http_port, type, 
		       username, password_enc, stream_url, sub_stream_url, 
		       onvif_profile_token, sub_stream_profile_token, supports_ptz, 
		       retention_gb_limit, is_active, created_at, updated_at 
		FROM cameras WHERE deleted=0 ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cameras []*models.Camera
	for rows.Next() {
		var c models.Camera
		var manufacturer, model, username, passwordEnc sql.NullString
		var macAddress sql.NullString
		var subStream, onvifToken, subStreamToken sql.NullString
		var retentionLimit sql.NullInt64
		var supportsPTZ, isActive int

		if err := rows.Scan(
			&c.ID, &c.UUID, &c.Name, &manufacturer, &model, &c.SerialNumber,
			&c.IPAddress, &macAddress, &c.HTTPPort, &c.Type, &username, &passwordEnc,
			&c.StreamURL, &subStream, &onvifToken, &subStreamToken,
			&supportsPTZ, &retentionLimit, &isActive, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if manufacturer.Valid { c.Manufacturer = manufacturer.String }
		if model.Valid { c.Model = model.String }
		if username.Valid { c.Username = username.String }
		if passwordEnc.Valid { c.PasswordEnc = passwordEnc.String }
		if macAddress.Valid { c.MACAddress = macAddress.String }
		if subStream.Valid { c.SubStreamURL = subStream.String }
		if onvifToken.Valid { c.OnvifProfileToken = onvifToken.String }
		if subStreamToken.Valid { c.SubStreamProfileToken = subStreamToken.String }
		if retentionLimit.Valid { 
			limit := int(retentionLimit.Int64)
			c.RetentionGBLimit = limit
		}
		c.SupportsPTZ = supportsPTZ == 1
		c.IsActive = isActive == 1

		cameras = append(cameras, &c)
	}

	return cameras, rows.Err()
}

/**
 * Get id, serial_number, ip_address, is_active
 * for in-system check
 * @return {([]*models.Camera, error)}
 */
func (r *cameraRepo) GetAllForInSystemCheck(ctx context.Context) ([]*models.Camera, error) {
	query := `
		SELECT id, serial_number, ip_address, is_active
		FROM cameras WHERE deleted=0 ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cameras []*models.Camera
	for rows.Next() {
		var c models.Camera
		var isActive int

		if err := rows.Scan(
			&c.ID, &c.SerialNumber,&c.IPAddress, &isActive,
		); err != nil {
			return nil, err
		}

		c.IsActive = isActive == 1

		cameras = append(cameras, &c)
	}

	return cameras, rows.Err()
}

// Deactivate performs a soft-delete to preserve evidence in the segments table.
func (r *cameraRepo) Deactivate(ctx context.Context, id int64) error {
	return r.SetActivate(ctx, id, 0)
}

func (r *cameraRepo) Activate(ctx context.Context, id int64) error {
	return r.SetActivate(ctx, id, 1)
}

// Deactivate performs a soft-delete to preserve evidence in the segments table.
func (r *cameraRepo) SetActivate(ctx context.Context, id int64, active int) error {
	query := `UPDATE cameras SET is_active = ?, updated_at = ? WHERE deleted=0 AND id = ?`

	result, err := r.db.ExecContext(ctx, query, active, time.Now().Unix(), id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrCameraNotFound
	}

	return nil
}

func (r *cameraRepo) Delete(ctx context.Context, id int64) error {
	query := `UPDATE cameras SET is_active = 0, deleted = 1, updated_at = ? WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, time.Now().Unix(), id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrCameraNotFound
	}

	return nil
}


func (r *cameraRepo) UpdatePartial(ctx context.Context, id int64, updates models.PartialUpdateInterfaces) error {

	// If the map is empty, there is nothing to update
	if len(updates) == 0 {
		return nil
	}

	// Stitch the query together safely
	query, args := JoinSetFieldsClause("UPDATE cameras SET ", updates, true)
	query = query + " WHERE deleted=0 AND id = ?"
	args = append(args, id)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		if isUniqueConstraintViolation(err) {
			errStr := err.Error()
			if strings.Contains(errStr, "serial_number") || strings.Contains(errStr, "ip_address") || strings.Contains(errStr, "idx_cam_dedup") {
				return ErrCameraHardwareConflict
			}
			return ErrCameraExists
		}

		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrCameraNotFound
	}

	return nil
}