// Package databasesql shows repository-style qrafter usage with database/sql.
//
// In a real application, import your database driver in the main package, for
// example _ "github.com/jackc/pgx/v5/stdlib".
package databasesql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	q "github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/dialect"
)

// ErrPostNotFound is returned when a write expected to affect a post did not
// match any row.
var ErrPostNotFound = errors.New("post not found")

// UserTable describes the users table once and reuses the typed columns in
// SELECT, INSERT, UPDATE, DELETE, and scanning code.
type UserTable struct {
	q.Table `table:"users"`

	ID   q.Column[int64]  `db:"id"`
	Name q.Column[string] `db:"name"`
}

// PostTable describes blog posts.
type PostTable struct {
	q.Table `table:"posts"`

	ID          q.Column[int64]      `db:"id"`
	AuthorID    q.Column[int64]      `db:"author_id"`
	Title       q.Column[string]     `db:"title"`
	Body        q.Column[string]     `db:"body"`
	PublishedAt q.Column[*time.Time] `db:"published_at"`
	DeletedAt   q.Column[*time.Time] `db:"deleted_at"`
}

// CommentTable describes comments attached to posts.
type CommentTable struct {
	q.Table `table:"comments"`

	ID        q.Column[int64] `db:"id"`
	PostID    q.Column[int64] `db:"post_id"`
	DeletedAt q.Column[*time.Time]
}

// PostSummary is an application-facing result type.
type PostSummary struct {
	ID           int64
	Title        string
	AuthorName   string
	CommentCount int64
	PublishedAt  *time.Time
}

// Store keeps database handles and already-bound table models together.
type Store struct {
	db       *sql.DB
	dialect  dialect.Renderer
	users    UserTable
	posts    PostTable
	comments CommentTable
}

// NewStore binds tables once. Query methods can then reuse the same columns.
func NewStore(db *sql.DB, renderer dialect.Renderer) *Store {
	return &Store{
		db:       db,
		dialect:  renderer,
		users:    q.MustNewTable[UserTable](),
		posts:    q.MustNewTable[PostTable](),
		comments: q.MustNewTable[CommentTable](),
	}
}

// ListPublishedPosts shows a larger SELECT used by a repository method.
func (s *Store) ListPublishedPosts(ctx context.Context, authorID int64, limit int) ([]PostSummary, error) {
	commentCount := q.Count(s.comments.ID).As("comment_count")

	sqlText, args, err := q.Select(
		s.posts.ID,
		s.posts.Title,
		s.users.Name,
		commentCount,
		s.posts.PublishedAt,
	).
		Join(s.users, s.posts.AuthorID.Eq(s.users.ID)).
		LeftJoin(s.comments, s.comments.PostID.Eq(s.posts.ID), s.comments.DeletedAt.IsNull()).
		Where(
			s.posts.AuthorID.Eq(authorID),
			s.posts.PublishedAt.IsNotNull(),
			s.posts.DeletedAt.IsNull(),
		).
		GroupBy(s.posts.ID, s.posts.Title, s.users.Name, s.posts.PublishedAt).
		OrderBy(s.posts.PublishedAt.Desc().NullsLast()).
		Limit(limit).
		Render(s.dialect)
	if err != nil {
		return nil, fmt.Errorf("render published posts query: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, sqlText, args...)
	if err != nil {
		return nil, fmt.Errorf("query published posts: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	posts := make([]PostSummary, 0)
	for rows.Next() {
		var post PostSummary
		if err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.AuthorName,
			&post.CommentCount,
			&post.PublishedAt,
		); err != nil {
			return nil, fmt.Errorf("scan published post: %w", err)
		}
		posts = append(posts, post)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate published posts: %w", err)
	}

	return posts, nil
}

// CreatePost shows INSERT ... RETURNING. Render returns an error if the chosen
// dialect cannot safely render RETURNING.
func (s *Store) CreatePost(ctx context.Context, authorID int64, title, body string) (int64, error) {
	sqlText, args, err := q.Insert(s.posts).
		Columns(s.posts.AuthorID, s.posts.Title, s.posts.Body, s.posts.PublishedAt).
		Values(authorID, title, body, q.Func("now")).
		Returning(s.posts.ID).
		Render(s.dialect)
	if err != nil {
		return 0, fmt.Errorf("render create post query: %w", err)
	}

	var id int64
	if err := s.db.QueryRowContext(ctx, sqlText, args...).Scan(&id); err != nil {
		return 0, fmt.Errorf("create post: %w", err)
	}

	return id, nil
}

// SoftDeletePost shows UPDATE with typed predicates and placeholders.
func (s *Store) SoftDeletePost(ctx context.Context, id int64) error {
	sqlText, args, err := q.Update(s.posts).
		Set(s.posts.DeletedAt, q.Func("now")).
		Where(s.posts.ID.Eq(id), s.posts.DeletedAt.IsNull()).
		Render(s.dialect)
	if err != nil {
		return fmt.Errorf("render soft delete query: %w", err)
	}

	result, err := s.db.ExecContext(ctx, sqlText, args...)
	if err != nil {
		return fmt.Errorf("soft delete post: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read affected rows: %w", err)
	}
	if affected == 0 {
		return ErrPostNotFound
	}

	return nil
}
