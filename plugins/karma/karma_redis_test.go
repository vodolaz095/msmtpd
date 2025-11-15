package karma

import (
	"context"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"sync"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/vodolaz095/msmtpd"
	"github.com/vodolaz095/msmtpd/internal"
	redisStorage "github.com/vodolaz095/msmtpd/plugins/karma/storage/redis"
)

func TestKarmaPluginRedisBad(t *testing.T) {
	var err error

	testRedisURL := os.Getenv("REDIS_URL")
	if testRedisURL == "" {
		t.Skipf("set redis connection string as REDIS_URL environmen variable")
	}
	t.Logf("Dialing redis via %s", testRedisURL)
	opts, err := redis.ParseURL(testRedisURL)
	if err != nil {
		t.Errorf("%s : while parsing redis url %s", err, testRedisURL)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	client := redis.NewClient(opts)
	err = client.Del(context.TODO(), "karma|8.8.8.8").Err()
	if err != nil {
		t.Errorf("%s : while deleting test key", err)
	}
	rs := redisStorage.Storage{Client: client}
	kh := Handler{
		InitialHate: 0,
		HateLimit:   5,
		KarmaLimit:  0,
		Storage:     &rs,
	}

	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		ConnectionCheckers: []msmtpd.ConnectionChecker{
			func(_ context.Context, tr *msmtpd.Transaction) error {
				tr.Addr = &net.TCPAddr{
					IP:   net.ParseIP("8.8.8.8"),
					Port: 60123,
				}
				return nil
			},
			kh.ConnectionChecker,
		},
		CloseHandlers: []msmtpd.CloseHandler{
			kh.CloseHandler,
			func(_ context.Context, transaction *msmtpd.Transaction) error {
				wg.Done()
				return nil
			},
		},
	})
	defer closer()
	_, err = smtp.Dial(addr)
	if err != nil {
		if err.Error() != "521 FUCK OFF!" {
			t.Errorf("%s : wrong error while performing dial", err)
		}
	}

	wg.Wait()
	var score redisStorage.Score
	err = client.HMGet(context.TODO(), "karma|8.8.8.8", "connections", "good", "bad").Scan(&score)
	if err != nil {
		t.Errorf("%s : while getting karma", err)
	}
	t.Logf("Score: %v connections, %v good and %v bad", score.Connections, score.Good, score.Bad)
	if score.Connections != 1 {
		t.Errorf("wrong connections %v isntead 1", score.Connections)
	}
	if score.Good != 0 {
		t.Errorf("wrong good connecetions %v isntead of 0", score.Good)
	}
	if score.Bad != 1 {
		t.Errorf("wrong bad connecetions %v isntead of 1", score.Bad)
	}
	err = rs.Close()
	if err != nil {
		t.Errorf("%s : while closing redis connection", err)
	}
}

func TestKarmaPluginRedisGood(t *testing.T) {
	var err error

	testRedisURL := os.Getenv("REDIS_URL")
	if testRedisURL == "" {
		t.Skipf("set redis connection string as REDIS_URL environmen variable")
	}
	t.Logf("Dialing redis via %s", testRedisURL)
	opts, err := redis.ParseURL(testRedisURL)
	if err != nil {
		t.Errorf("%s : while parsing redis url %s", err, testRedisURL)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	client := redis.NewClient(opts)
	err = client.Del(context.TODO(), "karma|1.1.1.1").Err()
	if err != nil {
		t.Errorf("%s : while deleting test key", err)
	}
	rs := redisStorage.Storage{Client: client}
	kh := Handler{
		InitialHate: DefaultInitialHate,
		HateLimit:   DefaultHateLimit,
		KarmaLimit:  DefaultKarmaLimit,
		Storage:     &rs,
	}

	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		ConnectionCheckers: []msmtpd.ConnectionChecker{
			func(_ context.Context, tr *msmtpd.Transaction) error {
				tr.Addr = &net.TCPAddr{
					IP:   net.ParseIP("1.1.1.1"),
					Port: 60123,
				}
				return nil
			},
			kh.ConnectionChecker,
		},
		CloseHandlers: []msmtpd.CloseHandler{
			kh.CloseHandler,
			func(_ context.Context, transaction *msmtpd.Transaction) error {
				wg.Done()
				return nil
			},
		},
	})
	defer closer()
	c, err := smtp.Dial(addr)
	if err != nil {
		if err.Error() != "521 FUCK OFF!" {
			t.Errorf("%s : wrong error while performing dial", err)
		}
	}
	err = c.Hello("localhost")
	if err != nil {
		t.Errorf("%s : while performing helo", err)
	}
	err = c.Mail("sender@example.org")
	if err != nil {
		t.Errorf("%s : while performing helo", err)
	}
	if err = c.Rcpt("recipient@example.net"); err != nil {
		t.Errorf("Rcpt failed: %v", err)
	}
	if err = c.Rcpt("recipient2@example.net"); err != nil {
		t.Errorf("Rcpt2 failed: %v", err)
	}
	wc, err := c.Data()
	if err != nil {
		t.Errorf("Data failed: %v", err)
	}
	_, err = fmt.Fprint(wc, internal.MakeTestMessage("sender@example.org", "recipient@example.net", "recipient2@example.net"))
	if err != nil {
		t.Errorf("Data body failed: %v", err)
	}
	err = wc.Close()
	if err != nil {
		t.Errorf("Data close failed: %v", err)
	}
	err = c.Quit()
	if err != nil {
		t.Errorf("%s : while performing helo", err)
	}
	wg.Wait()
	var score redisStorage.Score
	err = client.HMGet(context.TODO(), "karma|1.1.1.1", "connections", "good", "bad").Scan(&score)
	if err != nil {
		t.Errorf("%s : while getting karma", err)
	}
	t.Logf("Score: %v connections, %v good and %v bad", score.Connections, score.Good, score.Bad)
	if score.Connections != 1 {
		t.Errorf("wrong connections %v isntead 1", score.Connections)
	}
	if score.Good != 1 {
		t.Errorf("wrong good connecetions %v instead of 1", score.Good)
	}
	if score.Bad != 0 {
		t.Errorf("wrong bad connecetions %v instead of 0", score.Bad)
	}
	err = rs.Close()
	if err != nil {
		t.Errorf("%s : while closing redis connection", err)
	}
}
