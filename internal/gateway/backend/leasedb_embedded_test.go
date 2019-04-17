package backend

import (
	"bytes"
	"io/ioutil"
	"testing"
	"time"
)

const (
	maxLeaseTime time.Duration = 100 * time.Second
)

func TestEmbeddedLeaseDBOpen(t *testing.T) {
	tmp, err := ioutil.TempDir("", "testleasedb")
	if err != nil {
		t.Fatalf("could not create temp dir for test case")
	}

	db, err := NewEmbeddedLeaseDB(tmp, maxLeaseTime)
	if err != nil {
		t.Fatalf("could not create database: %v", err)
	}
	defer db.Close()
}

func TestEmbeddedLeaseDBCRUD(t *testing.T) {
	tmp, err := ioutil.TempDir("", "testleasedb")
	if err != nil {
		t.Fatalf("could not create temp dir for test case")
	}

	db, err := NewEmbeddedLeaseDB(tmp, maxLeaseTime)
	if err != nil {
		t.Fatalf("could not create database: %v", err)
	}
	defer db.Close()

	keyID1 := "key1"
	leasePath1 := "test.repo.org/path/one"
	token1, err := NewLeaseToken(leasePath1, maxLeaseTime)
	t.Run("new lease", func(t *testing.T) {
		if err != nil {
			t.Fatalf("could not generate session token: %v", err)
		}

		if err := db.NewLease(keyID1, leasePath1, *token1); err != nil {
			t.Fatalf("could not add new lease: %v", err)
		}
	})
	t.Run("get leases", func(t *testing.T) {
		leases, err := db.GetLeases()
		if err != nil {
			t.Fatalf("could not retrieve leases: %v", err)
		}
		if len(leases) != 1 {
			t.Fatalf("expected 1 lease")
		}
		_, present := leases[leasePath1]
		if !present {
			t.Fatalf("missing lease for %v", leasePath1)
		}
	})
	t.Run("get lease for path", func(t *testing.T) {
		lease, err := db.GetLeaseForPath(leasePath1)
		if err != nil {
			t.Fatalf("could not retrieve leases: %v", err)
		}
		if lease.KeyID != keyID1 ||
			lease.Token.TokenStr != token1.TokenStr ||
			!bytes.Equal(lease.Token.Secret, token1.Secret) {
			t.Fatalf("invalid lease returned: %v", lease)
		}
	})
	t.Run("get lease for token", func(t *testing.T) {
		_, lease, err := db.GetLeaseForToken(token1.TokenStr)
		if err != nil {
			t.Fatalf("could not retrieve leases: %v", err)
		}
		if lease.KeyID != keyID1 ||
			lease.Token.TokenStr != token1.TokenStr ||
			!bytes.Equal(lease.Token.Secret, token1.Secret) {
			t.Fatalf("invalid lease returned: %v", lease)
		}
	})
	t.Run("cancel leases", func(t *testing.T) {
		err := db.CancelLeases()
		if err != nil {
			t.Fatalf("could not cancel all leases")
		}
		leases, err := db.GetLeases()
		if err != nil {
			t.Fatalf("could not retrieve leases: %v", err)
		}
		if len(leases) > 0 {
			t.Fatalf("remaining leases after cancellation")
		}
	})
	t.Run("clear lease for path", func(t *testing.T) {
		leasePath := "test.repo.org/path/two"
		token, err := NewLeaseToken(leasePath, maxLeaseTime)
		if err != nil {
			t.Fatalf("could not generate session token: %v", err)
		}

		if err := db.NewLease(keyID1, leasePath, *token); err != nil {
			t.Fatalf("could not add new lease: %v", err)
		}

		if err := db.CancelLeaseForPath(leasePath); err != nil {
			t.Fatalf("could not clear lease for path")
		}

		leases, err := db.GetLeases()
		if err != nil {
			t.Fatalf("could not retrieve leases: %v", err)
		}
		if len(leases) > 0 {
			t.Fatalf("remaining leases after cancellation")
		}
	})
	t.Run("clear lease for token", func(t *testing.T) {
		leasePath := "test.repo.org/path/three"
		token, err := NewLeaseToken(leasePath, maxLeaseTime)
		if err != nil {
			t.Fatalf("could not generate session token: %v", err)
		}

		if err := db.NewLease(keyID1, leasePath, *token); err != nil {
			t.Fatalf("could not add new lease: %v", err)
		}

		if err := db.CancelLeaseForToken(token.TokenStr); err != nil {
			t.Fatalf("could not clear lease for token")
		}

		leases, err := db.GetLeases()
		if err != nil {
			t.Fatalf("could not retrieve leases: %v", err)
		}
		if len(leases) > 0 {
			t.Fatalf("remaining leases after cancellation")
		}
	})
}

func TestEmbeddedLeaseDBConflicts(t *testing.T) {
	tmp, err := ioutil.TempDir("", "testleasedb")
	if err != nil {
		t.Fatalf("could not create temp dir for test case")
	}

	db, err := NewEmbeddedLeaseDB(tmp, maxLeaseTime)
	if err != nil {
		t.Fatalf("could not create database: %v", err)
	}
	defer db.Close()

	keyID := "key1"
	leasePath1 := "test.repo.org/path/one"
	token1, err := NewLeaseToken(leasePath1, maxLeaseTime)
	if err != nil {
		t.Fatalf("could not generate session token: %v", err)
	}

	if err := db.NewLease(keyID, leasePath1, *token1); err != nil {
		t.Fatalf("could not add new lease: %v", err)
	}

	leasePath2 := "test.repo.org/path"
	token2, err := NewLeaseToken(leasePath2, maxLeaseTime)
	if err != nil {
		t.Fatalf("could not generate session token: %v", err)
	}

	err = db.NewLease(keyID, leasePath2, *token2)
	if _, ok := err.(PathBusyError); !ok {
		t.Fatalf("conflicting lease was added for path: %v", leasePath2)
	}

	leasePath3 := "test.repo.org/path/one/below"
	token3, err := NewLeaseToken(leasePath3, maxLeaseTime)
	if err != nil {
		t.Fatalf("could not generate session token: %v", err)
	}

	err = db.NewLease(keyID, leasePath3, *token3)
	if _, ok := err.(PathBusyError); !ok {
		t.Fatalf("conflicting lease was added for path: %v", leasePath3)
	}
}

func TestEmbeddedLeaseDBExpired(t *testing.T) {
	tmp, err := ioutil.TempDir("", "testleasedb")
	if err != nil {
		t.Fatalf("could not create temp dir for test case")
	}

	shortLeaseTime := 1 * time.Millisecond

	db, err := NewEmbeddedLeaseDB(tmp, shortLeaseTime)
	if err != nil {
		t.Fatalf("could not create database: %v", err)
	}
	defer db.Close()

	keyID := "key1"
	leasePath := "test.repo.org/path/one"
	token1, err := NewLeaseToken(leasePath, shortLeaseTime)
	if err != nil {
		t.Fatalf("could not generate session token: %v", err)
	}

	if err := db.NewLease(keyID, leasePath, *token1); err != nil {
		t.Fatalf("could not add new lease: %v", err)
	}

	time.Sleep(2 * shortLeaseTime)

	token2, err := NewLeaseToken(leasePath, shortLeaseTime)
	if err != nil {
		t.Fatalf("could not generate session token: %v", err)
	}

	if err := db.NewLease(keyID, leasePath, *token2); err != nil {
		t.Fatalf("could not add new lease in place of expired one")
	}
}