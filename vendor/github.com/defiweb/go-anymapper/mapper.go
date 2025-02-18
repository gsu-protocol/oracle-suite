package anymapper

import (
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// MapFunc is a function that maps a src value to a dst value. It returns an
// error if the mapping is not possible. The src and dst values are never
// pointers.
type MapFunc func(m *Mapper, ctx *Context, src, dst reflect.Value) error

// MapFuncProvider is a function that returns a MapFunc for given src and dst
// types. If mapping is not supported, it returns nil.
type MapFuncProvider func(m *Mapper, src, dst reflect.Type) MapFunc

// Default is the default Mapper used by the Map and MapRefl functions.
// It also provides additional mapping rules for time.Time, big.Int, big.Float
// and big.Rat. It can be modified to change the default behavior, but if the
// mapper is used by other packages, it is recommended to create a copy of the
// default mapper and modify the copy.
var Default = New()

// Context is a context that is passed to the mapping functions. It can be
// used to pass additional information to the mapping functions or to change
// the behavior of the mapper without modifying the global state or creating
// a copy of the mapper.
type Context struct {
	// StrictTypes enables strict type checking. If enabled, the source and
	// destination types must be exactly the same for the mapping to be
	// successful. However, mapping between different data structures, such as
	// `struct` ⇔ `struct`, `struct` ⇔ `map` and `map` ⇔ `map` is always
	// allowed. If the destination type is an empty interface, the source value
	// will be assigned to it regardless of the strict type check setting.
	StrictTypes bool

	// Tag is the name of the struct tag that is used by the mapper to
	// determine the name of the field to map to.
	Tag string

	// ByteOrder is the byte order used to map numbers to and from byte slices.
	ByteOrder binary.ByteOrder

	// DisableCache disables the cache of the type mappers.
	DisableCache bool

	// FieldMapper is a function that maps a struct field name to another name,
	// it is used only when the tag is not present.
	FieldMapper func(string) string

	// Custom is a custom value that can be used to pass additional information
	// to the mapping functions.
	Custom any
}

// WithStrictTypes returns a copy of the context with the StrictTypes field
// set to the given value.
func (c *Context) WithStrictTypes(strictTypes bool) *Context {
	cpy := *c
	cpy.StrictTypes = strictTypes
	return &cpy
}

// WithTag returns a copy of the context with the Tag field set to the given
// value.
func (c *Context) WithTag(tag string) *Context {
	cpy := *c
	cpy.Tag = tag
	return &cpy
}

// WithByteOrder returns a copy of the context with the ByteOrder field set
// to the given value.
func (c *Context) WithByteOrder(byteOrder binary.ByteOrder) *Context {
	cpy := *c
	cpy.ByteOrder = byteOrder
	return &cpy
}

// WithDisableCache returns a copy of the context with the DisableCache field
// set to the given value.
func (c *Context) WithDisableCache(disableCache bool) *Context {
	cpy := *c
	cpy.DisableCache = disableCache
	return &cpy
}

// WithFieldMapper returns a copy of the context with the FieldMapper field
// set to the given value.
func (c *Context) WithFieldMapper(fieldMapper func(string) string) *Context {
	cpy := *c
	cpy.FieldMapper = fieldMapper
	return &cpy
}

// WithCustom returns a copy of the context with the Custom field set to the
// given value.
func (c *Context) WithCustom(custom any) *Context {
	cpy := *c
	cpy.Custom = custom
	return &cpy
}

// Mapper hold the mapper configuration.
type Mapper struct {
	// Context is the default context used by the mapper.
	Context *Context

	// Mappers is a map of custom mapper providers. The key is the type that
	// the mapper can map to and from. The value is a function that returns
	// a MapFunc that maps the source type to the destination type. Provider
	// can return nil if the mapping is not possible.
	//
	// If both source and destination types have defined providers, then
	// the provider for source value is used first, and if it returns nil,
	// then the provider for destination value is used.
	Mappers map[reflect.Type]MapFuncProvider

	// Hooks are functions that are called during the mapping process. They
	// can modify the behavior of the mapper. See Hooks for more information.
	Hooks Hooks

	// Cache:
	cacheMu  sync.Mutex
	cacheMap map[typePair]*typeMapper
}

// Hooks are functions that are called during the mapping process. They can
// modify the behavior of the mapper.
type Hooks struct {
	// MapFuncHook allows to bypass the default mapping rules and use a custom
	// mapping function. If the hook returns nil, then the default mapping
	// rules are used.
	//
	// Returned MapFunc is cached.
	MapFuncHook MapFuncProvider

	// SourceValueHook returns a value that should be used as the source
	// value. It is called before the source value is used in the mapping.
	//
	// If the hook returns an invalid value, then the default function is used.
	//
	// By default, mapper unpacks pointers and dereferences interfaces. This
	// hook can be used to change this behavior.
	SourceValueHook func(reflect.Value) reflect.Value

	// DestinationValueHook returns a value that should be used as the destination
	// value. It is called before the destination value is used in the mapping.
	//
	// If the hook returns an invalid value, then the default function is used.
	//
	// By default, mapper unpacks pointers and dereferences interfaces. This
	// hook can be used to change this behavior.
	DestinationValueHook func(reflect.Value) reflect.Value
}

// New returns a new Mapper with default configuration.
func New() *Mapper {
	return &Mapper{
		Context: &Context{
			Tag:       `map`,
			ByteOrder: binary.BigEndian,
		},
		Mappers: map[reflect.Type]MapFuncProvider{
			timeTy:     timeTypeMapper,
			bigIntTy:   bigIntTypeMapper,
			bigFloatTy: bigFloatTypeMapper,
			bigRatTy:   bigRatTypeMapper,
		},
		cacheMap: make(map[typePair]*typeMapper, 0),
	}
}

// Map maps the source value to the destination value.
//
// It is shorthand for Default.mapRefl(src, dst).
func Map(src, dst any) error {
	return Default.Map(src, dst)
}

// MapContext maps the source value to the destination value.
//
// It is shorthand for Default.MapContext(ctx, src, dst).
func MapContext(ctx *Context, src, dst any) error {
	return Default.MapContext(ctx, src, dst)
}

// MapRefl maps the source value to the destination value.
//
// It is shorthand for Default.MapRefl(src, dst).
func MapRefl(src, dst reflect.Value) error {
	return Default.MapRefl(src, dst)
}

// MapReflContext maps the source value to the destination value.
//
// It is shorthand for Default.MapReflContext(ctx, src, dst).
func MapReflContext(ctx *Context, src, dst reflect.Value) error {
	return Default.MapReflContext(ctx, src, dst)
}

// Map maps the source value to the destination value.
func (m *Mapper) Map(src, dst any) error {
	return m.MapRefl(reflect.ValueOf(src), reflect.ValueOf(dst))
}

// MapContext maps the source value to the destination value.
func (m *Mapper) MapContext(ctx *Context, src, dst any) error {
	return m.MapReflContext(ctx, reflect.ValueOf(src), reflect.ValueOf(dst))
}

// MapRefl maps the source value to the destination value.
func (m *Mapper) MapRefl(src, dst reflect.Value) error {
	return m.MapReflContext(m.Context, src, dst)
}

// MapReflContext maps the source value to the destination value.
func (m *Mapper) MapReflContext(ctx *Context, src, dst reflect.Value) error {
	if ctx == nil {
		ctx = m.Context
	}
	srcVal := m.srcValue(src)
	dstVal := m.dstValue(dst)
	if !srcVal.IsValid() {
		return InvalidSrcErr
	}
	if !dstVal.IsValid() {
		return InvalidDstErr
	}
	return m.mapperFor(ctx, srcVal.Type(), dstVal.Type()).mapRefl(m, ctx, srcVal, dstVal)
}

// Copy creates a copy of the current Mapper with the same configuration.
func (m *Mapper) Copy() *Mapper {
	cpy := &Mapper{
		Context: &Context{
			StrictTypes:  m.Context.StrictTypes,
			Tag:          m.Context.Tag,
			ByteOrder:    m.Context.ByteOrder,
			DisableCache: m.Context.DisableCache,
			FieldMapper:  m.Context.FieldMapper,
			Custom:       m.Context.Custom,
		},
		Hooks:    m.Hooks,
		cacheMap: make(map[typePair]*typeMapper, 0),
	}
	if m.Mappers != nil {
		cpy.Mappers = make(map[reflect.Type]MapFuncProvider)
		for k, v := range m.Mappers {
			cpy.Mappers[k] = v
		}
	}
	return cpy
}

// mapperFor returns the typeMapper that can map values of the given types.
// If mapping is not possible, the returned typeMapper has a nil MapFunc.
func (m *Mapper) mapperFor(ctx *Context, src, dst reflect.Type) (tm *typeMapper) {
	if !ctx.DisableCache {
		m.cacheMu.Lock()
		if v, ok := m.cacheMap[typePair{src: src, dst: dst}]; ok {
			m.cacheMu.Unlock()
			return v
		}
		defer func() {
			m.cacheMap[typePair{src: src, dst: dst}] = tm
			m.cacheMu.Unlock()
		}()
	}

	tm = &typeMapper{
		SrcType: src,
		DstType: dst,
	}

	// If MapFuncHook is set, then use it to get the mapping function.
	if m.Hooks.MapFuncHook != nil {
		if fn := m.Hooks.MapFuncHook(m, src, dst); fn != nil {
			tm.MapFunc = fn
			return
		}
	}

	var isSrcSimple, isDstSimple, sameTypes bool
	if src == dst {
		isSrcSimple = isSimpleType(src)
		isDstSimple = isSrcSimple
		sameTypes = true
	} else {
		isSrcSimple = isSimpleType(src)
		isDstSimple = isSimpleType(dst)
	}

	// If both types are simple, e.g. int, string, etc. map the value directly
	// using reflect.Set.
	if sameTypes && isSrcSimple {
		tm.MapFunc = mapDirect
		return
	}

	// Try to find a mapper using mapper providers. It looks for providers
	// for src and dst types. First it tries to use providers for src. If
	// it returns a mapper, it uses it. If it returns nil, it tries to use
	// providers for dst. If both return nil, then mapping is not possible.
	var srcMapper, dstMapper MapFuncProvider
	var hasSrcMapper, hasDstMapper bool
	if !isSrcSimple {
		srcMapper, hasSrcMapper = m.Mappers[src]
	}
	if hasSrcMapper {
		tm.MapFunc = srcMapper(m, src, dst)
		if tm.MapFunc != nil {
			return
		}
	}
	if !sameTypes && !isDstSimple {
		dstMapper, hasDstMapper = m.Mappers[dst]
	}
	if hasDstMapper {
		tm.MapFunc = dstMapper(m, src, dst)
		if tm.MapFunc != nil {
			return
		}
	}
	if hasSrcMapper || hasDstMapper {
		return
	}

	// If destination type is an any interface, map the value directly using
	// reflect.Set, if the destination interface is not nil, map the value
	// to the same type as the value in the interface.
	if dst == anyTy {
		tm.MapFunc = mapAny
		return
	}

	// If there are no custom mappers and hooks, use the default mappers.
	tm.MapFunc = builtInTypesMapper(m, src, dst)
	return
}

// srcValue unpacks values from pointers and interfaces until it reaches a
// non-pointer or non-interface value, or a type that has a custom mapper.
func (m *Mapper) srcValue(v reflect.Value) reflect.Value {
	if !v.IsValid() {
		return v
	}
	if m.Hooks.SourceValueHook != nil {
		if v := m.Hooks.SourceValueHook(v); v.IsValid() {
			return v
		}
	}
	for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if isSimpleType(v.Type()) {
			return v
		}
		v = v.Elem()
	}
	return v
}

