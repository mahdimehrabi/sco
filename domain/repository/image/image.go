package image

import (
	"context"
	"errors"
	"m1-article-service/domain/entity"
)

var (
	ErrAlreadyExist = errors.New("already exist")
	ErrNotFound     = errors.New("not found")
)

type Image interface {
	CreateBatch(context.Context, []*entity.Image) error
	List(context.Context, uint64) ([]*entity.Image, error)
}
