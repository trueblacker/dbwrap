package dbwrap

type DBFunc1[T any] interface {
	Prepare(*T, *Params) error
}

type DBFunc2 interface {
	Prepare(*Params) error
}

type dbFuncsPreparer[T any] struct {
	target *T
}

func (this dbFuncsPreparer[T]) prepare(f any, prms *Params) error {
	switch stmtFunc := f.(type) {
	case DBFunc2:
		return stmtFunc.Prepare(prms)
	case DBFunc1[T]:
		return stmtFunc.Prepare(this.target, prms)
	}
	return nil
}

type preparer interface {
	prepare(f any, prms *Params) error
}

func Funcs(funcs ...any) []any {
	ret := make([]any, len(funcs))
	for i, f := range funcs {
		ret[i] = f
	}
	return ret
}
