package handlers

import (
	"net/http"

	"github.com/zhongshangwu/avatarai-social/pkg/config"
	"github.com/zhongshangwu/avatarai-social/pkg/mcp"
	"github.com/zhongshangwu/avatarai-social/pkg/repositories"
	"github.com/zhongshangwu/avatarai-social/pkg/services"
	"github.com/zhongshangwu/avatarai-social/types"
)

type ListMCPServersResponse struct {
	Servers []*mcp.MCPServerInfo `json:"servers"`
}

type InstallMCPServerRequest struct {
	Name     string                `json:"name"`
	Endpoint mcp.MCPServerEndpoint `json:"endpoint"`
}

type MCPMarketplaceHandler struct {
	mcpService *services.MCPService
}

func NewMCPMarketplaceHandler(config *config.SocialConfig, metaStore *repositories.MetaStore) *MCPMarketplaceHandler {
	return &MCPMarketplaceHandler{
		mcpService: services.NewMCPService(metaStore, config),
	}
}

func (h *MCPMarketplaceHandler) ListMCPServers(c *types.APIContext) error {
	userDid := c.User.Did
	servers, err := h.mcpService.ListMCPServers(userDid)
	if err != nil {
		return c.InternalServerError("获取MCP服务器列表失败")
	}
	return c.JSON(http.StatusOK, ListMCPServersResponse{Servers: servers})
}

func (h *MCPMarketplaceHandler) MCPServerDetail(c *types.APIContext) error {
	userDid := c.User.Did
	mcpId := c.Param("mcpId")
	server, err := h.mcpService.GetMCPServerDetail(mcpId, userDid)
	if err != nil {
		return c.InternalServerError("获取MCP服务器详情失败")
	}
	if server == nil {
		return c.NotFound("MCP服务器不存在")
	}
	return c.JSON(http.StatusOK, server)
}

func (h *MCPMarketplaceHandler) InstallMCPServer(c *types.APIContext) error {
	var req InstallMCPServerRequest
	if err := c.Bind(&req); err != nil {
		return c.InvalidRequest("invalid_request", "无效的请求参数")
	}

	userDid := c.User.Did
	mcpId, err := h.mcpService.AddMCPServer(req.Name, &req.Endpoint, userDid)
	if err != nil {
		return c.InternalServerError("添加MCP服务器失败")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "MCP服务器添加成功",
		"mcpId":   mcpId,
	})
}

func (h *MCPMarketplaceHandler) UninstallMCPServer(c *types.APIContext) error {
	mcpId := c.QueryParam("mcpId")

	userDid := c.User.Did

	if err := h.mcpService.DeleteMCPServer(mcpId, userDid); err != nil {
		return c.InternalServerError("删除MCP服务器失败")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "MCP服务器删除成功",
		"mcpId":   mcpId,
	})
}

func (h *MCPMarketplaceHandler) ToggleSyncResourcesStatus(c *types.APIContext) error {
	mcpId := c.Param("mcpId")
	var req struct {
		SyncResources bool `json:"syncResources"`
	}
	if err := c.Bind(&req); err != nil {
		return c.InvalidRequest("invalid_request", "无效的请求参数")
	}

	userDid := c.User.Did
	if err := h.mcpService.UpdateSyncResourcesStatus(mcpId, userDid, req.SyncResources); err != nil {
		return c.InternalServerError("更新同步状态失败")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"mcpId":         mcpId,
		"syncResources": req.SyncResources,
	})
}

func (h *MCPMarketplaceHandler) ToggleEnabled(c *types.APIContext) error {
	mcpId := c.Param("mcpId")
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.Bind(&req); err != nil {
		return c.InvalidRequest("invalid_request", "无效的请求参数")
	}

	userDid := c.User.Did

	serverInfo, err := h.mcpService.GetMCPServerDetail(mcpId, userDid)
	if err != nil {
		return c.InternalServerError("获取MCP服务器详情失败")
	}
	if serverInfo == nil {
		return c.NotFound("MCP服务器不存在")
	}

	if serverInfo.IsBuiltin {
		if err := h.mcpService.InstallBuiltinIfNotExists(serverInfo, userDid); err != nil {
			return c.InternalServerError("更新启用状态失败")
		}
	}

	if err := h.mcpService.UpdateEnabledStatus(mcpId, userDid, req.Enabled); err != nil {
		return c.InternalServerError("更新启用状态失败")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"mcpId":   mcpId,
		"enabled": req.Enabled,
	})
}
