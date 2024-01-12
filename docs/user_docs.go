// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplateuser = `{
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
        "/user/info": {
            "get": {
                "description": "查询用户信息",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "查询用户信息",
                "parameters": [
                    {
                        "type": "string",
                        "description": "用户id",
                        "name": "user_id",
                        "in": "query",
                        "required": true
                    },
                    {
                        "enum": [
                            0,
                            1
                        ],
                        "type": "integer",
                        "description": "指定根据id还是邮箱类型查找",
                        "name": "type",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "邮箱",
                        "name": "email",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/utils.Response"
                        }
                    }
                }
            }
        },
        "/user/info/modify": {
            "post": {
                "security": [
                    {
                        "BearerToken": []
                    }
                ],
                "description": "修改用户信息",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "修改用户信息",
                "parameters": [
                    {
                        "description": "request",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/http.UserInfoRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/utils.Response"
                        }
                    }
                }
            }
        },
        "/user/key/set": {
            "post": {
                "security": [
                    {
                        "BearerToken": []
                    }
                ],
                "description": "设置用户公钥",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "设置用户公钥",
                "parameters": [
                    {
                        "description": "request",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/http.SetPublicKeyRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/utils.Response"
                        }
                    }
                }
            }
        },
        "/user/login": {
            "post": {
                "description": "用户登录",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "用户登录",
                "parameters": [
                    {
                        "description": "request",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/http.LoginRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/utils.Response"
                        }
                    }
                }
            }
        },
        "/user/password/modify": {
            "post": {
                "security": [
                    {
                        "BearerToken": []
                    }
                ],
                "description": "修改用户密码",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "修改用户密码",
                "parameters": [
                    {
                        "description": "request",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/http.PasswordRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/utils.Response"
                        }
                    }
                }
            }
        },
        "/user/register": {
            "post": {
                "description": "用户注册",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "用户注册",
                "parameters": [
                    {
                        "description": "request",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/http.RegisterRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/utils.Response"
                        }
                    }
                }
            }
        },
        "/user/system/key/get": {
            "get": {
                "description": "获取系统pgp公钥",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "获取系统pgp公钥",
                "parameters": [
                    {
                        "enum": [
                            0,
                            1
                        ],
                        "type": "integer",
                        "description": "指定根据id还是邮箱类型查找",
                        "name": "type",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "邮箱",
                        "name": "email",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/utils.Response"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "http.LoginRequest": {
            "type": "object",
            "required": [
                "email",
                "password",
                "public_key"
            ],
            "properties": {
                "email": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                },
                "public_key": {
                    "type": "string"
                }
            }
        },
        "http.PasswordRequest": {
            "type": "object",
            "required": [
                "confirm_password",
                "old_password",
                "password"
            ],
            "properties": {
                "confirm_password": {
                    "type": "string"
                },
                "old_password": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                }
            }
        },
        "http.RegisterRequest": {
            "type": "object",
            "required": [
                "confirm_password",
                "email",
                "password",
                "public_key"
            ],
            "properties": {
                "confirm_password": {
                    "type": "string"
                },
                "email": {
                    "type": "string"
                },
                "nickname": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                },
                "public_key": {
                    "type": "string"
                }
            }
        },
        "http.SetPublicKeyRequest": {
            "type": "object",
            "required": [
                "public_key"
            ],
            "properties": {
                "public_key": {
                    "type": "string"
                }
            }
        },
        "http.UserInfoRequest": {
            "type": "object",
            "properties": {
                "avatar": {
                    "type": "string"
                },
                "nickname": {
                    "type": "string"
                },
                "signature": {
                    "type": "string"
                },
                "status": {
                    "type": "integer"
                },
                "tel": {
                    "type": "string"
                }
            }
        },
        "utils.Response": {
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
        }
    }
}`

// SwaggerInfouser holds exported Swagger Info so clients can modify it
var SwaggerInfouser = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "coss-user服务",
	Description:      "",
	InfoInstanceName: "user",
	SwaggerTemplate:  docTemplateuser,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfouser.InstanceName(), SwaggerInfouser)
}
