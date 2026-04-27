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
| `controller/kol_admin.go` | 管理员：修改 aff_code、查看提现列表、确认 PayPal 打款 |
| `model/affiliate_application.go` | AffiliateApplication 数据模型（申请状态机、token 一次性使用） |
| `model/commission.go` | CommissionRecord 数据模型（佣金记录、汇总查询） |
| `model/withdrawal.go` | WithdrawalRequest 数据模型（提现申请、余额原子扣减） |
| `service/commission.go` | ProcessCommission（幂等）、StartCommissionSettleTask（定时冻结 → 可提现） |
| `service/withdrawal.go` | CreateWithdrawalTransfer — Stripe Connect 打款封装（已停用，保留注释） |
| `setting/payment_commission.go` | KolCommissionRate=0.20、StripeConnectEnabled=false、MinWithdrawalAmount=5.0 |
| `setting/payment_stripe_connect.go` | StripeConnectClientId 配置变量 |
| `common/email_templates.go` | 6类多语言邮件模板（zh/en/ja/fr/es）：注册验证码、密码重置、达人申请验证码、审核通过、审核拒绝、打款通知 |
| `controller/topup_packages.go` | 公开充值套餐接口 `GetPublicTopupPackages` + 多语言辅助函数 `resolveI18nNames` / `localizedAmountOptions`；从上游 `controller/topup.go` 拆出以最小化上游合并冲突面 |
| `model/topup_quota_custom.go` | `CalcQuotaByAmount(topUp)` — 基于套餐积分数（`Amount × 5000`）计算 quota，使折扣仅影响价格不影响用户获得的积分数；`CreditDivisor()` — 返回冻结的 quota → 显示积分除数（5000 = QuotaPerUnit/100），确保与前端 `model_price×100` 的模型费用显示一致 |
| `web/src/helpers/creditQuota.js` | 后台管理端积分↔quota 换算辅助函数：`getCreditDivisor()` / `quotaToCredits()` / `creditsToQuota()` / `formatCredits()`，供管理员调额 UI 默认按积分输入 |
| `docs/rules/deconflict.md` | 解耦与上游同步规则（本文件的使用规范） |
| `docs/rules/review.md` | Code Review 规则与审阅清单 |
| `web/src/helpers/safeHtml.jsx` | 前端富文本安全渲染与 HTML 白名单清洗工具 |

---

## 二、修改的上游文件

### `middleware/kol.go`（新增文件，不修改上游）

- `KolAuth()` — 检查 `user.group == "kol"`，需在 `UserAuth()` 之后使用
- `ReviewerOrRootAuth()` — 允许 `reviewer` 分组或 `role >= RoleRootUser`，管理员（role=10）无权限

---

### `middleware/auth.go`

- **改动**：`authHelper()` 对 session 登录态新增当前用户回源校验，不再仅信任 cookie 中缓存的 `role/status/group/username`；被降权、换组或禁用后的旧 session 将在服务端立即失效。
- **改动**：`TokenOrUserAuth()` 的 session 分支同样改为读取当前用户状态，避免禁用用户继续通过旧会话访问混合鉴权接口。
- **风险点**：上游若修改 `authHelper()` 或 `TokenOrUserAuth()`，同步时需保留“session 登录态使用当前 DB 状态”的逻辑，不能回退为仅信任 session 缓存字段。

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
- **改动 2**：新增 `Lang` 字段（用于多语言邮件通知）：
  ```go
  Lang string `gorm:"type:varchar(8);default:'en';column:lang"`
  ```
  来源：用户通过 kol_token 注册时，从 AffiliateApplication.Lang 复制过来。
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
  - `ensureUserTableSQLite()` — users 表新增列（含 `lang`）
  - `ensureCommissionRecordTableSQLite()`
  - `ensureWithdrawalRequestTableSQLite()`
  - `ensureAffiliateApplicationTableSQLite()` — affiliate_applications 表新增列（含 `lang`）
- **⚠️ 风险点（同步必查）**：SQLite User 迁移是本项目最危险的改动，**绝对不能将 `&User{}` 重新加入 SQLite 的 AutoMigrate 列表**。

---

### `model/option.go`

