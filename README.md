# UserCore

统一身份认证（IAM）、租户管理、RBAC 权限与应用中心。

## 架构定位

- **UserCore**：登录、租户选择、应用中心、用户/角色管理（独立进程，端口 8091）
- **ProductCore**：商品 PIM（验证 UserCore 签发的 JWT，按 `tenant_id` 隔离数据）

一个公司（Company）可拥有多个租户（Tenant）；用户通过租户成员关系访问数据。

## 快速启动

```bash
# 1. 数据库（PostgreSQL）
createdb usercore  # 或使用 make init-db

# 2. 配置
cp configs/config.example.yaml configs/config.yaml
# jwt.secret 须与 ProductCore configs/config.yaml 中 auth.jwt_secret 一致

# 3. 后端
make run

# 4. 前端
cd web && npm install && npm run dev
```

- 应用中心：http://localhost:5174
- API：http://localhost:8091/api/v1

## 演示账号

| 邮箱 | 密码 | 说明 |
|------|------|------|
| admin@demo.com | demo123 | 平台管理员，可访问两个租户 |
| fashion@demo.com | demo123 | 服装商品库租户管理员 |
| digital@demo.com | demo123 | 3C 商品库租户管理员 |

## Phase B

- **PlatformGateway**（`~/projects/PlatformGateway`）：统一 API 入口 `:8088`
- 平台管理：公司 / 租户 CRUD（平台超管）
- 租户管理：用户 / 角色完整编辑 UI
- Portal 侧边栏布局

### 可选：经 Gateway 访问

在 `web/.env` 增加 `VITE_API_GATEWAY=http://localhost:8088`，并启动 PlatformGateway，则 UserCore 前端 `/api` 全部经 Gateway 转发。

## Phase C（进行中）

| 项 | 说明 | 状态 |
|----|------|------|
| Gateway 开发接入 | 前端 `.env` 切换 `VITE_API_GATEWAY` | ✅ |
| 用户邀请 | 已有账号加入其他租户 | 待做 |
| 平台进入租户 | 超管在租户列表一键切换 | 待做 |
| 统一退出 / Token 失效 | 登出黑名单或 refresh | 待做 |
| 审计日志 | 关键 IAM 操作留痕 | 待做 |


1. 在 `applications` 表注册应用（seed 已含 `productcore`）
2. 子应用前端从 URL 接收 `?token=` 写入 `localStorage`（key: `uc_access_token`）
3. 子应用 API 验证同一 JWT secret，从 claims 读取 `tid`（tenant_id）
