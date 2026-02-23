package repository

import (
    "database/sql"
    "errors"
    "time"

    "qr-tracker/internal/models"
)

type LinkRepository interface {
    Create(link *models.Link) error
    GetByCode(code string) (*models.Link, error)
    IncrementClick(code string) error
}

type sqliteLinkRepo struct {
    db *sql.DB
}

func NewSQLiteLinkRepository(db *sql.DB) LinkRepository {
    return &sqliteLinkRepo{db: db}
}

func (r *sqliteLinkRepo) Create(link *models.Link) error {
    _, err := r.db.Exec(`INSERT INTO links(id,code,target_url,click_count,created_at,is_active) VALUES(?,?,?,?,?,?)`,
        link.ID, link.Code, link.TargetURL, 0, link.CreatedAt.UTC().Format(time.RFC3339), 1)
    return err
}

func (r *sqliteLinkRepo) GetByCode(code string) (*models.Link, error) {
    row := r.db.QueryRow(`SELECT id,code,target_url,click_count,created_at,is_active FROM links WHERE code = ?`, code)
    var id, c, target, created string
    var clicks int64
    var isActive int
    if err := row.Scan(&id, &c, &target, &clicks, &created, &isActive); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, nil
        }
        return nil, err
    }
    t, _ := time.Parse(time.RFC3339, created)
    return &models.Link{ID: id, Code: c, TargetURL: target, Clicks: clicks, CreatedAt: t, IsActive: isActive == 1}, nil
}

func (r *sqliteLinkRepo) IncrementClick(code string) error {
    _, err := r.db.Exec(`UPDATE links SET click_count = click_count + 1 WHERE code = ?`, code)
    return err
}