- **位置**：`InitOptionMap()` 和 `UpdateOption()` 的 switch-case
- **改动 1**：注册 4 个 KOL 配置键：`KolCommissionRate`、`StripeConnectEnabled`、`MinWithdrawalAmount`、`StripeConnectClientId`
- **改动 2**：删除全局 `StripePriceId` 的 `InitOptionMap` 初始化和 `UpdateOption` case（checkout 改用 `price_data` 动态定价后不再需要）
- **风险点**：上游若改动同一区域的 `UpdateOption` switch 可能冲突；同步时检查新增 case 是否重名。

---

### `controller/user.go`

- **改动 1**：`Register()` 函数中：
  - 仅接受 `group == “kol”` 用户的邀请码（非 KOL 邀请人的 inviterId 被置零）
  - 处理 URL 参数 `kol_token`，调用 `applyKolInviteToken()`（非致命失败，仅 SysLog）
- **改动 2**：新增 `applyKolInviteToken()` 私有函数，委托 `model.ConsumeKolInviteToken()` 以事务方式原子完成”一次性 token 消耗 + 用户升级为 kol”；同时从 AffiliateApplication 复制 `lang` 字段到新注册用户，用于后续多语言邮件通知
- **规则收口**：文件内 9 处直接调用的 `encoding/json`（`json.NewDecoder().Decode`、`json.Marshal`、`json.Unmarshal`）已全部替换为 `common.DecodeJson` / `common.Marshal` / `common.Unmarshal`；`encoding/json` import 已移除。
- **风险点**：上游若修改 `Register()` 的注册流程（如 inviter 处理逻辑），需确认 KOL 过滤和 kol_token 处理仍在正确位置，且不要回退为”先查 token，再分步更新 user/group 与 token_used”的非原子写法。

---

### `model/affiliate_application.go`

- **改动 1**：新增 `ConsumeKolInviteToken()`，通过 GORM 事务和条件更新实现 `kol_invite_token` 的原子消费：
  - 仅允许 `status = approved && token_used = false` 的记录被成功消费一次
  - 在同一事务中同时更新 `affiliate_applications.token_used = true` 和 `users.group = 'kol'`
- **改动 2**：新增 `Lang varchar(8)` 字段（`default:'en'`），存储申请人提交时的界面语言；注册成功后复制到 User.Lang 用于邮件语言判断。
- **规则收口**：`GetSetting()` / `SetSetting()` / `GetUserDefaultConfig()` 中直接调用的 `encoding/json` 已替换为 `common.Unmarshal` / `common.Marshal`；`encoding/json` import 已移除。
- **风险点**：上游若修改 affiliate 申请模型或 token 字段语义，需保留”条件更新抢占 + 同事务内升级用户分组”的约束，不能拆回两次独立写入。

---

### `controller/affiliate.go`

- **改动 1**：达人申请提交现在要求邮箱验证码；验证码校验通过后才允许创建或更新申请记录。
- **改动 2**：状态查询接口已从公开 `GET /api/affiliate/apply?email=...` 收口为 `POST /api/affiliate/status`，必须同时提交 `email + verify_code` 才能查询该邮箱的申请状态；查询只校验验证码，不消费验证码，真正提交申请时才消费。
- **改动 3**：`GetAffiliateApplicationStatus` 在邮箱验证码校验通过后，额外返回申请人自己填写的完整信息（`name/phone/country/instagram/tiktok/youtube/other_social`），用于前端回填表单。因为调用方已通过验证码证明邮箱归属，此处暴露自身数据是安全且合理的。
- **改动 4**：申请提交和发验证码接口均接收 `lang` 字段；三个邮件发送函数改用 `common.BuildXxxEmail()` 多语言模板。
- **风险点**：上游若修改达人申请控制器或邮箱验证码逻辑，需保留”状态查询必须依赖邮箱验证码”的约束，不能恢复为仅凭邮箱公开查询状态。

---

### `controller/oauth.go`

- **改动**：OAuth 注册流程中，同样只接受 `group == "kol"` 用户的邀请码。
- **风险点**：上游若修改 OAuth 注册流程，需同步检查 KOL invite 过滤是否保留。

---

### `model/topup.go`

