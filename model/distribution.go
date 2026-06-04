package model

import (
	"errors"
	"fmt"
	"math"
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
			FreezeDays:                 0,
			MinSettlementQuota:         0,
			AdminRechargeTrigger:       false,
			RedeemTemporaryAttribution: true,
		}
		return setting, DB.Create(setting).Error
	}
	return setting, err
}

func UpdateDistributionSetting(setting *DistributionSetting) error {
	if setting.RewardMode != DistributionRewardModeFixed {
		setting.RewardMode = DistributionRewardModeRate
	}
	if setting.RewardRate < 0 {
		setting.RewardRate = 0
	}
	if setting.RewardRate > 50 {
		setting.RewardRate = 50
	}
	if setting.FixedRewardQuota < 0 {
		setting.FixedRewardQuota = 0
	}
	if setting.ReferralLimit < 0 {
		setting.ReferralLimit = 0
	}
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
	existing.FreezeDays = setting.FreezeDays
	existing.MinSettlementQuota = setting.MinSettlementQuota
	existing.AdminRechargeTrigger = setting.AdminRechargeTrigger
	existing.RedeemTemporaryAttribution = setting.RedeemTemporaryAttribution
	existing.UpdatedBy = setting.UpdatedBy
	*setting = *existing
	return DB.Save(existing).Error
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
	if err := DB.Select("id", "status").First(&referrer, "id = ?", referrerId).Error; err != nil || referrer.Status != common.UserStatusEnabled {
		return
	}
	reward := calculateDistributionReward(input.IncreasedQuota, setting)
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
		if setting.ReferralLimit > 0 {
			var count int64
			if err := tx.Model(&DistributionCommissionRecord{}).
				Where("referrer_user_id = ? AND referred_user_id = ? AND status <> ?", referrerId, input.UserId, DistributionCommissionStatusInvalid).
				Count(&count).Error; err != nil {
				return err
			}
			if int(count) >= setting.ReferralLimit {
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
			CommissionType:   setting.RewardMode,
			RewardRate:       setting.RewardRate,
			ReferralLimit:    setting.ReferralLimit,
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

func calculateDistributionReward(increasedQuota int, setting *DistributionSetting) int {
	if setting.RewardMode == DistributionRewardModeFixed {
		return setting.FixedRewardQuota
	}
	return int(math.Round(float64(increasedQuota) * float64(setting.RewardRate) / 100.0))
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
