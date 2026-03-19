package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresMembershipChecker checks if two users share at least one project via project_members.
type PostgresMembershipChecker struct {
	pool *pgxpool.Pool
}

func NewPostgresMembershipChecker(pool *pgxpool.Pool) *PostgresMembershipChecker {
	return &PostgresMembershipChecker{pool: pool}
}

func (c *PostgresMembershipChecker) ShareProject(ctx context.Context, userA, userB string) (bool, error) {
	query := `SELECT EXISTS(
		SELECT 1 FROM project_members a
		INNER JOIN project_members b ON a.project_id = b.project_id
		WHERE a.user_id = $1 AND b.user_id = $2
	)`
	var exists bool
	if err := c.pool.QueryRow(ctx, query, userA, userB).Scan(&exists); err != nil {
		return false, fmt.Errorf("membershipChecker.ShareProject: %w", err)
	}
	return exists, nil
}
