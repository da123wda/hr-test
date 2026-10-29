# HR 动态表单与通用接口引擎系统

本项目为面向企业复杂入职场景设计的动态表单设计与通用接口引擎，支持无代码表单定义、动态渲染联动、通用 CRUD 接口及敏感数据审计。

---

## 1. 技术架构

- **后端架构**：采用 Golang + Gin + xorm 架构体系，采用分层设计：
  - `src/main.go`：服务入口，负责日志、配置、存储、缓存与路由的生命周期初始化
  - `src/cfg/`：配置管理，解析 `config/` 目录下的 TOML 配置文件
  - `src/dao/`：数据访问层，管理 PostgreSQL 连接池、事务与数据实体读写
  - `src/rdb/`：Redis 连接池管理与缓存读写
  - `src/model/`：数据持久化模型，映射数据库表结构
  - `src/dto/`：请求与响应数据传输对象
  - `src/router/`：Gin 路由定义与路由组划分
  - `src/router/controller/`：接口控制层，处理参数绑定与请求上下文
  - `src/router/service/`：核心业务逻辑层
  - `src/engine/`：规则引擎与动态校验器
  - `src/log/`：基于 Zap 的全局日志记录与 Panic 异常恢复捕获
  - `src/cbase/` 与 `src/gin_utils/`：统一错误码结构与参数绑定封装
- **前端架构**：采用 Vue 3 + TypeScript + Vite + Element Plus 构建，包含动态渲染引擎与可视化设计器。
- **反向代理与网关**：采用 Nginx 统一代理前端静态页面与 `/api/` 接口转发。
- **数据库与缓存**：
  - PostgreSQL 16：核心数据采用 JSONB 结构化存储，辅以 GIN 倒排索引支持动态条件检索。
  - Redis 7：提供分布式缓存与会话状态支持。

---

## 2. 核心功能实现

### 2.1 动态表单 Schema 与渲染引擎
- **Schema 结构**：支持标准输入、数字、单选、多选、下拉、日期、文件上传，以及嵌套对象与嵌套数组（Repeater）。
- **字段废弃与版本演进**：已发布版本在快照表归档存储，删除字段置为 `deprecated: true`，新表单不渲染，历史提交记录保持完整展示。
- **规则引擎与环依赖检测**：
  - 支持字段显隐、必填动态切换及条件审批提醒。
  - 通过有向图深度优先遍历（DFS）实现静态循环依赖检测，拦截 $A \to B \to C \to A$ 死锁配置；单次动态联动设置执行深度上限，防止运行时递归死循环。

### 2.2 通用数据接口引擎
- **通用数据提交**：`POST /api/v1/dynamic-data/:formId`，自动匹配发布版本，执行服务端字段类型、必填及动态规则校验，校验失败返回字段级错误字典。
- **通用动态查询**：`GET /api/v1/dynamic-data/:formId`，基于 PostgreSQL JSONB 语法将过滤条件转换为底层 SQL 查询，支持类似 `experience > 3` 的动态数值及文本过滤。
- **安全文件上传**：`POST /api/v1/files/upload`，采用独立存储方案，限制 20MB 大小，校验扩展名与文件头 MIME，随机 UUID 重命名隔离存储，表单 JSON 仅引用文件元数据。
- **并发控制**：基于表单元数据的 `version` 字段与 CAS 条件更新实现乐观并发锁，并发保存版本不匹配时返回 409 冲突状态。
- **敏感数据审计**：记录更新前后的字段差异（Diff），如 `salary: 30000 -> 35000`，明确操作人、操作 IP 与时间。

---

## 3. 项目目录结构

```
├── backend/
│   ├── config/                   # TOML 配置文件目录
│   │   ├── config.toml           # 整合配置 (server, pgsql, redis)
│   │   ├── pgsql.toml            # PostgreSQL 连接与连接池配置
│   │   └── redis.toml            # Redis 缓存连接配置
│   ├── src/
│   │   ├── main.go               # 入口文件
│   │   ├── cfg/                  # 配置加载模块
│   │   ├── dao/                  # 数据访问层
│   │   ├── rdb/                  # Redis 访问模块
│   │   ├── model/                # 数据模型
│   │   ├── dto/                  # DTO 定义
│   │   ├── engine/               # 校验与规则引擎
│   │   ├── router/               # 路由模块
│   │   │   ├── controller/       # 控制器
│   │   │   └── service/          # 业务层
│   │   ├── log/                  # 日志与异常处理
│   │   ├── cbase/                # 基础返回结构
│   │   └── gin_utils/            # 请求参数绑定工具
│   ├── Dockerfile                # 后端 Dockerfile
│   ├── go.mod, go.sum
│   └── hr-server.exe
├── frontend/
│   ├── src/
│   │   ├── components/           # 动态表单渲染组件
│   │   ├── views/                # 页面视图
│   │   ├── api/                  # 接口请求
│   │   ├── router/               # 路由配置
│   │   └── types/                # TS 类型定义
│   ├── dist/                     # 前端打包静态资源
│   ├── package.json
│   └── vite.config.ts
├── nginx/
│   ├── conf.d/
│   │   └── default.conf          # Nginx 反向代理配置
│   └── logs/                     # Nginx 日志挂载目录
└── docker-compose.yml            # Docker 容器编排
```

---

## 4. 镜像打包与 Linux 部署

### 4.1 本地打包后端 Docker 镜像
在本地 `backend` 目录下通过 Dockerfile 构建 Linux amd64 镜像并导出为 tar 包：

```bash
# 1. 构建镜像
docker build -t hr-backend:latest ./backend

# 2. 导出镜像包
docker save -o hr-backend.tar hr-backend:latest
```

将 `hr-backend.tar` 传入 Linux 服务器。

### 4.2 Linux 服务器载入与启动

在 Linux 服务器上执行：

```bash
# 1. 导入后端镜像
docker load -i hr-backend.tar

# 2. 在项目根目录启动服务集群
docker compose up -d
```

### 4.3 挂载持久化说明
在 Linux 服务器上，所有数据卷与配置统一持久化在 `/home/ubuntu/hr_test/dockerData` 目录下：
- **PostgreSQL 数据卷挂载**：`/home/ubuntu/hr_test/dockerData/postgres` $\to$ `/var/lib/postgresql/data`
- **Redis 数据卷挂载**：`/home/ubuntu/hr_test/dockerData/redis` $\to$ `/data`
- **后端配置挂载**：`/home/ubuntu/hr_test/dockerData/config` $\to$ `/app/config:ro`
- **上传文件目录挂载**：`/home/ubuntu/hr_test/dockerData/uploads` $\to$ `/app/data/uploads`
- **Nginx 配置文件挂载**：`/home/ubuntu/hr_test/dockerData/nginx/conf.d` $\to$ `/etc/nginx/conf.d`
- **Nginx 日志目录挂载**：`/home/ubuntu/hr_test/dockerData/nginx/logs` $\to$ `/var/log/nginx`
- **前端静态资源挂载**：`/home/ubuntu/hr_test/dockerData/nginx/html` $\to$ `/usr/share/nginx/html:ro`

---

## 5. 说明事项

- 本系统当前聚焦于动态表单 Schema 引擎、通用 CRUD 数据接口、规则联动计算与审计归档，暂未实现 OAuth 2.0 登录认证。当前接口通过请求头传递操作人标识进行审计日志记录。
