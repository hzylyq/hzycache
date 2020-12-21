package hzycache

import (
	"fmt"
	"log"
	"testing"

	"hzycache/hzycachepb"

	"google.golang.org/protobuf/proto"
)

func TestHttpPool_ServerHttp(t *testing.T) {
	var db = map[string]string{
		"Tom":  "630",
		"Jack": "589",
		"Sam":  "567",
	}

	NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func TestHttpPool_Log(t *testing.T) {
	body, err := proto.Marshal(&hzycachepb.Response{Value: []byte("aaaaa")})
	if err != nil {
		t.Error(err)
	}

	t.Log(body)
}
