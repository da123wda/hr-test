package dao

import (
	"errors"
	"hr-dynamic-form-engine/src/model"

	"xorm.io/xorm"
)

type fileDao struct {
	conn *xorm.Engine
}

func FileDbDefault() fileDao {
	return fileDao{conn: DBDefault()}
}

func (d fileDao) InsertOne(file *model.FileMetadata) error {
	_, err := d.conn.InsertOne(file)
	return err
}

func (d fileDao) GetById(fileId string) (*model.FileMetadata, error) {
	f := new(model.FileMetadata)
	has, err := d.conn.Where("file_id = ?", fileId).Get(f)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("文件不存在")
	}
	return f, nil
}
