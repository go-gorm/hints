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
	fromClause := stmt.Clauses["FROM"]
	fromClause.Builder = indexHint.ModifyFromExpression
	stmt.Clauses["FROM"] = fromClause

	updateClause := stmt.Clauses["UPDATE"]
	if updateClause.AfterExpression == nil {
		updateClause.AfterExpression = indexHint
	} else {
		updateClause.AfterExpression = Exprs{updateClause.AfterExpression, indexHint}
	}
	stmt.Clauses["UPDATE"] = updateClause
}

func (indexHint IndexHint) ModifyFromExpression(c clause.Clause, builder clause.Builder) {
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

		// set indexHint in the middle between table and joins
		builder.WriteByte(' ')
		indexHint.Build(builder)

		for _, join := range joins {
			builder.WriteByte(' ')
			join.Build(builder)
		}
	} else {
		c.Expression.Build(builder)
	}

	if c.AfterExpression != nil {
		builder.WriteByte(' ')
		c.AfterExpression.Build(builder)
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
