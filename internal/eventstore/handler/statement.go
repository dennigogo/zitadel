package handler

import (
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/zitadel/logging"

	"github.com/dennigogo/zitadel/internal/eventstore"
)

var (
	ErrNoProjection    = errors.New("no projection")
	ErrNoValues        = errors.New("no values")
	ErrNoCondition     = errors.New("no condition")
	ErrSomeStmtsFailed = errors.New("some statements failed")
)

type Statements []Statement

func (stmts Statements) Len() int           { return len(stmts) }
func (stmts Statements) Swap(i, j int)      { stmts[i], stmts[j] = stmts[j], stmts[i] }
func (stmts Statements) Less(i, j int) bool { return stmts[i].Sequence < stmts[j].Sequence }

type Statement struct {
	AggregateType    eventstore.AggregateType
	Sequence         uint64
	PreviousSequence uint64
	InstanceID       string

	Execute func(ex Executer, projectionName string) error
}

func (s *Statement) IsNoop() bool {
	return s.Execute == nil
}

type Executer interface {
	Exec(string, ...interface{}) (sql.Result, error)
}

type Column struct {
	Name         string
	Value        interface{}
	ParameterOpt func(string) string
}

func NewCol(name string, value interface{}) Column {
	return Column{
		Name:  name,
		Value: value,
	}
}

func NewJSONCol(name string, value interface{}) Column {
	marshalled, err := json.Marshal(value)
	if err != nil {
		logging.WithFields("column", name).WithError(err).Panic("unable to marshal column")
	}

	return NewCol(name, marshalled)
}

type Condition Column

func NewCond(name string, value interface{}) Condition {
	return Condition{
		Name:  name,
		Value: value,
	}
}
