package main

import (
	"fmt"
	q "github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/dialect"
)

type Post struct {
	q.Table `table:"posts"`

	ID          q.Column[int] `db:"id"`
	Title       q.Column[string]
	Description q.Column[string]
}

type Comment struct {
	q.Table `table:"comments"`

	ID      q.Column[int] `db:"id"`
	PostId  q.Column[int] `db:"post_id"`
	Content q.Column[string]
}

func main() {

	posts := q.MustNewTable[Post]()
	comments := q.MustNewTable[Comment]()

	sql, args, err := q.Select(comments.Content, posts.Title, posts.Description).
		Where(
			posts.ID.Eq(1),
		).
		OrderBy(posts.ID.Asc()).
		Join(comments, comments.PostId.Eq(posts.ID)).
		Limit(10).
		Render(dialect.PostgreSQL{})

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(sql)
	fmt.Println(args)

}
