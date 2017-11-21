package compare

import (
	"reflect"
	"fmt"
)

type (
	// the diff between the expected and the actual on exactly the same field
	// if Field remains empty, that means the very root
	Diff struct {
		Field    string      // values of which field are different?
		Expected interface{} // expected value
		Actual   interface{} // actual value
	}

	// keep a record of visited pointers
	visitedPtrs map[uintptr]bool
)

// compare the given data, and returns the specific diffs
// if some expected's field does not appear in actual, an error would be thrown immediately
// NOTE: only public fields could be considered
// This function is supposed to be able to
// 1. compare fields recursively;
// 2. handle circular reference.
//
// This function CANNOT
// 1. be thread-safe.
// 2. expected and actual are not exactly the same type (fields in expected is not consistent with actual)
func Compare(expected, actual interface{}) (diffs []Diff, err error) {
	exp := reflect.ValueOf(expected)
	act := reflect.ValueOf(actual)

	expVisit := visitedPtrs{}
	actVisit := visitedPtrs{}

	return compareRecursively("", exp, act, expVisit, actVisit)
}

func compareRecursively(fieldPrefix string, exp, act reflect.Value, expVisit, actVisit visitedPtrs) (diffs []Diff, err error) {
	// 1. if the type mismatched
	if exp.Type() != act.Type() {
		return nil, fmt.Errorf("mismatched type of `%s`, exp.Type=%s, act.Type=%s", fieldPrefix, exp.Type(), act.Type())
	}
	// 2. non-pointer primitive type?
	// 2.1. just deep-equals!
	if isPrimitiveType(exp.Kind()) {
		if !reflect.DeepEqual(exp.Interface(), act.Interface()) {
			diffs = newSingleDiffList(fieldPrefix, exp, act)
		}
		return // anyway, we should return, there is nothing more to take care
	}
	// 3. struct? go deep!
	if isStructType(exp.Kind()) {
		return compareStruct(fieldPrefix, exp, act, expVisit, actVisit)
	}
	// 4. array? (array is different from slice, as it cannot be nil)
	if isArrayType(exp.Kind()) {
		return compareSliceOrArray(fieldPrefix, exp, act, expVisit, actVisit)
	}
	// 5. if the non-primitive/struct field is nil?
	if exp.IsNil() && act.IsNil() {
		return
	} else if exp.IsNil() || act.IsNil() {
		return newSingleDiffList(fieldPrefix, exp, act), nil
	}
	// 5. pointer? go deep!
	if isPointerType(exp.Kind()) {
		return comparePointer(fieldPrefix, exp, act, expVisit, actVisit)
	}
	// 7. slice?
	if isSliceType(exp.Kind()) {
		return compareSliceOrArray(fieldPrefix, exp, act, expVisit, actVisit)
	}
	// 7. map?
	if isMapType(exp.Kind()) {
		return compareMap(fieldPrefix, exp, act, expVisit, actVisit)
	}
	return nil, fmt.Errorf("invalid type found, field %s, exp.Type=%#v, act.Type=%#v", fieldPrefix, exp.Type(), act.Type())
}

func newSingleDiffList(field string, exp, act reflect.Value) []Diff {
	return []Diff{{
		Field:    field,
		Expected: exp.Interface(),
		Actual:   act.Interface(),
	}}
}

func comparePointer(fieldPrefix string, exp, act reflect.Value, expVisit, actVisit visitedPtrs) (diffs []Diff, err error) {
	// 1. if we have visited the pointers, just regard them as `equalled` FIXME: DOES NOT MAKE SENSE, AT ALL!
	if expVisit[exp.Pointer()] || actVisit[exp.Pointer()] {
		return // nothing is different
	}
	// 2. go deep, please
	return compareRecursively(fmt.Sprintf("*(%s)", fieldPrefix), reflect.Indirect(exp), reflect.Indirect(act), expVisit, actVisit)
}

