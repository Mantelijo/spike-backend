package cache

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Mantelijo/spike-backend/internal/dto"
	"github.com/redis/go-redis/v9"
	"golang.org/x/exp/slog"
)

var ErrorNotFound = errors.New("item not found in cache")

const (
	RedisKeyRecentlyUpdatedWidgets = "recently_updated_widgets"
)

type RedisCache interface {
	GetWidget(serialNumber string) (*dto.Widget, error)
	// SetWidget puts widget meta data into the cache
	SetWidget(widget *dto.Widget) error

	// GetConnections retrieves the widget's connections from cache.
	GetConnections(serialNumber string) (*dto.WidgetConnections, error)
	// SetConnections updates the widget's connections with non zero sn peers in
	// provided conns object.
	SetConnections(conns *dto.WidgetConnections) error

	// UnsetConnections removes the widget's (and peers') connections from
	// cache.
	UnsetConnections(serialNumber string, ports dto.WidgetPortBitmap) error

	// RetrieveRecentUpdates returns up to numUpdates or more (SSCAN count
	// behaviour) of recent widget connection updates, removes the serial
	// numbers of retrieves widgets from recently_updated_widgets set.
	RetrieveRecentUpdates(numUpdates int) ([]*dto.WidgetConnections, error)
}

func NewRedisCache(c *redis.Client) *redisCache {
	return &redisCache{c: c}
}

var _ RedisCache = (*redisCache)(nil)

type redisCache struct {
	c *redis.Client
}

func (r *redisCache) SetConnections(conns *dto.WidgetConnections) error {
	// TODO unit tests for this would be good

	// Build lua script parts to only update non empty peers from connn.
	commands := []string{}
	nArgs := 0
	argV := []interface{}{}
	if conns.P_PeerSerialNumber != "" {
		nArgs++
		argV = append(argV, conns.P_PeerSerialNumber)
		cmdOwnerWidgetConns := fmt.Sprintf(`redis.call('HSET', 'c:' .. KEYS[1], 'p_peer_sn', ARGV[%d])`, nArgs)
		cmdPeerWidgetConns := fmt.Sprintf(`redis.call('HSET', 'c:' .. ARGV[%d], 'p_peer_sn', KEYS[1])`, nArgs)
		commands = append(commands, cmdOwnerWidgetConns, cmdPeerWidgetConns)
	}
	if conns.R_PeerSerialNumber != "" {
		nArgs++
		argV = append(argV, conns.R_PeerSerialNumber)
		cmdOwnerWidgetConns := fmt.Sprintf(`redis.call('HSET', 'c:' .. KEYS[1], 'r_peer_sn', ARGV[%d])`, nArgs)
		cmdPeerWidgetConns := fmt.Sprintf(`redis.call('HSET', 'c:' .. ARGV[%d], 'r_peer_sn', KEYS[1])`, nArgs)
		commands = append(commands, cmdOwnerWidgetConns, cmdPeerWidgetConns)
	}
	if conns.Q_PeerSerialNumber != "" {
		nArgs++
		argV = append(argV, conns.Q_PeerSerialNumber)
		cmdOwnerWidgetConns := fmt.Sprintf(`redis.call('HSET', 'c:' .. KEYS[1], 'q_peer_sn', ARGV[%d])`, nArgs)
		cmdPeerWidgetConns := fmt.Sprintf(`redis.call('HSET', 'c:' .. ARGV[%d], 'q_peer_sn', KEYS[1])`, nArgs)
		commands = append(commands, cmdOwnerWidgetConns, cmdPeerWidgetConns)
	}

	// Set all the connections for the widget and its peers wiht non empty
	// values from conns. If all 3 ports are non empty, the lua script will resemble the following:
	// lua := `
	// 	redis.call('HSET', 'c:' .. KEYS[1], 'p_peer_sn', ARGV[1], 'q_peer_sn', ARGV[2], 'r_peer_sn', ARGV[3]) <-- This is split into 3 separate commands for each PRQ
	// 	redis.call('HSET', 'c:' .. ARGV[1], 'p_peer_sn', KEYS[1])
	// 	redis.call('HSET', 'c:' .. ARGV[2], 'r_peer_sn', KEYS[1])
	// 	redis.call('HSET', 'c:' .. ARGV[3], 'q_peer_sn', KEYS[1])
	// 	return redis.call('RPUSH', KEYS[2], KEYS[1])
	// `
	lua := fmt.Sprintf(`%s
		return redis.call('RPUSH', KEYS[2], KEYS[1])
	`, strings.Join(commands, "\n"))

	return r.c.Eval(
		context.Background(),
		lua,
		[]string{conns.SerialNumber, RedisKeyRecentlyUpdatedWidgets},
		argV...,
	).Err()
}

