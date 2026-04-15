package models

import (
	"time"
)

// =================================================================
// 1. 用户与权限模块 (User & Auth)
// =================================================================

// RegisterRequest 用户注册
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=4,max=32"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Phone    string `json:"phone" binding:"omitempty,mobile"` // 假设有自定义的手机号校验
	UserType string `json:"user_type" binding:"oneof=user master institution"`
}

// LoginRequest 登录
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UpdateProfileRequest 更新个人资料
type UpdateProfileRequest struct {
	Nickname    string                 `json:"nickname" binding:"omitempty,max=64"`
	AvatarURL   string                 `json:"avatar_url" binding:"omitempty,url"`
	Bio         string                 `json:"bio" binding:"omitempty,max=500"`
	Address     string                 `json:"address"`
	Workshop    string                 `json:"workshop_name"`
	ContactInfo map[string]interface{} `json:"contact_info"` // 对应 JSON 字段
}

// =================================================================
// 2. 非遗元数据检索 (ICH Metadata)
// =================================================================

// RegionQueryRequest 地区筛选
type RegionQueryRequest struct {
	ParentID uint `form:"parent_id"`
	Level    int8 `form:"level"`
}

// CraftSearchRequest 技艺搜索与筛选
type CraftSearchRequest struct {
	Keyword    string `form:"keyword"`     // 模糊搜索工艺名称
	CategoryID uint   `form:"category_id"` // 按分类筛选
	Difficulty int8   `form:"difficulty"`  // 按难度筛选
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"page_size,default=20"`
}

// =================================================================
// 3. 作品/内容发布模块 (Work/Content)
// =================================================================

// CreateWorkRequest 发布新作品/动态
type CreateWorkRequest struct {
	Title         string      `json:"title" binding:"required,max=200"`
	Content       string      `json:"content" binding:"required"`
	ContentType   int8        `json:"content_type" binding:"required,oneof=1 2 3"`
	CraftID       uint        `json:"craft_id"`
	CategoryID    uint        `json:"category_id"`
	RegionID      uint        `json:"region_id"`
	TechniqueTags []string    `json:"technique_tags"` // 技艺标签
	Materials     []string    `json:"materials"`      // 材料
	CreationTime  *time.Time  `json:"creation_time"`  // 作品实际创作时间
	Media         []MediaItem `json:"media" binding:"required,dive"`
}

type MediaItem struct {
	URL          string `json:"url" binding:"required,url"`
	MediaType    int8   `json:"media_type" binding:"required,oneof=1 2 3 4"`
	ThumbnailURL string `json:"thumbnail_url"`
	Description  string `json:"description"`
	SortOrder    int8   `json:"sort_order"`
}

// WorkListRequest 作品列表查询（带复杂筛选）
type WorkListRequest struct {
	UserID   uint   `query:"user_id"`
	CraftID  uint   `query:"craft_id"`
	RegionID uint   `query:"region_id"`
	IsMaster bool   `query:"is_master"` // 是否只看大师作品
	OrderBy  string `query:"order_by" binding:"oneof=newest hot weight"`
	Page     int    `query:"page"`
	PageSize int    `query:"page_size"`
}

// =================================================================
// 4. 互动系统 (Interaction)
// =================================================================

// CreateCommentRequest 发表评论
type CreateCommentRequest struct {
	WorkID   uint   `json:"work_id" binding:"required"`
	ParentID uint   `json:"parent_id"` // 回复某条评论
	RootID   uint   `json:"root_id"`   // 所属楼层根 ID
	Content  string `json:"content" binding:"required,max=1000"`
	MediaURL string `json:"media_url"`
}

// ToggleLikeRequest 点赞/取消点赞
type ToggleLikeRequest struct {
	TargetID   uint `json:"target_id" binding:"required"`
	TargetType int8 `json:"target_type" binding:"required,oneof=1 2"` // 1:作品, 2:评论
}

// FavoriteRequest 收藏操作
type FavoriteRequest struct {
	WorkID   uint   `json:"work_id" binding:"required"`
	FolderID uint   `json:"folder_id"`
	Remark   string `json:"remark"`
}

// FollowRequest 关注/取消关注
type FollowRequest struct {
	FollowingID uint `json:"following_id" binding:"required"`
}

// =================================================================
// 5. 管理后台/审核模块 (Admin/Audit)
// =================================================================

// UserAuditRequest 传承人认证审核
type UserAuditRequest struct {
	UserID uint   `json:"user_id" binding:"required"`
	Status int8   `json:"status" binding:"required,oneof=0 1"` // 0:拒绝, 1:通过
	Reason string `json:"reason"`                              // 拒绝原因
}

// ContentAuditRequest 作品审核
type ContentAuditRequest struct {
	WorkID uint `json:"work_id" binding:"required"`
	Status int8 `json:"status" binding:"required,oneof=1 3 4"` // 1:发布, 3:拒绝, 4:下架
}