- **位置**：`TopUp` struct 字段定义
- **改动**：新增 `OriginalMoney float64` 字段（`gorm:"default:0"`），存储折扣前原价（USD），供 ProcessCommission 计算净佣金使用。
- **改动 2**：`Recharge()` 函数第 133 行，充值 quota 计算公式从 `topUp.Money * common.QuotaPerUnit` 改为 `CalcQuotaByAmount(topUp)`（定义在新文件 `model/topup_quota_custom.go`）。原公式使用折扣后的实付金额计算 quota，导致不同折扣率的套餐充值后用户获得的积分数不等于套餐标称值（如 6000 积分套餐实际只显示 122.5）。新公式使用冻结乘数 `Amount × 5000`（= QuotaPerUnit/100），确保"买 N 积分就得 N 积分"且与前端 `model_price×100` 模型费用显示一致；乘数冻结后 StripeUnitPrice 调价不影响已有余额的积分显示。
- **迁移说明**：`TopUp` 表保留在所有数据库路径的 `AutoMigrate` 列表中，新列会自动添加，无需手工 `ALTER TABLE`。
- **⚠️ 风险点（同步必查）**：上游若修改 TopUp struct（如新增字段或修改类型），确认 OriginalMoney 字段仍存在且顺序无冲突。

---

### `controller/topup.go` / `controller/topup_creem.go` / `controller/topup_stripe.go` / `controller/topup_waffo.go`

- **改动 1**：每个支付回调成功后调用 `service.ProcessCommission(userId, tradeNo, paidUSD)` 3 参数版本（幂等，失败仅记日志）。折扣前原价由 `ProcessCommission` 内部按 `tradeNo` 反查 `TopUp.OriginalMoney` 获得，**调用方无需传 OriginalMoney**，最小化与上游 topup*.go 的 merge 冲突面。
- **改动 1.5**（仅 `topup_stripe.go` + 新增 `controller/topup_stripe_guard.go`）：上游 `RequestPay` 硬编码 `Amount > 10000` 上限被放宽为 `Amount > 10000000`（安全兜底），真正的入参校验改由新建文件 `topup_stripe_guard.go` 中的 `GuardedRequestStripePay` 白名单逻辑承担——只允许 `AmountOptions` 配置的金额通过，防止用户提交任意金额。上游若修改原始 10000 上限，需保留 10000000 阈值并确认 guard 白名单仍生效。
- **改动 2**（仅 `topup_stripe.go`）：
  - **保留上游 `genStripeLink(amount int64, ...)` 函数体不动**（含其使用的 `setting.StripePriceId`），仅作为上游兼容存根存在。
  - **新增并列函数 `genStripeLinkPriceData(payMoney float64, ...)`**：基于 Stripe `price_data` 的动态定价实现（`UnitAmount = payMoney × 100` 分），用于 KOL 推荐折扣 + 阶梯定价场景下每笔订单金额都不同的需求。
  - `RequestPay` 的 `payLink` 调用从 `genStripeLink` 改为 `genStripeLinkPriceData`，并叠加 KOL 推荐折扣率（`GetKolRebateForUser`），将 `OriginalMoney`（未折扣金额）写入 TopUp 记录，供 ProcessCommission 内部读取。
  - `RequestAmount` 同步返回 `original`（原价）和 `rebate_rate`（KOL 推荐折扣率）字段。
  - **解耦原则**：保留上游 `genStripeLink` 完整签名与函数体，意味着上游对该函数的任何修改（bug fix、Stripe SDK 升级）都能 merge 干净；我们的实现走完全独立的 `genStripeLinkPriceData`，互不影响。
