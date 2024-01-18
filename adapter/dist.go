/*
* @desc:磁盘缓存
* @company:云南奇讯科技有限公司
* @Author: yixiaohu<yxh669@qq.com>
* @Date:   2024/1/16 9:07
 */

package adapter

import (
	"context"
	"fmt"
	badger "github.com/dgraph-io/badger/v4"
	"github.com/gogf/gf/v2/container/gmap"
	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcache"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/tiger1103/gfast-token/instance"
	"reflect"
	"sync"
	"time"
)

const (
	DefaultGroupName = "default" // Default configuration group name.
	DistCacheName    = "distCache"
	// defaultMaxExpire is the default expire time for no expiring items.
	// It equals to math.MaxInt64/1000000.
	defaultMaxExpire time.Duration = 9223372036854
)

var (
	// Configuration groups.
	localConfigMap = gmap.NewStrAnyMap(true)
	ctx            = context.Background()
)

type distAdapter = gcache.Adapter

// Config 磁盘缓存配置
type Config struct {
	Dir string
}

// SetConfig sets the global configuration for specified group.
// If `name` is not passed, it sets configuration for the default group name.
func SetConfig(config *Config, name ...string) {
	group := DefaultGroupName
	if len(name) > 0 {
		group = name[0]
	}
	localConfigMap.Set(group, config)
	g.Log().Printf(context.TODO(), `SetConfig for group "%s": %+v`, group, config)
}

func New(name ...string) *Dist {
	var (
		group  = DefaultGroupName
		config *Config
		cache  *Dist
	)
	if len(name) > 0 && name[0] != "" {
		group = name[0]
	}
	instanceKey := fmt.Sprintf("%s.%s", DistCacheName, group)
	result := instance.GetOrSetFuncLock(instanceKey, func() interface{} {
		configMap := localConfigMap.Get(group)
		if configMap != nil {
			err := gconv.Struct(configMap, &config)
			if err != nil {
				panic(fmt.Sprintf(`missing configuration for distCache:"%+v"`, err))
			}
			if config == nil {
				panic(`missing configuration for distCache:"config unusable"`)
			}
		} else {
			panic(`missing configuration for distCache:"config not set"`)
		}
		option := badger.DefaultOptions(config.Dir).
			WithValueLogFileSize(100 << 20).
			WithMemTableSize(50 << 20).
			WithValueThreshold(512 << 10)
		db, err := badger.Open(option)
		if err != nil {
			panic(fmt.Sprintf(`loading dis db wrong:"%+v"`, err))
		}
		cache = &Dist{
			config: config,
			db:     db,
		}
		return cache
	})
	if result != nil {
		return result.(*Dist)
	}
	return nil
}

func NewDist() *Dist {
	return New()
}

type Dist struct {
	config *Config
	db     *badger.DB
	mu     sync.RWMutex
}

func (d *Dist) Set(ctx context.Context, key interface{}, value interface{}, duration time.Duration) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	duration = d.getInternalExpire(duration)
	err := d.db.Update(func(txn *badger.Txn) (err error) {
		value, err = d.convertOptionToArgs(value)
		if err != nil {
			return
		}
		e := badger.NewEntry(gconv.Bytes(key), gconv.Bytes(value)).WithTTL(duration)
		return txn.SetEntry(e)
	})
	return err
}

func (d *Dist) SetMap(ctx context.Context, data map[interface{}]interface{}, duration time.Duration) error {
	for k, v := range data {
		err := d.Set(ctx, k, v, duration)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Dist) SetIfNotExist(ctx context.Context, key interface{}, value interface{}, duration time.Duration) (ok bool, err error) {
	ok, err = d.Contains(ctx, key)
	if err != nil {
		return false, err
	}
	if ok {
		return false, nil
	}
	err = d.Set(ctx, key, value, duration)
	if err != nil {
		return
	}
	return true, nil
}

func (d *Dist) SetIfNotExistFunc(ctx context.Context, key interface{}, f gcache.Func, duration time.Duration) (ok bool, err error) {
	ok, err = d.Contains(ctx, key)
	if err != nil {
		return false, err
	}
	if ok {
		return false, nil
	}
	value, err := f(ctx)
	if err != nil {
		return false, err
	}
	err = d.Set(ctx, key, value, duration)
	if err != nil {
		return
	}
	return true, nil
}

func (d *Dist) SetIfNotExistFuncLock(ctx context.Context, key interface{}, f gcache.Func, duration time.Duration) (ok bool, err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	ok, err = d.SetIfNotExistFunc(ctx, key, f, duration)
	return
}

func (d *Dist) Get(ctx context.Context, key interface{}) (value *gvar.Var, err error) {
	err = d.db.View(func(txn *badger.Txn) error {
		item, e := txn.Get(gconv.Bytes(key))
		if e != nil {
			g.Log().Error(ctx, e)
			return nil
		}
		if item != nil {
			err = item.Value(func(val []byte) error {
				value = gvar.New(val)
				return nil
			})
		}
		return err
	})
	return
}

func (d *Dist) GetOrSet(ctx context.Context, key interface{}, value interface{}, duration time.Duration) (result *gvar.Var, err error) {
	result, _ = d.Get(ctx, key)
	if !result.IsEmpty() {
		return
	}
	result = gvar.New(value)
	err = d.Set(ctx, key, value, duration)
	return
}

func (d *Dist) GetOrSetFunc(ctx context.Context, key interface{}, f gcache.Func, duration time.Duration) (result *gvar.Var, err error) {
	result, _ = d.Get(ctx, key)
	if !result.IsEmpty() {
		return
	}
	var value interface{}
	value, err = f(ctx)
	result = gvar.New(value)
	err = d.Set(ctx, key, value, duration)
	return
}

func (d *Dist) GetOrSetFuncLock(ctx context.Context, key interface{}, f gcache.Func, duration time.Duration) (result *gvar.Var, err error) {
	return d.GetOrSetFunc(ctx, key, f, duration)
}

func (d *Dist) Contains(ctx context.Context, key interface{}) (b bool, err error) {
	var val *gvar.Var
	val, err = d.Get(ctx, key)
	if err != nil {
		return
	}
	b = !val.IsEmpty()
	return
}

func (d *Dist) Size(ctx context.Context) (size int, err error) {
	// 创建一个只读事务
	err = d.db.View(func(txn *badger.Txn) error {
		// 创建一个迭代器
		iterator := txn.NewIterator(badger.DefaultIteratorOptions)
		defer iterator.Close()
		// 遍历键值对并计数
		for iterator.Rewind(); iterator.Valid(); iterator.Next() {
			size++
		}
		return nil
	})
	return
}

func (d *Dist) Data(ctx context.Context) (data map[interface{}]interface{}, err error) {
	data = make(map[interface{}]interface{}, 1000)
	err = d.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			err = item.Value(func(v []byte) error {
				fmt.Printf("键值对的值分别是:%s-%s\n", k, v)
				data[gconv.String(k)] = v
				return nil
			})
			if err != nil {
				g.Log().Error(ctx, err)
				return err
			}
		}
		return nil
	})
	return
}

