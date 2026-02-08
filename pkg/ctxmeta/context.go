package ctxmeta

import (
	"context"
	"strings"
)

func normalize(v string) string {
	return strings.TrimSpace(v)
}

func with(ctx context.Context, key string, value string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	n := normalize(value)
	if n == "" {
		return ctx
	}
	return context.WithValue(ctx, key, n)
}

func get(ctx context.Context, key string) string {
	if ctx == nil {
		return ""
	}
	value, ok := ctx.Value(key).(string)
	if !ok {
		return ""
	}
	return normalize(value)
}

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return with(ctx, KeyTraceID, traceID)
}

func WithUserUUID(ctx context.Context, userUUID string) context.Context {
	return with(ctx, KeyUserUUID, userUUID)
}

func WithDeviceID(ctx context.Context, deviceID string) context.Context {
	return with(ctx, KeyDeviceID, deviceID)
}

func WithClientIP(ctx context.Context, clientIP string) context.Context {
	return with(ctx, KeyClientIP, clientIP)
}

func TraceID(ctx context.Context) string {
	return get(ctx, KeyTraceID)
}

func UserUUID(ctx context.Context) string {
	return get(ctx, KeyUserUUID)
}

func DeviceID(ctx context.Context) string {
	return get(ctx, KeyDeviceID)
}

func ClientIP(ctx context.Context) string {
	return get(ctx, KeyClientIP)
}

// CopyKnownFromParent copies canonical context metadata into a new background context.
func CopyKnownFromParent(parent context.Context) context.Context {
	ctx := context.Background()
	if parent == nil {
		return ctx
	}
	if traceID := TraceID(parent); traceID != "" {
		ctx = WithTraceID(ctx, traceID)
	}
	if userUUID := UserUUID(parent); userUUID != "" {
		ctx = WithUserUUID(ctx, userUUID)
	}
	if deviceID := DeviceID(parent); deviceID != "" {
		ctx = WithDeviceID(ctx, deviceID)
	}
	if clientIP := ClientIP(parent); clientIP != "" {
		ctx = WithClientIP(ctx, clientIP)
	}
	return ctx
}
