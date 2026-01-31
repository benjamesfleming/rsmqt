package rsmq

import (
	"errors"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Client struct {
	rdb *redis.Client
	ns  string
}

func NewClient(addr, password string, db int, ns string) *Client {
	if ns == "" {
		ns = "rsmq:"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &Client{
		rdb: rdb,
		ns:  ns,
	}
}

type QueueStats struct {
	Name       string
	Vt         int
	Delay      int
	MaxSize    int
	TotalRecv  uint64
	TotalSent  uint64
	Created    time.Time
	Modified   time.Time
	Msgs       int64
	HiddenMsgs int64
}

type Message struct {
	ID        string
	Body      string
	Rc        int
	Fr        time.Time
	Sent      time.Time
	VisibleAt time.Time
}

func (c *Client) TestConnection() error {
	return c.rdb.Ping().Err()
}

func (c *Client) ListQueues() ([]string, error) {
	return c.rdb.SMembers(c.ns + "QUEUES").Result()
}

func (c *Client) GetQueueStats(qname string) (*QueueStats, error) {
	key := c.ns + qname
	// Get Attributes from Hash
	// Fields: vt, delay, maxsize, totalrecv, totalsent, created, modified
	res, err := c.rdb.HMGet(key+":Q", "vt", "delay", "maxsize", "totalrecv", "totalsent", "created", "modified").Result()
	if err != nil {
		return nil, err
	}

	// If all nil, queue might not exist
	if len(res) == 0 || res[0] == nil {
		return nil, errors.New("queue not found")
	}

	toInt := func(v interface{}) int {
		if s, ok := v.(string); ok {
			i, _ := strconv.Atoi(s)
			return i
		}
		return 0
	}

	toUint64 := func(v interface{}) uint64 {
		if s, ok := v.(string); ok {
			i, _ := strconv.ParseUint(s, 10, 64)
			return i
		}
		return 0
	}

	toInt64 := func(v interface{}) int64 {
		if s, ok := v.(string); ok {
			i, _ := strconv.ParseInt(s, 10, 64)
			return i
		}
		return 0
	}

	stats := &QueueStats{
		Name:      qname,
		Vt:        toInt(res[0]),
		Delay:     toInt(res[1]),
		MaxSize:   toInt(res[2]),
		TotalRecv: toUint64(res[3]),
		TotalSent: toUint64(res[4]),
		Created:   time.Unix(toInt64(res[5]), 0),
		Modified:  time.Unix(toInt64(res[6]), 0),
	}

	// Get Msgs Count (ZCard)
	stats.Msgs, _ = c.rdb.ZCard(key).Result()

	// Get Hidden Msgs Count (ZCount where score > now)
	nowMs := time.Now().UnixNano() / 1e6
	stats.HiddenMsgs, _ = c.rdb.ZCount(key, strconv.FormatInt(nowMs, 10), "+inf").Result()

	return stats, nil
}

func (c *Client) ListMessages(qname string) ([]Message, error) {
	key := c.ns + qname

	// Get all members from ZSet with scores
	zres, err := c.rdb.ZRangeWithScores(key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	if len(zres) == 0 {
		return []Message{}, nil
	}

	msgs := make([]Message, len(zres))

	hashKey := key + ":Q"
	fields := make([]string, 0, len(zres)*3)
	for _, z := range zres {
		id := z.Member.(string)
		fields = append(fields, id, id+":rc", id+":fr")
	}

	hmres, err := c.rdb.HMGet(hashKey, fields...).Result()
	if err != nil {
		return nil, err
	}

	for i, z := range zres {
		id := z.Member.(string)

		// Parse ID to get Sent time
		// Match RSMQ implementation: parseInt(id.slice(0, 10), 36)
		sent := time.Time{}
		parseLen := 10
		if len(id) < 10 {
			parseLen = len(id)
		}

		if parseLen > 0 {
			tsMs, _ := strconv.ParseInt(id[:parseLen], 36, 64)
			sent = time.UnixMicro(tsMs)
		}

		body := ""
		if val := hmres[i*3]; val != nil {
			body = val.(string)
		}

		rc := 0
		if val := hmres[i*3+1]; val != nil {
			if s, ok := val.(string); ok {
				rc, _ = strconv.Atoi(s)
			}
		}

		fr := time.Time{}
		if val := hmres[i*3+2]; val != nil {
			if s, ok := val.(string); ok {
				frMs, _ := strconv.ParseInt(s, 10, 64)
				fr = time.UnixMilli(frMs)
			}
		}

		msgs[i] = Message{
			ID:        id,
			Body:      body,
			Rc:        rc,
			Fr:        fr,
			Sent:      sent,
			VisibleAt: time.UnixMilli(int64(z.Score)),
		}
	}

	return msgs, nil
}

func (c *Client) CreateQueue(qname string, vt, delay, maxsize int) error {
	key := c.ns + qname + ":Q"
	exists, err := c.rdb.Exists(key).Result()
	if err != nil {
		return err
	}
	if exists > 0 {
		return errors.New("queue already exists")
	}

	now := time.Now().Unix()

	pipe := c.rdb.TxPipeline()
	pipe.HMSet(key, map[string]interface{}{
		"vt":        vt,
		"delay":     delay,
		"maxsize":   maxsize,
		"created":   now,
		"modified":  now,
		"totalrecv": 0,
		"totalsent": 0,
	})
	pipe.SAdd(c.ns+"QUEUES", qname)
	_, err = pipe.Exec()
	return err
}

func (c *Client) DeleteQueue(qname string) error {
	pipe := c.rdb.TxPipeline()
	pipe.Del(c.ns + qname + ":Q")
	pipe.Del(c.ns + qname)
	pipe.SRem(c.ns+"QUEUES", qname)
	_, err := pipe.Exec()
	return err
}

func (c *Client) SetQueueAttributes(qname string, vt, delay, maxsize int) error {
	key := c.ns + qname + ":Q"
	now := time.Now().Unix()

	exists, err := c.rdb.Exists(key).Result()
	if err != nil {
		return err
	}
	if exists == 0 {
		return errors.New("queue not found")
	}

	_, err = c.rdb.HMSet(key, map[string]interface{}{
		"vt":       vt,
		"delay":    delay,
		"maxsize":  maxsize,
		"modified": now,
	}).Result()
	return err
}

func (c *Client) SendMessage(qname string, message string) error {
	stats, err := c.GetQueueStats(qname)
	if err != nil {
		return err
	}

	if len(message) > stats.MaxSize {
		return errors.New("message too long")
	}

	id := c.generateID()
	now := time.Now().UnixMilli()
	score := now + int64(stats.Delay*1000)

	keyQ := c.ns + qname + ":Q"
	keyZ := c.ns + qname

	pipe := c.rdb.TxPipeline()
	pipe.ZAdd(keyZ, redis.Z{Score: float64(score), Member: id})
	pipe.HMSet(keyQ, map[string]interface{}{
		id:           message,
		id + ":rc":   0,
		id + ":fr":   0,
		id + ":sent": now,
	})
	pipe.HIncrBy(keyQ, "totalsent", 1)

	_, err = pipe.Exec()
	return err
}

func (c *Client) DeleteMessage(qname string, id string) error {
	keyQ := c.ns + qname + ":Q"
	keyZ := c.ns + qname

	pipe := c.rdb.TxPipeline()
	pipe.ZRem(keyZ, id)
	pipe.HDel(keyQ, id, id+":rc", id+":fr", id+":sent")
	_, err := pipe.Exec()
	return err
}

func (c *Client) ClearQueue(qname string) error {
	keyQ := c.ns + qname + ":Q"
	keyZ := c.ns + qname

	// Get all message IDs
	ids, err := c.rdb.ZRange(keyZ, 0, -1).Result()
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}

	// Prepare fields to delete from Hash
	fields := make([]string, 0, len(ids)*4)
	for _, id := range ids {
		fields = append(fields, id, id+":rc", id+":fr", id+":sent")
	}

	pipe := c.rdb.TxPipeline()
	pipe.HDel(keyQ, fields...)
	pipe.Del(keyZ)
	_, err = pipe.Exec()
	return err
}

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

func (c *Client) generateID() string {
	// RSMQ uses microsecond precision for the ID timestamp part
	// Logic: Number(seconds + microseconds).toString(36)
	// Go's UnixMicro returns the number of microseconds elapsed since January 1, 1970 UTC.
	// This is equivalent to seconds*1e6 + microseconds, which matches the JS logic
	// assuming the JS logic intends to create a full microsecond timestamp.
	ts := strconv.FormatInt(time.Now().UnixMicro(), 36)
	if len(ts) < 10 {
		ts = strings.Repeat("0", 10-len(ts)) + ts
	}

	b := make([]byte, 22)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return ts + string(b)
}
