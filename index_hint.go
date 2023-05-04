package hints

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IndexHint struct {
	Type string
	Keys []string
}

func (indexHint IndexHint) ModifyStatement(stmt *gorm.Statement) {
	for _, name := range []string{"FROM", "UPDATE"} {
		clause := stmt.Clauses[name]

		if clause.AfterExpression == nil {
			clause.AfterExpression = indexHint
		} else {
			clause.AfterExpression = Exprs{clause.AfterExpression, indexHint}
		}

		if name == "FROM" {
			clause.Builder = IndexHintFromClauseBuilder
		}

		stmt.Clauses[name] = clause
	}
}

func (indexHint IndexHint) Build(builder clause.Builder) {
	if len(indexHint.Keys) > 0 {
		builder.WriteString(indexHint.Type)
		builder.WriteByte('(')
		for idx, key := range indexHint.Keys {
			if idx > 0 {
				builder.WriteByte(',')
			}
			builder.WriteQuoted(key)
		}
		builder.WriteByte(')')
	}
}

func UseIndex(names ...string) IndexHint {
	return IndexHint{Type: "USE INDEX ", Keys: names}
}

func IgnoreIndex(names ...string) IndexHint {
	return IndexHint{Type: "IGNORE INDEX ", Keys: names}
}

func ForceIndex(names ...string) IndexHint {
	return IndexHint{Type: "FORCE INDEX ", Keys: names}
}

func (indexHint IndexHint) ForJoin() IndexHint {
	indexHint.Type += "FOR JOIN "
	return indexHint
}

func (indexHint IndexHint) ForOrderBy() IndexHint {
	indexHint.Type += "FOR ORDER BY "
	return indexHint
}

func (indexHint IndexHint) ForGroupBy() IndexHint {
	indexHint.Type += "FOR GROUP BY "
	return indexHint
}

func IndexHintFromClauseBuilder(c clause.Clause, builder clause.Builder) {
	if c.BeforeExpression != nil {
		c.BeforeExpression.Build(builder)
		builder.WriteByte(' ')
	}

	if c.Name != "" {
		builder.WriteString(c.Name)
		builder.WriteByte(' ')
	}

	if c.AfterNameExpression != nil {
		c.AfterNameExpression.Build(builder)
		builder.WriteByte(' ')
	}

	if from, ok := c.Expression.(clause.From); ok {
		joins := from.Joins
		from.Joins = nil
		from.Build(builder)

		// set indexHints in the middle between table and joins
		squashExpression(c.AfterExpression, func(expression clause.Expression) {
			if indexHint, ok := expression.(IndexHint); ok { // pick
				builder.WriteByte(' ')
				indexHint.Build(builder)
			}
		})

		for _, join := range joins {
			builder.WriteByte(' ')
			join.Build(builder)
		}
	} else {
		c.Expression.Build(builder)
	}

	squashExpression(c.AfterExpression, func(expression clause.Expression) {
		if _, ok := expression.(IndexHint); ok {
			return
		}
		builder.WriteByte(' ')
		expression.Build(builder)
	})
}
