// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import "reflect"

// optionApplySnapshot captures an option's mutable apply-time state
// so it can be restored if a later error in the same config-file apply pass
// aborts the whole load.
type optionApplySnapshot struct {
	value                   reflect.Value
	isSet                   bool
	isSetDefault            bool
	preventDefault          bool
	clearReferenceBeforeSet bool
}

// configApplyTransaction records the pre-apply state of each option
// touched while applying a single config file (INI/JSON),
// so the whole apply pass can be rolled back to leave the bound struct exactly
// as it was if a later key fails.
//
// Options whose values are set through arbitrary user-supplied functions (isFunc)
// already ran that function by the time they are touched,
// so their own side effects cannot be undone;
// only their bookkeeping and the function reference itself are restored.
type configApplyTransaction struct {
	snapshots map[*Option]optionApplySnapshot
}

func newConfigApplyTransaction() *configApplyTransaction {
	return &configApplyTransaction{snapshots: make(map[*Option]optionApplySnapshot)}
}

// touch records opt's current state the first time it is seen in this transaction.
// Call it immediately before every Set/setDefault call so the state recorded
// is always the pre-apply state, regardless of how many times the same option is touched
// (e.g. once per element for a slice/map).
func (tx *configApplyTransaction) touch(opt *Option) {
	if _, ok := tx.snapshots[opt]; ok {
		return
	}

	valCopy := reflect.New(opt.value.Type()).Elem()
	valCopy.Set(opt.value)

	tx.snapshots[opt] = optionApplySnapshot{
		value:                   valCopy,
		isSet:                   opt.isSet,
		isSetDefault:            opt.isSetDefault,
		preventDefault:          opt.preventDefault,
		clearReferenceBeforeSet: opt.clearReferenceBeforeSet,
	}
}

// rollback restores every touched option to its pre-transaction state.
func (tx *configApplyTransaction) rollback() {
	for opt, snap := range tx.snapshots {
		opt.value.Set(snap.value)
		opt.isSet = snap.isSet
		opt.isSetDefault = snap.isSetDefault
		opt.preventDefault = snap.preventDefault
		opt.clearReferenceBeforeSet = snap.clearReferenceBeforeSet
	}
}
