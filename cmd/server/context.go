package main

import "context"

type redisKey struct{}

type authPermissions struct{}

func readKeyFromCtx(ctx context.Context) string {
	key, ok := ctx.Value(redisKey{}).(string)
	if !ok {
		return ""
	}
	return key
}

func writeKeyToCtx(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, redisKey{}, key)
}

func readPermissionsFromCtx(ctx context.Context) []string {
	permissions, ok := ctx.Value(authPermissions{}).([]string)
	if !ok {
		return []string{}
	}
	return permissions
}

func writePermissionsToCtx(ctx context.Context, permissions []string) context.Context {
	return context.WithValue(ctx, authPermissions{}, permissions)
}
