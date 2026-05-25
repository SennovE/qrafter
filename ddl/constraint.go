package ddl

import (
	"fmt"
	"strings"

	"github.com/SennovE/qrafter/dialect"
)

type constraintKind string

const (
	constraintPrimaryKey constraintKind = "primary_key"
	constraintUnique     constraintKind = "unique"
	constraintCheck      constraintKind = "check"
	constraintForeignKey constraintKind = "foreign_key"
)

// ReferenceAction is an ON DELETE or ON UPDATE action.
type ReferenceAction string

const (
	// NoAction renders NO ACTION.
	NoAction ReferenceAction = "NO ACTION"
	// Restrict renders RESTRICT.
	Restrict ReferenceAction = "RESTRICT"
	// Cascade renders CASCADE.
	Cascade ReferenceAction = "CASCADE"
	// SetNull renders SET NULL.
	SetNull ReferenceAction = "SET NULL"
	// SetDefault renders SET DEFAULT.
	SetDefault ReferenceAction = "SET DEFAULT"
)

// Constraint describes a table-level constraint.
type Constraint struct {
	state *constraintState
}

type constraintState struct {
	name       string
	kind       constraintKind
	columns    []string
	expr       string
	refTable   string
	refColumns []string
	onDelete   ReferenceAction
	onUpdate   ReferenceAction
}

// PrimaryKey creates a table-level PRIMARY KEY constraint.
func PrimaryKey(columns ...any) Constraint {
	return Constraint{state: &constraintState{kind: constraintPrimaryKey, columns: columnNames(columns)}}
}

// Unique creates a table-level UNIQUE constraint.
func Unique(columns ...any) Constraint {
	return Constraint{state: &constraintState{kind: constraintUnique, columns: columnNames(columns)}}
}

// Check creates a table-level CHECK constraint.
func Check(expr string) Constraint {
	return Constraint{state: &constraintState{kind: constraintCheck, expr: requireName("check expression", expr)}}
}

// ForeignKey creates a table-level FOREIGN KEY constraint.
func ForeignKey(columns ...any) Constraint {
	return Constraint{state: &constraintState{kind: constraintForeignKey, columns: columnNames(columns)}}
}

// Named names the constraint.
func (c Constraint) Named(name string) Constraint {
	state := c.currentState()
	state.name = requireName("constraint", name)
	return Constraint{state: &state}
}

// References sets the referenced table and columns for a FOREIGN KEY.
func (c Constraint) References(table any, columns ...any) Constraint {
	state := c.currentState()
	state.refTable = tableName(table)
	state.refColumns = columnNames(columns)
	return Constraint{state: &state}
}

// OnDelete sets the foreign key ON DELETE action.
func (c Constraint) OnDelete(action ReferenceAction) Constraint {
	state := c.currentState()
	state.onDelete = requireReferenceAction(action)
	return Constraint{state: &state}
}

// OnUpdate sets the foreign key ON UPDATE action.
func (c Constraint) OnUpdate(action ReferenceAction) Constraint {
	state := c.currentState()
	state.onUpdate = requireReferenceAction(action)
	return Constraint{state: &state}
}

func requireReferenceAction(action ReferenceAction) ReferenceAction {
	if action == "" {
		panic(fmt.Errorf("reference action is empty"))
	}
	return action
}

func (c Constraint) render(w *strings.Builder, d dialect.Renderer) {
	state := c.currentState()
	if state.name != "" {
		w.WriteString("CONSTRAINT ")
		w.WriteString(d.QuoteIdent(state.name))
		w.WriteString(" ")
	}

	switch state.kind {
	case constraintPrimaryKey:
		renderNamedColumnConstraint(w, d, "PRIMARY KEY", state.columns)
	case constraintUnique:
		renderNamedColumnConstraint(w, d, "UNIQUE", state.columns)
	case constraintCheck:
		w.WriteString("CHECK (")
		w.WriteString(state.expr)
		w.WriteString(")")
	case constraintForeignKey:
		renderForeignKeyConstraint(w, d, &state)
	default:
		panic(fmt.Errorf("unsupported constraint kind %q", state.kind))
	}
}

func (c Constraint) currentState() constraintState {
	if c.state == nil {
		return constraintState{}
	}
	return *c.state
}

func renderNamedColumnConstraint(w *strings.Builder, d dialect.Renderer, name string, columns []string) {
	if len(columns) == 0 {
		panic(fmt.Errorf("%s constraint must include at least one column", name))
	}
	w.WriteString(name)
	w.WriteString(" (")
	renderColumnList(w, d, columns)
	w.WriteString(")")
}

func renderForeignKeyConstraint(w *strings.Builder, d dialect.Renderer, c *constraintState) {
	if len(c.columns) == 0 {
		panic(fmt.Errorf("FOREIGN KEY constraint must include at least one column"))
	}
	if c.refTable == "" {
		panic(fmt.Errorf("FOREIGN KEY constraint must reference a table"))
	}

	w.WriteString("FOREIGN KEY (")
	renderColumnList(w, d, c.columns)
	w.WriteString(") REFERENCES ")
	w.WriteString(d.QuoteIdent(c.refTable))
	if len(c.refColumns) > 0 {
		w.WriteString(" (")
		renderColumnList(w, d, c.refColumns)
		w.WriteString(")")
	}
	if c.onDelete != "" {
		w.WriteString(" ON DELETE ")
		w.WriteString(string(c.onDelete))
	}
	if c.onUpdate != "" {
		w.WriteString(" ON UPDATE ")
		w.WriteString(string(c.onUpdate))
	}
}
