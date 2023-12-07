package dbwrap

import (
	"database/sql"
	"encoding/base64"
	"fmt"
)

func Bin2Str(src []byte) string {
	return base64.RawStdEncoding.EncodeToString(src)
}

func Str2Bin(src string) ([]byte, error) {
	return base64.RawStdEncoding.DecodeString(src)
}

type queryable interface {
	QueryRow(args ...any) *sql.Row
}

type query struct {
	args  []any
	dests []any
	strs  []*sql.NullString
}

func Query(args ...any) *query {
	return &query{
		args: args,
	}
}

func (this *query) Dests(dests ...any) *query {
	this.dests = dests
	return this
}

type stringable interface {
	FromDBString(string) error
}

func (this *query) getDests() []any {
	ret := make([]any, len(this.dests))
	this.strs = make([]*sql.NullString, len(this.dests))
	for i, d := range this.dests {
		switch d.(type) {
		case stringable, *string, *[]byte:
			this.strs[i] = new(sql.NullString)
			ret[i] = this.strs[i]
		default:
			ret[i] = d
		}
	}
	return ret
}

func (this query) parse() error {
	for i, arg := range this.strs {
		if arg == nil || !arg.Valid {
			continue
		}
		var err error
		switch v := this.dests[i].(type) {
		case *string:
			*v = arg.String
		case *[]byte:
			*v, err = Str2Bin(arg.String)
		case stringable:
			err = v.FromDBString(arg.String)
		default:
			err = fmt.Errorf("unsupported type")
		}
		if err != nil {
			return err
		}
	}
	return nil
}

type scannable interface {
	Scan(...any) error
}

func (this *query) Scan(s scannable) error {
	err := s.Scan(this.getDests()...)
	if err != nil {
		return err
	}
	return this.parse()
}

func (this *query) QueryRow(q queryable) error {
	err := q.QueryRow(this.args...).Scan(this.getDests()...)
	if err != nil {
		return err
	}
	return this.parse()
}
