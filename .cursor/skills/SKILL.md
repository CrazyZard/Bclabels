---
name: gva-plugin-development
description: "Gin-Vue-Admin 插件开发指南，用于快速开发独立的业务插件"
version: 1.0.0
author: Hermes Agent
tags: [gin-vue-admin, plugin, go, vue, backend, frontend]
related_skills: []
---

# Gin-Vue-Admin 插件开发指南

## 核心原则
✅ **插件代码只在插件目录实现，不修改母工程文件**

## 一、后端插件结构 (server/plugin/{plugin_name}/)

### 标准目录结构 (参考 announcement 插件)
```
plugin_name/
├── plugin.go                    # 插件入口，实现 Plugin 接口
├── config/
│   └── config.go               # 插件配置结构
├── initialize/
│   ├── api.go                  # API 初始化
│   ├── dictionary.go           # 字典初始化
│   ├── gorm.go                 # 数据库迁移
│   ├── menu.go                 # 菜单初始化
│   ├── router.go               # 路由初始化
│   └── viper.go                # 配置初始化
├── model/
│   ├── {entity}.go             # 数据模型
│   └── request/
│       └── {entity}.go         # 请求参数结构
├── api/
│   ├── enter.go                # API 入口
│   └── {entity}.go             # API 实现
├── router/
│   ├── enter.go                # Router 入口
│   └── {entity}.go             # 路由定义
├── service/
│   ├── enter.go                # Service 入口
│   └── {entity}.go             # 业务逻辑
└── gen/
    └── gen.go                  # 代码生成配置
```

### 简化结构 (参考 email 插件)
```
plugin_name/
├── main.go                      # 插件入口
├── config/
│   └── config.go               # 配置
├── global/
│   └── global.go               # 全局变量
├── model/
│   └── response/
│       └── {entity}.go
├── api/
│   ├── enter.go
│   └── {entity}.go
├── router/
│   ├── enter.go
│   └── {entity}.go
├── service/
│   ├── enter.go
│   └── {entity}.go
└── utils/
    └── {utils}.go
```

---

## 二、后端核心文件模板

### 1. plugin.go (插件入口)
```go
package plugin_name

import (
    "context"
    "github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin_name/initialize"
    interfaces "github.com/flipped-aurora/gin-vue-admin/server/utils/plugin/v2"
    "github.com/gin-gonic/gin"
)

var _ interfaces.Plugin = (*plugin)(nil)
var Plugin = new(plugin)
type plugin struct{}

func init() {
    interfaces.Register(Plugin)
}

func (p *plugin) Register(group *gin.Engine) {
    ctx := context.Background()
    initialize.Api(ctx)        // 注册 API
    initialize.Menu(ctx)       // 注册菜单
    initialize.Dictionary(ctx) // 注册字典
    initialize.Gorm(ctx)       // 数据库迁移
    initialize.Router(group)   // 注册路由
}
```

### 2. model/{entity}.go (数据模型)
```go
package model

import "github.com/flipped-aurora/gin-vue-admin/server/global"

type Info struct {
    global.GVA_MODEL
    Title   string `json:"title" form:"title" gorm:"column:title;comment:标题;"`
    Content string `json:"content" form:"content" gorm:"column:content;comment:内容;type:text;"`
    UserID  *int   `json:"userID" form:"userID" gorm:"column:user_id;comment:发布者;"`
}

func (Info) TableName() string {
    return "gva_announcements_info"
}
```

### 3. model/request/{entity}.go (搜索请求参数)
```go
package request

import "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"

type InfoSearch struct {
    request.PageInfo
    StartCreatedAt *int64 `json:"startCreatedAt" form:"startCreatedAt"`
    EndCreatedAt   *int64 `json:"endCreatedAt" form:"endCreatedAt"`
    // 其他搜索字段
    Title string `json:"title" form:"title"`
}
```