func (d *Dist) Keys(ctx context.Context) (keys []interface{}, err error) {
	keys = make([]interface{}, 0, 1000)
	err = d.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			keys = append(keys, k)
		}
		return nil
	})
	return
}

func (d *Dist) Values(ctx context.Context) (values []interface{}, err error) {
	values = make([]interface{}, 0, 1000)
	err = d.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			err = item.Value(func(v []byte) error {
				values = append(values, v)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return
}

func (d *Dist) Update(ctx context.Context, key interface{}, value interface{}) (oldValue *gvar.Var, exist bool, err error) {
	oldValue, _ = d.Get(ctx, key)
	if !oldValue.IsEmpty() {
		exist = true
	}
	var duration time.Duration
	duration, err = d.GetExpire(ctx, key)
	if err != nil {
		return
	}
	err = d.Set(ctx, key, value, duration)
	return
}

func (d *Dist) UpdateExpire(ctx context.Context, key interface{}, duration time.Duration) (oldDuration time.Duration, err error) {
	err = d.db.Update(func(txn *badger.Txn) error {
		// 获取键的元数据
		item, err := txn.Get(gconv.Bytes(key))
		if err != nil {
			return err
		}
		expire := gtime.NewFromTimeStamp(gconv.Int64(item.ExpiresAt()))
		now := gtime.Now()
		oldDuration = gconv.Duration(expire.Sub(now))
		err = item.Value(func(val []byte) error {
			duration = d.getInternalExpire(duration)
			e := badger.NewEntry(gconv.Bytes(key), val).WithTTL(duration)
			err = txn.SetEntry(e)
			return err
		})
		return err
	})
	return
}

func (d *Dist) GetExpire(ctx context.Context, key interface{}) (duration time.Duration, err error) {
	err = d.db.View(func(txn *badger.Txn) error {
		// 获取键的元数据
		item, err := txn.Get(gconv.Bytes(key))
		if err != nil {
			return err
		}
		expire := gtime.NewFromTimeStamp(gconv.Int64(item.ExpiresAt()))
		now := gtime.Now()
		duration = gconv.Duration(expire.Sub(now))
		return nil
	})
	return
}

func (d *Dist) Remove(ctx context.Context, keys ...interface{}) (lastValue *gvar.Var, err error) {
	err = d.db.Update(func(txn *badger.Txn) error {
		for index, key := range keys {
			if index == len(keys)-1 {
				item, err := txn.Get(gconv.Bytes(key))
				if err != nil {
					return err
				}
				err = item.Value(func(val []byte) error {
					lastValue = gvar.New(val)
					return nil
				})
				if err != nil {
					return err
				}
			}
			err = txn.Delete(gconv.Bytes(key))
			if err != nil {
				return err
			}
		}
		return nil
	})
	return
}

func (d *Dist) Clear(ctx context.Context) error {
	err := d.db.DropAll()
	return err
}

func (d *Dist) Close(ctx context.Context) error {
	err := d.db.Close()
	return err
}

// getInternalExpire converts and returns the expiration time with given expired duration in milliseconds.
func (d *Dist) getInternalExpire(duration time.Duration) time.Duration {
	if duration == 0 {
		return defaultMaxExpire * time.Millisecond
	}
	return duration
}

func (d *Dist) convertOptionToArgs(option interface{}) (result interface{}, err error) {
	if option == nil {
		return nil, nil
	}
	switch reflect.TypeOf(option).Kind() {
	case reflect.Ptr, reflect.Struct:
		result = gconv.Map(option)
	case reflect.Bool:
		result = gconv.String(option)
	case reflect.Slice, reflect.Array:
		optionSlice := gconv.SliceAny(option)
		var newOption = make(g.SliceAny, len(optionSlice))
		for k, v := range optionSlice {
			newOption[k], err = d.convertOptionToArgs(v)
			if err != nil {
				return
			}
		}
		option = newOption
	default:
		result = option
	}
	return
}
