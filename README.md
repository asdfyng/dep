
## 目录

- [语言规范](#语言规范)
	- [异常处理](#异常处理)
	- [格式化](#格式化)
	- [单行长度](#单行长度)
	- [反射](#反射)
	- [命名](#命名)
	- [json处理](#json处理)
	- [注释](#注释)
	- [更多参考](#更多参考)


- [包管理](#包管理)
	- [使用GOPATH](#使用GOPATH)
	- [使用govendor](#使用govendor)
	- [暂不使用gomod](#暂不使用gomod)


- [数据库](#数据库)
	- [Mysql](#Mysql)
		- [禁止使用ORM](#禁止使用ORM)
		- [确保连接释放](#确保连接释放)
		- [SQL语句防注入](#SQL语句防注入)
		- [数据删除](#数据删除)
		- [更多参考](#更多参考)
	- [Redis](#redis)
	- [数据库审核流程](#数据库审核流程)

- [日志](#日志)
	- [启动日志](#启动日志)
	- [日志切割](#日志切割)
	- [日志格式](#日志格式)
	- [异常日志](#异常日志)

- [工程结构](#工程结构)
	- [main包](#main包)
	- [公共包](#公共包)
	- [业务逻辑包](#业务逻辑包)
	- [配置](#配置)
	- [目录结构](#目录结构)
	- [示例工程](#示例工程)

- [版本信息](#版本信息)
	- [版本日志](#版本日志)
	- [版本查询接口](#版本查询接口)

- [测试](#测试)
	- [单元测试](#单元测试)
	- [功能测试](#功能测试)

- [工程构建、部署](#工程构建、部署)
	- [构建](#构建)
	- [部署](#部署)


## 语言规范

### 异常处理
- 除了启动阶段失败可以panic, 启动后所有代码不应该主动panic

- 为了防止预料外的异常导致整个进程挂掉, 所有协程必须recover panic

> 禁止裸go, 统一使用util.Go

> 其他异步回调比如time.AfterFunc回调时是新建一个协程, 回调函数中也要recover

```golang
import "github.com/temprory/util"

func callback() {
	defer util.HandlePanic()

	//...
	//可能panic	
}

time.AfterFunc(callback)
```

### 格式化
> 所有代码必须go fmt, 各种编辑器、IDE几乎都支持保存的时候自动fmt, 请开启

### 单行长度
> 单行不要过长, SQL语句分行, Printf类似的分行

### 反射
> 除非极特殊需要, 不直接使用反射

### 命名
> 包命名尽量小写

> 普通变量命名使用标准驼峰, 大小写区分是否导出包, 不要使用下划线

> 常量、枚举类型等可以大写+下划线

### json处理
> 尽量定义struct处理, 如非必须, 不要使用map[string]interface{}, 涉及option参数和默认空值的可以约定option参数传特殊值表示不使用此参数然后switch处理, 比如-1

### 注释
> 重要结构体、函数应加注释, 最好带上相关的业务的中文关键词

### 更多参考
- [Go官方编程规范翻译](http://www.gonglin91.com/2018/03/30/go-code-review-comments/)
- [Go编码规范指南](https://gocn.vip/article/1)
- [Golang代码规范](https://sheepbao.github.io/post/golang_code_specification)


## 包管理
### 使用GOPATH
> 为了避免多个GOPATH引起混乱, 只设置一个GOPATH目录, 所有代码都放到$GOPATH/src下面

### 使用govendor
- 使用govendor管理依赖包, 每个工程保存依赖包到项目代码中, 以免依赖包变更、删除导致兼容和风险

1. 安装govendor
```sh
go get github.com/kardianos/govendor
```

2. 导入依赖
```sh
cd $GOPATH/src/package

govendor init .
govendor add +external

# 保存到仓库
```

### 暂不使用gomod
> 1) gomod不方便引用私有仓库和$GOPATH/src下面的包, 多个使用gomod的项目间需要拷贝文件的方式共享代码

> 2) gomod缓存的文件并不在项目代码内, 如果刚好第三方删库了, 本地缓存也清理了, 项目依赖的源码找不到了会比较麻烦


## 数据库
### Mysql

#### 禁止使用ORM
> 1)不透明, 性能损失, 如果出问题也不好排查

> 2)复杂需求用ORM很难实现, 如果ORM和原生SQL都使用, 项目风格不统一显得乱

> 3)ORM不方便SQL客户端里调试SQL语句进行直接分析、优化

> 4)不方便DBA一起进行SQL代码审查

> 4)没有大厂大项目为ORM背书

#### 确保连接释放
1. 确保Query后释放rows.Close
```golang
rows, err := db.Query("select ...", args)
if err != nil {
	return err
}
//务必使用defer的方式, 避免后续代码可能产生panic
defer rows.Close()
```

2. 事务不论成功还是失败都要确保连接释放
```golang
import "github.com/temprory/util"

func someTx() error {
	//SQL事务独占一个连接, 
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	//务必使用defer的方式, 避免后续代码可能产生panic导致Rollback和Commit都没有执行造成连接泄露
	defer util.ClearTx(tx)

	//... may panic

	return tx.Commit()
}
```

#### 其他约束
1. 禁止使用浮点类型
> 数值类型一律用int64, 精度除以10^N的方式处理, 避免浮点精度对计算和显示的干扰

2. 禁止使用所有时间日期类型
> 一律用int64存时间戳进行处理, 不需要处理时区和各种转换计算, 展示由前端根据需要转换格式

3. 所有字段必须有not null约束


#### SQL语句防注入
- SQL语句禁止使用fmt.Sprintf等方式拼接参数, 尤其是string类型的参数，应使用占位符或者sql.Named, 避免注入风险

```golang
//mysql
rows, err := db.Query("select * from user where id=? and name=?",
		1,
		"name")
		
//sqlserver		
rows, err := db.Query("select * from user where id=@id and name=@name",
		sql.Named("id", 1),
		sql.Named("name", "name"))

//oracle
rows, err := db.Query("select * from user where id=:1 and name=:2",
		1,
		"name")
//其他...
```

#### 数据删除
> 重要数据禁止删除, 添加标记字段标识数据项是否废弃


#### 更多参考
- [SQL代码编码原则与规范](https://www.alibabacloud.com/help/zh/doc-detail/85305.htm)

- [MySQL SQL开发规范](https://www.kancloud.cn/wubx/mysql-sql-standard/600511)

### Redis

#### 参考
- [阿里云Redis开发规范](https://yq.aliyun.com/articles/531067)

### 数据库审核流程
- 数据库使用应增加审核流程（Mysql、Mongo、Redis或者其他数据库、缓存都包括）

> 1)开发人员设计、整理好相关数据结构及数据库、缓存涉及的操作

> 2)开发同事相应的审核人员进行review, 主要审核需求、安全性、性能

> 3)DBA审核, 主要审核安全性和性能


## 日志
### 启动日志
> 启动日志应输出程序版本信息, 方便确认线上问题与对应的代码版本

> 启动日志应该输出基本配置信息, 方便排查启动失败等问题, 敏感信息应打码(比如线上数据库配置)

> 启动过程中需要依赖的项应有清晰的日志, 如果有依赖项启动失败, 比如数据库初始化失败, 程序应该退出, 方便运维、开发及时发现和处理问题

### 日志切割
> 日志文件应进行切割, 避免单文件过大难处理

### 日志格式
> 日志应带时间、代码行数等信息, 方便排查问题

### 异常日志
> 所有协程应捕获panic并输出到日志


## 工程结构

### main包
> main包只包含一个main.go

### 公共包
> 涉及协议等内容的公共模块可以放到统一的仓库, 如grpc、其他proto定义等

### 业务逻辑包
> 业务逻辑统一放到app包, app包内尽量不再分包, 文件命名成URI的方式管理代码文件

> 我们以前也分很多包比如MVC方式分包, 但是功能复杂后可能导致的包之间循环引用等问题, 而且简单的分包可能只是一层函数封装, 分包的实际作用不大, 反而阅读代码时来回跳转, 可读性复杂度增加

> restful接口这种无状态服务, 接口之间基本无耦合, 尽量一个router一个go文件, 单文件内实现所有逻辑(包括db操作), 多人维护起来阅读也容易得多

### 配置
> 所有配置文件放到conf目录下, 程序启动可以通过命令行传入配置文件参数, 不传则应有默认的配置文件参数免去每次启动输入的麻烦

### 目录结构
- 工程以hello为例

> $GOPATH/src
>> hello

>>> app

>>>> app.go

>>>> other.go

>>> conf

>>>> config.json

>>> test

>>>> test.go

>>> tools

>>>> ...

>>> others...

>> main.go

### 示例工程

- [hello](https://github.com/temprory/hello)


## 版本信息
### 版本日志
> 启动日志应输出版本信息

### 版本查询接口
> 程序应提供版本信息查询接口


## 测试
### 单元测试
> 小段、基础算法应有单元测试

### 功能测试
> 具体业务逻辑可能比较复杂, 单元测试覆盖难、编写难如有需要, 自测代码放到工程的test目录中专门实现


## 工程构建、部署
### 构建

- 为方便确认线上程序与对应的代码版本，请用如下类似方式构建发布包

```sh
#!/bin/bash

BUILD_TIME=`date +%FT%T%z`
# echo $BUILD_TIME

BUILD_DIR=`pwd`
# echo $BUILD_DIR

GIT_VER=$(git rev-parse HEAD)
# echo $GIT_VER

# -tag=jsoniter开启jsoniter, -ldflags带上编译时间和git版本信息
go build -tags=jsoniter -ldflags "-X main.BuildTime=${BUILD_TIME} -X main.Version=${GIT_VER} -X github.com/temprory/log.BuildDir=${BUILD_DIR}"
```

- 日志中应输出git版本、构建时间等信息，查看git版本
```sh
#!/bin/bash

git log | grep c96e207a34a4a42139f520592899fc82372b0e51 -3 | tail -3
Author: george <george@yabotiyu.net>
Date:   Sat Mar 16 22:31:26 2019 +0800
```

- package main 中声明Version变量, 接收构建的Version, 并在输出日志
```golang
package main

import "fmt"

var Version = ""

func main() {
	fmt.Println("app version info: ", Version)
}
```

### 部署

- 内核参数调优, sysctl部分选项不同内核版本可能不支持, 如果设置失败忽略即可

```sh
ulimit -n 100000

sysctl -w net.ipv4.tcp_keepalive_intvl=75
sysctl -w net.ipv4.tcp_keepalive_probes=9
sysctl -w net.ipv4.tcp_keepalive_time=7200
sysctl -w net.ipv4.tcp_max_syn_backlog=65536
sysctl -w net.ipv4.tcp_max_tw_buckets=65536=75
sysctl -w net.ipv4.tcp_syn_retries=6
sysctl -w net.ipv4.tcp_synack_retries=5
sysctl -w net.ipv4.tcp_syncookies=1
sysctl -w net.ipv4.tcp_tw_reuse=1
sysctl -w net.ipv4.tcp_fin_timeout=30
sysctl -w net.ipv4.tcp_timestamps=1
sysctl -w net.ipv4.tcp_tw_recycle=1

sysctl -p

start proc ...
```

## 团队合作

- 以下都不是为了服务端省体力, 而是为了规范

### 与产品
1. 开发部门要搞清楚需求和实现, 充分评估需求对应的技术成本和技术风险, 不要产品提什么需求都接, 列清楚原因该拒绝就拒绝
2. 合理拒绝需求不是为了给自己省力, 所以不要随便拒绝合理需求

### 与前端
1. 服务端协议制定要合理, 协议设计合理的前提下以服务端为准, 所有协议规定好格式、字段类型
2. 服务端尽量提供原始数据的简单组装, 不做展示相关需求的加工比如数值浮点类型转成字符串再给客户端
