package repository

import (
	"go-project-278/Internal/dto"
	"context"
	"fmt"
	"database/sql"
)
type PostRepository interface {
	ListLinks(ctx context.Context) ([]*dto.LinkResponce, error)
	GetLinkByID(ctx context.Context, id int) (*dto.LinkResponce, error)
	GetLinkByShortName(ctx context.Context, shortName string) (*dto.LinkResponce, error) 
	DeleteLinkByID(ctx context.Context, id int) error
	CreateLink(ctx context.Context, link dto.LinkResponce) error
	UpdateLink(ctx context.Context, link dto.LinkResponce) error
	ListLinksLimited(ctx context.Context, start, limit int) ([]*dto.LinkResponce, error)
	RecordVisit(ctx context.Context, visit dto.Visit) error 
	ListVisits(ctx context.Context) ([]*dto.Visit, error) 
	ListVisitsLimited(ctx context.Context, start, limit int) ([]*dto.Visit, error) 
	CheckShortNameExists(ctx context.Context, shortName string) (bool, error)
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
func (r *Repository) ListLinksLimited(ctx context.Context, start, limit int) ([]*dto.LinkResponce, error) {
    query := `
        SELECT id, original_url, short_name, short_url 
        FROM links 
        ORDER BY id
        LIMIT $1 OFFSET $2
    `
    rows, err := r.db.QueryContext(ctx, query, limit, start)
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

func (r *Repository) GetLinkByShortName(ctx context.Context, shortName string) (*dto.LinkResponce, error) {
	query := `SELECT id, original_url, short_name, short_url FROM links WHERE short_name = $1;`
	var link dto.LinkResponce
	err := r.db.QueryRowContext(ctx, query, shortName).Scan(
		&link.Id, &link.Original_url, &link.Short_name, &link.Short_url,
	)
	if err != nil {
		return nil, fmt.Errorf("get link by short name: %w", err)
	}
	return &link, nil
}

func (r *Repository) RecordVisit(ctx context.Context, v dto.Visit) error {
	query := `
		INSERT INTO link_visits (link_id, ip, user_agent, referer, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6);
	`
	_, err := r.db.ExecContext(ctx, query, v.LinkID, v.IP, v.UserAgent,v.Status, v.CreatedAt)
	if err != nil {
		return fmt.Errorf("record visit: %w", err)
	}
	return nil
}

func (r *Repository) ListVisits(ctx context.Context) ([]*dto.Visit, error) {
	query := `SELECT id, link_id, ip, user_agent, referer, status, created_at FROM link_visits ORDER BY created_at DESC;`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var visits []*dto.Visit
	for rows.Next() {
		var v dto.Visit
		if err := rows.Scan(&v.Id, &v.LinkID, &v.IP, &v.UserAgent, &v.Status, &v.CreatedAt); err != nil {
			return nil, err
		}
		visits = append(visits, &v)
	}
	return visits, nil
}

func (r *Repository) ListVisitsLimited(ctx context.Context, start, limit int) ([]*dto.Visit, error) {
	query := `
		SELECT id, link_id, ip, user_agent, referer, status, created_at 
		FROM link_visits 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2;
	`
	rows, err := r.db.QueryContext(ctx, query, limit, start)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var visits []*dto.Visit
	for rows.Next() {
		var v dto.Visit
		if err := rows.Scan(&v.Id, &v.LinkID, &v.IP, &v.UserAgent,&v.Status, &v.CreatedAt); err != nil {
			return nil, err
		}
		visits = append(visits, &v)
	}
	return visits, nil
}

func (r *Repository) CheckShortNameExists(ctx context.Context, shortName string) (bool, error) {
    query := `SELECT EXISTS(SELECT 1 FROM links WHERE short_name = $1);`
    var exists bool
    err := r.db.QueryRowContext(ctx, query, shortName).Scan(&exists)
    if err != nil {
        return false, fmt.Errorf("check short name exists: %w", err)
    }
    return exists, nil
}