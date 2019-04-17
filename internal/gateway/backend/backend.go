package backend

import (
	"time"

	gw "github.com/cvmfs/gateway/internal/gateway"
	"github.com/pkg/errors"
)

// LeaseReturn is the response type of lease queries, handed
// back to the HTTP frontend
type LeaseReturn struct {
	KeyID     string `json:"key_id,omitempty"`
	LeasePath string `json:"path,omitempty"`
	TokenStr  string `json:"token,omitempty"`
	Expires   string `json:"expires,omitempty"`
}

// Services is a container for the various
// backend services
type Services struct {
	Access AccessConfig
	Leases LeaseDB
	Config gw.Config
}

// Start initializes the various backend services
func Start(cfg *gw.Config) (*Services, error) {
	ac, err := NewAccessConfig(cfg.AccessConfigFile)
	if err != nil {
		return nil, errors.Wrap(
			err, "loading repository access configuration failed")
	}

	leaseDBType := "embedded"
	if cfg.UseEtcd {
		leaseDBType = "etcd"
	}
	ldb, err := NewLeaseDB(leaseDBType, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "could not create lease DB")
	}

	return &Services{Access: *ac, Leases: ldb, Config: *cfg}, nil
}

// Close all the backend services
func (s *Services) Close() error {
	if err := s.Leases.Close(); err != nil {
		return errors.Wrap(err, "could not close lease database")
	}
	return nil
}

// NewLease for the specified path, using keyID
func (s *Services) NewLease(keyID, leasePath string) (string, error) {
	repoName, subPath, err := SplitLeasePath(leasePath)
	if err != nil {
		return "", errors.Wrap(err, "could not parse lease path")
	}

	// Check if keyID is allowed to request a lease in the repository
	// at the specified subpath
	if err := s.Access.Check(keyID, subPath, repoName); err != nil {
		return "", err
	}

	// Generate a new token for the lease
	token, err := NewLeaseToken(
		leasePath, time.Duration(s.Config.MaxLeaseTime)*time.Second)
	if err != nil {
		return "", errors.Wrap(err, "could not generate session token")
	}

	if err := s.Leases.NewLease(keyID, leasePath, *token); err != nil {
		return "", err
	}

	return token.TokenStr, nil
}

// GetLeases returns all active and valid leases
func (s *Services) GetLeases() (map[string]LeaseReturn, error) {
	leases, err := s.Leases.GetLeases()
	if err != nil {
		return nil, err
	}
	ret := make(map[string]LeaseReturn)
	for k, v := range leases {
		if _, err := CheckToken(v.Token.TokenStr, v.Token.Secret); err == nil {
			ret[k] = LeaseReturn{KeyID: v.KeyID, TokenStr: v.Token.TokenStr, Expires: v.Token.Expiration.String()}
		}
	}
	return ret, nil
}

// GetLease returns the lease associated with a token
func (s *Services) GetLease(tokenStr string) (*LeaseReturn, error) {
	leasePath, lease, err := s.Leases.GetLeaseForToken(tokenStr)
	if err != nil {
		return nil, err
	}

	if _, err := CheckToken(tokenStr, lease.Token.Secret); err != nil {
		return nil, err
	}

	ret := &LeaseReturn{
		KeyID:     lease.KeyID,
		LeasePath: leasePath,
		Expires:   lease.Token.Expiration.String(),
	}
	return ret, nil
}

// CancelLease associated with the token
func (s *Services) CancelLease(tokenStr string) error {
	_, lease, err := s.Leases.GetLeaseForToken(tokenStr)
	if err != nil {
		return err
	}

	if _, err := CheckToken(tokenStr, lease.Token.Secret); err != nil {
		return err
	}

	if err := s.Leases.CancelLeaseForToken(tokenStr); err != nil {
		return err
	}

	return nil
}