- **改动 3**：`topup.go` / `topup_stripe.go` 创建充值订单时，改为通过 `service.CreatePendingTopUpWithKolRebate()` 在事务内先锁定 invitee、按”已成功佣金单 + 当前 pending 充值单”计算推荐折扣资格，再落 pending 订单，堵住并发/提前下单绕过”前 3 单折扣”限制的问题；若拉起支付失败则立即将该 pending 订单标记为 `failed` 释放名额。
- **改动 6**（新文件 `controller/topup_stripe_guard.go` + `router/api-router.go` 1 行 + `topup_stripe.go` 1 处数值）：新增 `GuardedRequestStripePay` 包装函数，在调用上游 `RequestPay` 之前校验 `Amount` 是否在 `AmountOptions` 白名单中。路由 `/stripe/pay` 从 `RequestStripePay` 改为 `GuardedRequestStripePay`（`api-router.go` 仅改 1 行）。同时将 `topup_stripe.go` 中 `RequestPay` 内的硬编码上限从 `10000` 放宽为 `10000000`（仅改 1 个数字），使其不再阻断合法大额套餐；实际金额校验由 guard 的 `AmountOptions` 白名单负责。修复原因：`AmountOptions` 可配置超过 10000 的值（如 180000），上游硬编码上限导致大额套餐支付被拒。
- **改动 4**（仅 `topup.go`）：新增 `GetPublicTopupPackages()` — 无需鉴权的充值套餐接口，供落地页未登录访客查看实时定价。
  - **已拆离到 `controller/topup_packages.go`**：该函数及其私有辅助函数 `resolveI18nNames` / `localizedAmountOptions` 已全部迁移到新文件，上游 `controller/topup.go` 里只保留 `GetTopUpInfo` 顶部一行 `localizedAmountOptions(c)` 调用和响应 map 里两个字段。此拆分**将本条改动与上游 `topup.go` 的冲突面从 ~60 行降到 ~2 行**。
  - `PackageInfo` 新增 `Name string` 和 `Description string` 字段，分别来自 `AmountOptionNames` / `AmountOptionDescs` 配置，供落地页展示套餐标题和副标题。
  - `Credits` 字段直接等于 `amount`（即 `AmountOptions` 里的原始值），不再乘以 25；站点通过 `AmountOptions` 直接配置积分数量，无需换算。
  - `GetTopUpInfo` 同步返回 `amount_option_names` 和 `amount_option_descs` 字段，供已登录的充值页面使用。
  - 使用 `getStripePayMoney(amount, "default")` 计算标准价格（不含 KOL 折扣），仅返回最终价格，不暴露折扣表、StripeUnitPrice 等内部配置；Stripe 未配置时返回错误。
- **改动 5**（仅 `topup_packages.go`，新文件）：多语言套餐名称/描述支持。
  - `resolveI18nNames(raw, lang string) []string`：兼容纯 `[]string` 和多语言对象 `{"zh":[...],"en":[...]}` 两种存储格式；回退策略：指定语言 → "zh" → 第一个可用语言 → nil。
  - `GetPublicTopupPackages` 守卫条件由 `StripeApiSecret == "" || StripeWebhookSecret == ""` 改为 `setting.StripeUnitPrice <= 0`（价格展示不依赖 Stripe 密钥）。
  - `GetPublicTopupPackages` 响应新增 `credit_divisor` 字段（`model.CreditDivisor()`），前端通过该值将 quota 换算为显示积分数，无需硬编码除数。
  - `GetPublicTopupPackages` 与 `GetTopUpInfo` 均通过 `localizedAmountOptions(c)` 统一走 `common.OptionMapRWMutex.RLock()` 读取原始 `OptionMap` 字符串，再调用 `resolveI18nNames` 解析；两个接口均支持 `?lang=` 查询参数。
- **⚠️ 风险点**：
  - 上游若新增支付渠道或修改回调流程，只需在成功分支加一行 3 参数调用即可接入佣金系统。
  - 上游若修改 `RequestPay` / `RequestEpay` 的下单顺序，需保留”折扣资格判断与 pending 订单创建在同一事务内完成”的约束，不能回退为先查资格、后单独插入订单。
  - `topup.go` 中 `enable_stripe_topup` 检查已去除 `StripePriceId` 依赖，仅校验 `StripeApiSecret` + `StripeWebhookSecret`。`model/option.go` 已删除 `StripePriceId` 的 InitOptionMap/UpdateOption 注册，前端设置页（`SettingsPaymentGatewayStripe.jsx`、`PaymentSetting.jsx`）也删除了输入字段。**`setting/payment_stripe.go` 中 `var StripePriceId = “”` 已恢复并保留**——仅作为编译兼容存根存在，让上游 `genStripeLink` 函数体能编译通过；运行时该变量始终为空，我们的支付链路不依赖它。`model/subscription.go` 中各套餐自有的 `StripePriceId` 字段不受影响。
  - 上游若修改 `getStripePayMoney` 签名或 `AmountOptions` 结构，`GetPublicTopupPackages` 需同步调整。
  - `Credits` 字段语义已从”amount × 25 换算后的 token 数”改为”直接等于 amount”；如果未来需要换算比例，请修改 `GetPublicTopupPackages` 内的计算逻辑，**不要**回退为 `amount * 25`。
  - 上游若修改 `GetTopUpInfo` 或 `GetPublicTopupPackages` 的响应组装逻辑，需保留”持 RLock 读取原始 OptionMap + resolveI18nNames 调用”的路径，**不能**回退为 `operation_setting.GetPaymentSetting().AmountOptionNames/Descs`（该 `[]string` 字段在存储多语言 JSON 对象时会静默返回 nil）。