// dstValue unpacks values from pointers and interfaces until it reaches a
// settable non-pointer or non-interface value, value that has a custom mapper,
// or a value that is a map, slice or array. It returns an invalid value if it
// cannot find a value that meets these conditions. If the value is a pointer,
// map or slice, it will be initialized if needed.
func (m *Mapper) dstValue(v reflect.Value) reflect.Value {
	if !v.IsValid() {
		return v
	}
	if m.Hooks.DestinationValueHook != nil {
		if v := m.Hooks.DestinationValueHook(v); v.IsValid() {
			return v
		}
	}
	if v.Kind() != reflect.Interface && v.Kind() != reflect.Pointer && v.CanSet() {
		return v
	}
	settable := reflect.Value{}
	for {
		if !v.IsValid() {
			break
		}
		m.initValue(v)
		if v.CanSet() && isSimpleType(v.Type()) {
			return v
		}
		if m.Mappers[v.Type()] != nil {
			return v
		}
		if v.Kind() == reflect.Map && !v.IsNil() {
			return v
		}
		if v.CanSet() {
			settable = v
		}
		if v.Kind() != reflect.Interface && v.Kind() != reflect.Pointer {
			break
		}
		v = v.Elem()
	}
	return settable
}

// initValue initializes a value if it is a pointer, map or slice.
func (m *Mapper) initValue(v reflect.Value) {
	if v.Kind() < reflect.Map || v.Kind() > reflect.Slice || !v.IsNil() || !v.CanSet() {
		return
	}
	switch {
	case v.Kind() == reflect.Pointer:
		v.Set(reflect.New(v.Type().Elem()))
	case v.Kind() == reflect.Map:
		v.Set(reflect.MakeMap(v.Type()))
	case v.Kind() == reflect.Slice:
		v.Set(reflect.MakeSlice(v.Type(), 0, 0))
	}
}

