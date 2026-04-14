package models

import (
	"time"
)

// ============================================================
// 用户相关 DTO
// ============================================================

// UserResponse 用户信息响应（脱敏后）
type UserResponse struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone,omitempty"`
	UserType string `json:"user_type"` // "user" / "master" / "institution" / "admin"
	Status   int8   `json:"status"`    // 0:禁用 1:正常 2:待审核

	// 关联档案（可选，在详情接口中填充）
	Profile *UserProfileResponse `json:"profile,omitempty"`
}

// UserProfileResponse 用户详细档案响应
type UserProfileResponse struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	UserID uint `json:"user_id"`

	Nickname  string `json:"nickname,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
	Bio       string `json:"bio,omitempty"`
	Address   string `json:"address,omitempty"`

	IsMaster     bool   `json:"is_master"`
	MasterCertNo string `json:"master_cert_no,omitempty"`
	WorkshopName string `json:"workshop_name,omitempty"`
	RegionID     uint   `json:"region_id,omitempty"`

	CraftIDs    map[string]interface{} `json:"craft_ids,omitempty"`    // JSON 数组，如 [1,2,3]
	ContactInfo map[string]interface{} `json:"contact_info,omitempty"` // JSON 对象
}

// ============================================================
// 非遗元数据 DTO
// ============================================================

// RegionResponse 行政区划/地域文化空间响应
type RegionResponse struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	ParentID         uint   `json:"parent_id"`
	Name             string `json:"name"`
	Code             string `json:"code"`
	Level            int8   `json:"level"`
	IsHeritageCenter bool   `json:"is_heritage_center"`
	CultureDesc      string `json:"culture_desc,omitempty"`
}

// ICHCategoryResponse 非遗分类响应
type ICHCategoryResponse struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	ParentID    uint   `json:"parent_id"`
	Name        string `json:"name"`
	Level       int8   `json:"level"`
	RegionCode  string `json:"region_code,omitempty"`
	Description string `json:"description,omitempty"`
	IconURL     string `json:"icon_url,omitempty"`
	SortOrder   int    `json:"sort_order"`
	Status      int8   `json:"status"`
}

// CraftResponse 手工艺/技艺细项响应
type CraftResponse struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	CategoryID  uint   `json:"category_id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	History     string `json:"history,omitempty"`

	Tools          map[string]interface{} `json:"tools,omitempty"` // JSON 数组
	Difficulty     int8                   `json:"difficulty"`
	RegionFeatures map[string]interface{} `json:"region_features,omitempty"` // JSON 对象

	// 关联分类信息（可选）
	Category *ICHCategoryResponse `json:"category,omitempty"`
}

// ============================================================
// 作品/内容相关 DTO
// ============================================================

// WorkResponse 作品/动态响应
type WorkResponse struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	UserID      uint   `json:"user_id"`
	Title       string `json:"title"`
	Content     string `json:"content,omitempty"`
	ContentType int8   `json:"content_type"` // 1:图文 2:视频 3:纯文本

	CraftID    uint `json:"craft_id,omitempty"`
	CategoryID uint `json:"category_id,omitempty"`
	RegionID   uint `json:"region_id,omitempty"`

	TechniqueTags map[string]interface{} `json:"technique_tags,omitempty"` // JSON 数组
	Materials     map[string]interface{} `json:"materials,omitempty"`      // JSON 数组
	CreationTime  *time.Time             `json:"creation_time,omitempty"`

	ViewCount     uint `json:"view_count"`
	LikeCount     uint `json:"like_count"`
	CommentCount  uint `json:"comment_count"`
	FavoriteCount uint `json:"favorite_count"`
	ShareCount    uint `json:"share_count"`

	Status        int8       `json:"status"` // 0:草稿 1:已发布 2:审核中 3:未通过 4:下架
	IsTop         bool       `json:"is_top"`
	IsRecommended bool       `json:"is_recommended"`
	Weight        int        `json:"weight"`
	PublishedAt   *time.Time `json:"published_at,omitempty"`

	// 关联信息（可选）
	User  *UserResponse       `json:"user,omitempty"`
	Craft *CraftResponse      `json:"craft,omitempty"`
	Media []WorkMediaResponse `json:"media,omitempty"`
}

// WorkMediaResponse 作品附件响应
type WorkMediaResponse struct {
	ID           uint   `json:"id"`
	MediaType    int8   `json:"media_type"` // 1:图片 2:视频 3:音频 4:文档
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	Width        int    `json:"width,omitempty"`
	Height       int    `json:"height,omitempty"`
	Duration     int    `json:"duration,omitempty"`
	FileSize     int64  `json:"file_size,omitempty"`
	SortOrder    int8   `json:"sort_order"`
	Description  string `json:"description,omitempty"`
}

// ============================================================
// 互动系统 DTO
// ============================================================

// CommentResponse 评论响应（支持嵌套）
type CommentResponse struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`

	WorkID    uint   `json:"work_id"`
	UserID    uint   `json:"user_id"`
	ParentID  uint   `json:"parent_id,omitempty"`
	RootID    uint   `json:"root_id,omitempty"`
	Content   string `json:"content"`
	MediaURL  string `json:"media_url,omitempty"`
	LikeCount uint   `json:"like_count"`
	Status    int8   `json:"status"` // 0:删除 1:正常 2:审核中

	// 关联用户信息（可选）
	User *UserResponse `json:"user,omitempty"`
	// 子回复列表（可选，用于树形结构）
	Replies []CommentResponse `json:"replies,omitempty"`
}

// LikeResponse 点赞记录响应
type LikeResponse struct {
	UserID     uint      `json:"user_id"`
	TargetType int8      `json:"target_type"` // 1:作品 2:评论
	TargetID   uint      `json:"target_id"`
	CreatedAt  time.Time `json:"created_at"`
}

// FavoriteResponse 收藏夹响应
type FavoriteResponse struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	WorkID    uint      `json:"work_id"`
	FolderID  uint      `json:"folder_id,omitempty"`
	Remark    string    `json:"remark,omitempty"`
	CreatedAt time.Time `json:"created_at"`

	// 关联作品信息（可选）
	Work *WorkResponse `json:"work,omitempty"`
}

// FollowResponse 关注关系响应
type FollowResponse struct {
	FollowerID  uint      `json:"follower_id"`
	FollowingID uint      `json:"following_id"`
	CreatedAt   time.Time `json:"created_at"`
}
