package integrations

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/worker/core"
	"github.com/redis/go-redis/v9"
)

// RedisNode handles Redis database operations
type RedisNode struct{}

func (n *RedisNode) Type() string {
	return "integration.redis"
}

func (n *RedisNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config
	operation := getString(config, "operation", "get")

	credIDStr := getString(config, "credentialId", "")
	if credIDStr == "" {
		return nil, fmt.Errorf("Redis credential is required")
	}

	credID, err := uuid.Parse(credIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid credential ID")
	}

	cred, err := execCtx.GetCredential(credID)
	if err != nil {
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}

	client, err := n.connect(ctx, cred.Data)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}
	defer client.Close()

	switch operation {
	case "get":
		return n.get(ctx, client, config)
	case "set":
		return n.set(ctx, client, config, execCtx.Input)
	case "delete":
		return n.delete(ctx, client, config)
	case "exists":
		return n.exists(ctx, client, config)
	case "incr":
		return n.incr(ctx, client, config)
	case "decr":
		return n.decr(ctx, client, config)
	case "expire":
		return n.expire(ctx, client, config)
	case "ttl":
		return n.ttl(ctx, client, config)
	case "keys":
		return n.keys(ctx, client, config)
	case "hget":
		return n.hget(ctx, client, config)
	case "hset":
		return n.hset(ctx, client, config, execCtx.Input)
	case "hgetall":
		return n.hgetall(ctx, client, config)
	case "hdel":
		return n.hdel(ctx, client, config)
	case "lpush":
		return n.lpush(ctx, client, config, execCtx.Input)
	case "rpush":
		return n.rpush(ctx, client, config, execCtx.Input)
	case "lpop":
		return n.lpop(ctx, client, config)
	case "rpop":
		return n.rpop(ctx, client, config)
	case "lrange":
		return n.lrange(ctx, client, config)
	case "llen":
		return n.llen(ctx, client, config)
	case "sadd":
		return n.sadd(ctx, client, config, execCtx.Input)
	case "smembers":
		return n.smembers(ctx, client, config)
	case "sismember":
		return n.sismember(ctx, client, config)
	case "srem":
		return n.srem(ctx, client, config)
	case "publish":
		return n.publish(ctx, client, config, execCtx.Input)
	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}
}

func (n *RedisNode) connect(ctx context.Context, creds map[string]interface{}) (*redis.Client, error) {
	host := getString(creds, "host", "localhost")
	port := getInt(creds, "port", 6379)
	password := getString(creds, "password", "")
	db := getInt(creds, "db", 0)
	useTLS := getBool(creds, "tls", false)

	opts := &redis.Options{
		Addr:        fmt.Sprintf("%s:%d", host, port),
		Password:    password,
		DB:          db,
		DialTimeout: 10 * time.Second,
		PoolSize:    5,
	}

	if useTLS {
		opts.TLSConfig = nil // Use default TLS config
	}

	client := redis.NewClient(opts)

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return client, nil
}

