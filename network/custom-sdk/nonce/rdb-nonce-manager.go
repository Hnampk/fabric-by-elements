package nonce

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

// NonceManager ---
type NonceManager struct {
	rdb *redis.Client
}

const host = "localhost"
const port = "6379" // default Redis port
const pwd = ""      // no password set
const db = 0        // use default DB
const setNXDuration = time.Second * 30

var ctx = context.Background()

// Init connect to Redis server
func (nm *NonceManager) Init() {
	nm.rdb = redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: pwd,
		DB:       db,
	})
}

// GetLatestNonce return the latest signed nonce of an account
func (nm *NonceManager) GetLatestNonce(account string) (int, error) {
	if nm.rdb == nil {
		return -1, errors.Errorf("NonceManager did not initialized!")
	}

	value := nm.rdb.Get(ctx, account)
	return value.Int()
}

// IncreaseLatestNonce will be called after received the signed proposal from peers
func (nm *NonceManager) IncreaseLatestNonce(account string) error {
	if nm.rdb == nil {
		errMsg := "NonceManager did not initialized!"
		nm.publishIncreasingEvent(account, errMsg)
		return errors.Errorf(errMsg)
	}

	value, err := nm.rdb.Get(ctx, account).Int()

	if err != nil {
		if !nm.IsNilValue(err) {
			errMsg := err.Error()
			nm.publishIncreasingEvent(account, errMsg)
			return err
		}
		value = -1
	}

	increasedValue := value + 1
	rdbSetError := nm.rdb.Set(ctx, account, increasedValue, 0).Err()

	if rdbSetError != nil {
		errMsg := rdbSetError.Error()
		nm.publishIncreasingEvent(account, errMsg)
		return rdbSetError
	}

	nm.publishIncreasingEvent(account, "increased nonce "+account+" to "+strconv.Itoa(increasedValue))
	return nil
}

/*
SetNX make sure that only 1 service can propose to peer at a time
rdb.SetNX return false if the key is existed
	Key: account-nonce
*/
func (nm *NonceManager) SetNX(key string) (bool, error) {
	if nm.rdb == nil {
		return false, errors.Errorf("Redis connection problem!")
	}

	isSetNXOk := nm.rdb.SetNX(ctx, key, "in processing", setNXDuration)

	return isSetNXOk.Val(), nil
}

// IsNilValue maybe a response error from redis.get is not really an error
// with new account, redis.get will return nil
func (nm *NonceManager) IsNilValue(err error) bool {
	if err.Error() != "redis: nil" {
		return false
	}
	return true
}

func (nm *NonceManager) SubscribeChannel(rdbChannel string) *redis.PubSub {
	return nm.rdb.Subscribe(ctx, rdbChannel)
}

func (nm *NonceManager) UnSubscribeChannel(rdbPubSub *redis.PubSub, rdbChannel string) {
	rdbPubSub.Unsubscribe(ctx, rdbChannel)
	return
}

func (nm *NonceManager) publishIncreasingEvent(rdbChannel string, message string) {
	err := nm.rdb.Publish(ctx, rdbChannel, message)

	if err != nil {

	}

	return
}
