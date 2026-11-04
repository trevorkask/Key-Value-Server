package kvsrv

import (
	"fmt"
	"sync"
	"testing"

	"6.5840/kvsrv1/rpc"
)

func TestServerVersions(t *testing.T) {
	kv := MakeKVServer()
	var missing rpc.GetReply
	kv.Get(&rpc.GetArgs{Key: "k"}, &missing)
	if missing.Err != rpc.ErrNoKey {
		t.Fatalf("missing Get: got %v", missing.Err)
	}

	steps := []struct {
		value   string
		version rpc.Tversion
		err     rpc.Err
		stored  string
		current rpc.Tversion
	}{
		{"bad", 1, rpc.ErrNoKey, "", 0},
		{"first", 0, rpc.OK, "first", 1},
		{"duplicate", 0, rpc.ErrVersion, "first", 1},
		{"future", 2, rpc.ErrVersion, "first", 1},
		{"second", 1, rpc.OK, "second", 2},
		{"", 2, rpc.OK, "", 3},
	}
	for i, step := range steps {
		var put rpc.PutReply
		kv.Put(&rpc.PutArgs{Key: "k", Value: step.value, Version: step.version}, &put)
		if put.Err != step.err {
			t.Fatalf("step %d: Put got %v, want %v", i, put.Err, step.err)
		}
		var get rpc.GetReply
		kv.Get(&rpc.GetArgs{Key: "k"}, &get)
		wantErr := rpc.Err(rpc.OK)
		if step.current == 0 {
			wantErr = rpc.ErrNoKey
		}
		if get.Err != wantErr || get.Value != step.stored || get.Version != step.current {
			t.Fatalf("step %d: Get got %+v", i, get)
		}
	}

	var put rpc.PutReply
	kv.Put(&rpc.PutArgs{Key: "other", Value: "independent"}, &put)
	var get rpc.GetReply
	kv.Get(&rpc.GetArgs{Key: "k"}, &get)
	if put.Err != rpc.OK || get.Err != rpc.OK || get.Value != "" || get.Version != 3 {
		t.Fatalf("keys interfered: Put %+v, Get %+v", put, get)
	}
}

func TestServerConcurrentPut(t *testing.T) {
	kv := MakeKVServer()
	for version := rpc.Tversion(0); version < 10; version++ {
		start := make(chan struct{})
		winners := make(chan string, 32)
		var wg sync.WaitGroup
		for i := 0; i < 32; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				<-start
				value := fmt.Sprintf("%d/%d", version, i)
				var put rpc.PutReply
				kv.Put(&rpc.PutArgs{Key: "k", Value: value, Version: version}, &put)
				if put.Err == rpc.OK {
					winners <- value
				} else if put.Err != rpc.ErrVersion {
					t.Errorf("concurrent Put: got %v", put.Err)
				}
				var get rpc.GetReply
				kv.Get(&rpc.GetArgs{Key: "k"}, &get)
				if get.Err != rpc.OK || get.Version != version+1 {
					t.Errorf("concurrent Get: got %+v", get)
				}
			}(i)
		}
		close(start)
		wg.Wait()
		close(winners)
		if len(winners) != 1 {
			t.Fatalf("version %d: got %d successful writes, want 1", version, len(winners))
		}
		var get rpc.GetReply
		kv.Get(&rpc.GetArgs{Key: "k"}, &get)
		if winner := <-winners; get.Value != winner || get.Version != version+1 {
			t.Fatalf("version %d: Get got %+v, winner %q", version, get, winner)
		}
	}
}
