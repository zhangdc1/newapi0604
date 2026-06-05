package controller

import (
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
)

func distributionBaseURL(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil || strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https") {
		scheme = "https"
	}
	host := c.Request.Host
	if forwardedHost := c.GetHeader("X-Forwarded-Host"); forwardedHost != "" {
		host = forwardedHost
	}
	return scheme + "://" + host
}

func GetDistributionOverview(c *gin.Context) {
	userId := c.GetInt("id")
	overview, err := model.GetDistributionOverview(userId, distributionBaseURL(c))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, overview)
}

func GetDistributionInviteLink(c *gin.Context) {
	userId := c.GetInt("id")
	affCode, err := model.EnsureUserAffCode(userId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{
		"aff_code": affCode,
		"url":      distributionBaseURL(c) + "/register?aff=" + affCode,
	})
}

func GetDistributionInvites(c *gin.Context) {
	userId := c.GetInt("id")
	pageInfo := common.GetPageQuery(c)
	items, total, err := model.GetDistributionInvites(userId, pageInfo)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(items)
	common.ApiSuccess(c, pageInfo)
}

func GetDistributionCommissionRecords(c *gin.Context) {
	userId := c.GetInt("id")
	pageInfo := common.GetPageQuery(c)
	records, total, err := model.GetUserDistributionRecords(userId, pageInfo)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(records)
	common.ApiSuccess(c, pageInfo)
}

func SettleDistributionCommission(c *gin.Context) {
	userId := c.GetInt("id")
	quota, err := model.SettleDistributionCommission(userId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"quota": quota})
}

func AdminGetDistributionSetting(c *gin.Context) {
	setting, err := model.GetDistributionSetting()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, setting)
}

func AdminUpdateDistributionSetting(c *gin.Context) {
	var req model.DistributionSetting
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}
	req.UpdatedBy = c.GetInt("id")
	if err := model.UpdateDistributionSetting(&req); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, req)
}

func AdminGetDistributionRecords(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	referrerId, _ := strconv.Atoi(c.Query("referrer_user_id"))
	referredId, _ := strconv.Atoi(c.Query("referred_user_id"))
	status, _ := strconv.Atoi(c.Query("status"))
	records, total, err := model.SearchDistributionRecords(referrerId, referredId, status, pageInfo)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(records)
	common.ApiSuccess(c, pageInfo)
}

func AdminGetDistributionReferrers(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	items, total, err := model.SearchDistributionReferrers(c.Query("keyword"), c.Query("sort_by"), c.Query("sort_order"), pageInfo)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(items)
	common.ApiSuccess(c, pageInfo)
}

func AdminGetDistributionTransfers(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	userId, _ := strconv.Atoi(c.Query("user_id"))
	transfers, total, err := model.SearchDistributionTransfers(userId, pageInfo)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(transfers)
	common.ApiSuccess(c, pageInfo)
}

func AdminGetDistributionInvites(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	referrerId, _ := strconv.Atoi(c.Query("referrer_user_id"))
	items, total, err := model.SearchDistributionInvites(referrerId, pageInfo)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(items)
	common.ApiSuccess(c, pageInfo)
}

func AdminGetDistributionAgents(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	items, total, err := model.SearchDistributionAgents(c.Query("keyword"), pageInfo)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(items)
	common.ApiSuccess(c, pageInfo)
}

func AdminUpdateDistributionAgent(c *gin.Context) {
	userId, _ := strconv.Atoi(c.Param("user_id"))
	var req struct {
		IsAgent bool `json:"is_agent"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}
	if err := model.UpdateDistributionAgent(userId, req.IsAgent); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"user_id": userId, "is_agent": req.IsAgent})
}
