package atproto

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/bluesky-social/indigo/util"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/carlmjohnson/versioninfo"
	"github.com/go-jose/go-jose/v4"
	"github.com/zhongshangwu/avatarai-social/pkg/database"
	"gorm.io/gorm"
)

// This xrpc client is copied from the indigo xrpc client, with some tweaks:
// - There is no `AuthInfo` on the client. Instead, you pass auth _with the request_ in the `Do()` function
// - There is an `XrpcAuthedRequestArgs` struct that contains all the info you need to complete an authed request
// - There is a `OnDpopPdsNonceChanged` callback that will run when the dpop nonce receives an update. You can
//   use this to update a database, for example.
// - Requests are retried whenever the dpop nonce changes

type XrpcClient struct {
	// Client is an HTTP client to use. If not set, defaults to http.RobustHTTPClient().
	Client                *http.Client
	UserAgent             *string
	Headers               map[string]string
	OnDpopPdsNonceChanged func(did, newNonce string)
}

type XrpcAuthedRequestArgs struct {
	Did            string
	PdsUrl         string
	Issuer         string
	AccessToken    string
	DpopPdsNonce   string
	DpopPrivateJwk jose.JSONWebKey
}

func (c *XrpcClient) getClient() *http.Client {
	if c.Client == nil {
		return util.RobustHTTPClient()
	}
	return c.Client
}

func (c *XrpcClient) Do(ctx context.Context, authedArgs *XrpcAuthedRequestArgs, kind xrpc.XRPCRequestType, inpenc, method string, params map[string]any, bodyobj any, out any) error {
	// we might have to retry the request if we get a new nonce from the server
	for range 2 {
		var body io.Reader
		if bodyobj != nil {
			if rr, ok := bodyobj.(io.Reader); ok {
				body = rr
			} else {
				b, err := json.Marshal(bodyobj)
				if err != nil {
					return err
				}

				body = bytes.NewReader(b)
			}
		}

		var m string
		switch kind {
		case xrpc.Query:
			m = "GET"
		case xrpc.Procedure:
			m = "POST"
		default:
			return fmt.Errorf("unsupported request kind: %d", kind)
		}

		var paramStr string
		if len(params) > 0 {
			paramStr = "?" + makeParams(params)
		}

		ustr := authedArgs.PdsUrl + "/xrpc/" + method + paramStr
		req, err := http.NewRequest(m, ustr, body)
		if err != nil {
			return err
		}

		if bodyobj != nil && inpenc != "" {
			req.Header.Set("Content-Type", inpenc)
		}
		if c.UserAgent != nil {
			req.Header.Set("User-Agent", *c.UserAgent)
		} else {
			req.Header.Set("User-Agent", "atproto-oauth/"+versioninfo.Short())
		}

		if c.Headers != nil {
			for k, v := range c.Headers {
				req.Header.Set(k, v)
			}
		}

		if authedArgs != nil {
			dpopJwt, err := PDSDpopJWT(m, ustr, authedArgs.Issuer, authedArgs.AccessToken, authedArgs.DpopPdsNonce, authedArgs.DpopPrivateJwk)
			if err != nil {
				return err
			}

			req.Header.Set("DPoP", dpopJwt)
			req.Header.Set("Authorization", "DPoP "+authedArgs.AccessToken)
		}

		resp, err := c.getClient().Do(req.WithContext(ctx))
		if err != nil {
			return fmt.Errorf("request failed: %w", err)
		}

		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			var xe xrpc.XRPCError
			if err := json.NewDecoder(resp.Body).Decode(&xe); err != nil {
				return errorFromHTTPResponse(resp, fmt.Errorf("failed to decode xrpc error message: %w", err))
			}

			// if we get a new nonce, update the nonce and make the request again
			if (resp.StatusCode == 400 || resp.StatusCode == 401) && xe.ErrStr == "use_dpop_nonce" {
				authedArgs.DpopPdsNonce = resp.Header.Get("DPoP-Nonce")
				c.OnDpopPdsNonceChanged(authedArgs.Did, authedArgs.DpopPdsNonce)
				continue
			}

			return errorFromHTTPResponse(resp, &xe)
		}

		if out != nil {
			if buf, ok := out.(*bytes.Buffer); ok {
				if resp.ContentLength < 0 {
					_, err := io.Copy(buf, resp.Body)
					if err != nil {
						return fmt.Errorf("reading response body: %w", err)
					}
				} else {
					n, err := io.CopyN(buf, resp.Body, resp.ContentLength)
					if err != nil {
						return fmt.Errorf("reading length delimited response body (%d < %d): %w", n, resp.ContentLength, err)
					}
				}
			} else {
				if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
					return fmt.Errorf("decoding xrpc response: %w", err)
				}
			}
		}

		return nil
	}

	return nil
}

func GenerateCodeChallenge(pkceVerifier string) string {
	h := sha256.New()
	h.Write([]byte(pkceVerifier))
	hash := h.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(hash)
}

func errorFromHTTPResponse(resp *http.Response, err error) error {
	r := &xrpc.Error{
		StatusCode: resp.StatusCode,
		Wrapped:    err,
	}
	if resp.Header.Get("ratelimit-limit") != "" {
		r.Ratelimit = &xrpc.RatelimitInfo{
			Policy: resp.Header.Get("ratelimit-policy"),
		}
		if n, err := strconv.ParseInt(resp.Header.Get("ratelimit-reset"), 10, 64); err == nil {
			r.Ratelimit.Reset = time.Unix(n, 0)
		}
		if n, err := strconv.ParseInt(resp.Header.Get("ratelimit-limit"), 10, 64); err == nil {
			r.Ratelimit.Limit = int(n)
		}
		if n, err := strconv.ParseInt(resp.Header.Get("ratelimit-remaining"), 10, 64); err == nil {
			r.Ratelimit.Remaining = int(n)
		}
	}
	return r
}

// makeParams converts a map of string keys and any values into a URL-encoded string.
// If a value is a slice of strings, it will be joined with commas.
// Generally the values will be strings, numbers, booleans, or slices of strings
func makeParams(p map[string]any) string {
	params := url.Values{}
	for k, v := range p {
		if s, ok := v.([]string); ok {
			for _, v := range s {
				params.Add(k, v)
			}
		} else {
			params.Add(k, fmt.Sprint(v))
		}
	}

	return params.Encode()
}

func GetOauthSessionAuthArgs(session *database.OAuthSession) *XrpcAuthedRequestArgs {
	var dpopPrivateJWK jose.JSONWebKey
	err := dpopPrivateJWK.UnmarshalJSON([]byte(session.DpopPrivateJwk))
	if err != nil {
		log.Println("解析 DPoP 私钥失败: %w", err)
		return nil
	}
	return &XrpcAuthedRequestArgs{
		Did:            session.Did,
		AccessToken:    session.AccessToken,
		PdsUrl:         session.PdsUrl,
		Issuer:         session.AuthserverIss,
		DpopPdsNonce:   session.DpopPdsNonce,
		DpopPrivateJwk: dpopPrivateJWK,
	}
}

func NewXrpcClient(session *database.OAuthSession, db *gorm.DB) *XrpcClient {
	xrpcCli := &XrpcClient{
		Client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
		OnDpopPdsNonceChanged: func(did, newNonce string) {
			if err := database.UpdateOAuthSessionDpopPdsNonce(db, did, newNonce); err != nil {
				log.Println("error updating pds nonce", "err", err)
			}
		},
	}
	return xrpcCli
}
