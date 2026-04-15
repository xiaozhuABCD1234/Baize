package errors

import "errors"

// User errors
var (
	ErrUserNotFound    = errors.New("用户不存在")
	ErrEmailExists     = errors.New("邮箱已被注册")
	ErrUsernameExists  = errors.New("用户名已被使用")
	ErrInvalidPassword = errors.New("密码错误")
	ErrSamePassword    = errors.New("新密码不能与旧密码相同")
	ErrInvalidRole     = errors.New("无效的 UserType")
	ErrEmailNotChanged = errors.New("新邮箱与当前邮箱相同")
)

// Work errors
var (
	ErrWorkNotFound      = errors.New("作品不存在")
	ErrWorkMediaNotFound = errors.New("作品媒体不存在")
	ErrInvalidWorkStatus = errors.New("无效的作品状态")
	ErrCannotDeleteWork  = errors.New("无法删除已发布的作品")
)

// Resource errors
var (
	ErrCraftNotFound       = errors.New("技艺不存在")
	ErrCategoryNotFound    = errors.New("分类不存在")
	ErrRegionNotFound      = errors.New("地区不存在")
	ErrICHCategoryNotFound = errors.New("非遗分类不存在")
	ErrCommentNotFound     = errors.New("评论不存在")
	ErrFavoriteNotFound    = errors.New("收藏不存在")
	ErrFollowNotFound      = errors.New("关注关系不存在")
	ErrUserProfileNotFound = errors.New("用户档案不存在")
)

// Business logic errors
var (
	ErrCannotFollowSelf = errors.New("不能关注自己")
	ErrAlreadyFollowing = errors.New("已经关注过该用户")
	ErrNotFollowing     = errors.New("未关注该用户")
	ErrAlreadyFavorited = errors.New("已经收藏过该作品")
	ErrNotFavorited     = errors.New("未收藏该作品")
	ErrCannotReplyChild = errors.New("不能回复二级及以下的评论")
	ErrInvalidStatus    = errors.New("无效的评论状态")
)

// Region errors
var (
	ErrRegionCodeExists   = errors.New("地区代码已存在")
	ErrInvalidRegionLevel = errors.New("无效的地区级别")
	ErrCannotDeleteRegion = errors.New("无法删除有子节点的地区")
)

// Category errors
var (
	ErrCategoryNameExists   = errors.New("分类名称已存在")
	ErrInvalidCategoryLevel = errors.New("无效的分类级别")
	ErrCannotDeleteCategory = errors.New("无法删除有子节点的分类")
)

// Craft errors
var (
	ErrCraftNameExists   = errors.New("技艺名称已存在")
	ErrInvalidDifficulty = errors.New("无效的难度等级")
)
