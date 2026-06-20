package migrations

import (
	"time"

	q "github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/ddl"
)

type Organization struct {
	q.Table `table:"organizations"`

	ID        q.Column[string]
	Slug      q.Column[string]
	Name      q.Column[string]
	CreatedAt q.Column[time.Time]
}

func (o Organization) TableConfig() q.TableConfig { //nolint:gocritic // q.NewTable currently binds value table models.
	return q.TableConfig{
		Name:   "organizations",
		Schema: "public",
		Columns: q.ColumnsConfig{
			o.ID.DDLKey(): {
				Type:    ddl.UUID(),
				NotNull: true,
				Default: ddl.RawExpr("gen_random_uuid()"),
			},
			o.Slug.DDLKey(): {
				Type:    ddl.VarChar(64),
				NotNull: true,
			},
			o.Name.DDLKey(): {
				Type:    ddl.Text(),
				NotNull: true,
			},
			o.CreatedAt.DDLKey(): {
				Type:    ddl.TimestampTZ(),
				NotNull: true,
				Default: ddl.RawExpr("now()"),
			},
		},
		Constraints: q.ConstraintsConfig{
			ddl.PrimaryKey(o.ID.Name()).Named("pk_organizations"),
			ddl.Unique(o.Slug.Name()).Named("uq_organizations_slug"),
		},
	}
}

type User struct {
	q.Table `table:"users"`

	ID             q.Column[string]
	OrganizationID q.Column[string]
	Email          q.Column[string]
	DisplayName    q.Column[string]
	IsActive       q.Column[bool]
	CreatedAt      q.Column[time.Time]
	DeletedAt      q.Column[*time.Time]
}

func (u User) TableConfig() q.TableConfig { //nolint:gocritic // q.NewTable currently binds value table models.
	return q.TableConfig{
		Name:   "users",
		Schema: "public",
		Columns: q.ColumnsConfig{
			u.ID.DDLKey(): {
				Type:    ddl.UUID(),
				NotNull: true,
				Default: ddl.RawExpr("gen_random_uuid()"),
			},
			u.OrganizationID.DDLKey(): {
				Type:    ddl.UUID(),
				NotNull: true,
			},
			u.Email.DDLKey(): {
				Type:    ddl.VarChar(320),
				NotNull: true,
			},
			u.DisplayName.DDLKey(): {
				Type:    ddl.Text(),
				NotNull: true,
				Default: ddl.Literal(""),
			},
			u.IsActive.DDLKey(): {
				Type:    ddl.Boolean(),
				NotNull: true,
				Default: ddl.Literal(true),
			},
			u.CreatedAt.DDLKey(): {
				Type:    ddl.TimestampTZ(),
				NotNull: true,
				Default: ddl.RawExpr("now()"),
			},
			u.DeletedAt.DDLKey(): {
				Type: ddl.TimestampTZ(),
			},
		},
		Constraints: q.ConstraintsConfig{
			ddl.PrimaryKey(u.ID.Name()).Named("pk_users"),
			ddl.Unique(u.OrganizationID.Name(), u.Email.Name()).Named("uq_users_org_email"),
			ddl.ForeignKey(u.OrganizationID.Name()).
				References("organizations", "id").
				OnDelete(ddl.Cascade).
				Named("fk_users_organization"),
		},
		Indexes: q.IndexesConfig{
			ddl.IndexCols("ix_users_active_email", u.Email.Name()).
				Where(ddl.Col(u.DeletedAt.Name()).IsNull()),
		},
	}
}
