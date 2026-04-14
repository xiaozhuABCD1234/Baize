package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// =================================================================
// 用户体系：管理平台基础用户及非遗传承人核心身份
// =================================================================

// User 核心用户表
// 存储账户鉴权信息及基础状态
type User struct {
	gorm.Model

	Username string `gorm:"uniqueIndex;size:32;not null"` // 用户登录名，设为唯一索引
	Email    string `gorm:"uniqueIndex;size:128;not null"`
	Password string `gorm:"not null"` // 加密后的密码散列值
	Phone    string `gorm:"uniqueIndex;size:20"`

	// 权限与状态
	UserType UserType   `gorm:"size:20;default:'user';index"` // 用户类型：普通/大师/机构/管理员
	Status   UserStatus `gorm:"default:1;index"`              // 状态：禁用/激活/待审核

	// 关联：一对一扩展档案（使用级联删除保证数据完整性）
	Profile *UserProfile `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
}

// UserType 用户类型枚举值
type UserType string

const (
	UserTypeUser        UserType = "user"        // 普通爱好者
	UserTypeMaster      UserType = "master"      // 非遗传承人（大师）
	UserTypeInstitution UserType = "institution" // 非遗保护机构/工坊
	UserTypeAdmin       UserType = "admin"       // 平台管理人员
)

// UserStatus 用户账号生命周期状态
type UserStatus int8

const (
	UserStatusDisabled UserStatus = 0 // 封禁状态
	UserStatusActive   UserStatus = 1 // 正常可用
	UserStatusPending  UserStatus = 2 // 注册待激活或资料审核中
)

// UserProfile 用户详细档案
// 垂直拆分主表，存储非高频变动的非遗特色扩展信息
type UserProfile struct {
	ID     uint `gorm:"primarykey"`
	UserID uint `gorm:"uniqueIndex;not null"` // 关联 User.ID

	Nickname  string `gorm:"size:64"`
	AvatarURL string `gorm:"size:500"`
	Bio       string `gorm:"type:text"` // 个人简介/传承宣言
	Address   string `gorm:"size:255"`

	// 非遗特色属性
	IsMaster     bool   `gorm:"default:false;index:idx_profile_master"` // 是否为认证传承人
	MasterCertNo string `gorm:"size:64;index"`                          // 非遗传承人证书编号
	WorkshopName string `gorm:"size:128"`                               // 个人工作室/工坊名称
	RegionID     uint   `gorm:"index:idx_profile_region"`               // 所属行政区划 ID

	// 灵活字段：采用 JSON 存储
	CraftIDs    datatypes.JSON `gorm:"type:json"` // 掌握的工艺 ID 列表 (e.g. [1, 5, 12])
	ContactInfo datatypes.JSON `gorm:"type:json"` // 社交媒体、备用联系方式等扩展信息
}

// =================================================================
// 非遗元数据：构建非遗知识图谱的基础（地区、分类、工艺）
// =================================================================

// Region 行政区划与地域文化空间
// 用于标记非遗项目的发源地或流传地
type Region struct {
	gorm.Model

	ParentID         uint   `gorm:"default:0;index:idx_region_parent"` // 上级区域 ID
	Name             string `gorm:"size:64;not null"`
	Code             string `gorm:"uniqueIndex;size:20"`    // 行政区划代码（如 110101）
	Level            int8   `gorm:"index:idx_region_level"` // 级别：省/市/县
	IsHeritageCenter bool   `gorm:"default:false"`          // 是否为非遗重点保护区域
	CultureDesc      string `gorm:"type:text"`              // 地域文化背景描述
}

// ICHCategory 非遗项目分类体系
// 参考：民间文学、传统音乐、传统美术、传统技艺等
type ICHCategory struct {
	gorm.Model

	ParentID    uint   `gorm:"default:0;index:idx_category_parent"`
	Name        string `gorm:"size:64;not null"`
	Level       int8   `gorm:"index:idx_category_level"`
	RegionCode  string `gorm:"size:20;index:idx_category_region"` // 分类的地域特色标识
	Description string `gorm:"type:text"`
	IconURL     string `gorm:"size:500"`
	SortOrder   int    `gorm:"default:0;index:idx_category_sort"` // 排序权重
	Status      int8   `gorm:"default:1"`                         // 分类启用状态
}

// Craft 具体手工艺/技艺细项
type Craft struct {
	gorm.Model

	CategoryID  uint   `gorm:"index:idx_craft_category"`
	Name        string `gorm:"size:64;not null;index:idx_craft_name"`
	Description string `gorm:"type:text"` // 工艺概况
	History     string `gorm:"type:text"` // 历史沿革

	Tools          datatypes.JSON `gorm:"type:json"` // 所需核心工具列表
	Difficulty     int8           `gorm:"default:0"` // 技艺难度等级
	RegionFeatures datatypes.JSON `gorm:"type:json"` // 不同地域的表现形式差异

	Category ICHCategory `gorm:"foreignKey:CategoryID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

// =================================================================
// 作品/内容核心：论坛内容生产与社区互动
// =================================================================

// Work 核心作品/动态表
// 包含非遗作品展示、技艺教学视频或社区动态
type Work struct {
	gorm.Model

	UserID      uint        `gorm:"not null;index:idx_work_user_status"`
	Title       string      `gorm:"size:200;not null;index:idx_work_title"`
	Content     string      `gorm:"type:text"`
	ContentType ContentType `gorm:"default:1"` // 内容表现形式

	// 非遗元数据关联：将内容挂载到具体的技艺和地域下
	CraftID    uint `gorm:"index:idx_work_craft_status"`
	CategoryID uint `gorm:"index:idx_work_category"`
	RegionID   uint `gorm:"index:idx_work_region"`

	TechniqueTags datatypes.JSON `gorm:"type:json"`     // 技艺标签（如：苏绣、乱针绣）
	Materials     datatypes.JSON `gorm:"type:json"`     // 涉及材料（如：蚕丝、丝绒）
	CreationTime  *time.Time     `gorm:"type:datetime"` // 作品实际创作时间（非发布时间）

	// 计数统计：冗余设计以支持高性能列表查询
	ViewCount     uint `gorm:"default:0"`
	LikeCount     uint `gorm:"default:0"`
	CommentCount  uint `gorm:"default:0"`
	FavoriteCount uint `gorm:"default:0"`
	ShareCount    uint `gorm:"default:0"`

	// 状态与分发管理
	Status        WorkStatus `gorm:"default:1;index:idx_work_user_status,idx_work_craft_status,idx_work_status_created"`
	IsTop         bool       `gorm:"default:false;index:idx_work_top"`         // 是否置顶
	IsRecommended bool       `gorm:"default:false;index:idx_work_recommended"` // 是否入选精选推荐
	Weight        int        `gorm:"default:0;index:idx_work_weight"`          // 排序权重
	PublishedAt   *time.Time `gorm:"index:idx_work_published"`                 // 正式发布时间

	// 关联对象
	User  User        `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Craft Craft       `gorm:"foreignKey:CraftID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Media []WorkMedia `gorm:"foreignKey:WorkID;constraint:OnDelete:CASCADE;"`
}

// ContentType 标识作品的主体形式
type ContentType int8

const (
	ContentTypeImage ContentType = 1 // 图文动态
	ContentTypeVideo ContentType = 2 // 短视频/纪录片
	ContentTypeText  ContentType = 3 // 纯文本专栏
)

// WorkStatus 作品生命周期
type WorkStatus int8

const (
	WorkStatusDraft     WorkStatus = 0 // 草稿箱
	WorkStatusPublished WorkStatus = 1 // 已公开发布
	WorkStatusReviewing WorkStatus = 2 // 审核中（敏感词/合规检查）
	WorkStatusRejected  WorkStatus = 3 // 审核未通过
	WorkStatusOffline   WorkStatus = 4 // 管理员下架
)

// WorkMedia 作品附件库
// 支持一个作品拥有多张图片、视频或音频说明
type WorkMedia struct {
	ID           uint      `gorm:"primarykey"`
	WorkID       uint      `gorm:"not null;index:idx_media_work"`
	MediaType    MediaType `gorm:"default:1"`
	URL          string    `gorm:"size:500;not null"` // 媒体存储地址 (CDN)
	ThumbnailURL string    `gorm:"size:500"`          // 视频封面或图片缩略图
	Width        int       `gorm:"default:0"`
	Height       int       `gorm:"default:0"`
	Duration     int       `gorm:"default:0"` // 视音频时长（秒）
	FileSize     int64     `gorm:"default:0"`
	SortOrder    int8      `gorm:"default:0"` // 媒体排序（如长图文中的图片顺序）
	Description  string    `gorm:"size:255"`  // 单张媒体的文字说明
}

type MediaType int8

const (
	MediaTypeImage MediaType = 1
	MediaTypeVideo MediaType = 2
	MediaTypeAudio MediaType = 3
	MediaTypeDoc   MediaType = 4 // PDF/文档（常见于研究报告）
)

// =================================================================
// 互动系统：评论、点赞、关注
// =================================================================

// Comment 评论系统
// 支持两级嵌套（根评论 + 子回复）
type Comment struct {
	ID        uint      `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"type:datetime"`

	WorkID    uint          `gorm:"not null;index:idx_comment_work"`
	UserID    uint          `gorm:"not null;index:idx_comment_user"`
	ParentID  uint          `gorm:"default:0;index:idx_comment_parent"` // 被回复的评论 ID
	RootID    uint          `gorm:"default:0;index:idx_comment_root"`   // 属于哪条顶层评论
	Content   string        `gorm:"type:text;not null"`
	MediaURL  string        `gorm:"size:500"` // 评论支持带图（如交流作品细节）
	LikeCount uint          `gorm:"default:0"`
	Status    CommentStatus `gorm:"default:1;index:idx_comment_status"`

	User User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type CommentStatus int8

const (
	CommentStatusDeleted   CommentStatus = 0
	CommentStatusActive    CommentStatus = 1
	CommentStatusReviewing CommentStatus = 2
)

// Like 点赞记录
// 采用复合主键 (UserID, TargetType, TargetID) 防止重复点赞并优化查询性能
type Like struct {
	UserID     uint      `gorm:"primaryKey"`
	TargetType int8      `gorm:"primaryKey"` // 1:作品, 2:评论
	TargetID   uint      `gorm:"primaryKey"`
	CreatedAt  time.Time `gorm:"type:datetime"`
}

// Favorite 收藏夹
type Favorite struct {
	ID        uint      `gorm:"primarykey"`
	UserID    uint      `gorm:"not null;uniqueIndex:uk_favorite_user_work"`
	WorkID    uint      `gorm:"not null;uniqueIndex:uk_favorite_user_work"`
	FolderID  uint      `gorm:"default:0;index:idx_favorite_folder"` // 支持用户分类收藏
	Remark    string    `gorm:"size:255"`                            // 收藏时的个人备注
	CreatedAt time.Time `gorm:"type:datetime"`

	Work Work `gorm:"foreignKey:WorkID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// Follow 社交关系：关注传承人或同好
type Follow struct {
	FollowerID  uint      `gorm:"primaryKey"` // 发起关注者
	FollowingID uint      `gorm:"primaryKey"` // 被关注者
	CreatedAt   time.Time `gorm:"type:datetime"`
}

// =================================================================
// 非遗特色：传承谱系 (Lineage)
// 记录师傅与徒弟之间的技艺传承关系
// =================================================================

// type Lineage struct {
// 	gorm.Model

// 	CraftID        uint          `gorm:"not null;index:idx_lineage_craft"`  // 传承的技艺项
// 	MasterID       uint          `gorm:"not null;index:idx_lineage_master"` // 师傅 ID (User)
// 	ApprenticeID   uint          `gorm:"not null;uniqueIndex"`              // 徒弟 ID (User)
// 	Generation     int8          `gorm:"default:0"`                         // 传承代数（如：第 5 代传人）
// 	StartDate      *time.Time    `gorm:"type:datetime"`                     // 拜师/入门时间
// 	EndDate        *time.Time    `gorm:"type:datetime"`                     // 出师时间
// 	CertificateURL string        `gorm:"size:500"`                          // 传承证明/证书
// 	Status         LineageStatus `gorm:"default:1"`                         // 状态：在教/结业/中止

// 	Master     User `gorm:"foreignKey:MasterID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
// 	Apprentice User `gorm:"foreignKey:ApprenticeID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
// }

// type LineageStatus int8

// const (
// 	LineageStatusActive     LineageStatus = 1 // 传承中
// 	LineageStatusGraduated  LineageStatus = 2 // 已结业/已出师
// 	LineageStatusTerminated LineageStatus = 3 // 关系终止
// )

// =================================================================
// 数据库迁移逻辑
// =================================================================

// AutoMigrate 自动迁移所有模型
// 注意：迁移前请确保数据库字符集已设置为 utf8mb4 以支持 Emoji 表情
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&UserProfile{},
		&Region{},
		&ICHCategory{},
		&Craft{},
		&Work{},
		&WorkMedia{},
		&Comment{},
		&Like{},
		&Favorite{},
		&Follow{},
		// &Lineage{},
	)
}
