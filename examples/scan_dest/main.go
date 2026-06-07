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
		Where(posts.ID.Eq(1)).
		Render(dialect.PostgreSQL{})
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(sql)
	fmt.Println(args)

	dest, err := q.ScanDest(&posts)
	if err != nil {
		fmt.Println(err)
	}

	sampleRow := []any{int64(1), "ScanDest example", "ScanDest fills q.Column fields"}
	for i, value := range sampleRow {
		scanner, ok := dest[i].(interface{ Scan(any) error })
		if !ok {
			fmt.Printf("destination %d cannot scan values\n", i)
		}
		if err := scanner.Scan(value); err != nil {
			fmt.Println(err)
		}
	}

	fmt.Println(posts.ID.Get())
	fmt.Println(posts.Title.Get())
	fmt.Println(posts.Description.Get())
}
