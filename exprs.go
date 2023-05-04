package hints

import "gorm.io/gorm/clause"

type Exprs []clause.Expression

func (exprs Exprs) Build(builder clause.Builder) {
	for idx, expr := range exprs {
		if idx > 0 {
			builder.WriteByte(' ')
		}
		expr.Build(builder)
	}
}

func squashExpression(expression clause.Expression, do func(expression clause.Expression)) {
	if exprs, ok := expression.(Exprs); ok {
		for _, expr := range exprs {
			squashExpression(expr, do)
		}
	} else if expression != nil {
		do(expression)
	}
}
