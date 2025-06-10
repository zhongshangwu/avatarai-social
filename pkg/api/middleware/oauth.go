package middleware

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/zhongshangwu/avatarai-social/pkg/config"
	"github.com/zhongshangwu/avatarai-social/pkg/repositories"
	"github.com/zhongshangwu/avatarai-social/pkg/utils"
	"github.com/zhongshangwu/avatarai-social/types"
)

type WrapperFunc func(next ContextualHandlerFunc, mustAuth bool) echo.HandlerFunc
type ContextualHandlerFunc func(c *types.APIContext) error

type AuthenticationError struct {
	Code    string
	Message string
	Err     error
}

func (e *AuthenticationError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func convertRepositorySessionToTypes(repoSession *repositories.Session) *types.Session {
	if repoSession == nil {
		return nil
	}
	return &types.Session{
		ID:             repoSession.ID,
		UserDid:        repoSession.UserDid,
		AccessToken:    repoSession.AccessToken,
		RefreshToken:   repoSession.RefreshToken,
		OAuthSessionID: fmt.Sprintf("%d", repoSession.OAuthSessionID), // 转换 uint 到 string
		ExpiredAt:      repoSession.ExpiredAt,
		CreatedAt:      repoSession.CreatedAt,
	}
}

func convertRepositoryOAuthSessionToTypes(repoOAuthSession *repositories.OAuthSession) *types.OAuthSession {
	if repoOAuthSession == nil {
		return nil
	}
	return &types.OAuthSession{
		ID:                  fmt.Sprintf("%d", repoOAuthSession.ID), // 转换 uint 到 string
		Did:                 repoOAuthSession.Did,
		Handle:              repoOAuthSession.Handle,
		PdsUrl:              repoOAuthSession.PdsUrl,
		AuthserverIss:       repoOAuthSession.AuthserverIss,
		AccessToken:         repoOAuthSession.AccessToken,
		RefreshToken:        repoOAuthSession.RefreshToken,
		DpopAuthserverNonce: repoOAuthSession.DpopAuthserverNonce,
		DpopPdsNonce:        repoOAuthSession.DpopPdsNonce,
		DpopPrivateJwk:      repoOAuthSession.DpopPrivateJwk,
		ExpiresIn:           repoOAuthSession.ExpiresIn,
		CreatedAt:           repoOAuthSession.CreatedAt,
		Provider:            types.OAuthProviderType(repoOAuthSession.Platform),
		ReturnURI:           repoOAuthSession.ReturnURI,
	}
}

func convertRepositoryAvatarToUser(repoAvatar *repositories.Avatar) *types.User {
	if repoAvatar == nil {
		return nil
	}
	return &types.User{
		Did:         repoAvatar.Did,
		Handle:      repoAvatar.Handle,
		PdsUrl:      repoAvatar.PdsUrl,
		DisplayName: repoAvatar.DisplayName,
		AvatarCID:   repoAvatar.AvatarCID,
		Description: repoAvatar.Description,
		IsAster:     repoAvatar.IsAster,
		Creator:     repoAvatar.Creator,
		LastLoginAt: repoAvatar.LastLoginAt,
		UpdatedAt:   repoAvatar.UpdatedAt,
		CreatedAt:   repoAvatar.CreatedAt,
	}
}

func authenticateUser(metaStore *repositories.MetaStore, config *config.SocialConfig, token string) (*repositories.Session, *repositories.OAuthSession, *repositories.Avatar, *AuthenticationError) {
	if token == "" {
		return nil, nil, nil, &AuthenticationError{
			Code:    "missing_token",
			Message: "缺少认证令牌",
		}
	}

	claims, err := utils.ValidateAccessToken(config, token)
	if err != nil {
		return nil, nil, nil, &AuthenticationError{
			Code:    "invalid_token",
			Message: "无效的访问令牌",
			Err:     err,
		}
	}

	session, err := metaStore.UserRepo.GetSessionByID(claims.ID)
	if err != nil {
		return nil, nil, nil, &AuthenticationError{
			Code:    "session_not_found",
			Message: "会话不存在",
			Err:     err,
		}
	}

	oauthSession, err := metaStore.OAuthRepo.GetOAuthSessionByID(session.OAuthSessionID)
	if err != nil {
		return nil, nil, nil, &AuthenticationError{
			Code:    "oauth_session_not_found",
			Message: "OAuth会话不存在",
			Err:     err,
		}
	}

	if isOAuthSessionExpired(oauthSession) {
		return nil, nil, nil, &AuthenticationError{
			Code:    "oauth_session_expired",
			Message: "OAuth会话已过期",
		}
	}

	avatar, err := metaStore.UserRepo.GetAvatarByDID(session.UserDid)
	if err != nil {
		return nil, nil, nil, &AuthenticationError{
			Code:    "user_not_found",
			Message: "用户不存在",
			Err:     err,
		}
	}

	return session, oauthSession, avatar, nil
}

func handleAuthenticationFailure(c *types.APIContext, authErr *AuthenticationError, mustAuth bool) error {
	log.Errorf("认证失败 [%s]: %s", authErr.Code, authErr.Error())

	if !mustAuth {
		// 如果不强制认证，继续执行但不设置用户信息
		c.IsAuthenticated = false
		return nil
	}

	// 根据错误类型决定重定向策略
	switch authErr.Code {
	case "missing_token", "invalid_token":
		return c.RedirectToLogin("")
	case "session_not_found", "oauth_session_not_found", "user_not_found":
		return c.RedirectToLogin("")
	case "oauth_session_expired":
		// OAuth 会话过期，可以尝试自动刷新或重定向到登录
		log.Warnf("OAuth会话已过期，重定向到登录页面")
		return c.RedirectToLogin("")
	default:
		return c.RedirectToLogin("")
	}
}

func NewContextWrapper(metaStore *repositories.MetaStore, config *config.SocialConfig) WrapperFunc {
	return func(next ContextualHandlerFunc, mustAuth bool) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &types.APIContext{
				Context:         c,
				IsAuthenticated: false,
				User:            nil,
				Session:         nil,
				OauthProvider:   types.OAuthProviderType(c.QueryParam("provider")),
				OauthSession:    nil,
			}

			token := cc.AuthToken()

			session, oauthSession, avatar, authErr := authenticateUser(metaStore, config, token)

			if authErr != nil {
				if err := handleAuthenticationFailure(cc, authErr, mustAuth); err != nil {
					return err
				}

				if !mustAuth {
					return next(cc)
				}

				return nil
			}

			cc.IsAuthenticated = true
			cc.Session = convertRepositorySessionToTypes(session)
			cc.OauthSession = convertRepositoryOAuthSessionToTypes(oauthSession)
			cc.User = convertRepositoryAvatarToUser(avatar)

			log.Infof("用户认证成功: %s (%s)", avatar.Handle, avatar.Did)

			return next(cc)
		}
	}
}

// func NewRequireAuth(metaStore *repositories.MetaStore, config *config.SocialConfig) echo.MiddlewareFunc {
// 	wrapper := NewContextWrapper(metaStore, config)
// 	return func(next echo.HandlerFunc) echo.HandlerFunc {
// 		return wrapper(next, true)
// 	}
// }

func IsSessionExpired(session *repositories.OAuthSession) bool {
	return isOAuthSessionExpired(session)
}

func isOAuthSessionExpired(session *repositories.OAuthSession) bool {
	if session.ExpiresIn <= 0 {
		return false
	}
	expiryTime := time.Unix(session.CreatedAt, 0).Add(time.Duration(session.ExpiresIn) * time.Second)
	return time.Now().After(expiryTime)
}
