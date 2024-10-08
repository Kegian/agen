// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.24.0
// source: users.sql

package query

import (
	"context"

	"github.com/google/uuid"
)

const createUser = `-- name: CreateUser :one
INSERT INTO users (id, name) VALUES ($1, $2) RETURNING id, name
`

type CreateUserParams struct {
	ID   uuid.UUID
	Name string
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (*User, error) {
	row := q.db.QueryRow(ctx, createUser, arg.ID, arg.Name)
	var i User
	err := row.Scan(&i.ID, &i.Name)
	return &i, err
}

const getUser = `-- name: GetUser :one
SELECT id, name FROM users WHERE id = $1 LIMIT 1
`

func (q *Queries) GetUser(ctx context.Context, id uuid.UUID) (*User, error) {
	row := q.db.QueryRow(ctx, getUser, id)
	var i User
	err := row.Scan(&i.ID, &i.Name)
	return &i, err
}
