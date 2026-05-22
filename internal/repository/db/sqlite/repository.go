package db

import (
	"context"
	"database/sql"
	"reflect"

	"github.com/nextm/nextm/internal/model"
)

// ─── Auth Repository ──────────────────────────────────────

type AuthRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) CreateAccount(ctx context.Context, arg interface{}) (interface{}, error) {
	v := reflect.ValueOf(arg)
	acct := model.Account{
		ID:           v.FieldByName("ID").String(),
		Email:        v.FieldByName("Email").String(),
		Name:         v.FieldByName("Name").String(),
		PasswordHash: v.FieldByName("PasswordHash").String(),
		Locale:       v.FieldByName("Locale").String(),
		Timezone:     v.FieldByName("Timezone").String(),
		CreatedAt:    v.FieldByName("CreatedAt").Int(),
		UpdatedAt:    v.FieldByName("UpdatedAt").Int(),
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO accounts (id, email, name, password_hash, locale, timezone, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		acct.ID, acct.Email, acct.Name, acct.PasswordHash,
		acct.Locale, acct.Timezone, acct.CreatedAt, acct.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &acct, nil
}

func (r *AuthRepository) GetAccountByEmail(ctx context.Context, email string) (interface{}, error) {
	var a model.Account
	var lastLoginAt sql.NullInt64
	err := r.db.QueryRowContext(ctx,
		`SELECT id, email, name, avatar_url, auth_provider, password_hash, locale, timezone,
		        is_active, last_login_at, created_at, updated_at
		 FROM accounts WHERE email = ? AND is_active = 1`, email).
		Scan(&a.ID, &a.Email, &a.Name, &a.AvatarURL, &a.AuthProvider, &a.PasswordHash,
			&a.Locale, &a.Timezone, &a.IsActive, &lastLoginAt, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if lastLoginAt.Valid {
		a.LastLoginAt = &lastLoginAt.Int64
	}
	return &a, nil
}

func (r *AuthRepository) GetAccountByID(ctx context.Context, id string) (interface{}, error) {
	var a model.Account
	var lastLoginAt sql.NullInt64
	err := r.db.QueryRowContext(ctx,
		`SELECT id, email, name, avatar_url, auth_provider, password_hash, locale, timezone,
		        is_active, last_login_at, created_at, updated_at
		 FROM accounts WHERE id = ?`, id).
		Scan(&a.ID, &a.Email, &a.Name, &a.AvatarURL, &a.AuthProvider, &a.PasswordHash,
			&a.Locale, &a.Timezone, &a.IsActive, &lastLoginAt, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if lastLoginAt.Valid {
		a.LastLoginAt = &lastLoginAt.Int64
	}
	return &a, nil
}

func (r *AuthRepository) UpdateLastLogin(ctx context.Context, accountID string, lastLoginAt, updatedAt int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE accounts SET last_login_at = ?, updated_at = ? WHERE id = ?`,
		lastLoginAt, updatedAt, accountID)
	return err
}

func (r *AuthRepository) CreateRefreshToken(ctx context.Context, arg interface{}) error {
	v := reflect.ValueOf(arg)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO refresh_tokens (id, account_id, token_hash, device_id, expires_at, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		v.FieldByName("ID").String(),
		v.FieldByName("AccountID").String(),
		v.FieldByName("TokenHash").String(),
		"",
		v.FieldByName("ExpiresAt").Int(),
		v.FieldByName("CreatedAt").Int())
	return err
}

func (r *AuthRepository) GetRefreshToken(ctx context.Context, id string, now int64) (interface{}, error) {
	var rt model.RefreshToken
	err := r.db.QueryRowContext(ctx,
		`SELECT id, account_id, token_hash, device_id, expires_at, revoked, created_at
		 FROM refresh_tokens WHERE id = ? AND revoked = 0 AND expires_at > ?`, id, now).
		Scan(&rt.ID, &rt.AccountID, &rt.TokenHash, &rt.DeviceID, &rt.ExpiresAt, &rt.Revoked, &rt.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *AuthRepository) RevokeRefreshToken(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE refresh_tokens SET revoked = 1 WHERE id = ?`, id)
	return err
}

func (r *AuthRepository) RevokeAllAccountTokens(ctx context.Context, accountID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE refresh_tokens SET revoked = 1 WHERE account_id = ? AND revoked = 0`, accountID)
	return err
}

func (r *AuthRepository) CreateSpace(ctx context.Context, arg interface{}) (interface{}, error) {
	v := reflect.ValueOf(arg)
	s := model.Space{
		ID:        v.FieldByName("ID").String(),
		Name:      v.FieldByName("Name").String(),
		Type:      v.FieldByName("Type").String(),
		AccountID: v.FieldByName("AccountID").String(),
		CreatedAt: v.FieldByName("CreatedAt").Int(),
		UpdatedAt: v.FieldByName("UpdatedAt").Int(),
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO spaces (id, name, type, account_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		s.ID, s.Name, s.Type, s.AccountID, s.CreatedAt, s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *AuthRepository) GetPersonalSpace(ctx context.Context, accountID string) (interface{}, error) {
	var s model.Space
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, type, account_id, icon, description, object_count, sync_status, is_deleted, created_at, updated_at
		 FROM spaces WHERE account_id = ? AND type = 'personal' AND is_deleted = 0 LIMIT 1`, accountID).
		Scan(&s.ID, &s.Name, &s.Type, &s.AccountID, &s.Icon, &s.Description,
			&s.ObjectCount, &s.SyncStatus, &s.IsDeleted, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *AuthRepository) CreateDefaultTypes(ctx context.Context, arg interface{}) error {
	v := reflect.ValueOf(arg)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO object_types (id, space_id, name, icon, description, is_builtin, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, 1, ?, ?)`,
		v.FieldByName("ID").String(),
		v.FieldByName("SpaceID").String(),
		v.FieldByName("Name").String(),
		v.FieldByName("Icon").String(),
		v.FieldByName("Description").String(),
		v.FieldByName("CreatedAt").Int(),
		v.FieldByName("UpdatedAt").Int())
	return err
}

// ─── Object Repository ────────────────────────────────────

type ObjectRepository struct {
	db *sql.DB
}

func NewObjectRepository(db *sql.DB) *ObjectRepository {
	return &ObjectRepository{db: db}
}

func reflectObject(v reflect.Value) *model.KnowledgeObject {
	obj := &model.KnowledgeObject{
		ID:         v.FieldByName("ID").String(),
		SpaceID:    v.FieldByName("SpaceID").String(),
		TypeID:     v.FieldByName("TypeID").String(),
		Title:      v.FieldByName("Title").String(),
		Properties: v.FieldByName("Properties").String(),
		Tags:       v.FieldByName("Tags").String(),
		Source:     v.FieldByName("Source").String(),
		SourceMeta: v.FieldByName("SourceMeta").String(),
		WordCount:  int(v.FieldByName("WordCount").Int()),
		Version:    int(v.FieldByName("Version").Int()),
		CreatedAt:  v.FieldByName("CreatedAt").Int(),
		UpdatedAt:  v.FieldByName("UpdatedAt").Int(),
	}
	if cv := v.FieldByName("CoverImage"); cv.Kind() == reflect.Ptr && !cv.IsNil() {
		s := cv.Elem().String()
		obj.CoverImage = &s
	}
	return obj
}

func scanObject(scanner interface {
	Scan(dest ...interface{}) error
}, obj *model.KnowledgeObject) error {
	var coverImage, embeddingID sql.NullString
	var lastReadAt sql.NullInt64
	err := scanner.Scan(
		&obj.ID, &obj.SpaceID, &obj.TypeID, &obj.Title,
		&obj.Properties, &obj.Tags, &coverImage,
		&obj.Source, &obj.SourceMeta, &embeddingID,
		&obj.WordCount, &obj.Version, &obj.IsArchived, &obj.IsDeleted,
		&lastReadAt, &obj.SyncStatus, &obj.CreatedAt, &obj.UpdatedAt)
	if err != nil {
		return err
	}
	if coverImage.Valid {
		obj.CoverImage = &coverImage.String
	}
	if embeddingID.Valid {
		obj.EmbeddingID = &embeddingID.String
	}
	if lastReadAt.Valid {
		obj.LastReadAt = &lastReadAt.Int64
	}
	return nil
}

const objectColumns = `id, space_id, type_id, title, properties, tags,
	cover_image, source, source_meta, embedding_id,
	word_count, version, is_archived, is_deleted,
	last_read_at, sync_status, created_at, updated_at`

func (r *ObjectRepository) ListObjects(ctx context.Context, spaceID string, limit, offset int) ([]*model.KnowledgeObject, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+objectColumns+` FROM objects
		 WHERE space_id = ? AND is_deleted = 0
		 ORDER BY updated_at DESC LIMIT ? OFFSET ?`, spaceID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanObjects(rows)
}

func (r *ObjectRepository) ListObjectsByType(ctx context.Context, spaceID, typeID string, limit, offset int) ([]*model.KnowledgeObject, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+objectColumns+` FROM objects
		 WHERE space_id = ? AND type_id = ? AND is_deleted = 0
		 ORDER BY updated_at DESC LIMIT ? OFFSET ?`, spaceID, typeID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanObjects(rows)
}

func scanObjects(rows *sql.Rows) ([]*model.KnowledgeObject, error) {
	var result []*model.KnowledgeObject
	for rows.Next() {
		obj := &model.KnowledgeObject{}
		if err := scanObject(rows, obj); err != nil {
			return nil, err
		}
		result = append(result, obj)
	}
	return result, rows.Err()
}

func (r *ObjectRepository) GetObject(ctx context.Context, id, spaceID string) (*model.KnowledgeObject, error) {
	obj := &model.KnowledgeObject{}
	err := scanObject(r.db.QueryRowContext(ctx,
		`SELECT `+objectColumns+` FROM objects WHERE id = ? AND space_id = ? AND is_deleted = 0`, id, spaceID), obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (r *ObjectRepository) CreateObject(ctx context.Context, arg interface{}) (*model.KnowledgeObject, error) {
	v := reflect.ValueOf(arg)
	obj := reflectObject(v)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO objects (id, space_id, type_id, title, properties, tags,
		 cover_image, source, source_meta, word_count, version, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)`,
		obj.ID, obj.SpaceID, obj.TypeID, obj.Title, obj.Properties, obj.Tags,
		obj.CoverImage, obj.Source, obj.SourceMeta, obj.WordCount,
		obj.CreatedAt, obj.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (r *ObjectRepository) UpdateObject(ctx context.Context, arg interface{}) (*model.KnowledgeObject, error) {
	v := reflect.ValueOf(arg)
	obj := reflectObject(v)
	_, err := r.db.ExecContext(ctx,
		`UPDATE objects SET title = ?, properties = ?, tags = ?, cover_image = ?,
		 word_count = ?, version = version + 1, updated_at = ?
		 WHERE id = ? AND space_id = ? AND is_deleted = 0`,
		obj.Title, obj.Properties, obj.Tags, obj.CoverImage,
		obj.WordCount, obj.UpdatedAt, obj.ID, obj.SpaceID)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (r *ObjectRepository) SoftDeleteObject(ctx context.Context, id, spaceID string, updatedAt int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE objects SET is_deleted = 1, updated_at = ? WHERE id = ? AND space_id = ?`,
		updatedAt, id, spaceID)
	return err
}

func (r *ObjectRepository) ArchiveObject(ctx context.Context, id, spaceID string, updatedAt int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE objects SET is_archived = CASE WHEN is_archived = 0 THEN 1 ELSE 0 END, updated_at = ?
		 WHERE id = ? AND space_id = ?`, updatedAt, id, spaceID)
	return err
}

func (r *ObjectRepository) SearchObjects(ctx context.Context, spaceID, query string, limit, offset int) ([]*model.KnowledgeObject, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+objectColumns+` FROM objects
		 WHERE space_id = ? AND is_deleted = 0 AND title LIKE '%' || ? || '%'
		 ORDER BY updated_at DESC LIMIT ? OFFSET ?`, spaceID, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanObjects(rows)
}

func (r *ObjectRepository) CountObjectsBySpace(ctx context.Context, spaceID string) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM objects WHERE space_id = ? AND is_deleted = 0`, spaceID).Scan(&count)
	return count, err
}

func (r *ObjectRepository) CountObjectsByType(ctx context.Context, spaceID, typeID string) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM objects WHERE space_id = ? AND type_id = ? AND is_deleted = 0`, spaceID, typeID).Scan(&count)
	return count, err
}

// ─── Block Repository ─────────────────────────────────────

type BlockRepository struct {
	db *sql.DB
}

func NewBlockRepository(db *sql.DB) *BlockRepository {
	return &BlockRepository{db: db}
}

const blockColumns = `id, object_id, parent_id, type, content, properties,
	position, depth, collapsed, color, version, sync_status, created_at, updated_at`

func scanBlock(scanner interface {
	Scan(dest ...interface{}) error
}, b *model.Block) error {
	var parentID sql.NullString
	err := scanner.Scan(
		&b.ID, &b.ObjectID, &parentID, &b.Type, &b.Content, &b.Properties,
		&b.Position, &b.Depth, &b.Collapsed, &b.Color, &b.Version,
		&b.SyncStatus, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return err
	}
	if parentID.Valid {
		b.ParentID = &parentID.String
	}
	return nil
}

func (r *BlockRepository) ListBlocksByObject(ctx context.Context, objectID string) ([]*model.Block, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+blockColumns+` FROM blocks WHERE object_id = ? ORDER BY position ASC`, objectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*model.Block
	for rows.Next() {
		b := &model.Block{}
		if err := scanBlock(rows, b); err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	return result, rows.Err()
}

func (r *BlockRepository) GetBlock(ctx context.Context, id string) (*model.Block, error) {
	b := &model.Block{}
	err := scanBlock(r.db.QueryRowContext(ctx,
		`SELECT `+blockColumns+` FROM blocks WHERE id = ?`, id), b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (r *BlockRepository) CreateBlock(ctx context.Context, arg interface{}) (*model.Block, error) {
	v := reflect.ValueOf(arg)
	b := &model.Block{
		ID:         v.FieldByName("ID").String(),
		ObjectID:   v.FieldByName("ObjectID").String(),
		Type:       v.FieldByName("Type").String(),
		Content:    v.FieldByName("Content").String(),
		Properties: v.FieldByName("Properties").String(),
		Position:   v.FieldByName("Position").Float(),
		Depth:      int(v.FieldByName("Depth").Int()),
		Color:      v.FieldByName("Color").String(),
		Version:    1,
		SyncStatus: "synced",
		CreatedAt:  v.FieldByName("CreatedAt").Int(),
		UpdatedAt:  v.FieldByName("UpdatedAt").Int(),
	}
	if p := v.FieldByName("ParentID"); p.Kind() == reflect.Ptr && !p.IsNil() {
		s := p.Elem().String()
		b.ParentID = &s
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO blocks (id, object_id, parent_id, type, content, properties,
		 position, depth, color, version, sync_status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1, 'synced', ?, ?)`,
		b.ID, b.ObjectID, b.ParentID, b.Type, b.Content, b.Properties,
		b.Position, b.Depth, b.Color, b.CreatedAt, b.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (r *BlockRepository) UpdateBlock(ctx context.Context, arg interface{}) (*model.Block, error) {
	v := reflect.ValueOf(arg)
	b := &model.Block{
		ID:         v.FieldByName("ID").String(),
		ObjectID:   v.FieldByName("ObjectID").String(),
		Content:    v.FieldByName("Content").String(),
		Properties: v.FieldByName("Properties").String(),
		Type:       v.FieldByName("Type").String(),
		Position:   v.FieldByName("Position").Float(),
		Depth:      int(v.FieldByName("Depth").Int()),
		Color:      v.FieldByName("Color").String(),
		UpdatedAt:  v.FieldByName("UpdatedAt").Int(),
	}
	_, err := r.db.ExecContext(ctx,
		`UPDATE blocks SET content = ?, properties = ?, type = ?,
		 position = ?, depth = ?, color = ?,
		 version = version + 1, updated_at = ?
		 WHERE id = ? AND object_id = ?`,
		b.Content, b.Properties, b.Type, b.Position, b.Depth, b.Color,
		b.UpdatedAt, b.ID, b.ObjectID)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (r *BlockRepository) DeleteBlock(ctx context.Context, id, objectID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM blocks WHERE id = ? AND object_id = ?`, id, objectID)
	return err
}

func (r *BlockRepository) CountBlocksByObject(ctx context.Context, objectID string) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM blocks WHERE object_id = ?`, objectID).Scan(&count)
	return count, err
}
