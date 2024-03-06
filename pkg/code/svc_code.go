package code

var (
	// 用户服务状态码定义
	UserErrNotExistOrPassword                    = New(10000, "用户不存在或密码错误")
	UserErrLoginFailed                           = New(10001, "登录失败，请重试")
	UserErrLocked                                = New(10002, "用户已被锁定")
	UserErrDeleted                               = New(10003, "用户已被删除")
	UserErrDisabled                              = New(10004, "用户已被禁用")
	UserErrStatusException                       = New(10005, "用户状态异常")
	UserErrEmailAlreadyRegistered                = New(10006, "邮箱已被注册")
	UserErrRegistrationFailed                    = New(10007, "注册失败，请重试")
	UserErrGetUserInfoFailed                     = New(10008, "获取用户信息失败，请重试")
	UserErrUnableToGetUserListInfo               = New(10009, "无法获取用户列表信息")
	UserErrPublicKeyNotExist                     = New(10010, "用户公钥不存在")
	UserErrGetUserPublicKeyFailed                = New(10011, "获取用户公钥失败，请重试")
	UserErrSaveUserPublicKeyFailed               = New(10012, "保存用户公钥失败，请重试")
	UserErrNotExist                              = New(10013, "用户不存在")
	UserErrGetUserSecretBundleFailed             = New(10014, "获取用户秘钥包失败，请重试")
	UserErrSetUserSecretBundleFailed             = New(10015, "设置用户秘钥包失败，请重试")
	UserErrErrLogoutFailed                       = New(10016, "退出登录失败，请重试")
	UserErrOldPassword                           = New(10017, "旧密码错误")
	UserErrSwapPublicKeyFailed                   = New(10018, "用户交换公钥失败")
	UserErrGetUserLoginClientsFailed             = New(10019, "获取用户登录客户端失败，请重试")
	UserErrActivateUserFailed                    = New(10020, "激活用户失败")
	UserErrPasswordNotMatch                      = New(10021, "两次输入密码不匹配")
	UserErrNewPasswordAndOldPasswordEqual        = New(10022, "新密码和旧密码一致")
	UserErrResetPublicKeyFailed                  = New(10023, "重置公钥失败")
	UserErrSendEmailCodeFailed                   = New(10024, "发送邮箱验证码失败")
	UserErrCreateUserFailed                      = New(10025, "创建用户失败")
	UserErrCreateUserRollbackFailed              = New(10026, "创建用户回滚失败")
	UserErrUpdateUserLoginTokenFailed            = New(10027, "更新用户登录令牌失败")
	UserErrGetUserLoginByDriverIdAndUserIdFailed = New(10028, "获取用户登录信息失败")
	UserErrGetUserDriverTokenByUserIdFailed      = New(10029, "获取用户登录设备令牌失败")

	// 文件存储服务状态码定义
	StorageErrParseFilePathFailed    = New(11000, "解析文件路径失败")
	StorageErrCreateFileRecordFailed = New(11001, "保存文件失败")
	StorageErrGetFileInfoFailed      = New(11002, "获取文件信息失败")
	StorageErrDeleteFileFailed       = New(11003, "删除文件失败")

	// 关系服务状态码定义
	RelationErrUserNotFound                             = New(13000, "用户不存在")
	RelationUserErrFriendRelationNotFound               = New(13001, "好友关系不存在")
	RelationErrRelationNotFound                         = New(13002, "关系不存在")
	RelationErrAddFriendFailed                          = New(13003, "添加好友失败")
	RelationErrFriendRequestAlreadyPending              = New(13004, "好友请求已经发送，等待确认")
	RelationErrAlreadyFriends                           = New(13005, "已经是好友")
	RelationErrConfirmFriendFailed                      = New(13006, "管理好友申请失败")
	RelationErrDeleteFriendFailed                       = New(13007, "删除好友失败")
	RelationErrAddBlacklistFailed                       = New(13008, "添加黑名单失败")
	RelationErrDeleteBlacklistFailed                    = New(13009, "删除黑名单失败")
	RelationErrGetFriendListFailed                      = New(13010, "获取好友列表失败")
	RelationErrGetBlacklistFailed                       = New(13011, "获取黑名单列表失败")
	RelationErrGetUserRelationFailed                    = New(13012, "获取用户关系失败")
	RelationUserErrGetRequestListFailed                 = New(13013, "获取用户申请列表失败")
	RelationUserErrNoFriendRequestRecords               = New(13014, "未找到好友申请记录")
	RelationErrRejectFriendFailed                       = New(13015, "拒绝好友申请失败")
	RelationErrSetUserFriendSilentNotificationFailed    = New(13016, "设置用户好友静默通知失败")
	RelationErrNotInBlacklist                           = New(13017, "没有在黑名单")
	RelationErrSendFriendRequestFailed                  = New(13018, "发送好友请求失败")
	RelationErrManageFriendRequestFailed                = New(13019, "管理好友请求失败")
	RelationErrRequestAlreadyProcessed                  = New(13020, "该请求已处理")
	RelationErrSetUserOpenBurnAfterReadingFailed        = New(13021, "设置用户阅后即焚开关失败")
	RelationErrSetFriendRemarkFailed                    = New(13022, "设置好友备注失败")
	RelationErrSetUserOpenBurnAfterReadingTimeOutFailed = New(13023, "设置用户阅后即焚时间失败")

	RelationErrCreateGroupFailed                                  = New(13101, "创建群聊失败")
	RelationErrGetGroupIDsFailed                                  = New(13102, "获取群聊成员")
	RelationGroupErrRequestFailed                                 = New(13103, "申请加入群聊失败")
	RelationGroupErrRequestAlreadyPending                         = New(13104, "等待同意申请")
	RelationGroupErrAlreadyInGroup                                = New(13105, "已经在群聊中")
	RelationGroupErrApproveJoinGroupFailed                        = New(13106, "同意加入群聊失败")
	RelationGroupErrNoJoinRequestRecords                          = New(13107, "没有申请加入群聊记录")
	RelationGroupErrRejectJoinGroup                               = New(13108, "拒绝加入群聊")
	RelationGroupErrRemoveUserFromGroupFailed                     = New(13109, "将用户移除群聊失败")
	RelationGroupErrLeaveGroupFailed                              = New(13110, "退出群聊失败")
	RelationGroupErrGetJoinRequestListFailed                      = New(13111, "获取群聊申请列表失败")
	RelationGroupErrGroupRelationFailed                           = New(13112, "获取用户群组关系失败")
	RelationGroupErrGetGroupInfoFailed                            = New(13113, "获取群聊信息失败")
	RelationGroupErrManageJoinFailed                              = New(13114, "管理群聊申请失败")
	RelationGroupErrInviteFailed                                  = New(13115, "邀请入群失败")
	RelationGroupErrRelationNotFound                              = New(13116, "关系不存在")
	RelationGroupErrSetUserGroupSilentNotificationFailed          = New(13117, "设置群聊消息静默通知失败")
	RelationGroupErrNotInGroup                                    = New(13118, "没有在群聊中")
	RelationGroupErrDeleteUsersGroupRelationFailed                = New(13119, "删除多个用户群聊关系失败")
	RelationGroupRrrSetUserGroupOpenBurnAfterReadingFailed        = New(13120, "获取群聊阅后即焚开关失败")
	RelationGroupErrGetGroupAnnouncementListFailed                = New(13121, "获取群聊公告列表失败")
	RelationGroupErrCreateGroupAnnouncementFailed                 = New(13122, "创建群聊公告失败")
	RelationGroupErrGetGroupAnnouncementFailed                    = New(13123, "获取群聊公告失败")
	RelationGroupErrDeleteGroupAnnouncementFailed                 = New(13124, "删除群聊公告失败")
	RelationGroupErrUpdateGroupAnnouncementFailed                 = New(13125, "更新群聊公告失败")
	RelationGroupErrGroupAnnouncementNotFoundFailed               = New(13126, "群聊公告不存在")
	RelationGroupErrGroupFull                                     = New(13127, "群聊已满")
	RelationErrGroupRequestAlreadyProcessed                       = New(13128, "请求已存在")
	RelationErrGroupAnnouncementReadFailed                        = New(13129, "设置群聊公告已读失败")
	RelationErrGetGroupAnnouncementReadUsersFailed                = New(13130, "获取已读群聊公告用户失败")
	RelationErrGetGroupAnnouncementReadFailed                     = New(13131, "获取已读群聊公告失败")
	RelationErrGetGroupJoinRequestFailed                          = New(13132, "获取进群申请失败")
	RelationGroupErrSetUserGroupOpenBurnAfterReadingTimeOutFailed = New(13133, "获取群聊阅后即焚时间失败")

	// 消息服务错误码定义
	MsgErrInsertUserMessageFailed                   = New(14000, "发送消息失败")
	MsgErrInsertGroupMessageFailed                  = New(14001, "发送群聊消息失败")
	MsgErrGetUserMessageListFailed                  = New(14002, "获取用户消息列表失败")
	MsgErrGetLastMsgsForUserWithFriends             = New(14003, "获取消息失败")
	MsgErrGetLastMsgsForGroupsWithIDs               = New(14004, "获取群聊消息失败")
	MsgErrGetLastMsgsByDialogIds                    = New(14005, "获取消息失败")
	DialogErrCreateDialogFailed                     = New(14100, "创建对话失败")
	DialogErrJoinDialogFailed                       = New(14101, "加入对话失败")
	DialogErrGetUserDialogListFailed                = New(14102, "获取用户对话列表失败")
	DialogErrDeleteDialogFailed                     = New(14103, "删除对话失败")
	DialogErrDeleteDialogUsersFailed                = New(14104, "删除对话用户失败")
	DialogErrGetDialogUserByDialogIDAndUserIDFailed = New(14105, "获取对话用户失败")
	MsgErrEditUserMessageFailed                     = New(14006, "编辑用户消息失败")
	MsgErrEditGroupMessageFailed                    = New(14007, "编辑群聊消息失败")
	MsgErrDeleteUserMessageFailed                   = New(14008, "撤回用户消息失败")
	MsgErrDeleteGroupMessageFailed                  = New(14009, "撤回群聊消息失败")
	GetMsgErrGetGroupMsgByIDFailed                  = New(14010, "获取群聊消息失败")
	GetMsgErrGetUserMsgByIDFailed                   = New(14011, "获取用户消息失败")
	SetMsgErrSetUserMsgLabelFailed                  = New(14012, "设置用户消息标注失败")
	SetMsgErrSetGroupMsgLabelFailed                 = New(14013, "设置群聊消息标注失败")
	GetMsgErrGetUserMsgLabelByDialogIdFailed        = New(14014, "获取用户消息标注失败")
	GetMsgErrGetGroupMsgLabelByDialogIdFailed       = New(14015, "获取群聊消息标注失败")
	SetMsgErrSetUserMsgsReadStatusFailed            = New(14016, "批量已读消息失败")
	SetMsgErrSetUserMsgReadStatusFailed             = New(14017, "修改消息已读状态失败")
	GetMsgErrGetUnreadUserMsgsFailed                = New(14018, "获取未读消息失败")
	MsgErrTimeoutExceededCannotRevoke               = New(14019, "超过时间限制不能撤回")
	MsgErrGetGroupMsgListFailed                     = New(14020, "获取群聊消息列表失败")
	DialogErrCloseOrOpenDialogFailed                = New(14021, "关闭或打开对话失败")
	DialogErrGetDialogByIdFailed                    = New(14022, "获取对话信息失败")
	MsgErrSendMultipleFailed                        = New(14023, "批量发送消息失败")

	// 群组服务错误码定义
	GroupErrGetGroupInfoByGidFailed               = New(15000, "获取群聊信息失败")
	GroupErrGetBatchGroupInfoByIDsFailed          = New(15001, "获取群聊列表信息失败")
	GroupErrUpdateGroupFailed                     = New(15002, "更新群聊信息失败")
	GroupErrInsertGroupFailed                     = New(15003, "创建群组失败")
	GroupErrDeleteGroupFailed                     = New(15004, "删除群聊失败")
	GroupErrGroupNotFound                         = New(15005, "群聊不存在")
	GroupErrDeleteUserGroupRelationFailed         = New(15006, "删除用户群聊关系失败")
	GroupErrDeleteUserGroupRelationRevertFailed   = New(15007, "删除用户群聊关系回滚失败")
	GroupErrGroupStatusNotAvailable               = New(15008, "群聊状态不可用")
	GroupErrUserIsMuted                           = New(15009, "用户禁言中")
	GroupErrSetGroupMsgReadFailed                 = New(15010, "设置群聊消息已读状态失败")
	GroupErrGetGroupMsgReadersFailed              = New(15011, "获取群聊消息阅读者失败")
	GroupErrGetGroupMsgReadByMsgIdAndUserIdFailed = New(15012, "获取群聊消息阅读者失败")

	// 通话服务错误码定义
	LiveErrCreateCallFailed         = New(16000, "创建通话失败")
	LiveErrEndCallFailed            = New(16001, "结束通话失败")
	LiveErrGetCallInfoFailed        = New(16002, "获取通话信息失败")
	LiveErrGetUserCallListFailed    = New(16003, "获取用户通话列表失败")
	LiveErrGetGroupCallListFailed   = New(16004, "获取群聊通话列表失败")
	LiveErrGetCallByIdFailed        = New(16005, "获取通话信息失败")
	LiveErrJoinCallFailed           = New(16006, "加入通话失败")
	LiveErrLeaveCallFailed          = New(16007, "离开通话失败")
	LiveErrGetUserCallStatusFailed  = New(16008, "获取用户通话状态失败")
	LiveErrGetGroupCallStatusFailed = New(16009, "获取群聊通话状态失败")
	LiveErrUpdateCallStatusFailed   = New(16010, "更新通话状态失败")
	LiveErrCallNotFound             = New(16011, "通话不存在")
	LiveErrInvalidCallStatus        = New(16012, "无效的通话状态")
	LiveErrAlreadyInCall            = New(16013, "通话中")
	LiveErrUserNotInCall            = New(16014, "用户不在通话中")
	LiveErrGroupCallNotSupported    = New(16015, "群聊不支持通话")
	LiveErrMaxParticipantsExceeded  = New(16016, "超过通话最大参与人数")
	LiveErrInvalidParticipant       = New(16017, "无效的通话参与者")
	LiveErrInvalidMediaFormat       = New(16018, "无效的媒体格式")
	LiveErrMediaConnectionFailed    = New(16019, "媒体连接失败")
	LiveErrMediaPermissionDenied    = New(16020, "媒体权限被拒绝")
	LiveErrMediaTimeout             = New(16021, "媒体超时")
	LiveErrMediaDisconnected        = New(16022, "媒体断开连接")
	LiveErrMediaError               = New(16023, "媒体错误")
)
