// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"fmt"
	"reflect"
	"slices"
	"sort"
	"time"
)

func defaultOptionTypeRank() map[OptionTypeClass]int {
	return map[OptionTypeClass]int{
		OptionTypeBool:       0,
		OptionTypeNumber:     1,
		OptionTypeString:     2,
		OptionTypeDuration:   3,
		OptionTypeCollection: 4,
		OptionTypeCustom:     5,
	}
}

func buildOptionTypeRank(order []OptionTypeClass) (map[OptionTypeClass]int, error) {
	all := []OptionTypeClass{
		OptionTypeBool,
		OptionTypeNumber,
		OptionTypeString,
		OptionTypeDuration,
		OptionTypeCollection,
		OptionTypeCustom,
	}

	if len(order) == 0 {
		return defaultOptionTypeRank(), nil
	}

	rank := make(map[OptionTypeClass]int, len(all))
	for i, cls := range order {
		if !slices.Contains(all, cls) {
			return nil, fmt.Errorf("unknown option type class %d", cls)
		}
		if _, exists := rank[cls]; exists {
			return nil, fmt.Errorf("duplicate option type class %d", cls)
		}
		rank[cls] = i
	}

	next := len(rank)
	def := defaultOptionTypeRank()
	sort.Slice(all, func(i, j int) bool { return def[all[i]] < def[all[j]] })
	for _, cls := range all {
		if _, exists := rank[cls]; !exists {
			rank[cls] = next
			next++
		}
	}

	return rank, nil
}

func (p *Parser) sortedOptions(options []*Option) []*Option {
	if !p.shouldSortOptionsForDisplay(options) {
		return options
	}

	sort.SliceStable(options, func(i, j int) bool {
		return p.compareOptions(options[i], options[j]) < 0
	})

	return options
}

func (p *Parser) shouldSortOptionsForDisplay(options []*Option) bool {
	if p.optionSort != OptionSortByDeclaration {
		return true
	}

	for _, opt := range options {
		if opt.Order != 0 {
			return true
		}
	}

	return false
}

func (p *Parser) compareOptions(a *Option, b *Option) int {
	ab := orderBucket(a.Order)
	bb := orderBucket(b.Order)
	if ab != bb {
		return compareInt(ab, bb)
	}

	if a.Order != b.Order {
		// Higher order first within same bucket.
		return compareInt(b.Order, a.Order)
	}

	switch p.optionSort {
	case OptionSortByNameAsc:
		return compareString(optionSortName(a), optionSortName(b))
	case OptionSortByNameDesc:
		return compareString(optionSortName(b), optionSortName(a))
	case OptionSortByType:
		at := p.optionTypeRank[optionTypeClass(a)]
		bt := p.optionTypeRank[optionTypeClass(b)]
		if at != bt {
			return compareInt(at, bt)
		}
		return compareString(optionSortName(a), optionSortName(b))
	default:
		return 0
	}
}

func optionSortName(opt *Option) string {
	if opt.LongName != "" {
		return opt.LongNameWithNamespace()
	}
	if opt.ShortName != 0 {
		return string(opt.ShortName)
	}
	return opt.field.Name
}

func optionTypeClass(opt *Option) OptionTypeClass {
	tp := opt.value.Type()

	for tp.Kind() == reflect.Pointer {
		tp = tp.Elem()
	}

	switch tp.Kind() {
	case reflect.Bool:
		return OptionTypeBool
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		if tp == reflect.TypeFor[time.Duration]() {
			return OptionTypeDuration
		}
		return OptionTypeNumber
	case reflect.String:
		return OptionTypeString
	case reflect.Slice, reflect.Array, reflect.Map:
		return OptionTypeCollection
	default:
		return OptionTypeCustom
	}
}

func orderBucket(v int) int {
	switch {
	case v > 0:
		return 0
	case v < 0:
		return 2
	default:
		return 1
	}
}

func compareInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func compareString(a, b string) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}
