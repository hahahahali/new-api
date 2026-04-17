# UPSTREAM_SYNC_CHECKLIST.md

> **用途**：每次从上游 [QuantumNous/new-api](https://github.com/QuantumNous/new-api) 拉取更新前，**先通读本文件**，再按 `docs/CUSTOM_CHANGES.md` 逐条核对。
> 本文件记录高风险自定义逻辑，一旦上游修改同一区域，极易造成静默错误或数据安全问题。

---

## 🔴 极高风险 — 数据安全

### 1. SQLite User 表迁移安全

**文件**：`model/main.go`
**规则**：**`&User{}` 绝对不能出现在 SQLite 路径的 `AutoMigrate` 列表中。**

SQLite migrator（`glebarez/sqlite`）在 schema 存在历史差异时（如 unique 约束元数据不同）会尝试重建整张表，导致真实用户数据丢失。我们用 `ensureUserTableSQLite()` 替代，逐列执行 `ALTER TABLE ADD COLUMN`。

**同步时检查**：
- 上游是否向 `migrateDB()` 或 `migrateDBFast()` 的 AutoMigrate 列表中**重新添加** `&User{}`？
- 如果有，**绝对不要** 将该变更应用到 SQLite 路径，只保留 MySQL/PostgreSQL 路径中的变更。
- 上游是否向 User struct 新增了字段？如果有，**必须同时**在 `ensureUserTableSQLite()` 中添加对应的 `ALTER TABLE ADD COLUMN`。

---

## 🔴 极高风险 — 权限与安全

### 2. KolAuth / ReviewerOrRootAuth 中间件

**文件**：`middleware/kol.go`（新增文件，`middleware/auth.go` 未被修改）
**规则**：`KolAuth()` 和 `ReviewerOrRootAuth()` 是新增的权限门控中间件，必须在 `UserAuth()` 之后调用。

**同步时检查**：
- `middleware/auth.go` 无自定义改动，上游该文件可直接接受。
- `common.RoleRootUser` 常量是否有改动？ReviewerOrRootAuth 依赖此值。

---

### 3. KOL 邀请码过滤 — 仅接受 KOL 用户的邀请

**文件**：`controller/user.go`（`Register()`）、`controller/oauth.go`
**规则**：我们修改了邀请码逻辑，非 KOL 分组用户的 `inviterId` 被置零，不再触发旧的 quota 奖励流程。

**同步时检查**：
- 上游是否修改了 `Register()` 中 `inviterId` 的处理逻辑？确认我们的 KOL 过滤条件（`inviter.Group != "kol"` → 置零）仍然位于 inviterId 首次赋值之后、用于注册逻辑之前。
- 上游是否修改了 `FinalizeOAuthUserCreation()` 的 inviter 处理？同上。
- **⚠️ QuotaForInvitee/QuotaForInviter 逻辑已被我们删除**。如果上游重新加入/修改了该逻辑，需要评估是否与 KOL 佣金系统兼容（两套奖励系统不能同时运行）。

---

## 🟠 高风险 — 业务逻辑完整性

### 4. ProcessCommission 钩子 — 支付回调中的佣金触发

**文件**：`controller/topup.go`、`controller/topup_creem.go`、`controller/topup_stripe.go`、`controller/topup_waffo.go`
**规则**：每个支付成功回调中必须有 `service.ProcessCommission(userId, tradeNo, amount)` 调用。

**同步时检查**：
- 上游是否新增了支付渠道文件（如 `controller/topup_xxx.go`）？如果有，**必须在新文件的支付成功处**添加 ProcessCommission 调用。
- 上游是否修改了现有 topup controller 中支付成功的代码路径？确认 ProcessCommission 调用仍在正确位置（成功状态判断通过之后，quota 增加之前或之后均可，但必须存在）。

---

### 5. 新模型的数据库迁移

**文件**：`model/main.go`
**规则**：MySQL/PostgreSQL 路径中已注册 `CommissionRecord`、`WithdrawalRequest`、`AffiliateApplication` 的 AutoMigrate，以及对应的 SQLite 安全迁移函数。

**同步时检查**：
- 上游是否修改了 `migrateDB()` 或 `migrateDBFast()` 的结构（如改为批量 slice、添加前置检查等）？确认我们追加的模型注册代码仍然执行到。
- 我们的 SQLite 迁移函数（`ensureCommissionRecordTableSQLite` 等）是否依然在 SQLite 路径中被调用？

---

### 6. StartCommissionSettleTask — main.go 启动任务

**文件**：`main.go`
**规则**：`service.StartCommissionSettleTask()` 必须在应用启动时调用，负责每 5 分钟将 `available_at` 到期的 pending 佣金结算为可提现状态。

**同步时检查**：
- 上游是否修改了 `main()` 的初始化顺序（如数据库初始化时机变化）？确认 `StartCommissionSettleTask()` 仍在 DB 初始化完成之后调用。

---

## 🟡 中风险 — 配置与扩展

### 7. KOL 配置键注册

**文件**：`model/option.go`
**规则**：`InitOptionMap()` 和 `UpdateOption()` 中注册了 `KolCommissionRate`、`StripeConnectEnabled`、`MinWithdrawalAmount`、`StripeConnectClientId` 四个配置键。

**同步时检查**：
- 上游是否在同一区域新增 option key，导致 switch-case 冲突或 InitOptionMap 内容被覆盖？

---

### 8. 路由组冲突

**文件**：`router/api-router.go`
**规则**：末尾新增了 `/kol`、`/kol/admin`、`/affiliate` 路由组。

**同步时检查**：
- 上游是否在文件末尾新增了路由，导致文本 merge 冲突？通常可安全合并（各加各的）。
- 上游是否新增了与 `/kol` 或 `/affiliate` 同名的路由？需检查是否产生路由冲突。

---

## 同步操作流程

```bash
# 1. 保存本地改动
git stash

# 2. 拉取上游
git fetch upstream
git merge upstream/main

# 3. 解决冲突（参考本文件各条目）
# ...

# 4. 恢复本地改动
git stash pop

# 5. 重新跑迁移/测试
go test ./model/...
go build ./...

# 6. 更新 docs/CUSTOM_CHANGES.md 中的"上次上游同步版本"
```
