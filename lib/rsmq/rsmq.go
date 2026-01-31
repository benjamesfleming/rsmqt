package rsmq

import (
	"errors"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

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
	Created    int64
	Modified   int64
	Msgs       int64
	HiddenMsgs int64
}

type Message struct {
	ID        string
	Body      string
	Rc        int
	Fr        int64
	Sent      int64
	VisibleAt int64
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
		Created:   toInt64(res[5]),
		Modified:  toInt64(res[6]),
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
		// ID format: base36(timestamp_ms) + random(22)
		// We take first 10 chars usually?
		// Node rsmq says: parseInt(resp[0].slice(0, 10), 36)
		sent := int64(0)
		if len(id) >= 10 {
			tsMs, _ := strconv.ParseInt(id[:10], 36, 64)
			sent = tsMs // This is in ms
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

		fr := int64(0)
		if val := hmres[i*3+2]; val != nil {
			if s, ok := val.(string); ok {
				fr, _ = strconv.ParseInt(s, 10, 64)
			}
		}

		msgs[i] = Message{
			ID:        id,
			Body:      body,
			Rc:        rc,
			Fr:        fr,
			Sent:      sent,
			VisibleAt: int64(z.Score),
		}
	}

	return msgs, nil
}