---

### `setting/operation_setting/payment_setting.go`

- **改动**：`PaymentSetting` struct 新增两个字段：
  - `AmountOptionNames []string \`json:"amount_option_names"\`` — 每个充值档位的展示名称，与 `AmountOptions` 一一对应
  - `AmountOptionDescs []string \`json:"amount_option_descs"\`` — 每个充值档位的副标题描述，与 `AmountOptions` 一一对应
- **风险点**：上游若修改 `PaymentSetting` 结构体，merge 时确认两个新字段仍存在且 JSON key 无冲突。

---

### `router/api-router.go`

- **改动**：在现有路由末尾新增：
  - `POST /api/auth/logout` — 前端 POST 登出的别名
  - `POST /api/user/logout` — 同上
  - KOL 路由组 `/api/kol/*`（需 `UserAuth` + `KolAuth`）
  - KOL 管理路由组 `/api/kol/admin/*`（需 `AdminAuth`）
  - 公开达人申请路由 `POST /api/affiliate/apply`、`POST /api/affiliate/send-email-code`、`POST /api/affiliate/status`
  - 达人审核路由组 `/api/affiliate/*`（需 `UserAuth` + `ReviewerOrRootAuth`）
  - **`GET /api/topup/packages`** — 公开充值套餐定价接口（无需登录，加 `CriticalRateLimit()`）
- **改动**：为 `POST /api/user/topup/complete` 追加 `CriticalRateLimit()` + `SecureVerificationRequired()`。
- **风险点**：上游若在 apiRouter 末尾追加路由，merge 冲突概率较低但需检查。

---

### `main.go`

- **改动 1**：在 `main()` 初始化序列末尾追加 `service.StartCommissionSettleTask()`。
- **改动 2**：session cookie 的 `Secure` 改为**按请求动态判定**：默认 `false`，每次请求根据 `TLS` / `X-Forwarded-Proto` / `Forwarded` 等协议头自动决定；仅在当前请求明确为 HTTPS（或 `SESSION_COOKIE_SECURE=true` 强制开启）时才下发 `Secure` cookie，避免 `ServerAddress=https://...` 但实际通过 `http://局域网IP:端口` 访问时浏览器丢弃登录 cookie，出现“登录成功后访问任意鉴权接口立刻掉线”。
- **风险点**：
  1. 上游若在 main 末尾新增 goroutine/task 启动，需检查顺序是否合理。
  2. 上游若修改 session store 初始化，需保留“默认 `Secure=false` + 按请求动态覆盖”的逻辑，不能回退为固定 `Secure: false`，也不能恢复到“按 `ServerAddress` / release 模式一次性全局强开 Secure”的策略。

---

### `model/task_cas_test.go`（上游测试文件）

- **改动**：`TestMain` 的 `db.AutoMigrate(...)` 列表尾部追加 `&AffiliateApplication{}`；`truncateTables` 的 `DB.Exec` 序列追加 `DELETE FROM affiliate_applications`。
- **原因**：引入 `AffiliateApplication` 模型后，任务级 CAS 测试若不在测试库中注册该表，相关集成测试会因表缺失报错。
- **风险点**：上游若修改该测试文件（例如新增/移除 AutoMigrate 的模型、重构 truncate 流程），merge 时需保留我们追加的 `&AffiliateApplication{}` 注册和对应 DELETE 语句；同步新模型时也要在这里同步追加。

---

### `service/task_billing_test.go`（上游测试文件）

- **改动**：`TestMain` 的 `AutoMigrate(...)` 列表尾部追加 `&model.CommissionRecord{}` 和 `&model.TopUp{}`；`truncate` 的 `model.DB.Exec` 序列追加 `DELETE FROM commission_records` 与 `DELETE FROM top_ups`。
- **原因**：KOL 佣金系统在任务计费路径上会查询 `commission_records` 与 `top_ups` 表，若测试库未建表则计费相关测试失败。
- **风险点**：上游若修改该测试文件（如调整 AutoMigrate 列表或清理流程），merge 时必须保留两张表的注册与 DELETE 语句；后续若再新增与计费相关的模型，应在此处同步追加。

