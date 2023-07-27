package cache

import "context"

type Cache interface {
	Get(ctx context.Context, key string) (*string, error)
	Put(ctx context.Context, key string, value any) error
	Delete(ctx context.Context, key string) error
}

var impl Cache

func SetRepository(repository Cache) {
	impl = repository
}

func Get(ctx context.Context, key string) (*string, error) {
	return impl.Get(ctx, key)
}

func Put(ctx context.Context, key string, value any) error {
	return impl.Put(ctx, key, value)
}

func Delete(ctx context.Context, key string) error {
	return impl.Delete(ctx, key)
}
