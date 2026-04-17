# CUSTOM_CHANGES.md

> 记录本项目相对于上游 [QuantumNous/new-api](https://github.com/QuantumNous/new-api) 所做的全部自定义改动。
> 上次上游同步版本：commit `28347402`（2025-04-xx, pgx v5.9.0）
> 维护规则：每次修改上游文件时，在对应条目中更新说明；新增文件在"新增文件"小节注册。

---

## 一、新增文件（不与上游冲突）

| 文件 | 说明 |
|---|---|
| `controller/affiliate.go` | 达人申请公开接口 + 审核中心接口（ReviewerOrRootAuth 保护） |
| `controller/kol.go` | KOL 达人仪表盘、提现申请、Stripe Connect 入驻 |
| `controller/kol_admin.go` | 管理员：修改 aff_code、查看提现列表、Stripe 打款重试 |
| `model/affiliate_application.go` | AffiliateApplication 数据模型（申请状态机、token 一次性使用） |
| `model/commission.go` | CommissionRecord 数据模型（佣金记录、汇总查询） |
| `model/withdrawal.go` | WithdrawalRequest 数据模型（提现申请、余额原子扣减） |
| `service/commission.go` | ProcessCommission（幂等）、StartCommissionSettleTask（定时冻结 → 可提现） |
| `service/withdrawal.go` | CreateWithdrawalTransfer — Stripe Connect 打款封装 |
| `setting/payment_commission.go` | KolCommissionRate、StripeConnectEnabled、MinWithdrawalAmount 配置变量 |
| `setting/payment_stripe_connect.go` | StripeConnectClientId 配置变量 |
| `docs/rules/deconflict.md` | 解耦与上游同步规则（本文件的使用规范） |
| `docs/rules/review.md` | Code Review 规则与审阅清单 |

---

## 二、修改的上游文件

### `middleware/kol.go`（新增文件，不修改上游）

- `KolAuth()` — 检查 `user.group == "kol"`，需在 `UserAuth()` 之后使用
- `ReviewerOrRootAuth()` — 允许 `reviewer` 分组或 `role >= RoleRootUser`，管理员（role=10）无权限
- **`middleware/auth.go` 未被修改**，两个函数已提取到独立文件，与上游零冲突。

---

### `model/user.go`

- **位置**：User struct 字段定义末尾
- **改动**：新增 4 个 KOL 字段：
  ```go
  KolBalance             float64 `gorm:"type:decimal(10,6);default:0;column:kol_balance"`
  KolHistoryBalance      float64 `gorm:"type:decimal(10,6);default:0;column:kol_history_balance"`
  StripeConnectAccountId string  `gorm:"type:varchar(128);default:'';column:stripe_connect_account_id"`
  StripeConnectOnboarded bool    `gorm:"default:false;column:stripe_connect_onboarded"`
  ```
- **改动**：`Insert()` 和 `FinalizeOAuthUserCreation()` 中删除了原有的 `QuotaForInvitee` / `QuotaForInviter` 奖励逻辑（改为 KOL 佣金系统接管，按充值额计算佣金，原"注册赠送 quota"流程已停用）。`inviteUser()` 函数保留但不再调用。
- **⚠️ 风险点（同步必查）**：
  1. 上游若修改 `Insert()` 或 `FinalizeOAuthUserCreation()` 中的邀请逻辑，需确认我们的 KOL 分支仍完整。
  2. 上游若新增 User struct 字段，需在 SQLite `ensureUserTableSQLite()` 中同步添加对应 `ALTER TABLE ADD COLUMN`。

---

### `model/main.go`

- **改动 1**：`migrateDB()` 中 **SQLite 路径将 User 排除在 `AutoMigrate` 之外**，改为调用 `ensureUserTableSQLite()`（逐列 `ALTER TABLE ADD COLUMN`）。原因：SQLite migrator 在历史 schema 差异时会尝试重建整张表，曾造成真实生产数据丢失。
- **改动 2**：`migrateDB()` 和 `migrateDBFast()` 中为 MySQL/PostgreSQL 路径追加：
  - `CommissionRecord`、`WithdrawalRequest`、`AffiliateApplication` 的 AutoMigrate
- **改动 3**：新增 4 个 SQLite 安全迁移函数：
  - `ensureUserTableSQLite()` — users 表新增列
  - `ensureCommissionRecordTableSQLite()`
  - `ensureWithdrawalRequestTableSQLite()`
  - `ensureAffiliateApplicationTableSQLite()`
- **⚠️ 风险点（同步必查）**：SQLite User 迁移是本项目最危险的改动，**绝对不能将 `&User{}` 重新加入 SQLite 的 AutoMigrate 列表**。

---

### `model/option.go`

- **位置**：`InitOptionMap()` 和 `UpdateOption()` 的 switch-case
- **改动**：注册 4 个 KOL 配置键：`KolCommissionRate`、`StripeConnectEnabled`、`MinWithdrawalAmount`、`StripeConnectClientId`
- **风险点**：上游若改动同一区域的 `UpdateOption` switch 可能冲突；同步时检查新增 case 是否重名。

---

### `controller/user.go`

- **改动 1**：`Register()` 函数中：
  - 仅接受 `group == "kol"` 用户的邀请码（非 KOL 邀请人的 inviterId 被置零）
  - 处理 URL 参数 `kol_token`，调用 `applyKolInviteToken()`（非致命失败，仅 SysLog）
- **改动 2**：新增 `applyKolInviteToken()` 私有函数
- **风险点**：上游若修改 `Register()` 的注册流程（如 inviter 处理逻辑），需确认 KOL 过滤和 kol_token 处理仍在正确位置。

---

### `controller/oauth.go`

- **改动**：OAuth 注册流程中，同样只接受 `group == "kol"` 用户的邀请码。
- **风险点**：上游若修改 OAuth 注册流程，需同步检查 KOL invite 过滤是否保留。

---

### `controller/topup.go` / `controller/topup_creem.go` / `controller/topup_stripe.go` / `controller/topup_waffo.go`

- **改动**：每个支付回调成功后调用 `service.ProcessCommission(userId, tradeNo, amount)`（幂等，失败仅记日志）。
- **风险点**：上游若新增支付渠道或修改回调流程，需确认 ProcessCommission 钩子仍被调用。

---

### `router/api-router.go`

- **改动**：在现有路由末尾新增：
  - `POST /api/auth/logout` — 前端 POST 登出的别名
  - `POST /api/user/logout` — 同上
  - KOL 路由组 `/api/kol/*`（需 `UserAuth` + `KolAuth`）
  - KOL 管理路由组 `/api/kol/admin/*`（需 `AdminAuth`）
  - 公开达人申请路由 `POST /api/affiliate/apply`、`GET /api/affiliate/apply`
  - 达人审核路由组 `/api/affiliate/*`（需 `UserAuth` + `ReviewerOrRootAuth`）
- **风险点**：上游若在 apiRouter 末尾追加路由，merge 冲突概率较低但需检查。

---

### `main.go`

- **改动**：在 `main()` 初始化序列末尾追加 `service.StartCommissionSettleTask()`。
- **风险点**：上游若在 main 末尾新增 goroutine/task 启动，需检查顺序是否合理。

---

### 前端文件（`web/`）

| 文件 | 改动摘要 |
|---|---|
| `web/src/App.jsx` | 新增 KolDashboard 路由、AffiliateApply 路由、AffiliateReview 路由 |
| `web/src/components/auth/RegisterForm.jsx` | 支持 `kol_token` URL 参数，注册时传递给后端 |
| `web/src/components/layout/SiderBar.jsx` | 新增"达人中心"和"审核中心"菜单组（按 group/role 显示） |
| `web/src/components/table/users/UsersColumnDefs.jsx` | 用户列表新增 KOL 字段列 |
| `web/src/components/table/users/UsersTable.jsx` | 配合上述列变更 |
| `web/src/components/topup/index.jsx` | 充值页面适配 |
| `web/src/helpers/auth.jsx` | 新增 `ReviewerRoute` 路由守卫 |
| `web/src/helpers/utils.jsx` | 新增 `isReviewer()` 工具函数 |
| `web/src/i18n/locales/*.json` | 所有 7 个语言文件补齐 KOL/达人申请/审核中心相关 key |

---

## 三、已知待处理事项

| # | 问题 | 严重度 | 状态 |
|---|---|---|---|
| 1 | `QuotaForInvitee`/`QuotaForInviter` 管理面板配置项仍存在但逻辑已删除 | 中 | 待确认是否保留 UI 或彻底移除 |
| 2 | `kol_token` 应用失败时用户静默注册为普通用户（非 KOL） | 中 | 待确认是否改为报错阻断 |
| 3 | `GetKolInvitees` 存在 N+1 查询（每个 invitee 单独查 commission summary） | 低 | 待优化为 GROUP BY 批量查询 |
| 4 | 提现"申请"和"审核通过"两步未在同一事务内（可能状态短暂不一致） | 低 | 待确认是否需要合并为原子操作 |
