package model

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"

	"gorm.io/gorm"
)

const (
	DistributionRewardModeRate  = 1
	DistributionRewardModeFixed = 2

	DistributionCommissionStatusPending = 1
	DistributionCommissionStatusSettled = 2
	DistributionCommissionStatusInvalid = 3

	DistributionTransferStatusSuccess = 1
	DistributionTransferStatusFailed  = 2
)

const DistributionSourceCommissionSettle = "commission_settle"
const DistributionSourceAdminTopupComplete = "admin_topup_complete"

type DistributionSetting struct {
	Id                         int   `json:"id"`
	Enabled                    bool  `json:"enabled" gorm:"default:false"`
	RewardMode                 int   `json:"reward_mode" gorm:"default:1"`
	RewardRate                 int   `json:"reward_rate" gorm:"default:10"`
	FixedRewardQuota           int   `json:"fixed_reward_quota" gorm:"default:0"`
	ReferralLimit              int   `json:"referral_limit" gorm:"default:0"`
	NormalRewardMode           int   `json:"normal_reward_mode" gorm:"default:1"`
	NormalRewardRate           int   `json:"normal_reward_rate" gorm:"default:10"`
	NormalFixedRewardQuota     int   `json:"normal_fixed_reward_quota" gorm:"default:0"`
	NormalReferralLimit        int   `json:"normal_referral_limit" gorm:"default:5"`
	AgentRewardMode            int   `json:"agent_reward_mode" gorm:"default:1"`
	AgentRewardRate            int   `json:"agent_reward_rate" gorm:"default:10"`
	AgentFixedRewardQuota      int   `json:"agent_fixed_reward_quota" gorm:"default:0"`
	AgentReferralLimit         int   `json:"agent_referral_limit" gorm:"default:0"`
	SplitRuleInitialized       bool  `json:"split_rule_initialized" gorm:"default:false"`
	FreezeDays                 int   `json:"freeze_days" gorm:"default:0"`
	MinSettlementQuota         int   `json:"min_settlement_quota" gorm:"default:0"`
	AdminRechargeTrigger       bool  `json:"admin_recharge_trigger" gorm:"default:false"`
	RedeemTemporaryAttribution bool  `json:"redeem_temporary_attribution" gorm:"default:true"`
	UpdatedBy                  int   `json:"updated_by"`
	UpdatedAt                  int64 `json:"updated_at" gorm:"autoUpdateTime"`
	CreatedAt                  int64 `json:"created_at" gorm:"autoCreateTime"`
}

type DistributionCommissionRecord struct {
	Id               int    `json:"id"`
	ReferrerUserId   int    `json:"referrer_user_id" gorm:"index"`
	ReferredUserId   int    `json:"referred_user_id" gorm:"index"`
	Source           string `json:"source" gorm:"type:varchar(64);uniqueIndex:idx_distribution_source"`
	SourceId         string `json:"source_id" gorm:"type:varchar(128);uniqueIndex:idx_distribution_source"`
	SourceDetail     string `json:"source_detail" gorm:"type:varchar(255)"`
	IncreasedQuota   int    `json:"increased_quota"`
	CommissionQuota  int    `json:"commission_quota"`
	CommissionType   int    `json:"commission_type"`
	RewardRate       int    `json:"reward_rate"`
	ReferralLimit    int    `json:"referral_limit"`
	ReferrerIsAgent  bool   `json:"referrer_is_agent" gorm:"default:false;index"`
	FreezeUntil      int64  `json:"freeze_until" gorm:"index"`
	Status           int    `json:"status" gorm:"default:1;index"`
	SettledAt        int64  `json:"settled_at"`
	TemporaryAffCode string `json:"temporary_aff_code" gorm:"type:varchar(64)"`
	CreatedAt        int64  `json:"created_at" gorm:"autoCreateTime;index"`
	UpdatedAt        int64  `json:"updated_at" gorm:"autoUpdateTime"`
}

type DistributionTransfer struct {
	Id              int    `json:"id"`
	UserId          int    `json:"user_id" gorm:"index"`
	TransferQuota   int    `json:"transfer_quota"`
	SourceRecordIds string `json:"source_record_ids" gorm:"type:text"`
	Status          int    `json:"status" gorm:"default:1"`
	CreatedAt       int64  `json:"created_at" gorm:"autoCreateTime"`
	CompletedAt     int64  `json:"completed_at"`
}

