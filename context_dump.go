package synx

import (
	"context"
	"fmt"
	"unsafe"
)

// DumpContext values. Works for stdlib contexts only.
func DumpContext(ctx context.Context) map[any]any {
	if ctx == nil {
		return nil
	}

	values := map[any]any{}
	for {
		// cannot use type-switch here because those types are unexported
		switch fmt.Sprintf("%T", ctx) {
		case "*context.valueCtx":
			v := *(*valueCtx)((*iface)(unsafe.Pointer(&ctx)).data)
			values[v.key] = v.value
			ctx = v.Context

		case "*context.timerCtx", "*context.cancelCtx":
			v := *(*parentCtx)((*iface)(unsafe.Pointer(&ctx)).data)
			ctx = v.Context

		default:
			// know nothing about other types (or nil), so returning
			return values
		}
	}
}

// Same as runtime.(iface)
type iface struct {
	_    unsafe.Pointer
	data unsafe.Pointer
}

// Same as of context.(valueCtx)
type valueCtx struct {
	context.Context
	key, value any
}

// Same as context.(*timerCtx) and context.(*cancelCtx)
type parentCtx struct {
	context.Context
	_ struct{}
}
