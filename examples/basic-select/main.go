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

func main() {

	posts := q.MustNewTable[Post]()

	sql, args, err := q.Select(posts.ID, posts.Title, posts.Description).
		Where(
			posts.Title.Eq("test"),
		).
		OrderBy(posts.ID.Asc()).
		Limit(10).
		Render(dialect.PostgreSQL{})

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(sql)
	fmt.Println(args)
}
