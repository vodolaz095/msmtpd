package karma

import (
	"context"
	"net/smtp"
	"sync"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/vodolaz095/msmtpd"
	redisStorage "github.com/vodolaz095/msmtpd/plugins/karma/storage/redis"
)

func TestKarmaPluginRedisBad(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	client := redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    "127.0.0.1:6379",
	})

	rs := redisStorage.Storage{Client: client}
	kh := Handler{
		HateLimit: 5,
		Storage:   &rs,
	}

	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		ConnectionCheckers: []msmtpd.ConnectionChecker{
			kh.ConnectionChecker,
		},
		CloseHandlers: []msmtpd.CloseHandler{
			kh.CloseHandler,
			func(transaction *msmtpd.Transaction) error {
				wg.Done()
				return nil
			},
		},
	})
	defer closer()
	_, err := smtp.Dial(addr)
	if err != nil {
		if err.Error() != "521 FUCK OFF!" {
			t.Errorf("%s : wrong error while performing dial", err)
		}
	}

	wg.Wait()
	var score redisStorage.Score
	err = client.HMGet(context.TODO(), "karma|127.0.0.1", "connections", "good", "bad").Scan(&score)
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
	err = client.Del(context.TODO(), "karma|127.0.0.1").Err()
	if err != nil {
		t.Errorf("%s : while deleting test key", err)
	}
	err = rs.Close()
	if err != nil {
		t.Errorf("%s : while closing redis connection", err)
	}
}

func TestKarmaPluginRedisGood(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	client := redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    "127.0.0.1:6379",
	})

	rs := redisStorage.Storage{Client: client}
	kh := Handler{
		HateLimit: -2,
		Storage:   &rs,
	}

	addr, closer := msmtpd.RunTestServerWithoutTLS(t, &msmtpd.Server{
		ConnectionCheckers: []msmtpd.ConnectionChecker{
			kh.ConnectionChecker,
		},
		CloseHandlers: []msmtpd.CloseHandler{
			kh.CloseHandler,
			func(transaction *msmtpd.Transaction) error {
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
	err = c.Quit()
	if err != nil {
		t.Errorf("%s : while performing helo", err)
	}
	wg.Wait()
	var score redisStorage.Score
	err = client.HMGet(context.TODO(), "karma|127.0.0.1", "connections", "good", "bad").Scan(&score)
	if err != nil {
		t.Errorf("%s : while getting karma", err)
	}
	t.Logf("Score: %v connections, %v good and %v bad", score.Connections, score.Good, score.Bad)
	if score.Connections != 1 {
		t.Errorf("wrong connections %v isntead 1", score.Connections)
	}
	if score.Good != 1 {
		t.Errorf("wrong good connecetions %v isntead of 1", score.Good)
	}
	if score.Bad != 0 {
		t.Errorf("wrong bad connecetions %v isntead of 0", score.Bad)
	}
	err = client.Del(context.TODO(), "karma|127.0.0.1").Err()
	if err != nil {
		t.Errorf("%s : while deleting test key", err)
	}
	err = rs.Close()
	if err != nil {
		t.Errorf("%s : while closing redis connection", err)
	}
}
