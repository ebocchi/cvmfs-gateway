package frontend

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	gw "github.com/cvmfs/gateway/internal/gateway"
	be "github.com/cvmfs/gateway/internal/gateway/backend"
)

func forwardBody(w http.ResponseWriter, req *http.Request) {
	buf, _ := ioutil.ReadAll(req.Body)
	w.Write(buf)
}

type mockBackend struct {
}

func (b *mockBackend) GetSecret(keyID string) string {
	return "big_secret"
}

func (b *mockBackend) GetRepo(repoName string) be.KeyPaths {
	return be.KeyPaths{"keyid1": "/", "keyid2": "/restricted/to/subdir"}
}

func (b *mockBackend) GetRepos() map[string]be.KeyPaths {
	return map[string]be.KeyPaths{
		"test1.repo.org": be.KeyPaths{"keyid123": "/"},
		"test2.repo.org": be.KeyPaths{"keyid1": "/", "keyid2": "/restricted/to/subdir"},
	}
}

func (b *mockBackend) NewLease(ctx context.Context, keyID, leasePath string) (string, error) {
	return "lease_token_string", nil
}

func (b *mockBackend) GetLeases(ctx context.Context) (map[string]be.LeaseReturn, error) {
	return map[string]be.LeaseReturn{
		"test2.repo.org/some/path/one": be.LeaseReturn{
			KeyID:   "keyid1",
			Expires: time.Now().Add(60 * time.Second).String(),
		},
		"test2.repo.org/some/path/two": be.LeaseReturn{
			KeyID:   "keyid1",
			Expires: time.Now().Add(120 * time.Second).String(),
		},
	}, nil
}

func (b *mockBackend) GetLease(ctx context.Context, tokenStr string) (*be.LeaseReturn, error) {
	return &be.LeaseReturn{
		KeyID:     "keyid1",
		LeasePath: "test2.repo.org/some/path/one",
		Expires:   time.Now().Add(60 * time.Second).String(),
	}, nil
}

func (b *mockBackend) CancelLease(ctx context.Context, tokenStr string) error {
	return nil
}

func (b *mockBackend) CommitLease(ctx context.Context, tokenStr, oldRootHash, newRootHash string, tag gw.RepositoryTag) error {
	return nil
}

func (b *mockBackend) SubmitPayload(ctx context.Context, token string, payload io.Reader, digest string, headerSize int) error {
	return nil
}