package atproto

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	comatproto "github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/atproto/identity"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/bluesky-social/indigo/xrpc"

	"github.com/adrg/xdg"
)

var ErrNoAuthSession = errors.New("no auth session found")

type AuthSession struct {
	DID          syntax.DID `json:"did"`
	Password     string     `json:"password"`
	RefreshToken string     `json:"session_token"`
	PDS          string     `json:"pds"`
}

func CreateAccount(ctx context.Context, username, password, pdsURL string) error {
	return nil
}

func GetSessionByOauth(ctx context.Context, pdsURL string, accessJwt string, refreshJwt string) (*comatproto.ServerGetSession_Output, error) {
	return nil, nil
}

func GetSession(ctx context.Context, pdsURL string, accessJwt string, refreshJwt string) (*comatproto.ServerGetSession_Output, error) {
	atp := &xrpc.Client{
		Host: pdsURL,
		Client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
		Auth: &xrpc.AuthInfo{
			AccessJwt:  "eyJ0eXAiOiJhdCtqd3QiLCJhbGciOiJIUzI1NiJ9.eyJzY29wZSI6ImNvbS5hdHByb3RvLmFjY2VzcyIsImF1ZCI6ImRpZDp3ZWI6cGRzLmF2YXRhci5haSIsInN1YiI6ImRpZDpwbGM6bW9wN2FpcXgzZGd4Y2lvdnpteDdvNnhlIiwiaWF0IjoxNzQ0ODgwOTk4LCJleHAiOjE3NDQ4ODgxOTh9.pataea9eNFxBnYEggGvTns7xE41o-6SBu89Trzgr-TI",
			RefreshJwt: "eyJ0eXAiOiJyZWZyZXNoK2p3dCIsImFsZyI6IkhTMjU2In0.eyJzY29wZSI6ImNvbS5hdHByb3RvLnJlZnJlc2giLCJhdWQiOiJkaWQ6d2ViOnBkcy5hdmF0YXIuYWkiLCJzdWIiOiJkaWQ6cGxjOm1vcDdhaXF4M2RneGNpb3Z6bXg3bzZ4ZSIsImp0aSI6Ik1ON2lHQ25Tc0I2VUY5bG9wUWtmcjJhamwrR2RiL3N2Rlg5NWx1Mm1kdTQiLCJpYXQiOjE3NDQ4ODA5OTgsImV4cCI6MTc1MjY1Njk5OH0.mFNuHQgvsoVQIVEQrmKr_p641Xpcgl-jmB4Ot1YU6QI",
		},
	}

	resp, err := comatproto.ServerGetSession(ctx, atp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func persistAuthSession(sess *AuthSession) error {

	fPath, err := xdg.StateFile("goat/auth-session.json")
	if err != nil {
		return err
	}

	f, err := os.OpenFile(fPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	authBytes, err := json.MarshalIndent(sess, "", "  ")
	if err != nil {
		return err
	}
	_, err = f.Write(authBytes)
	return err
}

func loadAuthClient(ctx context.Context) (*xrpc.Client, error) {
	fPath, err := xdg.SearchStateFile("goat/auth-session.json")
	if err != nil {
		return nil, ErrNoAuthSession
	}

	fBytes, err := ioutil.ReadFile(fPath)
	if err != nil {
		return nil, err
	}

	var sess AuthSession
	err = json.Unmarshal(fBytes, &sess)
	if err != nil {
		return nil, err
	}

	client := xrpc.Client{
		Host: sess.PDS,
		Auth: &xrpc.AuthInfo{
			Did: sess.DID.String(),
			// NOTE: using refresh in access location for "refreshSession" call
			AccessJwt:  sess.RefreshToken,
			RefreshJwt: sess.RefreshToken,
		},
	}
	resp, err := comatproto.ServerRefreshSession(ctx, &client)
	if err != nil {
		// TODO: if failure, try creating a new session from password (2fa tokens are only valid once, so not reused)
		fmt.Println("trying to refresh auth from password...")
		as, err := refreshAuthSession(ctx, sess.DID.AtIdentifier(), sess.Password, sess.PDS, "")
		if err != nil {
			return nil, err
		}
		client.Auth.AccessJwt = as.RefreshToken
		client.Auth.RefreshJwt = as.RefreshToken
		resp, err = comatproto.ServerRefreshSession(ctx, &client)
		if err != nil {
			return nil, err
		}
	}
	client.Auth.AccessJwt = resp.AccessJwt
	client.Auth.RefreshJwt = resp.RefreshJwt

	return &client, nil
}

func refreshAuthSession(ctx context.Context, username syntax.AtIdentifier, password, pdsURL, authFactorToken string) (*AuthSession, error) {
	var did syntax.DID
	if pdsURL == "" {
		dir := identity.DefaultDirectory()
		ident, err := dir.Lookup(ctx, username)
		if err != nil {
			return nil, err
		}

		pdsURL = ident.PDSEndpoint()
		if pdsURL == "" {
			return nil, fmt.Errorf("empty PDS URL")
		}
		did = ident.DID
	}

	if did == "" && username.IsDID() {
		did, _ = username.AsDID()
	}

	client := xrpc.Client{
		Host: pdsURL,
	}
	var token *string
	if authFactorToken != "" {
		token = &authFactorToken
	}
	sess, err := comatproto.ServerCreateSession(ctx, &client, &comatproto.ServerCreateSession_Input{
		Identifier:      username.String(),
		Password:        password,
		AuthFactorToken: token,
	})
	if err != nil {
		return nil, err
	}

	// TODO: check account status?
	// TODO: warn if email isn't verified?
	// TODO: check that sess.Did matches username
	if did == "" {
		did, err = syntax.ParseDID(sess.Did)
		if err != nil {
			return nil, err
		}
	} else if sess.Did != did.String() {
		return nil, fmt.Errorf("session DID didn't match expected: %s != %s", sess.Did, did)
	}

	authSession := AuthSession{
		DID:          did,
		Password:     password,
		PDS:          pdsURL,
		RefreshToken: sess.RefreshJwt,
	}
	if err = persistAuthSession(&authSession); err != nil {
		return nil, err
	}
	return &authSession, nil
}