// parseTag parses the tag of the given field and returns the tag name and
// whether the field should be skipped.
func (m *Mapper) parseTag(ctx *Context, f reflect.StructField) (fields string, skip bool) {
	tag, ok := f.Tag.Lookup(ctx.Tag)
	if !ok {
		if ctx.FieldMapper != nil {
			return ctx.FieldMapper(f.Name), false
		} else {
			return f.Name, false
		}
	}
	if tag == "-" {
		return "", true
	}
	return tag, false
}

// isSimpleType indicates whether a type is simple type.
//
// A type is considered simple if it is a built-in type, or it is a slice,
// array or map that is composed of build-in types.
//
// Structs are never considered simple because they are rarely used without a
// custom type, and verifying if a struct is simple is too expensive.
func isSimpleType(p reflect.Type) bool {
	switch p.Kind() {
	case reflect.Bool:
		return p == boolTy
	case reflect.Int:
		return p == intTy
	case reflect.Int8:
		return p == int8Ty
	case reflect.Int16:
		return p == int16Ty
	case reflect.Int32:
		return p == int32Ty
	case reflect.Int64:
		return p == int64Ty
	case reflect.Uint:
		return p == uintTy
	case reflect.Uint8:
		return p == uint8Ty
	case reflect.Uint16:
		return p == uint16Ty
	case reflect.Uint32:
		return p == uint32Ty
	case reflect.Uint64:
		return p == uint64Ty
	case reflect.Float32:
		return p == float32Ty
	case reflect.Float64:
		return p == float64Ty
	case reflect.String:
		return p == stringTy
	case reflect.Slice:
		return strings.HasPrefix(p.String(), "[") && isSimpleType(p.Elem())
	case reflect.Array:
		return strings.HasPrefix(p.String(), "[") && isSimpleType(p.Elem())
	case reflect.Map:
		return strings.HasPrefix(p.String(), "map[") && isSimpleType(p.Elem()) && isSimpleType(p.Key())
	}
	return false
}

