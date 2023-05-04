package preferences

import (
	"fmt"
	"testing"
)

func TestSpliceConfig(t *testing.T) {
	d := GetStringSplice("db.list")
	fmt.Println(d)
}