// String operations
func (n *RedisNode) get(ctx context.Context, client *redis.Client, config map[string]interface{}) (map[string]interface{}, error) {
	key := getString(config, "key", "")
	if key == "" {
		return nil, fmt.Errorf("key is required")
	}

	val, err := client.Get(ctx, key).Result()
	if err == redis.Nil {
		return map[string]interface{}{
			"value": nil,
			"found": false,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get failed: %w", err)
	}

	return map[string]interface{}{
		"value": val,
		"found": true,
	}, nil
}

func (n *RedisNode) set(ctx context.Context, client *redis.Client, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	key := getString(config, "key", "")
	if key == "" {
		return nil, fmt.Errorf("key is required")
	}

	value := config["value"]
	if value == nil {
		value = input["value"]
	}

	ttlSeconds := getInt(config, "ttl", 0)
	var expiration time.Duration
	if ttlSeconds > 0 {
		expiration = time.Duration(ttlSeconds) * time.Second
	}

	err := client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return nil, fmt.Errorf("set failed: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"key":     key,
	}, nil
}

func (n *RedisNode) delete(ctx context.Context, client *redis.Client, config map[string]interface{}) (map[string]interface{}, error) {
	key := getString(config, "key", "")
	if key == "" {
		return nil, fmt.Errorf("key is required")
	}

	keys := []string{key}
	if keysArr := getArray(config, "keys"); len(keysArr) > 0 {
		keys = make([]string, len(keysArr))
		for i, k := range keysArr {
			keys[i] = fmt.Sprintf("%v", k)
		}
	}

	deleted, err := client.Del(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("delete failed: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"deleted": deleted,
	}, nil
}

func (n *RedisNode) exists(ctx context.Context, client *redis.Client, config map[string]interface{}) (map[string]interface{}, error) {
	key := getString(config, "key", "")
	if key == "" {
		return nil, fmt.Errorf("key is required")
	}

	count, err := client.Exists(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("exists failed: %w", err)
	}

	return map[string]interface{}{
		"exists": count > 0,
		"count":  count,
	}, nil
}

func (n *RedisNode) incr(ctx context.Context, client *redis.Client, config map[string]interface{}) (map[string]interface{}, error) {
	key := getString(config, "key", "")
	if key == "" {
		return nil, fmt.Errorf("key is required")
	}

	by := getInt(config, "by", 1)
	var val int64
	var err error

	if by == 1 {
		val, err = client.Incr(ctx, key).Result()
	} else {
		val, err = client.IncrBy(ctx, key, int64(by)).Result()
	}

	if err != nil {
		return nil, fmt.Errorf("incr failed: %w", err)
	}

	return map[string]interface{}{
		"value": val,
	}, nil
}

func (n *RedisNode) decr(ctx context.Context, client *redis.Client, config map[string]interface{}) (map[string]interface{}, error) {
	key := getString(config, "key", "")
	if key == "" {
		return nil, fmt.Errorf("key is required")
	}

	by := getInt(config, "by", 1)
	var val int64
	var err error

	if by == 1 {
		val, err = client.Decr(ctx, key).Result()
	} else {
		val, err = client.DecrBy(ctx, key, int64(by)).Result()
	}

	if err != nil {
		return nil, fmt.Errorf("decr failed: %w", err)
	}

	return map[string]interface{}{
		"value": val,
	}, nil
}

func (n *RedisNode) expire(ctx context.Context, client *redis.Client, config map[string]interface{}) (map[string]interface{}, error) {
	key := getString(config, "key", "")
	if key == "" {
		return nil, fmt.Errorf("key is required")
	}

	ttlSeconds := getInt(config, "ttl", 0)
	if ttlSeconds <= 0 {
		return nil, fmt.Errorf("ttl must be positive")
	}

	success, err := client.Expire(ctx, key, time.Duration(ttlSeconds)*time.Second).Result()
	if err != nil {
		return nil, fmt.Errorf("expire failed: %w", err)
	}

	return map[string]interface{}{
		"success": success,
	}, nil
}

func (n *RedisNode) ttl(ctx context.Context, client *redis.Client, config map[string]interface{}) (map[string]interface{}, error) {
	key := getString(config, "key", "")
	if key == "" {
		return nil, fmt.Errorf("key is required")
	}

	duration, err := client.TTL(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("ttl failed: %w", err)
	}

	ttl := int64(duration.Seconds())
	if duration == -1 {
		ttl = -1 // No expiry
	} else if duration == -2 {
		ttl = -2 // Key doesn't exist
	}

	return map[string]interface{}{
		"ttl": ttl,
	}, nil
}

func (n *RedisNode) keys(ctx context.Context, client *redis.Client, config map[string]interface{}) (map[string]interface{}, error) {
	pattern := getString(config, "pattern", "*")

	keys, err := client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("keys failed: %w", err)
	}

	return map[string]interface{}{
		"keys":  keys,
		"count": len(keys),
	}, nil
}

// Hash operations
func (n *RedisNode) hget(ctx context.Context, client *redis.Client, config map[string]interface{}) (map[string]interface{}, error) {
	key := getString(config, "key", "")
	field := getString(config, "field", "")
	if key == "" || field == "" {
		return nil, fmt.Errorf("key and field are required")
	}

	val, err := client.HGet(ctx, key, field).Result()
	if err == redis.Nil {
		return map[string]interface{}{
			"value": nil,
			"found": false,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("hget failed: %w", err)
	}

	return map[string]interface{}{
		"value": val,
		"found": true,
	}, nil
}

func (n *RedisNode) hset(ctx context.Context, client *redis.Client, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	key := getString(config, "key", "")
	if key == "" {
		return nil, fmt.Errorf("key is required")
	}

	field := getString(config, "field", "")
	value := config["value"]
	if value == nil {
		value = input["value"]
	}

	// Support single field or multiple fields
	if field != "" && value != nil {
		err := client.HSet(ctx, key, field, value).Err()
		if err != nil {
			return nil, fmt.Errorf("hset failed: %w", err)
		}
	} else if fields := getMap(config, "fields"); len(fields) > 0 {
		err := client.HSet(ctx, key, fields).Err()
		if err != nil {
			return nil, fmt.Errorf("hset failed: %w", err)
		}
	} else {
		return nil, fmt.Errorf("field/value or fields map is required")
	}

	return map[string]interface{}{
		"success": true,
	}, nil
}

func (n *RedisNode) hgetall(ctx context.Context, client *redis.Client, config map[string]interface{}) (map[string]interface{}, error) {
	key := getString(config, "key", "")
	if key == "" {
		return nil, fmt.Errorf("key is required")
	}

	result, err := client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("hgetall failed: %w", err)
	}

	return map[string]interface{}{
		"data":  result,
		"count": len(result),
	}, nil
}

func (n *RedisNode) hdel(ctx context.Context, client *redis.Client, config map[string]interface{}) (map[string]interface{}, error) {
	key := getString(config, "key", "")
	field := getString(config, "field", "")
	if key == "" || field == "" {
		return nil, fmt.Errorf("key and field are required")
	}

	fields := []string{field}
	if fieldsArr := getArray(config, "fields"); len(fieldsArr) > 0 {
		fields = make([]string, len(fieldsArr))
		for i, f := range fieldsArr {
			fields[i] = fmt.Sprintf("%v", f)
		}
	}

	deleted, err := client.HDel(ctx, key, fields...).Result()
	if err != nil {
		return nil, fmt.Errorf("hdel failed: %w", err)
	}

	return map[string]interface{}{
		"deleted": deleted,
	}, nil
}

// List operations
func (n *RedisNode) lpush(ctx context.Context, client *redis.Client, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	key := getString(config, "key", "")
	if key == "" {
		return nil, fmt.Errorf("key is required")
	}

	values := getArray(config, "values")
	if len(values) == 0 {
		if v := config["value"]; v != nil {
			values = []interface{}{v}
		} else if v := input["value"]; v != nil {
			values = []interface{}{v}
		}
	}

	if len(values) == 0 {
		return nil, fmt.Errorf("values are required")
	}

	length, err := client.LPush(ctx, key, values...).Result()
	if err != nil {
		return nil, fmt.Errorf("lpush failed: %w", err)
	}

	return map[string]interface{}{
		"length": length,
	}, nil
}

func (n *RedisNode) rpush(ctx context.Context, client *redis.Client, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	key := getString(config, "key", "")
	if key == "" {
		return nil, fmt.Errorf("key is required")
	}

	values := getArray(config, "values")
	if len(values) == 0 {
		if v := config["value"]; v != nil {
			values = []interface{}{v}
		} else if v := input["value"]; v != nil {
			values = []interface{}{v}
		}
	}

	if len(values) == 0 {
		return nil, fmt.Errorf("values are required")
	}

	length, err := client.RPush(ctx, key, values...).Result()
	if err != nil {
		return nil, fmt.Errorf("rpush failed: %w", err)
	}

	return map[string]interface{}{
		"length": length,
	}, nil
}

func (n *RedisNode) lpop(ctx context.Context, client *redis.Client, config map[string]interface{}) (map[string]interface{}, error) {
	key := getString(config, "key", "")
	if key == "" {
		return nil, fmt.Errorf("key is required")
	}

	count := getInt(config, "count", 1)
	if count <= 1 {
		val, err := client.LPop(ctx, key).Result()
		if err == redis.Nil {
			return map[string]interface{}{"value": nil}, nil
		}
		if err != nil {
			return nil, fmt.Errorf("lpop failed: %w", err)
		}
		return map[string]interface{}{"value": val}, nil
	}

	vals, err := client.LPopCount(ctx, key, count).Result()
	if err != nil {
		return nil, fmt.Errorf("lpop failed: %w", err)
	}

	return map[string]interface{}{
		"values": vals,
	}, nil
}

func (n *RedisNode) rpop(ctx context.Context, client *redis.Client, config map[string]interface{}) (map[string]interface{}, error) {
	key := getString(config, "key", "")
	if key == "" {
		return nil, fmt.Errorf("key is required")
	}

	count := getInt(config, "count", 1)
	if count <= 1 {
		val, err := client.RPop(ctx, key).Result()
		if err == redis.Nil {
			return map[string]interface{}{"value": nil}, nil
		}
		if err != nil {
			return nil, fmt.Errorf("rpop failed: %w", err)
		}
		return map[string]interface{}{"value": val}, nil
	}

	vals, err := client.RPopCount(ctx, key, count).Result()
	if err != nil {
		return nil, fmt.Errorf("rpop failed: %w", err)
	}

	return map[string]interface{}{
		"values": vals,
	}, nil
}

func (n *RedisNode) lrange(ctx context.Context, client *redis.Client, config map[string]interface{}) (map[string]interface{}, error) {
	key := getString(config, "key", "")
	if key == "" {
		return nil, fmt.Errorf("key is required")
	}

	start := int64(getInt(config, "start", 0))
	stop := int64(getInt(config, "stop", -1))

	vals, err := client.LRange(ctx, key, start, stop).Result()
	if err != nil {
		return nil, fmt.Errorf("lrange failed: %w", err)
	}

	return map[string]interface{}{
		"values": vals,
		"count":  len(vals),
	}, nil
}

func (n *RedisNode) llen(ctx context.Context, client *redis.Client, config map[string]interface{}) (map[string]interface{}, error) {
	key := getString(config, "key", "")
	if key == "" {
		return nil, fmt.Errorf("key is required")
	}

	length, err := client.LLen(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("llen failed: %w", err)
	}

	return map[string]interface{}{
		"length": length,
	}, nil
}

// Set operations
func (n *RedisNode) sadd(ctx context.Context, client *redis.Client, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	key := getString(config, "key", "")
	if key == "" {
		return nil, fmt.Errorf("key is required")
	}

	members := getArray(config, "members")
	if len(members) == 0 {
		if v := config["member"]; v != nil {
			members = []interface{}{v}
		} else if v := input["member"]; v != nil {
			members = []interface{}{v}
		}
	}

	if len(members) == 0 {
		return nil, fmt.Errorf("members are required")
	}

	added, err := client.SAdd(ctx, key, members...).Result()
	if err != nil {
		return nil, fmt.Errorf("sadd failed: %w", err)
	}

	return map[string]interface{}{
		"added": added,
	}, nil
}

func (n *RedisNode) smembers(ctx context.Context, client *redis.Client, config map[string]interface{}) (map[string]interface{}, error) {
	key := getString(config, "key", "")
	if key == "" {
		return nil, fmt.Errorf("key is required")
	}

	members, err := client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("smembers failed: %w", err)
	}

	return map[string]interface{}{
		"members": members,
		"count":   len(members),
	}, nil
}

