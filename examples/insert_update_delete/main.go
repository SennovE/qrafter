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

	//insert a row
	sql, args := q.Insert(posts).
		Set(posts.ID, 1).
		Set(posts.Title, "test").
		Set(posts.Description, "Test-Description").MustRender(dialect.PostgreSQL{})

	fmt.Println(sql)
	fmt.Println(args)

	//update a row
	sql, args, err := q.Update(posts).
		Set(posts.Title, "updated-title").
		Where(
			posts.ID.Eq(1)).
		Render(dialect.PostgreSQL{})

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(sql)
	fmt.Println(args)

	//delete a row
	sql, args, err = q.Delete(posts).
		Where(
			posts.ID.Eq(1),
		).
		Render(dialect.PostgreSQL{})

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(sql)
	fmt.Println(args)
}
