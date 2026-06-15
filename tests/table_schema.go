package tests

import (
	"time"

	q "github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/ddl"
)

// UserStatus is a test enum used by Person.
type UserStatus string

const (
	// UserStatusActive is an active user status.
	UserStatusActive UserStatus = "active"
	// UserStatusBlocked is a blocked user status.
	UserStatusBlocked UserStatus = "blocked"
	// UserStatusDeleted is a deleted user status.
	UserStatusDeleted UserStatus = "deleted"
	// UserStatusPending is a pending user status.
	UserStatusPending UserStatus = "pending"
)

const personTableName = "users"

// Person is a schema-rich test table model.
type Person struct {
	q.Table `table:"users"`

	ID q.Column[int64] `q:"type:bigint,pk"`

	OrgID q.Column[int64]

	UserName q.Column[string] `q:"uq"`
	Email    q.Column[string]

	DisplayName q.Column[*string]

	Age q.Column[*int]

	Status q.Column[UserStatus] `q:"default:'pending'"`

	IsAdmin    q.Column[bool] `q:"nn,default:false"`
	IsVerified q.Column[bool]

	Profile q.Column[[]byte]

	LastLoginAt q.Column[*time.Time] `db:"last_login"`

	CreatedAt q.Column[time.Time] `db:"time_created" q:"nn,default:now()"`
	UpdatedAt q.Column[time.Time] `db:"time_updated" q:"nn,default:now()"`

	DeletedAt q.Column[*time.Time]
}

// TableConfig returns explicit schema metadata for Person.
func (p Person) TableConfig() q.TableConfig { //nolint:gocritic // q.NewTable currently binds value table models.
	return q.TableConfig{
		Name:   personTableName,
		Schema: "public",

		Columns: q.ColumnsConfig{
			p.ID.DDLKey(): {
				Type:       ddl.BigInt(),
				NotNull:    false,
				PrimaryKey: true,
			},

			p.UserName.DDLKey(): {
				Type: ddl.VarChar(64),
			},

			p.DisplayName.DDLKey(): {
				Type:    ddl.VarChar(120),
				NotNull: true,
			},

			p.Status.DDLKey(): {
				Default: ddl.RawExpr("'pending'"),
			},

			p.IsVerified.DDLKey(): {
				Default: ddl.Literal(false),
			},

			p.Profile.DDLKey(): {
				Type:    ddl.JSONB(),
				Default: ddl.RawExpr("'{}'::jsonb"),
			},
		},

		Constraints: q.ConstraintsConfig{
			ddl.PrimaryKey(p.ID.Name()).Named("pk_users"),

			ddl.Unique(p.OrgID.Name(), p.Email.Name()).Named("uq_users_org_email"),

			ddl.Check(
				ddl.Or(
					ddl.Col(p.Age.Name()).IsNull(),
					ddl.Col(p.Age.Name()).Ge(0),
				),
			).Named("ck_users_age_valid"),

			ddl.Check(ddl.RawPred("status IN ('pending', 'active', 'blocked', 'deleted')")),

			ddl.ForeignKey(p.OrgID.Name()).
				References("organizations", "ID").
				Named("fk_users_org").
				OnUpdate(ddl.Cascade).
				OnDelete(ddl.Restrict),
		},

		Indexes: q.IndexesConfig{
			ddl.IndexCols("ix_users_org_id", p.OrgID.Name()),

			ddl.IndexCols("ix_users_status", p.Status.Name()),

			ddl.IndexCols("ix_users_org_active", p.OrgID.Name(), p.UserName.Name()).
				Where(ddl.RawPred("deleted_at IS NULL AND status = 'active'")),

			ddl.IndexCols("ux_users_org_active", p.OrgID.Name(), p.UserName.Name()).
				Unique().
				Where(ddl.Col(p.DeletedAt.Name()).IsNotNull()),

			ddl.Index(
				"ux_users_org_lower_email_active",
				ddl.KeyCol(p.OrgID.Name()).NullsFirst(),
				ddl.Key(
					ddl.Func("lower", ddl.Col(p.Email.Name())),
				),
			),

			ddl.Index(
				"ix_users_search_text_trgm",
				ddl.Key(ddl.RawExpr("search_text gin_trgm_ops")),
			).Using("gin"),
		},
	}
}
