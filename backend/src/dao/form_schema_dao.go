package dao

import (
	"encoding/json"
	"errors"
	"fmt"
	"hr-dynamic-form-engine/src/model"

	"xorm.io/xorm"
)

var ErrOptimisticLockConflict = errors.New("并发修改冲突：当前表单已被他人保存，版本不一致，请刷新后重试")

type formSchemaDao struct {
	conn *xorm.Engine
}

func FormSchemaDbDefault() formSchemaDao {
	return formSchemaDao{conn: DBDefault()}
}

// GetByFormId 根据 formId 获取表单定义
func (d formSchemaDao) GetByFormId(formId string) (*model.FormSchema, error) {
	schema := new(model.FormSchema)
	has, err := d.conn.Where("form_id = ?", formId).Get(schema)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("未找到表单定义")
	}
	return schema, nil
}

// ListAll 获取系统所有表单定义
func (d formSchemaDao) ListAll() ([]model.FormSchema, error) {
	var list []model.FormSchema
	err := d.conn.OrderBy("updated_at DESC").Find(&list)
	return list, err
}

// SaveSchema 保存或更新 Schema，执行乐观锁校验
func (d formSchemaDao) SaveSchema(schema *model.FormSchema, expectedVersion int) error {
	existing, err := d.GetByFormId(schema.FormId)
	if err != nil {
		schema.Version = 1
		schema.Status = model.SchemaStatusDraft
		_, err = d.conn.InsertOne(schema)
		return err
	}

	if expectedVersion > 0 && existing.Version != expectedVersion {
		return fmt.Errorf("%w (您的版本: v%d, 当前数据库版本: v%d)", ErrOptimisticLockConflict, expectedVersion, existing.Version)
	}

	fieldsJSON, err := json.Marshal(schema.Fields)
	if err != nil {
		return fmt.Errorf("序列化字段列表失败: %w", err)
	}

	migrationsJSON, _ := json.Marshal(schema.Migrations)
	if schema.Migrations == nil {
		migrationsJSON = []byte("{}")
	}

	res, err := d.conn.Exec(
		`UPDATE form_schema 
		 SET title = ?, description = ?, fields = ?::jsonb, migrations = ?::jsonb, version = version + 1, updated_at = NOW() 
		 WHERE form_id = ? AND version = ?`,
		schema.Title, schema.Description, string(fieldsJSON), string(migrationsJSON), schema.FormId, existing.Version,
	)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return ErrOptimisticLockConflict
	}
	schema.Version = existing.Version + 1
	return nil
}

// PublishSchema 发布新版本并写入历史快照表
func (d formSchemaDao) PublishSchema(formId string, expectedVersion int, operator string) (*model.FormSchema, error) {
	current, err := d.GetByFormId(formId)
	if err != nil {
		return nil, err
	}
	if expectedVersion > 0 && current.Version != expectedVersion {
		return nil, fmt.Errorf("%w (期望版本: v%d, 实际版本: v%d)", ErrOptimisticLockConflict, expectedVersion, current.Version)
	}

	session := d.conn.NewSession()
	defer session.Close()
	if err = session.Begin(); err != nil {
		return nil, err
	}

	newPublishedVer := current.PublishedVersion + 1
	res, err := session.Exec(
		`UPDATE form_schema 
		 SET published_version = ?, status = 'published', version = version + 1, updated_at = NOW() 
		 WHERE form_id = ? AND version = ?`,
		newPublishedVer, formId, current.Version,
	)
	if err != nil {
		_ = session.Rollback()
		return nil, err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		_ = session.Rollback()
		return nil, ErrOptimisticLockConflict
	}

	fieldsJSON, _ := json.Marshal(current.Fields)
	snapSQL := `INSERT INTO form_schema_version (form_id, version, title, description, fields, published_at, published_by)
	           VALUES (?, ?, ?, ?, ?::jsonb, NOW(), ?)`
	if _, err = session.Exec(snapSQL, formId, newPublishedVer, current.Title, current.Description, string(fieldsJSON), operator); err != nil {
		_ = session.Rollback()
		return nil, err
	}

	if err = session.Commit(); err != nil {
		return nil, err
	}

	current.PublishedVersion = newPublishedVer
	current.Version++
	current.Status = model.SchemaStatusPublished
	return current, nil
}

func (d formSchemaDao) GetVersion(formId string, version int) (*model.FormSchemaVersion, error) {
	snap := new(model.FormSchemaVersion)
	has, err := d.conn.Where("form_id = ? AND version = ?", formId, version).Get(snap)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("历史版本未找到")
	}
	return snap, nil
}

func (d formSchemaDao) ListVersions(formId string) ([]model.FormSchemaVersion, error) {
	var list []model.FormSchemaVersion
	err := d.conn.Where("form_id = ?", formId).OrderBy("version DESC").Find(&list)
	return list, err
}
