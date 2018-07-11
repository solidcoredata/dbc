package parser

import (
	"testing"
)

func TestConst(t *testing.T) {
	var f = "%[1]s: %[1]d"
	t.Logf(f, AllowNone)
	t.Logf(f, AllowRead)
	t.Logf(f, AllowReturn)
	t.Logf(f, AllowInsert)
	t.Logf(f, AllowUpdate)
	t.Logf(f, AllowDelete)
	t.Logf(f, AllowFull)
}