### 4. initialize/gorm.go (数据库迁移)
```go
package initialize

import (
    "context"
    "fmt"
    "github.com/flipped-aurora/gin-vue-admin/server/global"
    "github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin_name/model"
    "github.com/pkg/errors"
    "go.uber.org/zap"
)

func Gorm(ctx context.Context) {
    err := global.GVA_DB.WithContext(ctx).AutoMigrate(
        new(model.Entity),
    )
    if err != nil {
        err = errors.Wrap(err, "注册表失败!")
        zap.L().Error(fmt.Sprintf("%+v", err))
    }
}
```

### 4. service/{entity}.go (业务逻辑)
```go
package service

import (
    "github.com/flipped-aurora/gin-vue-admin/server/global"
    "github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin_name/model"
    "github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin_name/model/request"
)

var Entity = new(entity)
type entity struct{}

// 创建
func (s *entity) CreateEntity(entity *model.Entity) error {
    return global.GVA_DB.Create(entity).Error
}

// 删除
func (s *entity) DeleteEntity(ID string) error {
    return global.GVA_DB.Delete(&model.Entity{}, "id = ?", ID).Error
}

// 批量删除
func (s *entity) DeleteEntityByIds(IDs []string) error {
    return global.GVA_DB.Delete(&[]model.Entity{}, "id in ?", IDs).Error
}

// 更新
func (s *entity) UpdateEntity(entity model.Entity) error {
    return global.GVA_DB.Model(&model.Entity{}).Where("id = ?", entity.ID).Updates(&entity).Error
}

// 查询单个
func (s *entity) GetEntity(ID string) (model.Entity, error) {
    var entity model.Entity
    err := global.GVA_DB.Where("id = ?", ID).First(&entity).Error
    return entity, err
}

// 分页查询
func (s *entity) GetEntityList(info request.EntitySearch) (list []model.Entity, total int64, err error) {
    limit := info.PageSize
    offset := info.PageSize * (info.Page - 1)
    db := global.GVA_DB.Model(&model.Entity{})
    
    // 条件搜索
    if info.Keyword != "" {
        db = db.Where("field1 LIKE ?", "%"+info.Keyword+"%")
    }
    
    err = db.Count(&total).Error
    if err != nil {
        return
    }
    if limit != 0 {
        db = db.Limit(limit).Offset(offset)
    }
    err = db.Find(&list).Error
    return
}
```

### 5. api/{entity}.go (API 层)
```go
package api

import (
    "github.com/flipped-aurora/gin-vue-admin/server/global"
    "github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
    "github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin_name/model"
    "github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin_name/model/request"
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

var Entity = new(entity)
type entity struct{}

// 创建
func (a *entity) CreateEntity(c *gin.Context) {
    var entity model.Entity
    if err := c.ShouldBindJSON(&entity); err != nil {
        response.FailWithMessage(err.Error(), c)
        return
    }
    if err := serviceEntity.CreateEntity(&entity); err != nil {
        global.GVA_LOG.Error("创建失败!", zap.Error(err))
        response.FailWithMessage("创建失败", c)
        return
    }
    response.OkWithMessage("创建成功", c)
}

// 删除
func (a *entity) DeleteEntity(c *gin.Context) {
    ID := c.Query("ID")
    if err := serviceEntity.DeleteEntity(ID); err != nil {
        global.GVA_LOG.Error("删除失败!", zap.Error(err))
        response.FailWithMessage("删除失败", c)
        return
    }
    response.OkWithMessage("删除成功", c)
}

// 批量删除
func (a *entity) DeleteEntityByIds(c *gin.Context) {
    IDs := c.QueryArray("IDs[]")
    if err := serviceEntity.DeleteEntityByIds(IDs); err != nil {
        global.GVA_LOG.Error("批量删除失败!", zap.Error(err))
        response.FailWithMessage("批量删除失败", c)
        return
    }
    response.OkWithMessage("批量删除成功", c)
}

// 更新
func (a *entity) UpdateEntity(c *gin.Context) {
    var entity model.Entity
    if err := c.ShouldBindJSON(&entity); err != nil {
        response.FailWithMessage(err.Error(), c)
        return
    }
    if err := serviceEntity.UpdateEntity(entity); err != nil {
        global.GVA_LOG.Error("更新失败!", zap.Error(err))
        response.FailWithMessage("更新失败", c)
        return
    }
    response.OkWithMessage("更新成功", c)
}

// 查询单个
func (a *entity) FindEntity(c *gin.Context) {
    ID := c.Query("ID")
    entity, err := serviceEntity.GetEntity(ID)
    if err != nil {
        global.GVA_LOG.Error("查询失败!", zap.Error(err))
        response.FailWithMessage("查询失败", c)
        return
    }
    response.OkWithData(entity, c)
}

// 分页查询
func (a *entity) GetEntityList(c *gin.Context) {
    var pageInfo request.EntitySearch
    if err := c.ShouldBindQuery(&pageInfo); err != nil {
        response.FailWithMessage(err.Error(), c)
        return
    }
    list, total, err := serviceEntity.GetEntityList(pageInfo)
    if err != nil {
        global.GVA_LOG.Error("获取失败!", zap.Error(err))
        response.FailWithMessage("获取失败", c)
        return
    }
    response.OkWithDetailed(response.PageResult{
        List:     list,
        Total:    total,
        Page:     pageInfo.Page,
        PageSize: pageInfo.PageSize,
    }, "获取成功", c)
}
```