func (r *redisCache) RetrieveRecentUpdates(numUpdates int) ([]*dto.WidgetConnections, error) {
	// retrieve recently_updated_widgets, their connections, remove
	// their sns from recently_updated_widgets list and return the result for furher
	// processing
	//
	// FUTURE TODO ensure retrieved records are actually updated in db by
	// pushing data to persistent queue, implementing retry logic, etc.
	lua := `
		local updated_serial_nums = redis.call('LPOP', KEYS[1], ARGV[1])
		if type(updated_serial_nums) ~= 'table' then
			return {}
		end 
		-- each entry in widget_connections will have the serial number of owner widget as the first element
		local widget_connections = {}
		for i, serial_num in ipairs(updated_serial_nums) do
			local connections = redis.call('HGETALL', 'c:' .. serial_num)
			-- insert current widget's sn as the first element
			table.insert(connections, 1, serial_num)
			table.insert(widget_connections, connections)
		end
		return widget_connections
	`

	res := r.c.Eval(
		context.Background(),
		lua,
		[]string{RedisKeyRecentlyUpdatedWidgets},
		numUpdates,
	)

	if res.Err() != nil {
		return nil, res.Err()
	}

	// Parse the result. First element is the serial number of widget, rest 6
	// elements are connection hash result (p, q, r key/val pairs).
	//
	resVal := res.Val().([]any)
	ret := make([]*dto.WidgetConnections, 0, len(resVal))
	// sn: index in ret
	presentSerialNumbers := map[string]int{}
	for _, v := range resVal {
		v := v.([]any)
		el := &dto.WidgetConnections{}
		el.SerialNumber = v[0].(string)

		for j := 1; j < len(v); j += 2 {
			key := v[j].(string)
			val := v[j+1].(string)
			switch key {
			case "p_peer_sn":
				el.P_PeerSerialNumber = val
			case "q_peer_sn":
				el.Q_PeerSerialNumber = val
			case "r_peer_sn":
				el.R_PeerSerialNumber = val
			}
		}

		// Ensure that one serial number (owner widget) is present in the result
		// only once.
		if indexInRet, present := presentSerialNumbers[el.SerialNumber]; present {
			ret[indexInRet] = el
		} else {
			ret = append(ret, el)
			presentSerialNumbers[el.SerialNumber] = len(ret) - 1
		}

	}

	return ret, nil
}

func (r *redisCache) GetWidget(serialNumber string) (*dto.Widget, error) {
	slog.Error("[GetWidget] not implemented")
	return nil, nil
}

func (r *redisCache) SetWidget(widget *dto.Widget) error {
	slog.Error("[SetWidget] not implemented")
	return nil
}

func (r *redisCache) UnsetConnections(serialNumber string, ports dto.WidgetPortBitmap) error {
	slog.Error("[UnsetConnections] not implemented")
	return nil
}

func (r *redisCache) GetConnections(serialNumber string) (*dto.WidgetConnections, error) {
	slog.Error("[GetConnections] not implemented")
	return nil, nil
}
