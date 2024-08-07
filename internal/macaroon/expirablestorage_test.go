// Copyright 2024 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package macaroon_test

import (
	"context"
	"time"

	"github.com/go-macaroon-bakery/macaroon-bakery/v3/bakery"
	"github.com/go-macaroon-bakery/macaroon-bakery/v3/bakery/checkers"
	"github.com/go-macaroon-bakery/macaroon-bakery/v3/bakery/dbrootkeystore"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/macaroon.v2"

	domainmacaroon "github.com/juju/juju/domain/macaroon"
	macaroonservice "github.com/juju/juju/domain/macaroon/service"
	macaroonstate "github.com/juju/juju/domain/macaroon/state"
	domaintesting "github.com/juju/juju/internal/changestream/testing"
	internalmacaroon "github.com/juju/juju/internal/macaroon"
)

type expirableStorageSuite struct {
	domaintesting.ControllerSuite
	macaroonService *macaroonservice.Service
	clock           domainmacaroon.Clock
	now             time.Time
}

var _ = gc.Suite(&expirableStorageSuite{})

var moment = time.Now()

func (s *expirableStorageSuite) SetUpTest(c *gc.C) {
	s.ControllerSuite.SetUpTest(c)
	s.now = moment
	s.clock = clockVal(&s.now)
	s.macaroonService = macaroonservice.NewService(
		macaroonstate.NewState(s.TxnRunnerFactory()),
		s.clock,
	)
}

func (s *expirableStorageSuite) TestNewExpirableStorage(c *gc.C) {
	expireableStorage := internalmacaroon.NewExpirableStorage(s.macaroonService, time.Minute, s.clock)

	key1, id, err := expireableStorage.RootKey(context.Background())
	c.Assert(err, jc.ErrorIsNil)

	key2, err := expireableStorage.Get(context.Background(), id)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(key1, gc.DeepEquals, key2)
}

func (s *expirableStorageSuite) TestExpiredRootKeyRemovedGet(c *gc.C) {
	err := s.macaroonService.InsertKeyContext(context.Background(), dbrootkeystore.RootKey{
		Id:      []byte("id"),
		Created: s.now,
		Expires: s.now.Add(-time.Second),
		RootKey: []byte("key"),
	})
	c.Assert(err, jc.ErrorIsNil)

	expireableStorage := internalmacaroon.NewExpirableStorage(s.macaroonService, time.Minute, s.clock)
	_, err = expireableStorage.Get(context.Background(), []byte("id"))
	c.Assert(err, jc.ErrorIs, bakery.ErrNotFound)

	_, err = s.macaroonService.GetKeyContext(context.Background(), []byte("id"))
	c.Assert(err, jc.ErrorIs, bakery.ErrNotFound)
}

func (s *expirableStorageSuite) TestExpiredRootKeyRemovedRootKey(c *gc.C) {
	err := s.macaroonService.InsertKeyContext(context.Background(), dbrootkeystore.RootKey{
		Id:      []byte("id"),
		Created: s.now,
		Expires: s.now.Add(-time.Second),
		RootKey: []byte("key"),
	})
	c.Assert(err, jc.ErrorIsNil)

	expireableStorage := internalmacaroon.NewExpirableStorage(s.macaroonService, time.Minute, s.clock)
	_, _, err = expireableStorage.RootKey(context.Background())
	c.Assert(err, jc.ErrorIsNil)

	_, err = s.macaroonService.GetKeyContext(context.Background(), []byte("id"))
	c.Assert(err, jc.ErrorIs, bakery.ErrNotFound)
}

func (s *expirableStorageSuite) TestCheckNewMacaroon(c *gc.C) {
	expireableStorage := internalmacaroon.NewExpirableStorage(s.macaroonService, time.Minute, s.clock)

	b := bakery.New(bakery.BakeryParams{
		RootKeyStore: expireableStorage,
	})

	cav := []checkers.Caveat{{Condition: "something"}}
	mac, err := b.Oven.NewMacaroon(context.Background(), bakery.LatestVersion, cav, bakery.NoOp)
	c.Assert(err, jc.ErrorIsNil)

	op, conditions, err := b.Oven.VerifyMacaroon(context.Background(), macaroon.Slice{mac.M()})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(op, jc.DeepEquals, []bakery.Op{bakery.NoOp})
	c.Assert(conditions, jc.DeepEquals, []string{"something"})
}

func (s *expirableStorageSuite) TestExpiryTime(c *gc.C) {
	expireableStorage := internalmacaroon.NewExpirableStorage(s.macaroonService, 5*time.Millisecond, s.clock)

	b := bakery.New(bakery.BakeryParams{
		RootKeyStore: expireableStorage,
	})

	mac, err := b.Oven.NewMacaroon(context.Background(), bakery.LatestVersion, nil, bakery.NoOp)
	c.Assert(err, jc.ErrorIsNil)

	for i := 0; i < 15; i++ {
		_, _, err = b.Oven.VerifyMacaroon(context.Background(), macaroon.Slice{mac.M()})
		if err == nil {
			s.now = s.now.Add(500 * time.Millisecond)
			continue
		}
		c.Assert(err, gc.ErrorMatches, "verification failed: macaroon not found in storage")
		return
	}
	c.Fatal("timed out waiting for storage expiry")
}

