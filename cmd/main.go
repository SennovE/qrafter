package main

import (
	"fmt"

	q "github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/dialect"
)

type User struct {
	UserName q.Column[string]
	Age      q.Column[int] `db:"user_age"`

	Meta string
	meta q.Column[int]
}

func (User) TableConfig() q.TableConfig {
	return q.TableConfig{
		Name: "user_table",
	}
}

func makeUser() User {
	var user User
	q.Bind(&user)
	return user
}

var user = makeUser()
var empl, _ = q.TableAlias(user, "empl")

func main() {
	q := q.Select(
		user.UserName,
		q.As(user.Age, "gage"),
		empl.UserName,
		q.Sum(
			q.Const(123),
			q.Const(321),
		),
	).Where(
		q.Eq(user.UserName, q.Const("ABC")),
		q.Or(
			q.Ge(user.Age, q.Const("1")),
			q.Eq(q.Const("Test"), user.UserName),
		),
	)
	fmt.Println(q.Render(dialect.PostgreSQL{}))
}
