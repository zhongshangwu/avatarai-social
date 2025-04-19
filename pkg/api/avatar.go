package api

import (
	"errors"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/zhongshangwu/avatarai-social/pkg/database"
	"github.com/zhongshangwu/avatarai-social/pkg/utils"
)

type AsterProfileResponse struct {
	Aster       *database.Avatar `json:"aster"`
	Initialized bool             `json:"initialized"`
}

func (a *AvatarAIAPI) HandleAsterProfile(c echo.Context) error {
	ac := c.(*utils.AvatarAIContext)
	avatar := ac.Avatar
	aster, err := database.GetAsterByCreatorDid(a.metaStore.DB, avatar.Did)
	if err != nil {
		if errors.Is(err, database.ErrAsterNotFound) {
			return c.JSON(200, AsterProfileResponse{
				Aster:       nil,
				Initialized: false,
			})
		}
		// 其他错误则返回 500
		return c.JSON(500, map[string]string{
			"error": "获取Aster信息失败: " + err.Error(),
		})
	}

	return c.JSON(200, AsterProfileResponse{
		Aster:       aster,
		Initialized: true,
	})
}

func (a *AvatarAIAPI) HandleAsterMint(c echo.Context) error {
	ac := c.(*utils.AvatarAIContext)
	avatar := ac.Avatar

	existingAster, err := database.GetAsterByCreatorDid(a.metaStore.DB, avatar.Did)
	if err == nil && existingAster != nil {
		return c.JSON(400, map[string]string{
			"error": "已经拥有Aster，不能重复铸造",
		})
	} else if err != nil && !errors.Is(err, database.ErrAsterNotFound) {
		return c.JSON(500, map[string]string{
			"error": "检查Aster状态失败: " + err.Error(),
		})
	}

	did, _, err := utils.GenerateDIDKey()
	if err != nil {
		return c.JSON(500, map[string]string{
			"error": "生成 didKey 失败: " + err.Error(),
		})
	}

	aster := &database.Avatar{
		Did:        did,
		CreatorDid: avatar.Did,
		IsAster:    true,
		CreatedAt:  time.Now(),
	}

	if err := database.CreateAster(a.metaStore.DB, aster); err != nil {
		return c.JSON(500, map[string]string{
			"error": "保存Aster失败: " + err.Error(),
		})
	}

	return c.JSON(200, aster)
}