var isValidWithPolicyTests = []struct {
	about  string
	policy dbrootkeystore.Policy
	now    time.Time
	key    dbrootkeystore.RootKey
	expect bool
}{{
	about: "success",
	policy: dbrootkeystore.Policy{
		GenerateInterval: 3 * time.Minute,
		ExpiryDuration:   3 * time.Minute,
	},
	now: moment.Add(20 * time.Minute),
	key: dbrootkeystore.RootKey{
		Created: moment.Add(19 * time.Minute),
		Expires: moment.Add(25 * time.Minute),
		Id:      []byte("id"),
		RootKey: []byte("key"),
	},
	expect: true,
}, {
	about: "empty root key",
	policy: dbrootkeystore.Policy{
		GenerateInterval: 3 * time.Minute,
		ExpiryDuration:   3 * time.Minute,
	},
	now:    moment.Add(20 * time.Minute),
	key:    dbrootkeystore.RootKey{},
	expect: false,
}, {
	about: "created too early",
	policy: dbrootkeystore.Policy{
		GenerateInterval: 3 * time.Minute,
		ExpiryDuration:   3 * time.Minute,
	},
	now: moment.Add(20 * time.Minute),
	key: dbrootkeystore.RootKey{
		Created: moment.Add(17*time.Minute - time.Millisecond),
		Expires: moment.Add(24 * time.Minute),
		Id:      []byte("id"),
		RootKey: []byte("key"),
	},
	expect: false,
}, {
	about: "expires too early",
	policy: dbrootkeystore.Policy{
		GenerateInterval: 3 * time.Minute,
		ExpiryDuration:   3 * time.Minute,
	},
	now: moment.Add(20 * time.Minute),
	key: dbrootkeystore.RootKey{
		Created: moment.Add(19 * time.Minute),
		Expires: moment.Add(21 * time.Minute),
		Id:      []byte("id"),
		RootKey: []byte("key"),
	},
	expect: false,
}, {
	about: "expires too late",
	policy: dbrootkeystore.Policy{
		GenerateInterval: 3 * time.Minute,
		ExpiryDuration:   3 * time.Minute,
	},
	now: moment.Add(20 * time.Minute),
	key: dbrootkeystore.RootKey{
		Created: moment.Add(19 * time.Minute),
		Expires: moment.Add(26*time.Minute + time.Millisecond),
		Id:      []byte("id"),
		RootKey: []byte("key"),
	},
	expect: false,
}}

func (s *expirableStorageSuite) TestIsValidWithPolicy(c *gc.C) {
	for i, test := range isValidWithPolicyTests {
		c.Logf("test %d: %v", i, test.about)
		c.Check(test.key.IsValidWithPolicy(test.policy, test.now), gc.Equals, test.expect)
	}
}

func (s *expirableStorageSuite) TestRootKeyUsesKeysValidWithPolicy(c *gc.C) {
	// We re-use the TestIsValidWithPolicy tests so that we
	// know that the mongo logic uses the same behaviour.
	for _, test := range isValidWithPolicyTests {
		s.SetUpTest(c)

		if test.key.RootKey == nil {
			// We don't store empty root keys in the database.
			c.Log("skipping test with empty root key")
			continue
		}
		s.now = test.now
		// Prime the collection with the root key document.
		err := s.macaroonService.InsertKeyContext(context.Background(), test.key)
		c.Assert(err, jc.ErrorIsNil, gc.Commentf(test.about))

		store := internalmacaroon.NewExpirableStorage(s.macaroonService, test.policy.ExpiryDuration, s.clock)
		key, id, err := store.RootKey(context.Background())
		c.Assert(err, jc.ErrorIsNil, gc.Commentf(test.about))
		if test.expect {
			c.Assert(string(id), gc.Equals, "id", gc.Commentf(test.about))
			c.Assert(string(key), gc.Equals, "key", gc.Commentf(test.about))
		} else {
			// If it didn't match then RootKey will have
			// generated a new key.
			c.Assert(key, gc.HasLen, 24, gc.Commentf(test.about))
			c.Assert(id, gc.HasLen, 32, gc.Commentf(test.about))
		}
		c.Assert(err, jc.ErrorIsNil, gc.Commentf(test.about))

		s.TearDownTest(c)
	}
}