type DistributionOverview struct {
	InviteLink       string `json:"invite_link"`
	AffCode          string `json:"aff_code"`
	InviteCount      int64  `json:"invite_count"`
	EffectiveInvites int64  `json:"effective_invites"`
	PendingQuota     int64  `json:"pending_quota"`
	AvailableQuota   int64  `json:"available_quota"`
	SettledQuota     int64  `json:"settled_quota"`
	TotalEarnedQuota int64  `json:"total_earned_quota"`
	MinSettlement    int    `json:"min_settlement_quota"`
	Enabled          bool   `json:"enabled"`
}

type DistributionInviteSummary struct {
	Id                  int    `json:"id"`
	Username            string `json:"username"`
	DisplayName         string `json:"display_name"`
	CreatedAt           int64  `json:"created_at"`
	TotalIncreasedQuota int64  `json:"total_increased_quota"`
	TotalCommission     int64  `json:"total_commission_quota"`
	RewardCount         int64  `json:"reward_count"`
}

type DistributionAdminInviteSummary struct {
	DistributionInviteSummary
	ReferrerUserId      int    `json:"referrer_user_id"`
	ReferrerUsername    string `json:"referrer_username"`
	ReferrerDisplayName string `json:"referrer_display_name"`
}

type DistributionReferrerSummary struct {
	Id               int    `json:"id"`
	Username         string `json:"username"`
	DisplayName      string `json:"display_name"`
	Email            string `json:"email"`
	Status           int    `json:"status"`
	IsAgent          bool   `json:"is_agent"`
	CreatedAt        int64  `json:"created_at"`
	InviteCount      int64  `json:"invite_count"`
	EffectiveInvites int64  `json:"effective_invites"`
	PendingQuota     int64  `json:"pending_quota"`
	AvailableQuota   int64  `json:"available_quota"`
	SettledQuota     int64  `json:"settled_quota"`
	TotalEarnedQuota int64  `json:"total_earned_quota"`
	LastRewardAt     int64  `json:"last_reward_at"`
}