func compareSliceOrArray(fieldPrefix string, exp, act reflect.Value, expVisit, actVisit visitedPtrs) (diffs []Diff, err error) {
	// 1. check the len pls
	if exp.Len() != act.Len() {
		return nil, fmt.Errorf("mismatched len of `%s`, exp.Len=%d, act.Len=%d", fieldPrefix, exp.Len(), act.Len())
	}
	// 2. empty? great!
	if exp.Len() == 0 {
		return
	}
	// 3. compare element one by one
	for i, l := 0, exp.Len(); i < l; i++ {
		if diffs, err = compareElem(
			diffs,
			fmt.Sprintf("%s[%d]", fieldPrefix, i),
			exp.Index(i), act.Index(i),
			expVisit, actVisit,
		); err != nil {
			return
		}
	}
	return
}

func compareMap(fieldPrefix string, exp, act reflect.Value, expVisit, actVisit visitedPtrs) (diffs []Diff, err error) {
	// 1. length matters
	if exp.Len() != act.Len() {
		return nil, fmt.Errorf("mismatched len of `%s`, exp.Len=%d, act.Len=%d", fieldPrefix, exp.Len(), act.Len())
	}
	// 2. empty is always good news
	if exp.Len() == 0 {
		return
	}
	// 3. let us check every elem
	for _, key := range exp.MapKeys() {
		if diffs, err = compareElem(
			diffs,
			fmt.Sprintf("%s[%v]", fieldPrefix, key.String()),
			exp.MapIndex(key), act.MapIndex(key),
			expVisit, actVisit,
		); err != nil {
			return
		}
	}
	return
}

func compareStruct(fieldPrefix string, exp, act reflect.Value, expVisit, actVisit visitedPtrs) (diffs []Diff, err error) {
	// 0. prepare: getting field list
	typ := exp.Type()
	// 1. traverse all fields
	var fieldName string
	var expField, actField reflect.Value
	var subFieldPrefix string
	for i, l := 0, typ.NumField(); i < l; i++ {
		// NOTE: we use name here, because we are not sure if fields' orders are exactly matched
		fieldName = typ.Field(i).Name
		subFieldPrefix = fmt.Sprintf("%s.%s", fieldPrefix, fieldName)
		expField = exp.FieldByName(fieldName)
		actField = act.FieldByName(fieldName)
		if actField == (reflect.Value{}) {
			// zero-valued
			return nil, fmt.Errorf("field=`%s` not found in act", subFieldPrefix)
		}
		if diffs, err = compareElem(diffs, subFieldPrefix, expField, actField, expVisit, actVisit); err != nil {
			return
		}
	}
	return
}

// util function for comparing and updating elements inside a slice, array, map or structure
// it accepts an array of diff, and append elements if necessary
func compareElem(diffs []Diff, fieldPrefix string, exp, act reflect.Value, expVisit, actVisit visitedPtrs) ([]Diff, error) {
	elemDiffs, err := compareRecursively(
		fieldPrefix,
		exp, act,
		expVisit, actVisit,
	)
	if err != nil {
		return diffs, err
	} else if len(elemDiffs) > 0 {
		diffs = append(diffs, elemDiffs...)
	}
	return diffs, err
}

func isPrimitiveType(kind reflect.Kind) bool {
	// chan, invalid and unsafe-pointer are all primitive types
	return !isPointerType(kind) && !isSliceOrArrayType(kind) && !isMapType(kind) && !isStructType(kind)
}

func isPointerType(kind reflect.Kind) bool {
	return kind == reflect.Ptr
}

func isArrayType(kind reflect.Kind) bool {
	return kind == reflect.Array
}

func isSliceType(kind reflect.Kind) bool {
	return kind == reflect.Slice
}


func isSliceOrArrayType(kind reflect.Kind) bool {
	return isArrayType(kind) || isSliceType(kind)
}

func isMapType(kind reflect.Kind) bool {
	return kind == reflect.Map
}

func isStructType(kind reflect.Kind) bool {
	return kind == reflect.Struct
}