---

### `controller/misc.go`

- **位置**：`SendEmailVerification()` 和 `SendPasswordResetEmail()` 两个函数内
- **改动 1**（`SendEmailVerification`）：删除原本硬编码的中文邮件主题/正文赋值，改为读取 `lang` 查询参数（`c.DefaultQuery("lang", "zh")`）并调用 `common.BuildRegistrationVerificationEmail(lang, ...)` 多语言模板。
- **改动 2**（`SendPasswordResetEmail`）：同上，改为调用 `common.BuildPasswordResetEmail(lang, ...)`。
- **背景**：前端 `api.js` 已在发送验证码/重置密码请求时附带 `&lang=<locale>` 参数，但后端两个函数之前硬编码中文，导致切换语言后收到的邮件仍为中文。两个多语言模板函数（zh/en/ja/fr/es）已实现在 `common/email_templates.go`（我方新增文件）中，`misc.go` 仅增加 2 行 lang 读取 + 1 行模板调用，改动范围最小化。
- **⚠️ 风险点（同步必查）**：上游若修改 `SendEmailVerification` 或 `SendPasswordResetEmail` 的邮件发送逻辑，merge 时需保留"读取 `lang` 参数 + 调用 `common.Build*Email`"的两行，不能回退为硬编码中文内容。

---

### `controller/payment_webhook_availability.go`

- **改动**：`isStripeTopUpEnabled()` 移除 `StripePriceId != ""` 条件，仅保留 `StripeApiSecret` + `StripeWebhookSecret` 两项校验。原因：项目已改用 `genStripeLinkPriceData`（动态定价），不再使用全局 `StripePriceId`；而 `StripePriceId` 已从 `InitOptionMap` / `UpdateOption` 中删除，`setting.StripePriceId` 始终为空，导致 `isStripeWebhookEnabled()` 永远返回 `false`，Stripe webhook 被 403 拒绝，用户付款后积分无法到账。
- **同步修改**：`payment_webhook_availability_test.go` 中 `TestStripeWebhookEnabledRequiresTopUpAndWebhookConfig` 同步去除 `StripePriceId` 相关保存/恢复/断言。
- **⚠️ 风险点（同步必查）**：上游若修改 `isStripeTopUpEnabled` 条件（如新增其他必填配置项），merge 时需确认不恢复 `StripePriceId` 条件。

---

### `.gitignore`

- **改动**：新增 `PROJECT_MASTER.md` 忽略规则，将本地联合开发统领文件排除在 git 之外，避免误提交本地开发/部署信息。
- **风险点**：若上游后续修改根目录 `.gitignore`，同步时需保留 `PROJECT_MASTER.md` 的本地忽略规则。

---

### 前端文件（`web/`）

