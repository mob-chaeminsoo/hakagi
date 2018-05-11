package guess

import (
	"github.com/jinzhu/inflection"
	"github.com/syucream/hakagi/src/constraint"
	"github.com/syucream/hakagi/src/database"
)

const (
	idColumn           = "id"
	targetColumnSuffix = "_id"
)

var primaryKeyTypes = map[string]bool{
	"tinyint":   true,
	"smallint":  true,
	"meriumint": true,
	"int":       true,
	"bigint":    true,
}

type GuessOption func(database.Schema, database.PrimaryKey) bool

func isPrimaryKeyType(dataType string) bool {
	_, ok := primaryKeyTypes[dataType]
	return ok
}

// Recongnize a column thats same name of other table's primary key is a foreign key
// This base idea refers to SchemaSpy DbAnalyzer:
//   https://github.com/schemaspy/schemaspy/blob/master/src/main/java/org/schemaspy/DbAnalyzer.java
func GuessByPrimaryKey() GuessOption {
	return func(s database.Schema, pk database.PrimaryKey) bool {
		return isPrimaryKeyType(pk.DataType) && s.DataType == pk.DataType && s.Column == pk.Column && pk.Column != idColumn
	}
}

func GuessByTableAndColumn() GuessOption {
	return func(s database.Schema, pk database.PrimaryKey) bool {
		if !isPrimaryKeyType(pk.DataType) && s.DataType != pk.DataType {
			return false
		}

		cLen := len(s.Column)
		tLen := len(targetColumnSuffix)
		if !(cLen >= tLen && s.Column[cLen-tLen:] == targetColumnSuffix) {
			return false
		}

		return inflection.Plural(s.Column[:cLen-tLen]) == pk.Table && pk.Column == idColumn
	}
}

func GuessConstraints(schemas []database.Schema, primaryKeys []database.PrimaryKey, guessOptions ...GuessOption) []constraint.Constraint {
	var constraints []constraint.Constraint

	for _, s := range schemas {
		for _, pk := range primaryKeys {
			if s.Table != pk.Table {
				for _, guesser := range guessOptions {
					if guesser(s, pk) {
						constraints = append(constraints, constraint.Constraint{s.Table, s.Column, pk.Table, pk.Column})
					}
				}
			}
		}
	}

	return constraints
}
