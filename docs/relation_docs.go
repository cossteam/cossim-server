// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplaterelation = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/relation/delete_blacklist": {
            "post": {
                "description": "删除黑名单",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "UserRelation"
                ],
                "summary": "删除黑名单",
                "parameters": [
                    {
                        "description": "request",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.DeleteBlacklistRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.Response"
                        }
                    }
                }
            }
        },
        "/relation/group/admin/manage/join": {
            "post": {
                "description": "管理员管理加入群聊 action (0=拒绝, 1=同意)",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "GroupRelation"
                ],
                "summary": "管理员管理加入群聊",
                "parameters": [
                    {
                        "description": "Action (0: rejected, 1: joined)",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.AdminManageJoinGroupRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.Response"
                        }
                    }
                }
            }
        },
        "/relation/group/admin/manage/remove": {
            "post": {
                "description": "将用户从群聊移除",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "GroupRelation"
                ],
                "summary": "将用户从群聊移除",
                "parameters": [
                    {
                        "description": "request",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.RemoveUserFromGroupRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.Response"
                        }
                    }
                }
            }
        },
        "/relation/group/invite": {
            "post": {
                "description": "邀请加入群聊",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "GroupRelation"
                ],
                "summary": "邀请加入群聊",
                "parameters": [
                    {
                        "description": "request",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.InviteGroupRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.Response"
                        }
                    }
                }
            }
        },
        "/relation/group/join": {
            "post": {
                "description": "加入群聊",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "GroupRelation"
                ],
                "summary": "加入群聊",
                "parameters": [
                    {
                        "description": "request",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.JoinGroupRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.Response"
                        }
                    }
                }
            }
        },
        "/relation/group/list": {
            "get": {
                "description": "群聊列表",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "GroupRelation"
                ],
                "summary": "群聊列表",
                "responses": {
                    "200": {
                        "description": "status 0:正常状态；1:被封禁状态；2:被删除状态",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/model.Response"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "data": {
                                            "type": "array",
                                            "items": {
                                                "$ref": "#/definitions/usersorter.CustomGroupData"
                                            }
                                        }
                                    }
                                }
                            ]
                        }
                    }
                }
            }
        },
        "/relation/group/manage_join": {
            "post": {
                "description": "用户管理加入群聊 action (0=拒绝, 1=同意)",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "GroupRelation"
                ],
                "summary": "用户管理加入群聊",
                "parameters": [
                    {
                        "description": "Action (0: rejected, 1: joined)",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.ManageJoinGroupRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.Response"
                        }
                    }
                }
            }
        },
        "/relation/group/member": {
            "get": {
                "description": "群聊成员列表",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "GroupRelation"
                ],
                "summary": "群聊成员列表",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "群聊ID",
                        "name": "group_id",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.Response"
                        }
                    }
                }
            }
        },
        "/relation/group/quit": {
            "post": {
                "description": "退出群聊",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "GroupRelation"
                ],
                "summary": "退出群聊",
                "parameters": [
                    {
                        "description": "request",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.QuitGroupRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.Response"
                        }
                    }
                }
            }
        },
        "/relation/group/request_list": {
            "get": {
                "security": [
                    {
                        "Bearer": []
                    }
                ],
                "description": "获取用户的群聊申请列表",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "GroupRelation"
                ],
                "summary": "获取群聊申请列表",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Bearer JWT",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "status (0=申请中, 1=待通过, 2=已加入, 3=已删除, 4=被拒绝, 5=被封禁)",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/model.Response"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "data": {
                                            "$ref": "#/definitions/model.GroupRequestListResponse"
                                        }
                                    }
                                }
                            ]
                        }
                    }
                }
            }
        },
        "/relation/group/silent": {
            "post": {
                "description": "设置群聊静默通知",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "GroupRelation"
                ],
                "summary": "设置群聊静默通知",
                "parameters": [
                    {
                        "description": "request",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.SetGroupSilentNotificationRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.Response"
                        }
                    }
                }
            }
        },
        "/relation/user/add_blacklist": {
            "post": {
                "description": "添加黑名单",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "UserRelation"
                ],
                "summary": "添加黑名单",
                "parameters": [
                    {
                        "description": "request",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.AddBlacklistRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.Response"
                        }
                    }
                }
            }
        },
        "/relation/user/add_friend": {
            "post": {
                "description": "添加好友",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "UserRelation"
                ],
                "summary": "添加好友",
                "parameters": [
                    {
                        "description": "request",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.AddFriendRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.Response"
                        }
                    }
                }
            }
        },
        "/relation/user/blacklist": {
            "get": {
                "description": "黑名单",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "UserRelation"
                ],
                "summary": "黑名单",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.Response"
                        }
                    }
                }
            }
        },
        "/relation/user/delete_friend": {
            "post": {
                "description": "删除好友",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "UserRelation"
                ],
                "summary": "删除好友",
                "parameters": [
                    {
                        "description": "request",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.DeleteFriendRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.Response"
                        }
                    }
                }
            }
        },
        "/relation/user/friend_list": {
            "get": {
                "description": "好友列表",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "UserRelation"
                ],
                "summary": "好友列表",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.Response"
                        }
                    }
                }
            }
        },
        "/relation/user/manage_friend": {
            "post": {
                "description": "管理好友请求  action (0=拒绝, 1=同意)",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "UserRelation"
                ],
                "summary": "管理好友请求",
                "parameters": [
                    {
                        "description": "request",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.ManageFriendRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.Response"
                        }
                    }
                }
            }
        },
        "/relation/user/request_list": {
            "get": {
                "description": "好友申请列表",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "UserRelation"
                ],
                "summary": "好友申请列表",
                "responses": {
                    "200": {
                        "description": "UserStatus 申请状态 (0=申请中, 1=待通过, 2=已添加, 3=被拒绝, 4=已删除, 5=已拒绝)",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/model.Response"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "data": {
                                            "type": "array",
                                            "items": {
                                                "$ref": "#/definitions/model.UserRequestListResponse"
                                            }
                                        }
                                    }
                                }
                            ]
                        }
                    }
                }
            }
        },
        "/relation/user/silent": {
            "post": {
                "description": "设置私聊静默通知",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "UserRelation"
                ],
                "summary": "设置私聊静默通知",
                "parameters": [
                    {
                        "description": "request",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.SetUserSilentNotificationRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.Response"
                        }
                    }
                }
            }
        },
        "/relation/user/switch/e2e/key": {
            "post": {
                "security": [
                    {
                        "BearerToken": []
                    }
                ],
                "description": "交换用户端到端公钥",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "UserRelation"
                ],
                "summary": "交换用户端到端公钥",
                "parameters": [
                    {
                        "description": "request",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.SwitchUserE2EPublicKeyRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.Response"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "model.ActionEnum": {
            "type": "integer",
            "enum": [
                0,
                1
            ],
            "x-enum-comments": {
                "ActionAccepted": "同意",
                "ActionRejected": "拒绝"
            },
            "x-enum-varnames": [
                "ActionRejected",
                "ActionAccepted"
            ]
        },
        "model.AddBlacklistRequest": {
            "type": "object",
            "required": [
                "user_id"
            ],
            "properties": {
                "user_id": {
                    "type": "string"
                }
            }
        },
        "model.AddFriendRequest": {
            "type": "object",
            "required": [
                "user_id"
            ],
            "properties": {
                "e2e_public_key": {
                    "type": "string"
                },
                "msg": {
                    "type": "string"
                },
                "user_id": {
                    "type": "string"
                }
            }
        },
        "model.AdminManageJoinGroupRequest": {
            "type": "object",
            "required": [
                "group_id",
                "user_id"
            ],
            "properties": {
                "action": {
                    "$ref": "#/definitions/model.ActionEnum"
                },
                "group_id": {
                    "type": "integer"
                },
                "user_id": {
                    "type": "string"
                }
            }
        },
        "model.DeleteBlacklistRequest": {
            "type": "object",
            "required": [
                "user_id"
            ],
            "properties": {
                "user_id": {
                    "type": "string"
                }
            }
        },
        "model.DeleteFriendRequest": {
            "type": "object",
            "required": [
                "user_id"
            ],
            "properties": {
                "user_id": {
                    "type": "string"
                }
            }
        },
        "model.GroupRequestListResponse": {
            "type": "object",
            "properties": {
                "creator_id": {
                    "type": "string"
                },
                "group_avatar": {
                    "type": "string"
                },
                "group_id": {
                    "type": "integer"
                },
                "group_name": {
                    "type": "string"
                },
                "group_status": {
                    "type": "integer"
                },
                "group_type": {
                    "type": "integer"
                },
                "max_members_limit": {
                    "type": "integer"
                },
                "msg": {
                    "type": "string"
                },
                "user_avatar": {
                    "type": "string"
                },
                "user_id": {
                    "type": "string"
                },
                "user_name": {
                    "type": "string"
                }
            }
        },
        "model.InviteGroupRequest": {
            "type": "object",
            "required": [
                "group_id",
                "member"
            ],
            "properties": {
                "group_id": {
                    "type": "integer"
                },
                "member": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                }
            }
        },
        "model.JoinGroupRequest": {
            "type": "object",
            "required": [
                "group_id"
            ],
            "properties": {
                "group_id": {
                    "type": "integer"
                }
            }
        },
        "model.ManageFriendRequest": {
            "type": "object",
            "required": [
                "user_id"
            ],
            "properties": {
                "action": {
                    "$ref": "#/definitions/model.ActionEnum"
                },
                "e2e_public_key": {
                    "type": "string"
                },
                "user_id": {
                    "type": "string"
                }
            }
        },
        "model.ManageJoinGroupRequest": {
            "type": "object",
            "required": [
                "group_id"
            ],
            "properties": {
                "action": {
                    "$ref": "#/definitions/model.ActionEnum"
                },
                "group_id": {
                    "type": "integer"
                },
                "inviter_id": {
                    "type": "string"
                }
            }
        },
        "model.QuitGroupRequest": {
            "type": "object",
            "required": [
                "group_id"
            ],
            "properties": {
                "group_id": {
                    "type": "integer"
                }
            }
        },
        "model.RemoveUserFromGroupRequest": {
            "type": "object",
            "required": [
                "group_id",
                "user_id"
            ],
            "properties": {
                "group_id": {
                    "type": "integer"
                },
                "user_id": {
                    "type": "string"
                }
            }
        },
        "model.Response": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "integer"
                },
                "data": {},
                "msg": {
                    "type": "string"
                }
            }
        },
        "model.SetGroupSilentNotificationRequest": {
            "type": "object",
            "required": [
                "group_id"
            ],
            "properties": {
                "group_id": {
                    "description": "群ID",
                    "type": "integer"
                },
                "is_silent": {
                    "description": "是否开启静默通知",
                    "allOf": [
                        {
                            "$ref": "#/definitions/model.SilentNotificationType"
                        }
                    ]
                }
            }
        },
        "model.SetUserSilentNotificationRequest": {
            "type": "object",
            "required": [
                "user_id"
            ],
            "properties": {
                "is_silent": {
                    "description": "是否开启静默通知",
                    "allOf": [
                        {
                            "$ref": "#/definitions/model.SilentNotificationType"
                        }
                    ]
                },
                "user_id": {
                    "description": "用户ID",
                    "type": "string"
                }
            }
        },
        "model.SilentNotificationType": {
            "type": "integer",
            "enum": [
                0,
                1
            ],
            "x-enum-comments": {
                "IsSilent": "开启静默通知",
                "NotSilent": "静默通知关闭"
            },
            "x-enum-varnames": [
                "NotSilent",
                "IsSilent"
            ]
        },
        "model.SwitchUserE2EPublicKeyRequest": {
            "type": "object",
            "required": [
                "public_key",
                "user_id"
            ],
            "properties": {
                "public_key": {
                    "type": "string"
                },
                "user_id": {
                    "type": "string"
                }
            }
        },
        "model.UserRequestListResponse": {
            "type": "object",
            "properties": {
                "msg": {
                    "type": "string"
                },
                "request_at": {
                    "type": "string"
                },
                "user_avatar": {
                    "type": "string"
                },
                "user_id": {
                    "type": "string"
                },
                "user_name": {
                    "type": "string"
                },
                "user_status": {
                    "type": "integer"
                }
            }
        },
        "usersorter.CustomGroupData": {
            "type": "object",
            "properties": {
                "avatar": {
                    "type": "string"
                },
                "dialog_id": {
                    "type": "integer"
                },
                "group_id": {
                    "type": "integer"
                },
                "name": {
                    "type": "string"
                },
                "status": {
                    "type": "integer"
                }
            }
        }
    }
}`

// SwaggerInforelation holds exported Swagger Info so clients can modify it
var SwaggerInforelation = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "coss-relation-bff服务",
	Description:      "",
	InfoInstanceName: "relation",
	SwaggerTemplate:  docTemplaterelation,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInforelation.InstanceName(), SwaggerInforelation)
}