// mapAny map src to dst assuming dst is an empty interface.
func mapAny(m *Mapper, _ *Context, src, dst reflect.Value) error {
	if !dst.IsNil() && !dst.Elem().CanSet() {
		// Mapper always tries to reuse the destination value if possible, but
		// if destination value is not settable, we need to cheat a little and
		// create a new value of the same type and then set it back to the
		// destination.
		auxVal := reflect.New(dst.Elem().Type())
		auxDst := m.dstValue(auxVal)
		if err := m.MapRefl(src, auxDst); err != nil {
			return NewInvalidMappingError(src.Type(), dst.Type(), "")
		}
		dst.Set(auxVal.Elem())
		return nil
	}
	dst.Set(src)
	return nil
}

// mapDirect maps src to dst using a direct assignment.
func mapDirect(_ *Mapper, _ *Context, src, dst reflect.Value) error {
	dst.Set(src)
	return nil
}

type typeMapper struct {
	SrcType reflect.Type
	DstType reflect.Type
	MapFunc MapFunc
}

func (tm *typeMapper) match(src, dst reflect.Type) bool {
	if tm == nil {
		return false
	}
	return tm.SrcType == src && tm.DstType == dst
}

func (tm *typeMapper) mapRefl(m *Mapper, ctx *Context, src, dst reflect.Value) error {
	if tm == nil {
		return NewInvalidMappingError(src.Type(), dst.Type(), "unknown mapper")
	}
	if tm.MapFunc == nil {
		return NewInvalidMappingError(src.Type(), dst.Type(), "")
	}
	return tm.MapFunc(m, ctx, src, dst)
}

