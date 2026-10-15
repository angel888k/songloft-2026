package database

import (
	"database/sql/driver"
	"strings"

	"github.com/mozillazg/go-pinyin"
	"modernc.org/sqlite"
)

func init() {
	// Register before Open so every pooled connection, including transaction
	// connections, can sort before LIMIT/OFFSET. No schema or stored titles change.
	sqlite.MustRegisterDeterministicScalarFunction("songloft_text_sort_key", 1,
		func(_ *sqlite.FunctionContext, args []driver.Value) (driver.Value, error) {
			if args[0] == nil {
				return nil, nil
			}
			return textSortKey(args[0].(string)), nil
		})
}

// textSortKey interleaves Chinese pinyin with Latin titles, ignoring case and
// surrounding whitespace. Use the dictionary's default reading for polyphones.
// Keep digits, punctuation and unsupported scripts rather than dropping them.
func textSortKey(text string) string {
	args := pinyin.NewArgs()
	args.Fallback = func(r rune, _ pinyin.Args) []string { return []string{string(r)} }
	return strings.Join(pinyin.LazyPinyin(strings.ToLower(strings.TrimSpace(text)), args), "")
}
