package main

import (
	"fmt"

	"github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/expr"
	"github.com/SennovE/qrafter/pred"
	"github.com/SennovE/qrafter/query"
)

type User struct {
	UserName expr.Column[string]
	Age      expr.Column[int] `db:"user_age"`

	Meta string
	meta expr.Column[int]
}

func (User) TableConfig() qrafter.TableConfig {
	return qrafter.TableConfig{
		Name: "user_table",
	}
}

func makeUser() User {
	var user User
	qrafter.Bind(&user)
	return user
}

var user = makeUser()
var empl, _ = qrafter.TableAlias(user, "empl")

func main() {
	q := query.Select(
		user.UserName,
		expr.As(user.Age, "gage"),
		empl.UserName,
		expr.Sum(
			expr.ConstExpr(123),
			expr.ConstExpr(321),
		),
	).Where(
		pred.Eq(user.UserName, expr.ConstExpr("ABC")),
		pred.And(
			pred.Ge(expr.ConstExpr("1"), expr.ConstExpr("3")),
			pred.Ge(expr.ConstExpr("100"), expr.ConstExpr("md")),
		),
	)
	fmt.Println(q.Render())
}