func (s *expirableStorageSuite) TestRootKey(c *gc.C) {
	store := internalmacaroon.NewExpirableStorage(s.macaroonService, 5*time.Minute, s.clock)
	key, id, err := store.RootKey(context.Background())
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(key, gc.HasLen, 24)
	c.Assert(id, gc.HasLen, 32)

	// If we get a key within the generate interval, we should
	// get the same one.
	s.now = moment.Add(time.Minute)
	key1, id1, err := store.RootKey(context.Background())
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(key1, gc.DeepEquals, key)
	c.Assert(id1, gc.DeepEquals, id)

	// A different store instance should get the same root key.
	store1 := internalmacaroon.NewExpirableStorage(s.macaroonService, 5*time.Minute, s.clock)
	key1, id1, err = store1.RootKey(context.Background())
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(key1, gc.DeepEquals, key)
	c.Assert(id1, gc.DeepEquals, id)

	// After the generation interval has passed, we should generate a new key.
	s.now = moment.Add(5*time.Minute + time.Second)
	key1, id1, err = store.RootKey(context.Background())
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(key, gc.HasLen, 24)
	c.Assert(id, gc.HasLen, 32)
	c.Assert(key1, gc.Not(gc.DeepEquals), key)
	c.Assert(id1, gc.Not(gc.DeepEquals), id)

	// The other store should pick it up too.
	key2, id2, err := store1.RootKey(context.Background())
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(key2, gc.DeepEquals, key1)
	c.Assert(id2, gc.DeepEquals, id1)
}

var preferredRootKeyTests = []struct {
	about          string
	now            time.Time
	keys           []dbrootkeystore.RootKey
	expiryDuration time.Duration
	expectId       []byte
}{{
	about: "latest creation time is preferred",
	now:   moment.Add(5 * time.Minute),
	keys: []dbrootkeystore.RootKey{{
		Created: moment.Add(4 * time.Minute),
		Expires: moment.Add(15 * time.Minute),
		Id:      []byte("id0"),
		RootKey: []byte("key0"),
	}, {
		Created: moment.Add(5*time.Minute + 30*time.Second),
		Expires: moment.Add(16 * time.Minute),
		Id:      []byte("id1"),
		RootKey: []byte("key1"),
	}, {
		Created: moment.Add(5 * time.Minute),
		Expires: moment.Add(16 * time.Minute),
		Id:      []byte("id2"),
		RootKey: []byte("key2"),
	}},
	expiryDuration: 7 * time.Minute,
	expectId:       []byte("id1"),
}, {
	about: "ineligible keys are exluded",
	now:   moment.Add(5 * time.Minute),
	keys: []dbrootkeystore.RootKey{{
		Created: moment.Add(4 * time.Minute),
		Expires: moment.Add(15 * time.Minute),
		Id:      []byte("id0"),
		RootKey: []byte("key0"),
	}, {
		Created: moment.Add(5 * time.Minute),
		Expires: moment.Add(16*time.Minute + 30*time.Second),
		Id:      []byte("id1"),
		RootKey: []byte("key1"),
	}, {
		Created: moment.Add(6 * time.Minute),
		Expires: moment.Add(time.Hour),
		Id:      []byte("id2"),
		RootKey: []byte("key2"),
	}},
	expiryDuration: 7 * time.Minute,
	expectId:       []byte("id1"),
}}

func (s *expirableStorageSuite) TestPreferredRootKeyFromDatabase(c *gc.C) {
	for _, test := range preferredRootKeyTests {
		s.SetUpTest(c)

		for _, key := range test.keys {
			err := s.macaroonService.InsertKeyContext(context.Background(), key)
			c.Assert(err, jc.ErrorIsNil, gc.Commentf(test.about))
		}
		store := internalmacaroon.NewExpirableStorage(s.macaroonService, test.expiryDuration, s.clock)
		s.now = test.now
		_, id, err := store.RootKey(context.Background())
		c.Assert(err, jc.ErrorIsNil, gc.Commentf(test.about))
		c.Assert(id, gc.DeepEquals, test.expectId, gc.Commentf(test.about))

		s.TearDownTest(c)
	}
}

func (s *expirableStorageSuite) TestGet(c *gc.C) {
	store := internalmacaroon.NewExpirableStorage(s.macaroonService, 30*time.Minute, s.clock)
	type idKey struct {
		id  string
		key []byte
	}
	var keys []idKey
	keyIds := make(map[string]bool)
	for i := 0; i < 20; i++ {
		key, id, err := store.RootKey(context.Background())
		c.Assert(err, jc.ErrorIsNil)
		c.Assert(keyIds[string(id)], gc.Equals, false)
		keys = append(keys, idKey{string(id), key})
		s.now = s.now.Add(time.Minute + time.Second)
	}
	for i, k := range keys {
		key, err := store.Get(context.Background(), []byte(k.id))
		c.Assert(err, jc.ErrorIsNil, gc.Commentf("key %d (%s)", i, k.id))
		c.Assert(key, gc.DeepEquals, k.key, gc.Commentf("key %d (%s)", i, k.id))
	}
}

func clockVal(t *time.Time) dbrootkeystore.Clock {
	return clockFunc(func() time.Time {
		return *t
	})
}

type clockFunc func() time.Time

func (f clockFunc) Now() time.Time {
	return f()
}