| 文件 | 改动摘要 |
|---|---|
| `web/src/App.jsx` | 新增 KolDashboard 路由、AffiliateApply 路由、AffiliateReview 路由 |
| `web/src/helpers/safeHtml.jsx` | 新增 `sanitizeRichTextHtml` / `renderMarkdownToSafeHtml` / `SafeHtml`，统一收口富文本渲染的 XSS 面 |
| `web/src/components/auth/RegisterForm.jsx` | 支持 `kol_token` URL 参数，注册时传递给后端 |
| `web/src/components/layout/SiderBar.jsx` | 新增"达人中心"和"审核中心"菜单组（按 group/role 显示） |
| `web/src/components/settings/PaymentSetting.jsx` | 删除 `StripePriceId` 初始 state；新增 `AmountOptionNames` / `AmountOptionDescs` 初始 state 及对应 switch case（key `payment_setting.amount_option_names` / `payment_setting.amount_option_descs`） |
| `web/src/pages/Setting/Payment/SettingsGeneralPayment.jsx` | 新增 `AmountOptionNames` / `AmountOptionDescs` state、验证、提交逻辑（key `payment_setting.amount_option_names` / `payment_setting.amount_option_descs`）及 TextArea UI 组件 |
| `web/src/components/table/users/UsersColumnDefs.jsx` | 用户列表新增 KOL 字段列 |
| `web/src/components/table/users/UsersTable.jsx` | 配合上述列变更 |
| `web/src/components/common/DocumentRenderer/index.jsx` | HTML 文档页不再直接注入原始 HTML，统一改为白名单清洗后的安全渲染 |
| `web/src/components/dashboard/AnnouncementsPanel.jsx` / `web/src/components/dashboard/FaqPanel.jsx` / `web/src/components/layout/NoticeModal.jsx` | 公告/FAQ/通知的 `marked + dangerouslySetInnerHTML` 改为安全 HTML 渲染，阻断脚本注入 |
| `web/src/pages/Home/index.jsx` / `web/src/pages/About/index.jsx` / `web/src/components/layout/Footer.jsx` / `web/src/components/settings/OtherSetting.jsx` / `web/src/helpers/utils.jsx` | 首页/关于/页脚/更新弹窗/HTML toast 全部切换到统一安全渲染器，保留富文本能力但移除危险标签、事件属性和不安全链接 |
| `web/src/components/topup/index.jsx` | 新增 `stripeRebateRate` / `stripeOriginalAmount` state，从 `RequestAmount` 响应中解析并传入 `PaymentConfirmModal` |
| `web/src/components/topup/modals/PaymentConfirmModal.jsx` | 新增 `stripeRebateRate` / `stripeOriginalAmount` props；当 `payWay === 'stripe'` 且存在推荐折扣时展示原价删除线 + "推荐折扣" Tag |
| `web/src/components/table/users/modals/EditUserModal.jsx` | 管理员“调整额度”弹窗默认改为按“积分”输入；新增高级折叠项“金额输入”“原生额度输入”，三者双向换算并统一提交 raw quota 给 `/api/user/manage`，以减少将金额误当积分导致的错调 |
| `web/src/components/table/users/modals/EditAffCodeModal.jsx` | 引入 `useSecureVerification`，修改达人邀请码（`PUT /api/kol/admin/users/:id/aff_code`）前强制触发二次身份校验 |
| `web/src/components/topup/modals/TopupHistoryModal.jsx` | 引入 `useSecureVerification`，手动补单（`POST /api/user/topup/complete`）前强制触发二次身份校验 |
| `web/src/helpers/auth.jsx` | 新增 `ReviewerRoute` 路由守卫 |
| `web/src/helpers/utils.jsx` | 新增 `isReviewer()` 工具函数 |
| `web/src/pages/Setting/Payment/SettingsPaymentGatewayStripe.jsx` | 删除 `StripePriceId` 输入字段及提交逻辑；第一行 Col 宽度从 `md={8}` 调整为 `md={12}`；`StripeUnitPrice` 输入框 `precision` 从 `2` 改为 `7`，支持填写 0.0008167 这类 7 位小数单价（后端 `StripeUnitPrice` 为 `float64`，天然支持；merge 上游时若该行被改动，需保留 `precision={7}`） |
| `web/src/pages/KolDashboard/index.jsx` | （新增文件）"返佣比例"卡片改为实时显示 `commission_rate * 100 - rebateRate`，随推荐折扣滑块联动更新，反映让利后的净佣金比例 |
| `web/src/i18n/locales/*.json` | 所有 7 个语言文件补齐 KOL/达人申请/审核中心相关 key；补充"推荐折扣"、"推荐折扣已更新"及折扣说明/示例文案 key；新增后台调额相关文案 key（"积分"、"输入积分"、"使用金额输入"、"收起金额输入"） |

---

## 三、已知待处理事项

| # | 问题 | 严重度 | 状态 |
|---|---|---|---|
| 1 | `QuotaForInvitee`/`QuotaForInviter` 管理面板配置项仍存在但逻辑已删除 | 中 | 待确认是否保留 UI 或彻底移除 |
| 2 | `kol_token` 应用失败时用户静默注册为普通用户（非 KOL） | 中 | 待确认是否改为报错阻断 |
| 3 | `GetKolInvitees` 存在 N+1 查询（每个 invitee 单独查 commission summary） | 低 | 待优化为 GROUP BY 批量查询 |
| 4 | 提现"申请"和"审核通过"两步未在同一事务内（可能状态短暂不一致） | 低 | 待确认是否需要合并为原子操作 |