### 6. router/{entity}.go (路由)
```go
package router

import (
    "github.com/flipped-aurora/gin-vue-admin/server/middleware"
    "github.com/gin-gonic/gin"
)

var Entity = new(entity)
type entity struct{}

func (r *entity) Init(public *gin.RouterGroup, private *gin.RouterGroup) {
    {
        group := private.Group("entity").Use(middleware.OperationRecord())
        group.POST("createEntity", apiEntity.CreateEntity)
        group.DELETE("deleteEntity", apiEntity.DeleteEntity)
        group.DELETE("deleteEntityByIds", apiEntity.DeleteEntityByIds)
        group.PUT("updateEntity", apiEntity.UpdateEntity)
    }
    {
        group := private.Group("entity")
        group.GET("findEntity", apiEntity.FindEntity)
        group.GET("getEntityList", apiEntity.GetEntityList)
    }
    {
        group := public.Group("entity")
        // 公开接口
    }
}
```

### 7. initialize/router.go (路由初始化)
```go
package initialize

import (
    "github.com/flipped-aurora/gin-vue-admin/server/global"
    "github.com/flipped-aurora/gin-vue-admin/server/middleware"
    "github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin_name/router"
    "github.com/gin-gonic/gin"
)

func Router(engine *gin.Engine) {
    public := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("")
    private := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("")
    private.Use(middleware.JWTAuth()).Use(middleware.CasbinHandler())
    router.Router.Entity.Init(public, private)
}
```

---

## 三、前端插件结构 (web/src/plugin/{plugin_name}/)

```
plugin_name/
├── api/
│   └── {entity}.js            # API 调用
├── view/
│   └── index.vue              # 列表页面
└── form/
    └── {entity}.vue           # 表单弹窗
```

### 1. api/{entity}.js
```javascript
import service from '@/utils/request'

// 创建
export const createEntity = (data) => {
  return service({
    url: '/entity/createEntity',
    method: 'post',
    data
  })
}

// 删除
export const deleteEntity = (params) => {
  return service({
    url: '/entity/deleteEntity',
    method: 'delete',
    params
  })
}

// 批量删除
export const deleteEntityByIds = (params) => {
  return service({
    url: '/entity/deleteEntityByIds',
    method: 'delete',
    params
  })
}

// 更新
export const updateEntity = (data) => {
  return service({
    url: '/entity/updateEntity',
    method: 'put',
    data
  })
}

// 查询单个
export const findEntity = (params) => {
  return service({
    url: '/entity/findEntity',
    method: 'get',
    params
  })
}

// 分页查询
export const getEntityList = (params) => {
  return service({
    url: '/entity/getEntityList',
    method: 'get',
    params
  })
}
```

---

## 四、开发步骤

### 1. 初始化后端插件
```bash
mkdir -p server/plugin/{plugin_name}/{api,router,service,model/request,config,initialize}
```

