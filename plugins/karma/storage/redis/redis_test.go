package redis

import (
	"context"
	"net"
	"os"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/vodolaz095/msmtpd"
)

var testRedisUrl string

func TestRedisEnvironmentIsSet(t *testing.T) {
	testRedisUrl = os.Getenv("REDIS_URL")
	if testRedisUrl == "" {
		t.Skipf("set redis connection string as REDIS_URL environmen variable")
	}
	t.Logf("Dialing redis via %s", testRedisUrl)
}

func TestStorage(t *testing.T) {
	var score int
	if testRedisUrl == "" {
		t.Skipf("set redis connection string as REDIS_URL environmen variable")
	}
	opts, err := redis.ParseURL(testRedisUrl)
	if err != nil {
		t.Errorf("%s : while parsing redis url %s", err, testRedisUrl)
	}
	client := redis.NewClient(opts)

	err = client.Ping(context.TODO()).Err()
	if err != nil {
		t.Errorf("%s : while pinging redis", err)
	}
	err = client.Del(context.TODO(), "karma|192.168.1.3").Err()
	if err != nil {
		t.Errorf("%s : while cleaning data from redis by client", err)
	}
	storage := Storage{Client: client}
	err = storage.Ping(context.TODO())
	if err != nil {
		t.Errorf("%s : while pinging storage", err)
	}
	tr := msmtpd.Transaction{
		Addr: &net.TCPAddr{IP: net.ParseIP("192.168.1.3"), Port: 25},
	}
	err = storage.SaveGood(&tr)
	if err != nil {
		t.Errorf("%s : while saving transaction as good", err)
	}
	err = storage.SaveGood(&tr)
	if err != nil {
		t.Errorf("%s : while saving transaction as good", err)
	}
	score, err = storage.Get(&tr)
	if err != nil {
		t.Errorf("%s : while geting transaction", err)
	}
	t.Logf("Score %v", score)
	if score != 2 {
		t.Errorf("wrong score %v instead of 2", score)
	}
	err = storage.SaveBad(&tr)
	if err != nil {
		t.Errorf("%s : while saving transaction as bad", err)
	}
	err = storage.SaveBad(&tr)
	if err != nil {
		t.Errorf("%s : while saving transaction as bad", err)
	}
	score, err = storage.Get(&tr)
	if err != nil {
		t.Errorf("%s : while geting transaction", err)
	}
	t.Logf("Score %v", score)
	if score != 0 {
		t.Errorf("wrong score %v instead of 0", score)
	}
	var raw Score
	err = client.
		HMGet(context.Background(), "karma|192.168.1.3", "connections", "good", "bad").
		Scan(&raw)
	if err != nil {
		t.Errorf("%s : while getting data from redis client", err)
	}
	t.Logf("Raw: %v", raw)
	if raw.Connections != 4 {
		t.Errorf("Wrong connections %v, 4 expected", raw.Connections)
	}
	if raw.Bad != 2 {
		t.Errorf("Wrong bad connections %v, 2 expected", raw.Bad)
	}
	if raw.Good != 2 {
		t.Errorf("Wrong good connections %v, 2 expected", raw.Good)
	}
	err = client.Del(context.TODO(), "karma|192.168.1.3").Err()
	if err != nil {
		t.Errorf("%s : while cleaning data from redis by client", err)
	}

	err = storage.Close()
	if err != nil {
		t.Errorf("%s : while closing storage", err)
	}
}
