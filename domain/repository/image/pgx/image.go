package pgx

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"m1-article-service/domain/entity"
)

const pageSize = 10

type ImageRepository struct {
	conn *pgxpool.Pool
}

func NewImageRepository(conn *pgxpool.Pool) *ImageRepository {
	ur := &ImageRepository{
		conn: conn,
	}
	return ur
}

func (r ImageRepository) List(ctx context.Context, offset uint64) ([]*entity.Image, error) {
	images := make([]*entity.Image, 0)
	rows, err := r.conn.Query(ctx, `SELECT * FROM images LIMIT $1 OFFSET $2 `, pageSize, offset)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		image := &entity.Image{}
		if err := rows.Scan(&image.File); err != nil {
			return nil, err
		}
		images = append(images, image)
	}

	if rows.Err() != nil {
		return nil, err
	}
	return images, nil
}

func (r ImageRepository) CreateBatch(ctx context.Context, images []*entity.Image) error {
	batch := &pgx.Batch{}

	for _, image := range images {
		sql := `INSERT INTO images (file) VALUES($1)`
		batch.Queue(sql, image.File)
	}
	br := r.conn.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < batch.Len(); i++ {
		_, err := br.Exec()
		if err != nil {
			return err
		}
	}

	return nil
}
