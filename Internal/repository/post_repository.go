package repository

import (
	"go-project-278/Internal/dto"
	"context"
	"fmt"
	"database/sql"
)
type PostRepository interface {
	ListLinks(ctx context.Context) ([]*dto.LinkResponce, error)
	GetLinkByID(ctx context.Context,id int) (*dto.LinkResponce, error)
	DeleteLinkByID(ctx context.Context,id int) (error)
	CreateLink(ctx context.Context,link dto.LinkResponce) (error)
	UpdateLink(ctx context.Context,link dto.LinkResponce) (error)
}
type Repository struct {
	db *sql.DB
}

func NewLinkRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListLinks(ctx context.Context) ([]*dto.LinkResponce, error) {
	query := `
		SELECT * FROM links;
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list links: %w", err)
	}
	defer rows.Close()
	var links []*dto.LinkResponce
	for rows.Next() {
		var link dto.LinkResponce
		err := rows.Scan(
			&link.Id,
			&link.Original_url,
			&link.Short_name,
			&link.Short_url,
		)
		if err != nil {
			return nil, fmt.Errorf("scan link: %w", err)
		}
		links = append(links, &link)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return links, nil
}

func (r *Repository) GetLinkByID(ctx context.Context,id int) (*dto.LinkResponce, error) {
	query := `
		SELECT * FROM links
		WHERE id = $1;
	`
	var link dto.LinkResponce
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&link.Id,
		&link.Original_url,
		&link.Short_name,
		&link.Short_url,
	)
	if err != nil {
		return nil, fmt.Errorf("get link: %w", err)
	}
	
	return &link, nil
}

func (r *Repository) DeleteLinkByID(ctx context.Context,id int) (error) {
	query := `
		DELETE FROM links
		WHERE id = $1;
	`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete link: %w", err)
	}
	
	return nil
}

func (r *Repository) CreateLink(ctx context.Context,link dto.LinkResponce) (error) {
	query := `
		INSERT INTO links (original_url, short_name, short_url)
		VALUES ($1, $2, $3);
	`
	_, err := r.db.ExecContext(ctx, query, link.Original_url, link.Short_name, link.Short_url)
	if err != nil {
		return fmt.Errorf("create link: %w", err)
	}
	
	return  nil
}

func (r *Repository) UpdateLink(ctx context.Context,link dto.LinkResponce) (error) {
	query := `
		UPDATE links
		SET 
    	original_url = COALESCE($2, original_url),
    	short_name = COALESCE($3, short_name),
    	short_url = COALESCE($4, short_url)
		WHERE id = $1;
	`
	_, err :=  r.db.ExecContext(ctx, query, link.Id, link.Original_url,link.Short_name,link.Short_url)
	if err != nil {
		return fmt.Errorf("update link: %w", err)
	}
	return  nil
}