// InvalidSrcErr is returned when reflect.IsValid returns false for the source
// value.
var InvalidSrcErr = errors.New("mapper: invalid source value")

// InvalidDstErr is returned when reflect.IsValid returns false for the
// destination value. It may happen when the destination value was not
// passed as a pointer.
var InvalidDstErr = errors.New("mapper: invalid destination value")

type InvalidMappingErr struct {
	From, To reflect.Type
	Reason   string
}

func NewStrictMappingError(from, to reflect.Type) *InvalidMappingErr {
	return &InvalidMappingErr{From: from, To: to, Reason: "strict mode"}
}

func NewInvalidMappingError(from, to reflect.Type, reason string) *InvalidMappingErr {
	return &InvalidMappingErr{From: from, To: to, Reason: reason}
}

func (e *InvalidMappingErr) Error() string {
	if len(e.Reason) == 0 {
		return fmt.Sprintf("mapper: cannot map %v to %v", e.From, e.To)
	}
	return fmt.Sprintf("mapper: cannot map %v to %v: %s", e.From, e.To, e.Reason)
}

type typePair struct {
	src reflect.Type
	dst reflect.Type
}

var (
	anyTy     = reflect.TypeOf((*any)(nil)).Elem()
	boolTy    = reflect.TypeOf((*bool)(nil)).Elem()
	intTy     = reflect.TypeOf((*int)(nil)).Elem()
	int8Ty    = reflect.TypeOf((*int8)(nil)).Elem()
	int16Ty   = reflect.TypeOf((*int16)(nil)).Elem()
	int32Ty   = reflect.TypeOf((*int32)(nil)).Elem()
	int64Ty   = reflect.TypeOf((*int64)(nil)).Elem()
	uintTy    = reflect.TypeOf((*uint)(nil)).Elem()
	uint8Ty   = reflect.TypeOf((*uint8)(nil)).Elem()
	uint16Ty  = reflect.TypeOf((*uint16)(nil)).Elem()
	uint32Ty  = reflect.TypeOf((*uint32)(nil)).Elem()
	uint64Ty  = reflect.TypeOf((*uint64)(nil)).Elem()
	float32Ty = reflect.TypeOf((*float32)(nil)).Elem()
	float64Ty = reflect.TypeOf((*float64)(nil)).Elem()
	stringTy  = reflect.TypeOf((*string)(nil)).Elem()
)