### 2. 创建核心文件（按顺序）
1. **model/{entity}.go** - 数据模型定义
2. **model/request/{entity}.go** - 搜索请求参数结构
3. **service/{entity}.go** - 业务逻辑实现
4. **service/enter.go** - Service 入口
5. **api/{entity}.go** - API 层实现
6. **api/enter.go** - API 入口
7. **router/{entity}.go** - 路由定义
8. **router/enter.go** - Router 入口
9. **initialize/*.go** - 各初始化器（gorm/router/menu/api）
10. **plugin.go** - 插件入口

### 3. 注册插件
在 `server/plugin/register.go` 中导入（这是唯一需要修改的母工程文件）：
```go
import (
    _ "github.com/flipped-aurora/gin-vue-admin/server/plugin/announcement"
    _ "github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin_name"
)
```

### 4. 创建前端插件
```bash
mkdir -p web/src/plugin/{plugin_name}/{api,view,form}
```

### 5. 创建前端文件
- **api/{entity}.js** - API 调用封装
- **view/{entity}.vue** - 列表页面（包含搜索、表格、增删改查按钮）
- **form/{entity}.vue** - 表单弹窗组件

---

## 五、关键实现机制详解

### 插件自动注册原理
1. **init() 函数**：每个插件的 `plugin.go` 中的 `init()` 函数会在包被导入时自动执行
2. **interfaces.Register(Plugin)**：将插件实例注册到全局插件注册表
3. **系统启动时**：主程序会遍历所有已注册的插件，调用 `Register()` 方法
4. **Register() 方法**：按顺序执行所有初始化函数

### 菜单/API 自动注册
通过 `plugin-tool/utils` 提供的工具函数：
- `utils.RegisterMenus(entities...)`：将菜单自动写入 `sys_base_menus` 表
- `utils.RegisterApis(entities...)`：将 API 自动写入 `sys_apis` 表

**注意**：这些操作是幂等的，重复注册不会产生重复数据。

### 数据库自动迁移
`initialize.Gorm()` 调用 `AutoMigrate()` 自动创建或更新表结构，无需手动执行 SQL。

### 前端动态路由
- 后端菜单注册时指定 `Component: "plugin/announcement/view/info.vue"`
- 前端动态路由系统会自动加载插件目录下的 vue 文件
- 无需在前端路由文件中手动配置

---

## 五、关键注意事项

### 数据库表命名规范
```
gva_{plugin_name}_{entity}
例: gva_announcements_info
```
表名必须以 `gva_` 开头，遵循蛇形命名法。

### 路由前缀
- 自动使用 `global.GVA_CONFIG.System.RouterPrefix`
- 通常是 `/api` 或空
- 实际路由格式：`{RouterPrefix}/info/createInfo`

### 权限控制
- **private 路由**：使用 `JWTAuth()` + `CasbinHandler()` 中间件
  - 需要登录验证
  - 需要 Casbin 权限分配
- **public 路由**：无权限验证，公开访问
- **OperationRecord()**：操作记录中间件，记录用户操作日志

### 核心原则：不修改母工程文件
✅ **只在 `server/plugin/{plugin_name}/` 下写后端代码**
✅ **只在 `web/src/plugin/{plugin_name}/` 下写前端代码**
✅ **唯一需要修改的母工程文件：`server/plugin/register.go` 添加导入语句**

### 文件命名规范
- 数据模型：单数形式（`info.go` 不是 `infos.go`）
- 表名：`gva_{plugin}_{entity}` 格式
- Go 结构体：大驼峰命名（`Info`）
- JSON 字段：小驼峰命名（`title`）

### 开发检查清单
开发完成后，检查以下内容：
- [ ] 所有 CRUD 接口实现完整
- [ ] 数据模型包含 `global.GVA_MODEL`
- [ ] 表名通过 `TableName()` 方法指定
- [ ] 菜单已在 `initialize/menu.go` 注册
- [ ] API 已在 `initialize/api.go` 注册
- [ ] 数据库迁移在 `initialize/gorm.go` 配置
- [ ] 路由在 `initialize/router.go` 正确初始化
- [ ] `server/plugin/register.go` 已添加插件导入
- [ ] 前端 API 文件已创建
- [ ] 前端列表页面已创建
- [ ] 前端表单页面已创建
