coss-server
==============
`cossim/coss-server` 是用于支持coss-client的API服务。

---------------------------------------
* [特性](#特性)
* [服务](#服务)
* [快速启动](#快速启动)
* [配置](#配置)
* [文档](#文档)
* [更多](#更多)

---------------------------------------

## 特性
* 高性能
* 纯Golang实现
* 分布式服务架构（服务间通信使用grpc）
* DDD项目架构
* 支持动态扩缩服务实例与负载均衡
* 服务实例支持pprof调试和prometheus metrics(可接入prometheus和grafana实现可视化的服务监控)
* 支持服务动态注册发现和配置中心（基于consul）
* 实时+离线推送支持（SocketIO+RabbitMQ）
* 采用高性能API网关（apisix）
* 服务生命周期实现（manage）
* OSS对象存储（minio）
* 传输加密（openpgp）

## 服务
> coss-server 包含以下服务：

* user: 👤用户服务，处理用户注册、登录等功能。
* group: 👬群组服务，管理用户之间的群组关系。
* push: ✈️消息推送服务，负责实时消息推送功能。
* msg: 📩消息服务，处理用户间消息的收发功能。
* live: ☎️实时通讯服务，支持语音通话和视频通话功能。
* storage: 🗃存储服务，负责文件存储和管理。
* relation: 🧚‍关系服务，管理用户之间的社交关系。
* admin: 👷‍管理员服务，用于管理系统用户和权限。

## 快速启动
> 以下两种方式都需要安装docker-compose
> 请安装[coss-cli工具](https://github.com/zwtesttt/coss-cli/releases)
### 源码启动
```
1.拉取最新代码
git clone https://github.com/cossim/coss-server
mv ./coss-cli-xxx coss-server/coss-cli
cd coss-server

2.生成配置文件
chmod a+x coss-cli
coss-cli gen --path ./deploy/docker/

3.启动必需中间件
docker-compose -f deploy/docker/docker-compose.base.yaml up -d

4.启动服务
这里只拿user举例
go run ./cmd/user/main.go -config deploy/docker/config/service/user.yaml
```
### cli工具启动
> ⚠️请注意：cli工具启动时，会自动生成配置文件在当前目录下，如有需要请创建文件夹
```
mkdir coss-server
cd coss-server
coss-cli start
```
## 配置
**config/common**
> 存放公共中间件配置文件

**config/pgp**
> 存放pgp公私钥

**config/service**
> 存放所有服务配置文件

## 文档
TODO

## 更多
TODO