type DistributionAgentUser struct {
	Id          int    `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Status      int    `json:"status"`
	IsAgent     bool   `json:"is_agent"`
	CreatedAt   int64  `json:"created_at"`
}

type distributionRewardRule struct {
	Mode          int
	Rate          int
	FixedQuota    int
	ReferralLimit int
	IsAgent       bool
}

type DistributionGrantInput struct {
	UserId           int
	IncreasedQuota   int
	Source           string
	SourceId         string
	SourceDetail     string
	TemporaryAffCode string
}

func GetDistributionSetting() (*DistributionSetting, error) {
	setting := &DistributionSetting{}
	err := DB.First(setting, "id = ?", 1).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		setting = &DistributionSetting{
			Id:                         1,
			Enabled:                    false,
			RewardMode:                 DistributionRewardModeRate,
			RewardRate:                 10,
			FixedRewardQuota:           0,
			ReferralLimit:              0,
			NormalRewardMode:           DistributionRewardModeRate,
			NormalRewardRate:           10,
			NormalFixedRewardQuota:     0,
			NormalReferralLimit:        5,
			AgentRewardMode:            DistributionRewardModeRate,
			AgentRewardRate:            10,
			AgentFixedRewardQuota:      0,
			AgentReferralLimit:         0,
			SplitRuleInitialized:       true,
			FreezeDays:                 0,
			MinSettlementQuota:         0,
			AdminRechargeTrigger:       false,
			RedeemTemporaryAttribution: true,
		}
		return setting, DB.Create(setting).Error
	}
	if err == nil && !setting.SplitRuleInitialized {
		initializeSplitDistributionRules(setting)
		if saveErr := DB.Save(setting).Error; saveErr != nil {
			return setting, saveErr
		}
	}
	return setting, err
}

func UpdateDistributionSetting(setting *DistributionSetting) error {
	setting.RewardMode, setting.RewardRate, setting.FixedRewardQuota, setting.ReferralLimit = normalizeDistributionRule(
		setting.RewardMode,
		setting.RewardRate,
		setting.FixedRewardQuota,
		setting.ReferralLimit,
	)
	setting.NormalRewardMode, setting.NormalRewardRate, setting.NormalFixedRewardQuota, setting.NormalReferralLimit = normalizeDistributionRule(
		setting.NormalRewardMode,
		setting.NormalRewardRate,
		setting.NormalFixedRewardQuota,
		setting.NormalReferralLimit,
	)
	setting.AgentRewardMode, setting.AgentRewardRate, setting.AgentFixedRewardQuota, setting.AgentReferralLimit = normalizeDistributionRule(
		setting.AgentRewardMode,
		setting.AgentRewardRate,
		setting.AgentFixedRewardQuota,
		setting.AgentReferralLimit,
	)
	if setting.FreezeDays < 0 {
		setting.FreezeDays = 0
	}
	if setting.MinSettlementQuota < 0 {
		setting.MinSettlementQuota = 0
	}

	existing, err := GetDistributionSetting()
	if err != nil {
		return err
	}
	existing.Enabled = setting.Enabled
	existing.RewardMode = setting.RewardMode
	existing.RewardRate = setting.RewardRate
	existing.FixedRewardQuota = setting.FixedRewardQuota
	existing.ReferralLimit = setting.ReferralLimit
	existing.NormalRewardMode = setting.NormalRewardMode
	existing.NormalRewardRate = setting.NormalRewardRate
	existing.NormalFixedRewardQuota = setting.NormalFixedRewardQuota
	existing.NormalReferralLimit = setting.NormalReferralLimit
	existing.AgentRewardMode = setting.AgentRewardMode
	existing.AgentRewardRate = setting.AgentRewardRate
	existing.AgentFixedRewardQuota = setting.AgentFixedRewardQuota
	existing.AgentReferralLimit = setting.AgentReferralLimit
	existing.SplitRuleInitialized = true
	existing.FreezeDays = setting.FreezeDays
	existing.MinSettlementQuota = setting.MinSettlementQuota
	existing.AdminRechargeTrigger = setting.AdminRechargeTrigger
	existing.RedeemTemporaryAttribution = setting.RedeemTemporaryAttribution
	existing.UpdatedBy = setting.UpdatedBy
	*setting = *existing
	return DB.Save(existing).Error
}

func initializeSplitDistributionRules(setting *DistributionSetting) {
	setting.NormalRewardMode = setting.RewardMode
	if setting.NormalRewardMode == 0 {
		setting.NormalRewardMode = DistributionRewardModeRate
	}
	setting.NormalRewardRate = setting.RewardRate
	if setting.NormalRewardRate == 0 {
		setting.NormalRewardRate = 10
	}
	setting.NormalFixedRewardQuota = setting.FixedRewardQuota
	setting.NormalReferralLimit = 5

	setting.AgentRewardMode = setting.RewardMode
	if setting.AgentRewardMode == 0 {
		setting.AgentRewardMode = DistributionRewardModeRate
	}
	setting.AgentRewardRate = setting.RewardRate
	if setting.AgentRewardRate == 0 {
		setting.AgentRewardRate = 10
	}
	setting.AgentFixedRewardQuota = setting.FixedRewardQuota
	setting.AgentReferralLimit = 0
	setting.SplitRuleInitialized = true
}

func EnsureUserAffCode(userId int) (string, error) {
	user, err := GetUserById(userId, true)
	if err != nil {
		return "", err
	}
	if user.AffCode != "" {
		return user.AffCode, nil
	}
	for i := 0; i < 5; i++ {
		user.AffCode = common.GetRandomString(8)
		if err := user.Update(false); err == nil {
			return user.AffCode, nil
		}
	}
	return "", errors.New("failed to generate invite code")
}

func ResolveDistributionReferrer(userId int, temporaryAffCode string, setting *DistributionSetting) (referrerId int, tempCode string) {
	if setting != nil && setting.RedeemTemporaryAttribution && temporaryAffCode != "" {
		if id, err := GetUserIdByAffCode(temporaryAffCode); err == nil && id != 0 && id != userId {
			return id, temporaryAffCode
		}
	}
	var user User
	if err := DB.Select("id", "inviter_id").First(&user, "id = ?", userId).Error; err == nil && user.InviterId != 0 && user.InviterId != userId {
		return user.InviterId, ""
	}
	return 0, ""
}

func GrantDistributionForQuotaIncrease(input DistributionGrantInput) {
	if input.IncreasedQuota <= 0 || input.UserId == 0 || input.Source == DistributionSourceCommissionSettle {
		return
	}
	setting, err := GetDistributionSetting()
	if err != nil {
		common.SysLog("failed to load distribution setting: " + err.Error())
		return
	}
	if !setting.Enabled {
		return
	}
	if isAdminDistributionSource(input.Source) && !setting.AdminRechargeTrigger {
		return
	}
	if input.SourceId == "" {
		input.SourceId = fmt.Sprintf("%s:%d:%d", input.Source, input.UserId, time.Now().UnixNano())
	}
	referrerId, tempCode := ResolveDistributionReferrer(input.UserId, input.TemporaryAffCode, setting)
	if referrerId == 0 {
		return
	}
	var referrer User
	if err := DB.Select("id", "status", "is_agent").First(&referrer, "id = ?", referrerId).Error; err != nil || referrer.Status != common.UserStatusEnabled {
		return
	}
	rule := distributionRuleForReferrer(referrer, setting)
	reward := calculateDistributionReward(input.IncreasedQuota, rule)
	if reward <= 0 {
		return
	}
	freezeUntil := common.GetTimestamp() + int64(setting.FreezeDays*24*60*60)

	err = DB.Transaction(func(tx *gorm.DB) error {
		var exists int64
		if err := tx.Model(&DistributionCommissionRecord{}).
			Where("source = ? AND source_id = ?", input.Source, input.SourceId).
			Count(&exists).Error; err != nil {
			return err
		}
		if exists > 0 {
			return nil
		}
		if rule.ReferralLimit > 0 {
			var count int64
			if err := tx.Model(&DistributionCommissionRecord{}).
				Where("referrer_user_id = ? AND referred_user_id = ? AND status <> ?", referrerId, input.UserId, DistributionCommissionStatusInvalid).
				Count(&count).Error; err != nil {
				return err
			}
			if int(count) >= rule.ReferralLimit {
				return nil
			}
		}
		record := &DistributionCommissionRecord{
			ReferrerUserId:   referrerId,
			ReferredUserId:   input.UserId,
			Source:           input.Source,
			SourceId:         input.SourceId,
			SourceDetail:     input.SourceDetail,
			IncreasedQuota:   input.IncreasedQuota,
			CommissionQuota:  reward,
			CommissionType:   rule.Mode,
			RewardRate:       rule.Rate,
			ReferralLimit:    rule.ReferralLimit,
			ReferrerIsAgent:  rule.IsAgent,
			FreezeUntil:      freezeUntil,
			Status:           DistributionCommissionStatusPending,
			TemporaryAffCode: tempCode,
		}
		if err := tx.Create(record).Error; err != nil {
			return err
		}
		updates := map[string]interface{}{
			"aff_history": gorm.Expr("aff_history + ?", reward),
		}
		if freezeUntil <= common.GetTimestamp() {
			updates["aff_quota"] = gorm.Expr("aff_quota + ?", reward)
		}
		if err := tx.Model(&User{}).Where("id = ?", referrerId).Updates(updates).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		common.SysLog("failed to grant distribution commission: " + err.Error())
	}
}

func isAdminDistributionSource(source string) bool {
	return source == "admin" || source == DistributionSourceAdminTopupComplete
}

func normalizeDistributionRule(mode int, rate int, fixedQuota int, referralLimit int) (int, int, int, int) {
	if mode != DistributionRewardModeFixed {
		mode = DistributionRewardModeRate
	}
	if rate < 0 {
		rate = 0
	}
	if rate > 50 {
		rate = 50
	}
	if fixedQuota < 0 {
		fixedQuota = 0
	}
	if referralLimit < 0 {
		referralLimit = 0
	}
	return mode, rate, fixedQuota, referralLimit
}

func distributionRuleForReferrer(referrer User, setting *DistributionSetting) distributionRewardRule {
	if setting == nil {
		return distributionRewardRule{
			Mode:          DistributionRewardModeRate,
			Rate:          10,
			FixedQuota:    0,
			ReferralLimit: 5,
			IsAgent:       referrer.IsAgent,
		}
	}
	baseMode := setting.RewardMode
	baseRate := setting.RewardRate
	baseFixed := setting.FixedRewardQuota
	if baseRate == 0 {
		baseRate = 10
	}
	if referrer.IsAgent {
		mode := setting.AgentRewardMode
		rate := setting.AgentRewardRate
		fixed := setting.AgentFixedRewardQuota
		limit := setting.AgentReferralLimit
		if mode == 0 {
			mode = baseMode
		}
		if rate == 0 {
			rate = baseRate
		}
		if fixed == 0 {
			fixed = baseFixed
		}
		mode, rate, fixed, limit = normalizeDistributionRule(mode, rate, fixed, limit)
		return distributionRewardRule{Mode: mode, Rate: rate, FixedQuota: fixed, ReferralLimit: limit, IsAgent: true}
	}
	mode := setting.NormalRewardMode
	rate := setting.NormalRewardRate
	fixed := setting.NormalFixedRewardQuota
	limit := setting.NormalReferralLimit
	if mode == 0 {
		mode = baseMode
	}
	if rate == 0 {
		rate = baseRate
	}
	if fixed == 0 {
		fixed = baseFixed
	}
	if limit == 0 && setting.ReferralLimit > 0 {
		limit = setting.ReferralLimit
	}
	if limit == 0 {
		limit = 5
	}
	mode, rate, fixed, limit = normalizeDistributionRule(mode, rate, fixed, limit)
	return distributionRewardRule{Mode: mode, Rate: rate, FixedQuota: fixed, ReferralLimit: limit, IsAgent: false}
}

func calculateDistributionReward(increasedQuota int, rule distributionRewardRule) int {
	if rule.Mode == DistributionRewardModeFixed {
		return rule.FixedQuota
	}
	return int(math.Round(float64(increasedQuota) * float64(rule.Rate) / 100.0))
}

func GetDistributionOverview(userId int, baseURL string) (*DistributionOverview, error) {
	affCode, err := EnsureUserAffCode(userId)
	if err != nil {
		return nil, err
	}
	setting, err := GetDistributionSetting()
	if err != nil {
		return nil, err
	}
	now := common.GetTimestamp()
	overview := &DistributionOverview{
		AffCode:       affCode,
		InviteLink:    fmt.Sprintf("%s/register?aff=%s", baseURL, affCode),
		MinSettlement: setting.MinSettlementQuota,
		Enabled:       setting.Enabled,
	}
	DB.Model(&User{}).Where("inviter_id = ?", userId).Count(&overview.InviteCount)
	DB.Model(&DistributionCommissionRecord{}).Where("referrer_user_id = ? AND status <> ?", userId, DistributionCommissionStatusInvalid).
		Distinct("referred_user_id").Count(&overview.EffectiveInvites)
	DB.Model(&DistributionCommissionRecord{}).Where("referrer_user_id = ? AND status = ? AND freeze_until > ?", userId, DistributionCommissionStatusPending, now).
		Select("COALESCE(SUM(commission_quota), 0)").Scan(&overview.PendingQuota)
	DB.Model(&DistributionCommissionRecord{}).Where("referrer_user_id = ? AND status = ? AND freeze_until <= ?", userId, DistributionCommissionStatusPending, now).
		Select("COALESCE(SUM(commission_quota), 0)").Scan(&overview.AvailableQuota)
	DB.Model(&DistributionCommissionRecord{}).Where("referrer_user_id = ? AND status = ?", userId, DistributionCommissionStatusSettled).
		Select("COALESCE(SUM(commission_quota), 0)").Scan(&overview.SettledQuota)
	overview.TotalEarnedQuota = overview.PendingQuota + overview.AvailableQuota + overview.SettledQuota
	return overview, nil
}

func GetDistributionInvites(userId int, pageInfo *common.PageInfo) ([]DistributionInviteSummary, int64, error) {
	var total int64
	if err := DB.Model(&User{}).Where("inviter_id = ?", userId).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var users []User
	if err := DB.Select("id", "username", "display_name", "created_at").
		Where("inviter_id = ?", userId).
		Order("id desc").
		Limit(pageInfo.GetPageSize()).
		Offset(pageInfo.GetStartIdx()).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}
	items := make([]DistributionInviteSummary, 0, len(users))
	for _, u := range users {
		item := DistributionInviteSummary{
			Id:          u.Id,
			Username:    u.Username,
			DisplayName: u.DisplayName,
			CreatedAt:   u.CreatedAt,
		}
		DB.Model(&DistributionCommissionRecord{}).Where("referrer_user_id = ? AND referred_user_id = ? AND status <> ?", userId, u.Id, DistributionCommissionStatusInvalid).
			Select("COALESCE(SUM(increased_quota), 0)").Scan(&item.TotalIncreasedQuota)
		DB.Model(&DistributionCommissionRecord{}).Where("referrer_user_id = ? AND referred_user_id = ? AND status <> ?", userId, u.Id, DistributionCommissionStatusInvalid).
			Select("COALESCE(SUM(commission_quota), 0)").Scan(&item.TotalCommission)
		DB.Model(&DistributionCommissionRecord{}).Where("referrer_user_id = ? AND referred_user_id = ? AND status <> ?", userId, u.Id, DistributionCommissionStatusInvalid).
			Count(&item.RewardCount)
		items = append(items, item)
	}
	return items, total, nil
}

func SearchDistributionInvites(referrerId int, pageInfo *common.PageInfo) ([]DistributionAdminInviteSummary, int64, error) {
	query := DB.Model(&User{}).Where("inviter_id <> 0")
	if referrerId > 0 {
		query = query.Where("inviter_id = ?", referrerId)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var users []User
	if err := query.Select("id", "username", "display_name", "created_at", "inviter_id").
		Order("id desc").
		Limit(pageInfo.GetPageSize()).
		Offset(pageInfo.GetStartIdx()).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}
	referrerIds := make([]int, 0, len(users))
	for _, u := range users {
		referrerIds = append(referrerIds, u.InviterId)
	}
	referrers := map[int]User{}
	if len(referrerIds) > 0 {
		var referrerUsers []User
		if err := DB.Select("id", "username", "display_name").Where("id IN ?", referrerIds).Find(&referrerUsers).Error; err != nil {
			return nil, 0, err
		}
		for _, referrer := range referrerUsers {
			referrers[referrer.Id] = referrer
		}
	}
	items := make([]DistributionAdminInviteSummary, 0, len(users))
	for _, u := range users {
		item := DistributionAdminInviteSummary{
			DistributionInviteSummary: DistributionInviteSummary{
				Id:          u.Id,
				Username:    u.Username,
				DisplayName: u.DisplayName,
				CreatedAt:   u.CreatedAt,
			},
			ReferrerUserId: u.InviterId,
		}
		if referrer, ok := referrers[u.InviterId]; ok {
			item.ReferrerUsername = referrer.Username
			item.ReferrerDisplayName = referrer.DisplayName
		}
		DB.Model(&DistributionCommissionRecord{}).Where("referrer_user_id = ? AND referred_user_id = ? AND status <> ?", u.InviterId, u.Id, DistributionCommissionStatusInvalid).
			Select("COALESCE(SUM(increased_quota), 0)").Scan(&item.TotalIncreasedQuota)
		DB.Model(&DistributionCommissionRecord{}).Where("referrer_user_id = ? AND referred_user_id = ? AND status <> ?", u.InviterId, u.Id, DistributionCommissionStatusInvalid).
			Select("COALESCE(SUM(commission_quota), 0)").Scan(&item.TotalCommission)
		DB.Model(&DistributionCommissionRecord{}).Where("referrer_user_id = ? AND referred_user_id = ? AND status <> ?", u.InviterId, u.Id, DistributionCommissionStatusInvalid).
			Count(&item.RewardCount)
		items = append(items, item)
	}
	return items, total, nil
}

func GetUserDistributionRecords(userId int, pageInfo *common.PageInfo) ([]DistributionCommissionRecord, int64, error) {
	var records []DistributionCommissionRecord
	var total int64
	query := DB.Model(&DistributionCommissionRecord{}).Where("referrer_user_id = ?", userId)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("id desc").Limit(pageInfo.GetPageSize()).Offset(pageInfo.GetStartIdx()).Find(&records).Error
	return records, total, err
}

func SettleDistributionCommission(userId int) (int, error) {
	setting, err := GetDistributionSetting()
	if err != nil {
		return 0, err
	}
	now := common.GetTimestamp()
	var records []DistributionCommissionRecord
	if err := DB.Where("referrer_user_id = ? AND status = ? AND freeze_until <= ?", userId, DistributionCommissionStatusPending, now).
		Order("id asc").Find(&records).Error; err != nil {
		return 0, err
	}
	total := 0
	ids := make([]int, 0, len(records))
	for _, record := range records {
		total += record.CommissionQuota
		ids = append(ids, record.Id)
	}
	if total <= 0 {
		return 0, errors.New("no available commission")
	}
	if setting.MinSettlementQuota > 0 && total < setting.MinSettlementQuota {
		return 0, fmt.Errorf("available commission is below minimum settlement: %s", logger.LogQuota(setting.MinSettlementQuota))
	}
	err = DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&DistributionCommissionRecord{}).Where("id IN ?", ids).Updates(map[string]interface{}{
			"status":     DistributionCommissionStatusSettled,
			"settled_at": now,
		}).Error; err != nil {
			return err
		}
		if err := tx.Model(&User{}).Where("id = ?", userId).Updates(map[string]interface{}{
			"quota":     gorm.Expr("quota + ?", total),
			"aff_quota": gorm.Expr("CASE WHEN aff_quota >= ? THEN aff_quota - ? ELSE 0 END", total, total),
		}).Error; err != nil {
			return err
		}
		transfer := &DistributionTransfer{
			UserId:          userId,
			TransferQuota:   total,
			SourceRecordIds: common.GetJsonString(ids),
			Status:          DistributionTransferStatusSuccess,
			CompletedAt:     now,
		}
		if err := tx.Create(transfer).Error; err != nil {
			return err
		}
		return nil
	})
	if err == nil {
		RecordLog(userId, LogTypeSystem, fmt.Sprintf("分销奖励转入余额 %s", logger.LogQuota(total)))
	}
	return total, err
}

func SearchDistributionRecords(referrerId int, referredId int, status int, pageInfo *common.PageInfo) ([]DistributionCommissionRecord, int64, error) {
	query := DB.Model(&DistributionCommissionRecord{})
	if referrerId > 0 {
		query = query.Where("referrer_user_id = ?", referrerId)
	}
	if referredId > 0 {
		query = query.Where("referred_user_id = ?", referredId)
	}
	if status > 0 {
		query = query.Where("status = ?", status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var records []DistributionCommissionRecord
	err := query.Order("id desc").Limit(pageInfo.GetPageSize()).Offset(pageInfo.GetStartIdx()).Find(&records).Error
	return records, total, err
}

func SearchDistributionTransfers(userId int, pageInfo *common.PageInfo) ([]DistributionTransfer, int64, error) {
	query := DB.Model(&DistributionTransfer{})
	if userId > 0 {
		query = query.Where("user_id = ?", userId)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var transfers []DistributionTransfer
	err := query.Order("id desc").Limit(pageInfo.GetPageSize()).Offset(pageInfo.GetStartIdx()).Find(&transfers).Error
	return transfers, total, err
}

func SearchDistributionAgents(keyword string, pageInfo *common.PageInfo) ([]DistributionAgentUser, int64, error) {
	query := DB.Model(&User{}).Where("deleted_at IS NULL")
	keyword = strings.TrimSpace(keyword)
	if keyword != "" {
		if id, err := strconv.Atoi(keyword); err == nil {
			query = query.Where("id = ? OR username LIKE ? OR display_name LIKE ?", id, keyword+"%", keyword+"%")
		} else {
			query = query.Where("username LIKE ? OR display_name LIKE ? OR email LIKE ?", keyword+"%", keyword+"%", keyword+"%")
		}
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var users []DistributionAgentUser
	err := query.Select("id", "username", "display_name", "email", "status", "is_agent", "created_at").
		Order("id desc").
		Limit(pageInfo.GetPageSize()).
		Offset(pageInfo.GetStartIdx()).
		Find(&users).Error
	return users, total, err
}

func SearchDistributionReferrers(keyword string, sortBy string, sortOrder string, pageInfo *common.PageInfo) ([]DistributionReferrerSummary, int64, error) {
	query := DB.Model(&User{}).Where("deleted_at IS NULL")
	keyword = strings.TrimSpace(keyword)
	if keyword != "" {
		if id, err := strconv.Atoi(keyword); err == nil {
			query = query.Where("id = ? OR username LIKE ? OR display_name LIKE ?", id, keyword+"%", keyword+"%")
		} else {
			query = query.Where("username LIKE ? OR display_name LIKE ? OR email LIKE ?", keyword+"%", keyword+"%", keyword+"%")
		}
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var users []DistributionReferrerSummary
	if err := query.Select("id", "username", "display_name", "email", "status", "is_agent", "created_at").Find(&users).Error; err != nil {
		return nil, 0, err
	}
	now := common.GetTimestamp()
	for idx := range users {
		userId := users[idx].Id
		DB.Model(&User{}).Where("inviter_id = ?", userId).Count(&users[idx].InviteCount)
		DB.Model(&DistributionCommissionRecord{}).Where("referrer_user_id = ? AND status <> ?", userId, DistributionCommissionStatusInvalid).
			Distinct("referred_user_id").Count(&users[idx].EffectiveInvites)
		DB.Model(&DistributionCommissionRecord{}).Where("referrer_user_id = ? AND status = ? AND freeze_until > ?", userId, DistributionCommissionStatusPending, now).
			Select("COALESCE(SUM(commission_quota), 0)").Scan(&users[idx].PendingQuota)
		DB.Model(&DistributionCommissionRecord{}).Where("referrer_user_id = ? AND status = ? AND freeze_until <= ?", userId, DistributionCommissionStatusPending, now).
			Select("COALESCE(SUM(commission_quota), 0)").Scan(&users[idx].AvailableQuota)
		DB.Model(&DistributionCommissionRecord{}).Where("referrer_user_id = ? AND status = ?", userId, DistributionCommissionStatusSettled).
			Select("COALESCE(SUM(commission_quota), 0)").Scan(&users[idx].SettledQuota)
		users[idx].TotalEarnedQuota = users[idx].PendingQuota + users[idx].AvailableQuota + users[idx].SettledQuota
		DB.Model(&DistributionCommissionRecord{}).Where("referrer_user_id = ?", userId).
			Select("COALESCE(MAX(created_at), 0)").Scan(&users[idx].LastRewardAt)
	}
	sortDistributionReferrers(users, sortBy, sortOrder)
	start := pageInfo.GetStartIdx()
	if start >= len(users) {
		return []DistributionReferrerSummary{}, total, nil
	}
	end := pageInfo.GetEndIdx()
	if end > len(users) {
		end = len(users)
	}
	return users[start:end], total, nil
}

func sortDistributionReferrers(items []DistributionReferrerSummary, sortBy string, sortOrder string) {
	sortBy = strings.ToLower(strings.TrimSpace(sortBy))
	sortOrder = strings.ToLower(strings.TrimSpace(sortOrder))
	if sortOrder != "asc" {
		sortOrder = "desc"
	}
	value := func(item DistributionReferrerSummary) int64 {
		switch sortBy {
		case "total_earned_quota":
			return item.TotalEarnedQuota
		case "invite_count":
			return item.InviteCount
		case "effective_invites":
			return item.EffectiveInvites
		case "last_reward_at":
			return item.LastRewardAt
		default:
			return item.AvailableQuota
		}
	}
	sort.SliceStable(items, func(i, j int) bool {
		left := value(items[i])
		right := value(items[j])
		if left == right {
			if sortOrder == "asc" {
				return items[i].Id < items[j].Id
			}
			return items[i].Id > items[j].Id
		}
		if sortOrder == "asc" {
			return left < right
		}
		return left > right
	})
}

func UpdateDistributionAgent(userId int, isAgent bool) error {
	if userId <= 0 {
		return errors.New("invalid user id")
	}
	return DB.Model(&User{}).Where("id = ? AND deleted_at IS NULL", userId).Update("is_agent", isAgent).Error
}
