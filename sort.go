package main

import (
	"fmt"
	"sort"
	"time"
)

type sorter struct {
	records []Record
	key     string
	desc    bool
}

func (sorter *sorter) Len() int {
	return len(sorter.records)
}

func (sorter *sorter) Swap(a, b int) {
	sorter.records[a], sorter.records[b] = sorter.records[b], sorter.records[a]
}

func (sorter *sorter) Less(a, b int) bool {
	valueA, _ := sorter.records[a][sorter.key]
	valueB, _ := sorter.records[b][sorter.key]
	return compare(valueA, valueB, sorter.desc)
}

func compare(a, b interface{}, desc bool) bool {
	if a == nil && b == nil {
		return false
	}

	if a == nil && b != nil {
		return desc
	}

	if a != nil && b == nil {
		return !desc
	}

	switch typedA := a.(type) {
	case bool:
		typedB := b.(bool)
		if typedA == typedB {
			return false
		}

		if !typedA && typedB {
			return desc
		}

		return !desc

	case int64:
		typedB := b.(int64)

		if typedA == typedB {
			return false
		}

		if typedA < typedB {
			return desc
		}

		return !desc

	case int:
		typedB := b.(int)

		if typedA == typedB {
			return false
		}

		if typedA < typedB {
			return desc
		}

		return !desc

	case string:
		typedB := b.(string)

		if typedA == typedB {
			return false
		}

		if sort.StringsAreSorted([]string{typedA, typedB}) {
			return desc
		}

		return desc

	case time.Time:
		return compare(
			typedA.UnixNano(),
			b.(time.Time).UnixNano(),
			desc,
		)

	case time.Duration:
		return compare(
			typedA.Nanoseconds(),
			b.(time.Duration).Nanoseconds(),
			desc,
		)
	}

	panic(fmt.Sprintf("unexpected comparison: %#v vs %#v", a, b))
}