func (n *RedisNode) sismember(ctx context.Context, client *redis.Client, config map[string]interface{}) (map[string]interface{}, error) {
	key := getString(config, "key", "")
	member := getString(config, "member", "")
	if key == "" || member == "" {
		return nil, fmt.Errorf("key and member are required")
	}

	isMember, err := client.SIsMember(ctx, key, member).Result()
	if err != nil {
		return nil, fmt.Errorf("sismember failed: %w", err)
	}

	return map[string]interface{}{
		"isMember": isMember,
	}, nil
}

func (n *RedisNode) srem(ctx context.Context, client *redis.Client, config map[string]interface{}) (map[string]interface{}, error) {
	key := getString(config, "key", "")
	if key == "" {
		return nil, fmt.Errorf("key is required")
	}

	members := getArray(config, "members")
	if len(members) == 0 {
		if v := config["member"]; v != nil {
			members = []interface{}{v}
		}
	}

	if len(members) == 0 {
		return nil, fmt.Errorf("members are required")
	}

	removed, err := client.SRem(ctx, key, members...).Result()
	if err != nil {
		return nil, fmt.Errorf("srem failed: %w", err)
	}

	return map[string]interface{}{
		"removed": removed,
	}, nil
}

// Pub/Sub
func (n *RedisNode) publish(ctx context.Context, client *redis.Client, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	channel := getString(config, "channel", "")
	if channel == "" {
		return nil, fmt.Errorf("channel is required")
	}

	message := config["message"]
	if message == nil {
		message = input["message"]
	}
	if message == nil {
		return nil, fmt.Errorf("message is required")
	}

	receivers, err := client.Publish(ctx, channel, message).Result()
	if err != nil {
		return nil, fmt.Errorf("publish failed: %w", err)
	}

	return map[string]interface{}{
		"receivers": receivers,
	}, nil
}


