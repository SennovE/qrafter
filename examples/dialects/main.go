package main

import (
	"fmt"
	"log"

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

	query := q.Select(posts.ID, posts.Title, posts.Description).
		Where(posts.Title.Eq("test"), posts.ID.Eq(10)).
		OrderBy(posts.ID.Desc()).
		Limit(5).
		Offset(10)

	render("PostgreSQL", query, dialect.PostgreSQL{})
	render("MySQL", query, dialect.MySQL{})
	render("SQLite", query, dialect.SQLite{})
}

func render(name string, query q.SelectQuery, renderer dialect.Renderer) {
	sql, args, err := query.Render(renderer)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(name)
	fmt.Println(sql)
	fmt.Println(args)
	fmt.Println()
}
