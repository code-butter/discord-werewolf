package testlib

type NoopLogger struct{}

func (NoopLogger) Print(...interface{})          {}
func (NoopLogger) Printf(string, ...interface{}) {}
func (NoopLogger) Println(...interface{})        {}
func (NoopLogger) Fatalf(string, ...interface{}) {}
