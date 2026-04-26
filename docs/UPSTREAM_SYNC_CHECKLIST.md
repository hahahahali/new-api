# UPSTREAM_SYNC_CHECKLIST.md

> **用途**：每次从上游 [QuantumNous/new-api](https://github.com/QuantumNous/new-api) 拉取更新前，**先通读本文件**，再按 `docs/CUSTOM_CHANGES.md` 逐条核对。
> 本文件记录高风险自定义逻辑，一旦上游修改同一区域，极易造成静默错误或数据安全问题。

---

## 🔴 极高风险 — 数据安全

### 1. SQLite User 表迁移安全

**文件**：`model/main.go`
**规则**：**`&User{}` 绝对不能出现在 SQLite 路径的 `AutoMigrate` 列表中。**

SQLite migrator（`glebarez/sqlite`）在 schema 存在历史差异时（如 unique 约束元数据不同）会尝试重建整张表，导致真实用户数据丢失。我们用 `ensureUserTableSQLite()` 替代，逐列执行 `ALTER TABLE ADD COLUMN`。

**当前我们已在 `ensureUserTableSQLite()` 中追加的列（不可遗漏）**：
- `kol_balance`（decimal）
- `kol_history_balance`（decimal）
- `stripe_connect_account_id`（varchar）
- `stripe_connect_onboarded`（bool）
- `kol_rebate_rate`（decimal(5,4)，默认 0）— KOL 给被邀请者的推荐折扣率
- `lang`（varchar(8)，默认 'en'）— 用户邮件语言

**同步时检查**：
- 上游是否向 `migrateDB()` 或 `migrateDBFast()` 的 AutoMigrate 列表中**重新添加** `&User{}`？
- 如果有，**绝对不要** 将该变更应用到 SQLite 路径，只保留 MySQL/PostgreSQL 路径中的变更。
- 上游是否向 User struct 新增了字段？如果有，**必须同时**在 `ensureUserTableSQLite()` 中添加对应的 `ALTER TABLE ADD COLUMN`。
- 我们追加的上述 6 个字段是否仍然存在于 `ensureUserTableSQLite()` 函数中？merge 冲突时保留所有字段。

---

## 🔴 极高风险 — 权限与安全

### 2. KolAuth / ReviewerOrRootAuth 中间件

**文件**：`middleware/kol.go`
**规则**：`KolAuth()` 和 `ReviewerOrRootAuth()` 是新增的权限门控中间件，必须在 `UserAuth()` 之后调用。

**同步时检查**：
- `common.RoleRootUser` 常量是否有改动？ReviewerOrRootAuth 依赖此值。

---

### 2.1 Session 登录态必须回源校验当前权限状态

**文件**：`middleware/auth.go`
**规则**：session 登录态不能只信任 cookie 中缓存的 `role/status/group/username`，必须按 `id` 读取当前用户状态；否则用户被禁用、降权或换组后，旧 session 仍会继续生效。

**同步时检查**：
- 上游若修改 `authHelper()`，确认 session 分支仍会读取当前用户，而不是直接使用 `session.Get("role/status/group")` 做鉴权。
- 上游若修改 `TokenOrUserAuth()`，确认其 session 分支同样保留当前用户状态校验。

### 2.2 生产环境 session cookie 必须默认启用 Secure

**文件**：`main.go`
**规则**：session cookie 不能固定为 `Secure: false`，也不能按 `release` 模式或单一 `ServerAddress` 配置无脑开启；必须按**当前请求**的协议动态判定，只有在请求本身明确为 HTTPS（或 `SESSION_COOKIE_SECURE=true` 强制覆盖）时才应开启 `Secure`，否则默认关闭，避免 `ServerAddress=https://...` 但运维/开发实际通过 `http://局域网IP:端口` 访问时出现登录态丢失。

**同步时检查**：
- 上游若重构 `cookie.NewStore`、`store.Options(...)` 或 session 中间件，确认“默认 `Secure=false` + 按请求动态覆盖”的逻辑仍然生效。
- 若上游新增 cookie 选项配置，不能恢复成“release 模式默认强开 Secure”或“仅依赖 `ServerAddress` 全局判定”的策略。

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

### 3.1 前端富文本渲染必须经过白名单清洗

**文件**：`web/src/helpers/safeHtml.jsx` 及调用它的公告/FAQ/首页/关于/页脚/协议页组件
**规则**：后台可配的 Markdown/HTML 内容不能直接 `dangerouslySetInnerHTML` 到页面，必须经过 `sanitizeRichTextHtml` / `SafeHtml` 清洗，移除危险标签、事件属性和不安全链接。

**同步时检查**：
- 上游若新增 `marked.parse(...)+dangerouslySetInnerHTML` 的页面，必须改接 `renderMarkdownToSafeHtml()` 或 `SafeHtml`。
- 上游若修改这些页面的富文本渲染逻辑，确认没有回退为原始 HTML 直出。

### 3.2 GetTopUpInfo / GetPublicTopupPackages — 多语言套餐名称读取路径

**文件**：`controller/topup_packages.go`（新建文件，含 `resolveI18nNames` / `localizedAmountOptions` / `GetPublicTopupPackages`）；`controller/topup.go` 中 `GetTopUpInfo` 仅保留 1 行 `localizedAmountOptions(c)` 调用。
**规则**：`GetTopUpInfo`（已登录充值页）和 `GetPublicTopupPackages`（公开落地页）均通过 `localizedAmountOptions(c)` 统一走 `common.OptionMapRWMutex.RLock()` 读取原始 OptionMap 字符串，再调用 `resolveI18nNames(raw, lang)` 解析；**不能**回退为 `operation_setting.GetPaymentSetting().AmountOptionNames/Descs`（该 `[]string` 字段在存储多语言 JSON 对象 `{"zh":[...],"en":[...]}` 时会静默返回 nil）。

**同步时检查**：
- 上游若修改 `GetTopUpInfo` 的响应组装逻辑（响应 map 中的 `amount_option_names`/`amount_option_descs` 两个 key），需保留对 `localizedAmountOptions(c)` 的调用，以及 `?lang=` 参数支持。
- 上游若修改 `GetPublicTopupPackages` 的守卫条件，需保持改为 `setting.StripeUnitPrice <= 0`（原始条件为 `StripeApiSecret == "" || StripeWebhookSecret == ""`，该条件过严）。
- `resolveI18nNames` 和 `localizedAmountOptions` 位于 `controller/topup_packages.go`（我们新建的文件，上游不会修改），回退策略：指定语言 → "zh" → 第一个可用语言 → nil。

### 3.3 管理员调额 UI 默认按积分输入

**文件**：`web/src/components/table/users/modals/EditUserModal.jsx`、`web/src/helpers/creditQuota.js`
**规则**：管理员“调整额度”弹窗默认输入语义已从“金额”改为“积分”，并保留“金额输入”“原生额度输入”两个高级折叠项。最终提交给 `/api/user/manage` 的仍是 raw quota。

**同步时检查**：
- 上游若修改 `EditUserModal` 的调额弹窗，确认默认主输入仍是“积分”，不要回退成“金额”。
- 上游若修改 `quota_per_unit` 前端读取逻辑，确认 `creditQuota.js` 中 `getCreditDivisor()` 的 `quotaPerUnit / 100` 推导仍成立。
- 上游若新增管理员调额 API 字段，保持现有 `/api/user/manage` 的 raw quota 提交路径不变，避免前后端口径再次分叉。

---

### 4. ProcessCommission 钩子 — 支付回调中的佣金触发

**文件**：`controller/topup.go`、`controller/topup_creem.go`、`controller/topup_stripe.go`、`controller/topup_waffo.go`
**规则**：每个支付成功回调中必须有 `service.ProcessCommission(userId, tradeNo, amount)` 调用。

**同步时检查**：
- 上游是否新增了支付渠道文件（如 `controller/topup_xxx.go`）？如果有，**必须在新文件的支付成功处**添加 ProcessCommission 调用。
- 上游是否修改了现有 topup controller 中支付成功的代码路径？确认 ProcessCommission 调用仍在正确位置（成功状态判断通过之后，quota 增加之前或之后均可，但必须存在）。

### 4.1 KOL 推荐折扣必须在 pending 订单创建时占位

**文件**：`service/commission.go`、`controller/topup.go`、`controller/topup_stripe.go`
**规则**：KOL 推荐折扣的“前 3 单”资格不能只看已成功佣金单，必须同时计入当前 `pending` 充值订单；并且资格判断与 pending 订单创建必须在同一事务内完成，避免并发/提前下单绕过限制。

**同步时检查**：
- 上游若修改 `GetKolRebateForUser()` 或新增推荐折扣入口，确认其统计口径仍为“commission_records + pending top_ups”。
- 上游若调整 `RequestEpay` / `RequestPay` 的下单顺序，确认仍通过 `CreatePendingTopUpWithKolRebate()` 这类事务化入口创建 pending 订单，而不是先查资格、后单独插入订单。

---

### 4.2 Stripe 动态定价：`genStripeLinkPriceData` 与上游 `genStripeLink` 必须并列存在

**文件**：`controller/topup_stripe.go`、`setting/payment_stripe.go`
**规则**：上游的 `genStripeLink(amount int64, ...)` 函数体与签名**保留原样不动**（含其使用的 `setting.StripePriceId`），仅作为编译兼容存根存在；我们的 KOL 推荐折扣 + 阶梯定价场景一律走新增的 `genStripeLinkPriceData(payMoney float64, ...)`。`setting.StripePriceId` 变量在 `setting/payment_stripe.go` 中保留为 `var StripePriceId = ""` 仅用于编译，运行时始终为空，`model/option.go` 不再注册其为可配置项。

**同步时检查**：
- 上游若修改 `genStripeLink` 的签名、函数体或调用方（除 `RequestPay` 之外的其他入口），让它原样合并即可，**不要**把上游的修改往 `genStripeLinkPriceData` 里搬，两套实现独立维护。
- 上游若修改 `RequestPay` 中 `payLink, err := genStripeLink(...)` 这行调用，需确认我们的本地版本仍调用 `genStripeLinkPriceData(...)`，并且参数列表已改为 `payMoney float64`（来自 `topUp.Money`）。
- 上游若删除或重命名 `setting.StripePriceId` 变量，需在本地保留该变量（即使空字符串），否则上游 `genStripeLink` 函数体编译失败。
- `RequestAmount` 的 JSON 响应已新增 `original`（原价）和 `rebate_rate`（KOL 推荐折扣率）字段；前端依赖这两个字段渲染折扣 UI。上游若修改 `RequestAmount` 的响应结构，需保留这两个字段。

---

### 4.3 TopUp 模型新增 `OriginalMoney` 字段

**文件**：`model/topup.go`、`service/commission.go`
**规则**：`TopUp` 结构体新增 `OriginalMoney float64`（折扣前的美元原价），由 `CreatePendingTopUpWithKolRebate` 写入，`ProcessCommission` 反查后用于计算佣金，避免按折后金额返佣给 KOL 造成差额。

**同步时检查**：
- 上游若修改 `TopUp` 结构体（如新增字段、重排序、改类型），确认 `OriginalMoney` 字段仍存在且未被覆盖。
- 上游若新增对 `TopUp` 的 INSERT 入口（绕过 `CreatePendingTopUpWithKolRebate`），需评估是否需要补写 `OriginalMoney`，否则该订单的佣金计算会按折后金额走（可能少返佣给 KOL）。

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

### 8. 路由组冲突与速率限制保护

**文件**：`router/api-router.go`
**规则**：末尾新增了 `/kol`、`/kol/admin`、`/affiliate` 路由组，并对所有"会产生副作用"的接口加了速率限制中间件。

**同步时检查**：
- 上游是否在文件末尾新增了路由，导致文本 merge 冲突？通常可安全合并（各加各的）。
- 上游是否新增了与 `/kol` 或 `/affiliate` 同名的路由？需检查是否产生路由冲突。
- 以下高敏接口必须继续保留 `SecureVerificationRequired()`：
  - `POST /api/user/topup/complete`
- 以下接口必须继续保留 `CriticalRateLimit()`（防刷/防恶意调用）：
  - `POST /api/user/topup/complete`
  - `GET /api/topup/packages`（公开接口，未登录可调）
  - `POST /api/affiliate/status`（公开接口，邮箱+验证码查询）
  - `POST /api/affiliate/send-email-code`（实际用 `EmailVerificationRateLimit()`，更严格的邮件限流）
  - `POST /api/kol/withdraw`（提现申请，避免重复提交）
  - `PUT /api/kol/aff_code`（KOL 修改自己的邀请码）
  - `PUT /api/kol/admin/users/:id/aff_code`（管理员修改 KOL 邀请码）
  - `POST /api/kol/admin/withdrawals/:id/pay`（标记提现已打款）
- 上游若调整这些路由的中间件链，**不要**删除上述限流/二次校验中间件。

### 9. KOL 一次性邀请链接

**文件**：`controller/user.go`、`model/affiliate_application.go`
**规则**：`kol_token` 的消费必须保持原子性，当前实现依赖：
- 条件更新 `affiliate_applications`：仅 `status = approved && token_used = false` 时允许成功消费一次
- 在同一事务内同步完成 `users.group = 'kol'`

**同步时检查**：
- 上游若调整注册流程或 affiliate 模型，是否把 token 消费重新拆成“先查询、再分步更新”的非原子流程？
- 上游若改动 `token_used` / `kol_invite_token` 字段语义，是否仍能保证并发下只有一个账号能成功成为达人？

### 10. 达人申请状态查询必须依赖邮箱验证码

**文件**：`controller/affiliate.go`、`router/api-router.go`
**规则**：达人申请状态查询不能再提供”仅凭邮箱即可查询”的公开接口；当前实现要求通过 `POST /api/affiliate/status` 同时提交 `email + verify_code`，并且状态查询只校验验证码，不消费验证码，真正提交申请时才消费。验证通过后，接口同时返回该邮箱对应的完整申请数据（name/phone/country/instagram 等），因为调用方已证明邮箱归属，前端用这些数据回填表单。

**同步时检查**：
- 上游若修改达人申请路由，确认没有恢复 `GET /api/affiliate/apply?email=...` 这类可枚举邮箱状态的公开接口。
- 上游若修改邮箱验证码逻辑，确认状态查询仍需要验证码，且查询成功后验证码不会被提前消费，避免影响后续提交。
- 上游若修改 `GetAffiliateApplicationStatus` 的返回字段，确认仍包含完整申请数据（`name/phone/country/instagram/tiktok/youtube/other_social`），否则前端回填功能失效。

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
