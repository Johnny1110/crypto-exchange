package security

import (
	"errors"
	"github.com/johnny1110/crypto-exchange/dto"
)

type CredentialCache struct {
	// token: userId
	tokenCache map[string]*dto.User
}

func NewCredentialCache() *CredentialCache {
	return &CredentialCache{
		tokenCache: make(map[string]*dto.User),
	}
}

func (c *CredentialCache) Put(token string, user *dto.User) {
	c.tokenCache[token] = user
}

func (c *CredentialCache) Delete(token string) {
	delete(c.tokenCache, token)
}

func (c *CredentialCache) Get(token string) (*dto.User, error) {
	user, ok := c.tokenCache[token]
	if !ok {
		return nil, errors.New("invalid credential")
	}
	return user, nil
}
