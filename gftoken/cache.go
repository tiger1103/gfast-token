package gftoken

import (
	"context"
	"time"
)

func (m *GfToken) contains(ctx context.Context, key string) bool {
	ok, _ := m.cache.Contains(ctx, key)
	return ok
}

func (m *GfToken) setCache(ctx context.Context, key string, value interface{}) error {
	return m.cache.Set(ctx, key, value, time.Duration(m.Timeout+m.MaxRefresh)*time.Second)
}

func (m *GfToken) getCache(ctx context.Context, key string) (string, error) {
	result, err := m.cache.Get(ctx, key)
	if err != nil {
		return "", err
	}
	if result != nil {
		return result.String(), nil
	} else {
		return "", nil
	}
}

func (m *GfToken) removeCache(ctx context.Context, key string) (err error) {
	_, err = m.cache.Remove(ctx, key)
	return
}

func (m *GfToken) getUserKeyCache(ctx context.Context, key string) ([]string, error) {
	result, err := m.cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	return result.Strings(), nil
}

func (m *GfToken) addUserKeyCache(ctx context.Context, key string, value string) (err error) {
	result, err := m.cache.Get(ctx, key)
	if err != nil {
		return
	}
	list := result.Strings()
	list = append(list, value)
	err = m.cache.Set(ctx, key, list, time.Duration(m.Timeout+m.MaxRefresh)*time.Second)
	return
}

func (m *GfToken) UpdateExpireUserKeyCache(ctx context.Context, key string) error {
	_, err := m.cache.UpdateExpire(ctx, key, time.Duration(m.Timeout+m.MaxRefresh)*time.Second)

	return err
}
