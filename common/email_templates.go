package common

import (
	"fmt"
	"strings"
)

// normalizeEmailLang returns a supported email language code, falling back to "en".
// Handles locale variants like zh-CN/zh-TW → zh.
func normalizeEmailLang(lang string) string {
	// Normalize zh variants (zh-CN, zh-TW, zh-Hans, zh-Hant, etc.)
	if strings.HasPrefix(lang, "zh") {
		return "zh"
	}
	switch lang {
	case "en", "ja", "fr", "es", "ru", "vi":
		return lang
	default:
		return "en"
	}
}

// appendLangQuery appends ?lng=<code> (or &lng=<code>) to a URL, preserving any existing fragment.
func appendLangQuery(rawURL, code string) string {
	if rawURL == "" {
		return rawURL
	}
	base, frag := rawURL, ""
	if i := strings.Index(rawURL, "#"); i >= 0 {
		base, frag = rawURL[:i], rawURL[i:]
	}
	sep := "?"
	if strings.Contains(base, "?") {
		sep = "&"
	}
	return base + sep + "lng=" + code + frag
}

// ── v7 email helpers ──────────────────────────────────────────────────────────

const emailLogoBase64 = `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`

func emailWrapper(mainContent string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"><meta name="format-detection" content="telephone=no,date=no,address=no,email=no,url=no"><meta name="x-apple-disable-message-reformatting"><meta name="color-scheme" content="light only"><meta name="supported-color-schemes" content="light only"></head>
<body style="margin:0;padding:0;background:linear-gradient(145deg,#E8EAFF 0%%,#F2ECFF 45%%,#E5F0FF 100%%);font-family:-apple-system,BlinkMacSystemFont,'SF Pro Display','Segoe UI',Arial,sans-serif;">
<table width="100%%" cellpadding="0" cellspacing="0" border="0" style="background:linear-gradient(145deg,#E8EAFF 0%%,#F2ECFF 45%%,#E5F0FF 100%%);padding:56px 16px;">
  <tr><td align="center">
  <table width="560" cellpadding="0" cellspacing="0" border="0" style="max-width:560px;width:100%%;background:#FFFFFF;border-radius:28px;box-shadow:0 0 0 1px rgba(180,168,240,0.18),0 8px 24px rgba(120,100,210,0.08),0 32px 72px rgba(110,90,220,0.12);overflow:hidden;">
    <tr>
      <td align="center" style="padding:44px 52px 28px;">
        <table cellpadding="0" cellspacing="0" border="0"><tr>
          <td style="vertical-align:middle;"><img src="data:image/png;base64,%s" width="30" height="30" alt="Logo" style="display:block;border-radius:7px;"></td>
          <td style="padding-left:9px;vertical-align:middle;"><span style="font-size:15px;font-weight:700;color:#1D1D1F;letter-spacing:-0.3px;">api.loriqo</span></td>
        </tr></table>
      </td>
    </tr>
    <tr><td style="padding:0 52px;"><div style="height:1px;background:rgba(195,186,242,0.28);"></div></td></tr>
    %s
    <tr><td style="padding:0 52px 40px;">
      <div style="height:1px;background:rgba(195,186,242,0.28);margin-bottom:22px;"></div>
      <p style="margin:0;font-size:11px;color:#AEAEB2;text-align:center;line-height:2;">
        &copy; 2026 api.loriqo &nbsp;&middot;&nbsp;
        <a href="#" style="color:#6E6E73;text-decoration:none;">帮助中心</a>&nbsp;&middot;&nbsp;
        <a href="#" style="color:#6E6E73;text-decoration:none;">隐私政策</a>
      </p>
    </td></tr>
  </table>
  </td></tr>
</table>
</body>
</html>`, emailLogoBase64, mainContent)
}

func emailBadgePurple(label string) string {
	return fmt.Sprintf(`<span style="display:inline-block;background:rgba(79,70,229,0.08);color:#4F46E5;border:1px solid rgba(79,70,229,0.18);font-size:11px;font-weight:600;border-radius:100px;padding:4px 13px;letter-spacing:0.04em;">%s</span>`, label)
}

func emailBadgeGreen(label string) string {
	return fmt.Sprintf(`<span style="display:inline-block;background:rgba(16,185,129,0.08);color:#059669;border:1px solid rgba(16,185,129,0.2);font-size:11px;font-weight:600;border-radius:100px;padding:4px 13px;letter-spacing:0.04em;">%s</span>`, label)
}

func emailBadgeAmber(label string) string {
	return fmt.Sprintf(`<span style="display:inline-block;background:rgba(245,158,11,0.09);color:#D97706;border:1px solid rgba(245,158,11,0.2);font-size:11px;font-weight:600;border-radius:100px;padding:4px 13px;letter-spacing:0.04em;">%s</span>`, label)
}

func emailPrimaryBtn(href, label string) string {
	return fmt.Sprintf(`<table width="100%%" cellpadding="0" cellspacing="0" border="0"><tr><td align="center"><a href="%s" style="display:inline-block;background:linear-gradient(180deg,#5B54F2 0%%,#4F46E5 100%%);box-shadow:0 4px 14px rgba(79,70,229,0.32),inset 0 1px 0 rgba(255,255,255,0.18);color:#FFFFFF;font-size:14px;font-weight:600;text-decoration:none;padding:15px 48px;border-radius:14px;letter-spacing:0.01em;">%s</a></td></tr></table>`, href, label)
}

func emailSecondaryBtn(href, label string) string {
	return fmt.Sprintf(`<table width="100%%" cellpadding="0" cellspacing="0" border="0"><tr><td align="center"><a href="%s" style="display:inline-block;background:linear-gradient(180deg,rgba(255,255,255,0.96) 0%%,rgba(246,242,255,0.82) 100%%);border:1px solid rgba(195,186,242,0.45);color:#475569;font-size:14px;font-weight:600;text-decoration:none;padding:15px 48px;border-radius:14px;box-shadow:inset 0 1px 0 rgba(255,255,255,1),0 2px 8px rgba(120,100,210,0.07);letter-spacing:0.01em;">%s</a></td></tr></table>`, href, label)
}

func emailInfoBox(body string) string {
	return fmt.Sprintf(`<div style="background:linear-gradient(180deg,rgba(255,255,255,0.96) 0%%,rgba(246,242,255,0.82) 100%%);border:1px solid rgba(195,186,242,0.45);border-radius:13px;padding:14px 18px;box-shadow:inset 0 1px 0 rgba(255,255,255,1),0 2px 8px rgba(120,100,210,0.07);"><p style="margin:0;font-size:12px;color:#6E6E73;line-height:1.75;">%s</p></div>`, body)
}

func emailWarningBox(title, body string) string {
	return fmt.Sprintf(`<div style="background:rgba(254,243,199,0.55);border:1px solid rgba(251,191,36,0.28);border-radius:14px;padding:18px 20px;margin-bottom:24px;box-shadow:inset 0 1px 0 rgba(255,255,255,0.9);"><p style="margin:0 0 5px;font-size:11px;font-weight:600;color:#B45309;letter-spacing:0.05em;text-transform:uppercase;">%s</p><p style="margin:0;font-size:13px;color:#6E6E73;line-height:1.65;">%s</p></div>`, title, body)
}

func emailStatCard2(label1, val1, label2, val2 string) string {
	cell := `<div style="background:linear-gradient(180deg,rgba(255,255,255,0.96) 0%%,rgba(246,242,255,0.82) 100%%);border:1px solid rgba(195,186,242,0.45);border-radius:14px;padding:16px 18px;box-shadow:inset 0 1px 0 rgba(255,255,255,1),0 2px 8px rgba(120,100,210,0.07);"><p style="margin:0 0 4px;font-size:11px;font-weight:600;color:#AEAEB2;letter-spacing:0.05em;text-transform:uppercase;">%s</p><p style="margin:0;font-size:20px;font-weight:700;color:#1D1D1F;">%s</p></div>`
	return fmt.Sprintf(`<table width="100%%" cellpadding="0" cellspacing="0" border="0"><tr><td width="50%%" style="padding:0 7px 0 0;vertical-align:top;">`+cell+`</td><td width="50%%" style="padding:0 0 0 7px;vertical-align:top;">`+cell+`</td></tr></table>`,
		label1, val1, label2, val2)
}

func emailStatCard3(label1, val1, label2, val2, label3, val3 string) string {
	const cellStyle = `width="30%%" height="124" valign="middle" align="center" style="width:30%%;height:124px;background:#F8F5FF;border:1px solid rgba(195,186,242,0.45);border-radius:14px;text-align:center;vertical-align:middle;padding:14px 8px;"`
	const gap = `<td width="5%%" style="width:5%%;font-size:0;line-height:0;">&nbsp;</td>`
	const valStyle = `style="display:block;margin:0 0 6px;font-size:22px;font-weight:800;color:#4F46E5;letter-spacing:-0.5px;line-height:1.1;font-family:-apple-system,BlinkMacSystemFont,'SF Pro Display','Segoe UI',Arial,sans-serif;"`
	const labStyle = `style="display:block;margin:0;font-size:11px;color:#8E8E93;font-weight:500;line-height:1.4;font-family:-apple-system,BlinkMacSystemFont,'SF Pro Display','Segoe UI',Arial,sans-serif;"`
	cell := `<td ` + cellStyle + `><span ` + valStyle + `>%s</span><span ` + labStyle + `>%s</span></td>`
	return fmt.Sprintf(`<table width="100%%" cellpadding="0" cellspacing="0" border="0" style="margin-bottom:28px;border-collapse:separate;"><tr>`+cell+gap+cell+gap+cell+`</tr></table>`,
		val1, label1, val2, label2, val3, label3)
}

func buildCodeDigitCells(code string) string {
	const digitStyle = `width="54" height="64" style="width:54px;height:64px;background:linear-gradient(180deg,rgba(255,255,255,0.96) 0%%,rgba(246,242,255,0.82) 100%%);border:1px solid rgba(195,186,242,0.45);border-radius:14px;text-align:center;vertical-align:middle;font-size:32px;font-weight:800;color:#1D1D1F;box-shadow:inset 0 1px 0 rgba(255,255,255,1),0 2px 8px rgba(120,100,210,0.07);font-variant-numeric:tabular-nums;"`
	const sp4 = `<td width="4" style="width:4px;font-size:0;line-height:0;">&nbsp;</td>`
	const sp20 = `<td width="20" style="width:20px;font-size:0;line-height:0;">&nbsp;</td>`
	var out strings.Builder
	out.WriteString(`<table cellpadding="0" cellspacing="0" border="0" align="center"><tr>`)
	for i, ch := range code {
		if i > 0 {
			if i == 3 {
				out.WriteString(sp20)
			} else {
				out.WriteString(sp4)
			}
		}
		out.WriteString(fmt.Sprintf(`<td %s>%c</td>`, digitStyle, ch))
	}
	out.WriteString(`</tr></table>`)
	return out.String()
}

// ── end helpers ───────────────────────────────────────────────────────────────

// BuildAffiliateVerificationEmail returns subject and HTML body for the email verification code.
func BuildAffiliateVerificationEmail(lang, systemName, code string) (subject, content string) {
	lang = normalizeEmailLang(lang)

	type copy struct {
		subject  string
		badge    string
		heading  string
		subtitle string
		security string
		noShare  string
	}

	var t copy
	switch lang {
	case "zh":
		t = copy{
			subject:  fmt.Sprintf("【%s】邮箱验证码", systemName),
			badge:    "达人申请验证",
			heading:  "邮箱验证码",
			subtitle: fmt.Sprintf("您正在申请加入 %s 达人合作计划，请使用下方验证码完成验证。验证码 10 分钟内有效。", systemName),
			security: "如非本人操作，可忽略此邮件，账户不会受到任何影响。",
			noShare:  fmt.Sprintf("请勿将验证码分享给任何人，包括 %s 工作人员。", systemName),
		}
	case "ja":
		t = copy{
			subject:  fmt.Sprintf("【%s】メール認証コード", systemName),
			badge:    "アフィリエイト申請",
			heading:  "メール認証コード",
			subtitle: fmt.Sprintf("%s のアフィリエイト申請中です。下記の認証コードをご入力ください。コードは 10 分間有効です。", systemName),
			security: "お心当たりのない場合は、このメールを無視してください。",
			noShare:  fmt.Sprintf("認証コードを %s スタッフを含む第三者に共有しないでください。", systemName),
		}
	case "fr":
		t = copy{
			subject:  fmt.Sprintf("【%s】Code de vérification", systemName),
			badge:    "Demande d'affiliation",
			heading:  "Code de vérification",
			subtitle: fmt.Sprintf("Vous postulez au programme d'affiliation %s. Utilisez le code ci-dessous pour finaliser la vérification. Valable 10 minutes.", systemName),
			security: "Si vous n'êtes pas à l'origine de cette action, ignorez cet e-mail.",
			noShare:  fmt.Sprintf("Ne partagez ce code avec personne, y compris l'équipe %s.", systemName),
		}
	case "ru":
		t = copy{
			subject:  fmt.Sprintf("【%s】Код подтверждения", systemName),
			badge:    "Заявка партнёра",
			heading:  "Код подтверждения",
			subtitle: fmt.Sprintf("Вы подаёте заявку на участие в партнёрской программе %s. Используйте код ниже. Действителен 10 минут.", systemName),
			security: "Если вы не запрашивали этот код, просто проигнорируйте письмо.",
			noShare:  fmt.Sprintf("Не передавайте код третьим лицам, включая сотрудников %s.", systemName),
		}
	case "vi":
		t = copy{
			subject:  fmt.Sprintf("【%s】Mã xác minh", systemName),
			badge:    "Đăng ký CTV",
			heading:  "Mã xác minh email",
			subtitle: fmt.Sprintf("Bạn đang đăng ký chương trình cộng tác viên %s. Dùng mã bên dưới để hoàn tất. Có hiệu lực 10 phút.", systemName),
			security: "Nếu bạn không thực hiện hành động này, hãy bỏ qua email này.",
			noShare:  fmt.Sprintf("Không chia sẻ mã này với bất kỳ ai, kể cả nhân viên %s.", systemName),
		}
	default: // en
		t = copy{
			subject:  fmt.Sprintf("【%s】Email verification code", systemName),
			badge:    "Affiliate Application",
			heading:  "Email Verification Code",
			subtitle: fmt.Sprintf("You are applying to the %s affiliate program. Use the code below to complete verification. Valid for 10 minutes.", systemName),
			security: "If you didn't request this, you can safely ignore this email.",
			noShare:  fmt.Sprintf("Never share this code with anyone, including %s staff.", systemName),
		}
	}

	subject = t.subject
	digits := buildCodeDigitCells(code)
	main := fmt.Sprintf(`
    <tr><td style="padding:40px 52px 36px;">
      <p style="margin:0 0 14px;">%s</p>
      <h1 style="margin:0 0 10px;font-size:26px;font-weight:700;color:#1D1D1F;line-height:1.25;letter-spacing:-0.6px;">%s</h1>
      <p style="margin:0 0 32px;font-size:15px;color:#6E6E73;line-height:1.7;">%s</p>
      <table width="100%%" cellpadding="0" cellspacing="0" border="0">
        <tr><td align="center">%s</td></tr>
      </table>
      <p style="margin:16px 0 0;font-size:12px;color:#AEAEB2;text-align:center;">10 min</p>
    </td></tr>
    <tr><td style="padding:0 52px 28px;">%s</td></tr>`,
		emailBadgePurple(t.badge),
		t.heading,
		t.subtitle,
		digits,
		emailInfoBox(t.security+`<br><span style="color:#AEAEB2;">`+t.noShare+`</span>`),
	)
	content = emailWrapper(main)
	return
}

// BuildAffiliateApprovalEmail returns subject and HTML body for the approval notification.
func BuildAffiliateApprovalEmail(lang, name, systemName, link string) (subject, content string) {
	lang = normalizeEmailLang(lang)

	type copy struct {
		subject      string
		badge        string
		heading      string
		subtitle     string
		btnLabel     string
		linkWarning  string
		statLabel1   string
		statVal1     string
		statLabel2   string
		security     string
		noShare      string
	}

	var t copy
	switch lang {
	case "zh":
		t = copy{
			subject:     fmt.Sprintf("【%s】您的达人合作申请已通过", systemName),
			badge:       "申请已通过",
			heading:     "恭喜，您的申请已通过！",
			subtitle:    fmt.Sprintf("Hi %s，您的达人合作申请已审核通过。点击下方链接完成账户注册，即刻开始您的合作之旅。", name),
			btnLabel:    "立即注册账户",
			linkWarning: "此链接仅可使用一次，请勿转发",
			statLabel1:  "打款周期",
			statVal1:    "每月 1 & 15 日",
			statLabel2:  "结算方式",
			security:    "如非本人操作，请忽略此邮件，账户不会受到任何影响。",
			noShare:     "请勿将任何链接或验证信息分享给他人。",
		}
	case "ja":
		t = copy{
			subject:     fmt.Sprintf("【%s】アフィリエイト申請が承認されました", systemName),
			badge:       "申請承認",
			heading:     "おめでとうございます！申請が承認されました",
			subtitle:    fmt.Sprintf("%s 様、アフィリエイト申請が承認されました。下のボタンからアカウント登録を完了してください。", name),
			btnLabel:    "アカウントを登録",
			linkWarning: "このリンクは一度のみ使用可能です。転送しないでください",
			statLabel1:  "支払いスケジュール",
			statVal1:    "毎月 1 & 15 日",
			statLabel2:  "支払い方法",
			security:    "お心当たりのない場合は、このメールを無視してください。",
			noShare:     "リンクや認証情報を第三者と共有しないでください。",
		}
	case "fr":
		t = copy{
			subject:     fmt.Sprintf("【%s】Votre demande d'affiliation a été approuvée", systemName),
			badge:       "Demande approuvée",
			heading:     "Félicitations, votre demande est approuvée !",
			subtitle:    fmt.Sprintf("Bonjour %s, votre demande d'affiliation a été approuvée. Cliquez ci-dessous pour finaliser votre inscription.", name),
			btnLabel:    "Créer mon compte",
			linkWarning: "Ce lien est à usage unique, ne le transférez pas",
			statLabel1:  "Calendrier de paiement",
			statVal1:    "1er & 15 de chaque mois",
			statLabel2:  "Mode de paiement",
			security:    "Si vous n'êtes pas à l'origine de cette action, ignorez cet e-mail.",
			noShare:     "Ne partagez aucun lien ni information de vérification avec qui que ce soit.",
		}
	case "ru":
		t = copy{
			subject:     fmt.Sprintf("【%s】Ваша партнёрская заявка одобрена", systemName),
			badge:       "Заявка одобрена",
			heading:     "Поздравляем, ваша заявка одобрена!",
			subtitle:    fmt.Sprintf("Здравствуйте, %s! Ваша партнёрская заявка одобрена. Нажмите кнопку ниже для завершения регистрации.", name),
			btnLabel:    "Зарегистрировать аккаунт",
			linkWarning: "Ссылка одноразовая, не пересылайте её",
			statLabel1:  "График выплат",
			statVal1:    "1-го и 15-го каждого месяца",
			statLabel2:  "Способ оплаты",
			security:    "Если вы не запрашивали это, просто проигнорируйте письмо.",
			noShare:     "Не передавайте ссылки или данные верификации третьим лицам.",
		}
	case "vi":
		t = copy{
			subject:     fmt.Sprintf("【%s】Đơn đăng ký cộng tác viên của bạn đã được chấp thuận", systemName),
			badge:       "Đơn đã được duyệt",
			heading:     "Chúc mừng, đơn của bạn đã được duyệt!",
			subtitle:    fmt.Sprintf("Xin chào %s, đơn đăng ký cộng tác viên của bạn đã được chấp thuận. Nhấp bên dưới để hoàn tất đăng ký.", name),
			btnLabel:    "Đăng ký tài khoản",
			linkWarning: "Liên kết chỉ dùng được một lần, không chuyển tiếp",
			statLabel1:  "Lịch thanh toán",
			statVal1:    "Ngày 1 & 15 hàng tháng",
			statLabel2:  "Phương thức thanh toán",
			security:    "Nếu bạn không thực hiện hành động này, hãy bỏ qua email này.",
			noShare:     "Không chia sẻ bất kỳ liên kết hay thông tin xác minh nào với người khác.",
		}
	default: // en
		t = copy{
			subject:     fmt.Sprintf("【%s】Your affiliate application has been approved", systemName),
			badge:       "Application Approved",
			heading:     "Congratulations, your application is approved!",
			subtitle:    fmt.Sprintf("Hi %s, your affiliate application has been approved. Click below to complete your account registration.", name),
			btnLabel:    "Register Account",
			linkWarning: "This link can only be used once, do not forward it",
			statLabel1:  "Payout Schedule",
			statVal1:    "1st & 15th of each month",
			statLabel2:  "Payment Method",
			security:    "If you didn't request this, you can safely ignore this email.",
			noShare:     "Never share any links or verification info with others.",
		}
	}

	subject = t.subject
	main := fmt.Sprintf(`
    <tr><td style="padding:40px 52px 8px;">
      <p style="margin:0 0 14px;">%s</p>
      <h1 style="margin:0 0 10px;font-size:26px;font-weight:700;color:#1D1D1F;line-height:1.25;letter-spacing:-0.6px;">%s</h1>
      <p style="margin:0 0 32px;font-size:15px;color:#6E6E73;line-height:1.7;">%s</p>
      %s
      <p style="margin:18px 0 28px;font-size:12px;color:#AEAEB2;text-align:center;">%s</p>
      <div style="height:1px;background:rgba(195,186,242,0.28);margin-bottom:24px;"></div>
      %s
    </td></tr>
    <tr><td style="padding:24px 0 0;"></td></tr>
    <tr><td style="padding:0 52px 28px;">%s</td></tr>`,
		emailBadgeGreen(t.badge),
		t.heading,
		t.subtitle,
		emailPrimaryBtn(link, t.btnLabel),
		t.linkWarning,
		emailStatCard2(t.statLabel1, t.statVal1, t.statLabel2, "PayPal"),
		emailInfoBox(t.security+`<br><span style="color:#AEAEB2;">`+t.noShare+`</span>`),
	)
	content = emailWrapper(main)
	return
}

// BuildAffiliateRejectionEmail returns subject and HTML body for the rejection notification.
func BuildAffiliateRejectionEmail(lang, name, systemName, reason string) (subject, content string) {
	lang = normalizeEmailLang(lang)

	type copy struct {
		subject      string
		badge        string
		heading      string
		subtitle     string
		reasonLabel  string
		reapplyLabel string
		reapplyHref  string
		bodyText     string
		security     string
		noShare      string
	}

	var t copy
	switch lang {
	case "zh":
		t = copy{
			subject:      fmt.Sprintf("【%s】关于您的达人合作申请", systemName),
			badge:        "申请未通过",
			heading:      "关于您的达人申请",
			subtitle:     fmt.Sprintf("Hi %s，感谢您申请 %s 达人合作计划。经过审核，我们暂时无法批准您的申请。", name, systemName),
			reasonLabel:  "未通过原因",
			reapplyLabel: "重新申请",
			bodyText:     "您可以在满足申请条件后重新提交申请，我们非常欢迎您再次尝试。",
			security:     "如非本人操作，请忽略此邮件，账户不会受到任何影响。",
			noShare:      "请勿将任何链接或验证信息分享给他人。",
		}
	case "ja":
		t = copy{
			subject:      fmt.Sprintf("【%s】アフィリエイト申請について", systemName),
			badge:        "申請未承認",
			heading:      "アフィリエイト申請について",
			subtitle:     fmt.Sprintf("%s 様、%s のアフィリエイト申請をご検討いただきありがとうございます。審査の結果、今回は承認できませんでした。", name, systemName),
			reasonLabel:  "未承認の理由",
			reapplyLabel: "再申請する",
			bodyText:     "条件を満たした後に再度ご申請いただけます。",
			security:     "お心当たりのない場合は、このメールを無視してください。",
			noShare:      "リンクや認証情報を第三者と共有しないでください。",
		}
	case "fr":
		t = copy{
			subject:      fmt.Sprintf("【%s】Concernant votre demande d'affiliation", systemName),
			badge:        "Demande non retenue",
			heading:      "Concernant votre demande",
			subtitle:     fmt.Sprintf("Bonjour %s, merci d'avoir postulé au programme d'affiliation %s. Après examen, nous ne pouvons pas approuver votre demande pour le moment.", name, systemName),
			reasonLabel:  "Motif du refus",
			reapplyLabel: "Postuler à nouveau",
			bodyText:     "Vous pouvez soumettre une nouvelle demande une fois les conditions remplies.",
			security:     "Si vous n'êtes pas à l'origine de cette action, ignorez cet e-mail.",
			noShare:      "Ne partagez aucun lien ni information de vérification avec qui que ce soit.",
		}
	case "ru":
		t = copy{
			subject:      fmt.Sprintf("【%s】О вашей партнёрской заявке", systemName),
			badge:        "Заявка не одобрена",
			heading:      "О вашей партнёрской заявке",
			subtitle:     fmt.Sprintf("Здравствуйте, %s! Благодарим за интерес к партнёрской программе %s. По результатам рассмотрения мы не можем одобрить вашу заявку на данный момент.", name, systemName),
			reasonLabel:  "Причина отказа",
			reapplyLabel: "Подать заявку повторно",
			bodyText:     "Вы можете повторно подать заявку, когда будете соответствовать требованиям.",
			security:     "Если вы не запрашивали это, просто проигнорируйте письмо.",
			noShare:      "Не передавайте ссылки или данные верификации третьим лицам.",
		}
	case "vi":
		t = copy{
			subject:      fmt.Sprintf("【%s】Về đơn đăng ký cộng tác viên của bạn", systemName),
			badge:        "Đơn chưa được duyệt",
			heading:      "Về đơn đăng ký của bạn",
			subtitle:     fmt.Sprintf("Xin chào %s, cảm ơn bạn đã đăng ký chương trình cộng tác viên %s. Sau khi xem xét, chúng tôi chưa thể chấp thuận đơn của bạn lần này.", name, systemName),
			reasonLabel:  "Lý do chưa duyệt",
			reapplyLabel: "Đăng ký lại",
			bodyText:     "Bạn có thể nộp lại đơn khi đủ điều kiện yêu cầu.",
			security:     "Nếu bạn không thực hiện hành động này, hãy bỏ qua email này.",
			noShare:      "Không chia sẻ bất kỳ liên kết hay thông tin xác minh nào với người khác.",
		}
	default: // en
		t = copy{
			subject:      fmt.Sprintf("【%s】Regarding your affiliate application", systemName),
			badge:        "Application Not Approved",
			heading:      "Regarding your application",
			subtitle:     fmt.Sprintf("Hi %s, thank you for applying to the %s affiliate program. After review, we are unable to approve your application at this time.", name, systemName),
			reasonLabel:  "Reason",
			reapplyLabel: "Apply Again",
			bodyText:     "You are welcome to reapply once you meet the requirements.",
			security:     "If you didn't request this, you can safely ignore this email.",
			noShare:      "Never share any links or verification info with others.",
		}
	}

	subject = t.subject

	warningBlock := ""
	if reason != "" {
		warningBlock = emailWarningBox(t.reasonLabel, reason)
	}

	main := fmt.Sprintf(`
    <tr><td style="padding:40px 52px 8px;">
      <p style="margin:0 0 14px;">%s</p>
      <h1 style="margin:0 0 10px;font-size:26px;font-weight:700;color:#1D1D1F;line-height:1.25;letter-spacing:-0.6px;">%s</h1>
      <p style="margin:0 0 32px;font-size:15px;color:#6E6E73;line-height:1.7;">%s</p>
      %s
      <p style="margin:0 0 28px;font-size:13px;color:#6E6E73;line-height:1.75;">%s</p>
      %s
    </td></tr>
    <tr><td style="padding:24px 0 0;"></td></tr>
    <tr><td style="padding:0 52px 28px;">%s</td></tr>`,
		emailBadgeAmber(t.badge),
		t.heading,
		t.subtitle,
		warningBlock,
		t.bodyText,
		emailSecondaryBtn("#", t.reapplyLabel),
		emailInfoBox(t.security+`<br><span style="color:#AEAEB2;">`+t.noShare+`</span>`),
	)
	content = emailWrapper(main)
	return
}

// BuildRegistrationVerificationEmail returns subject and HTML body for the registration email verification code.
func BuildRegistrationVerificationEmail(lang, systemName, code string, validMinutes int) (subject, content string) {
	lang = normalizeEmailLang(lang)

	type copy struct {
		subject  string
		badge    string
		heading  string
		subtitle string
		validity string
		security string
		noShare  string
	}

	var t copy
	switch lang {
	case "zh":
		t = copy{
			subject:  fmt.Sprintf("【%s】邮箱验证码", systemName),
			badge:    "验证码",
			heading:  "邮箱验证码",
			subtitle: fmt.Sprintf("您正在注册 %s，请使用下方验证码完成验证。", systemName),
			validity: fmt.Sprintf("%d 分钟内有效", validMinutes),
			security: "如非本人操作，可忽略此邮件，账户不会受到任何影响。",
			noShare:  fmt.Sprintf("请勿将验证码分享给任何人，包括 %s 工作人员。", systemName),
		}
	case "ja":
		t = copy{
			subject:  fmt.Sprintf("【%s】メール認証コード", systemName),
			badge:    "認証コード",
			heading:  "メール認証コード",
			subtitle: fmt.Sprintf("%s の登録中です。下記の認証コードをご入力ください。", systemName),
			validity: fmt.Sprintf("%d 分間有効", validMinutes),
			security: "お心当たりのない場合は、このメールを無視してください。",
			noShare:  fmt.Sprintf("認証コードを %s スタッフを含む第三者に共有しないでください。", systemName),
		}
	case "fr":
		t = copy{
			subject:  fmt.Sprintf("【%s】Code de vérification", systemName),
			badge:    "Code de vérification",
			heading:  "Code de vérification",
			subtitle: fmt.Sprintf("Vous créez un compte sur %s. Utilisez le code ci-dessous pour finaliser la vérification.", systemName),
			validity: fmt.Sprintf("Valable %d minutes", validMinutes),
			security: "Si vous n'êtes pas à l'origine de cette action, ignorez cet e-mail.",
			noShare:  fmt.Sprintf("Ne partagez ce code avec personne, y compris l'équipe %s.", systemName),
		}
	case "ru":
		t = copy{
			subject:  fmt.Sprintf("【%s】Код подтверждения", systemName),
			badge:    "Код подтверждения",
			heading:  "Код подтверждения",
			subtitle: fmt.Sprintf("Вы регистрируетесь на %s. Используйте код ниже для подтверждения.", systemName),
			validity: fmt.Sprintf("Действителен %d минут", validMinutes),
			security: "Если вы не запрашивали этот код, просто проигнорируйте письмо.",
			noShare:  fmt.Sprintf("Не передавайте код третьим лицам, включая сотрудников %s.", systemName),
		}
	case "vi":
		t = copy{
			subject:  fmt.Sprintf("【%s】Mã xác minh email", systemName),
			badge:    "Mã xác minh",
			heading:  "Mã xác minh email",
			subtitle: fmt.Sprintf("Bạn đang đăng ký tài khoản trên %s. Dùng mã bên dưới để hoàn tất xác minh.", systemName),
			validity: fmt.Sprintf("Có hiệu lực trong %d phút", validMinutes),
			security: "Nếu bạn không thực hiện hành động này, hãy bỏ qua email này.",
			noShare:  fmt.Sprintf("Không chia sẻ mã này với bất kỳ ai, kể cả nhân viên %s.", systemName),
		}
	default: // en
		t = copy{
			subject:  fmt.Sprintf("【%s】Email verification code", systemName),
			badge:    "Verification Code",
			heading:  "Email Verification Code",
			subtitle: fmt.Sprintf("You are registering on %s. Use the code below to complete verification.", systemName),
			validity: fmt.Sprintf("Valid for %d minutes", validMinutes),
			security: "If you didn't request this, you can safely ignore this email.",
			noShare:  fmt.Sprintf("Never share this code with anyone, including %s staff.", systemName),
		}
	}

	subject = t.subject
	digits := buildCodeDigitCells(code)
	main := fmt.Sprintf(`
    <tr><td style="padding:40px 52px 36px;">
      <p style="margin:0 0 14px;">%s</p>
      <h1 style="margin:0 0 10px;font-size:26px;font-weight:700;color:#1D1D1F;line-height:1.25;letter-spacing:-0.6px;">%s</h1>
      <p style="margin:0 0 32px;font-size:15px;color:#6E6E73;line-height:1.7;">%s</p>
      <table width="100%%" cellpadding="0" cellspacing="0" border="0">
        <tr><td align="center">%s</td></tr>
      </table>
      <p style="margin:16px 0 0;font-size:12px;color:#AEAEB2;text-align:center;">%s</p>
    </td></tr>
    <tr><td style="padding:0 52px 28px;">%s</td></tr>`,
		emailBadgePurple(t.badge),
		t.heading,
		t.subtitle,
		digits,
		t.validity,
		emailInfoBox(t.security+`<br><span style="color:#AEAEB2;">`+t.noShare+`</span>`),
	)
	content = emailWrapper(main)
	return
}

// BuildPasswordResetEmail returns subject and HTML body for the password reset email.
func BuildPasswordResetEmail(lang, systemName, link string, validMinutes int) (subject, content string) {
	lang = normalizeEmailLang(lang)

	type copy struct {
		subject   string
		badge     string
		heading   string
		subtitle  string
		btnLabel  string
		validity  string
		fallback  string
		security  string
		noShare   string
	}

	var t copy
	switch lang {
	case "zh":
		t = copy{
			subject:  fmt.Sprintf("【%s】密码重置", systemName),
			badge:    "密码重置",
			heading:  "重置您的密码",
			subtitle: "我们收到了重置密码的请求，请点击下方按钮设置新密码。",
			btnLabel: "重置密码",
			validity: fmt.Sprintf("链接 %d 分钟内有效 · 仅可使用一次", validMinutes),
			fallback: "如按钮无法点击，请复制下方链接到浏览器：",
			security: "如非本人操作，请忽略此邮件，账户不会受到任何影响。",
			noShare:  "请勿将任何链接或验证信息分享给他人。",
		}
	case "ja":
		t = copy{
			subject:  fmt.Sprintf("【%s】パスワードリセット", systemName),
			badge:    "パスワードリセット",
			heading:  "パスワードをリセット",
			subtitle: "パスワードリセットのリクエストを受け付けました。下のボタンをクリックして新しいパスワードを設定してください。",
			btnLabel: "パスワードをリセット",
			validity: fmt.Sprintf("リンクは %d 分間有効 · 一度のみ使用可能", validMinutes),
			fallback: "ボタンが開かない場合は、以下のURLをブラウザにコピーしてください：",
			security: "お心当たりのない場合は、このメールを無視してください。",
			noShare:  "リンクや認証情報を第三者と共有しないでください。",
		}
	case "fr":
		t = copy{
			subject:  fmt.Sprintf("【%s】Réinitialisation du mot de passe", systemName),
			badge:    "Réinitialisation",
			heading:  "Réinitialisez votre mot de passe",
			subtitle: "Nous avons reçu une demande de réinitialisation de mot de passe. Cliquez sur le bouton ci-dessous pour définir un nouveau mot de passe.",
			btnLabel: "Réinitialiser le mot de passe",
			validity: fmt.Sprintf("Lien valable %d minutes · usage unique", validMinutes),
			fallback: "Si le bouton ne s'ouvre pas, copiez ce lien dans votre navigateur :",
			security: "Si vous n'êtes pas à l'origine de cette demande, ignorez cet e-mail.",
			noShare:  "Ne partagez aucun lien ni information de vérification avec qui que ce soit.",
		}
	case "ru":
		t = copy{
			subject:  fmt.Sprintf("【%s】Сброс пароля", systemName),
			badge:    "Сброс пароля",
			heading:  "Сбросьте ваш пароль",
			subtitle: "Мы получили запрос на сброс пароля. Нажмите кнопку ниже, чтобы задать новый пароль.",
			btnLabel: "Сбросить пароль",
			validity: fmt.Sprintf("Ссылка действительна %d минут · одноразовая", validMinutes),
			fallback: "Если кнопка не работает, скопируйте ссылку в браузер:",
			security: "Если вы не запрашивали сброс пароля, проигнорируйте это письмо.",
			noShare:  "Не передавайте ссылки или данные верификации третьим лицам.",
		}
	case "vi":
		t = copy{
			subject:  fmt.Sprintf("【%s】Đặt lại mật khẩu", systemName),
			badge:    "Đặt lại mật khẩu",
			heading:  "Đặt lại mật khẩu của bạn",
			subtitle: "Chúng tôi nhận được yêu cầu đặt lại mật khẩu. Nhấp vào nút bên dưới để thiết lập mật khẩu mới.",
			btnLabel: "Đặt lại mật khẩu",
			validity: fmt.Sprintf("Liên kết có hiệu lực %d phút · chỉ dùng một lần", validMinutes),
			fallback: "Nếu nút không hoạt động, hãy sao chép liên kết sau vào trình duyệt:",
			security: "Nếu bạn không thực hiện yêu cầu này, hãy bỏ qua email này.",
			noShare:  "Không chia sẻ bất kỳ liên kết hay thông tin xác minh nào với người khác.",
		}
	default: // en
		t = copy{
			subject:  fmt.Sprintf("【%s】Password reset", systemName),
			badge:    "Password Reset",
			heading:  "Reset your password",
			subtitle: "We received a request to reset your password. Click the button below to set a new password.",
			btnLabel: "Reset Password",
			validity: fmt.Sprintf("Link valid for %d minutes · one-time use", validMinutes),
			fallback: "If the button doesn't work, copy the link below into your browser:",
			security: "If you didn't request this, you can safely ignore this email.",
			noShare:  "Never share any links or verification info with others.",
		}
	}

	subject = t.subject
	main := fmt.Sprintf(`
    <tr><td style="padding:40px 52px 8px;">
      <p style="margin:0 0 14px;">%s</p>
      <h1 style="margin:0 0 10px;font-size:26px;font-weight:700;color:#1D1D1F;line-height:1.25;letter-spacing:-0.6px;">%s</h1>
      <p style="margin:0 0 32px;font-size:15px;color:#6E6E73;line-height:1.7;">%s</p>
      %s
      <p style="margin:20px 0 28px;font-size:12px;color:#AEAEB2;text-align:center;">%s</p>
      <div style="height:1px;background:rgba(195,186,242,0.28);margin-bottom:24px;"></div>
      <p style="margin:0 0 4px;font-size:12px;color:#AEAEB2;line-height:1.75;text-align:center;">%s</p>
      <p style="margin:0;font-size:12px;color:#6E6E73;word-break:break-all;text-align:center;font-family:ui-monospace,SFMono-Regular,monospace;line-height:1.6;">%s</p>
    </td></tr>
    <tr><td style="padding:28px 52px 0;"></td></tr>
    <tr><td style="padding:0 52px 28px;">%s</td></tr>`,
		emailBadgePurple(t.badge),
		t.heading,
		t.subtitle,
		emailPrimaryBtn(link, t.btnLabel),
		t.validity,
		t.fallback,
		link,
		emailInfoBox(t.security+`<br><span style="color:#AEAEB2;">`+t.noShare+`</span>`),
	)
	content = emailWrapper(main)
	return
}

// BuildKolWelcomeEmail returns subject and HTML body sent to a new KOL after they register via invite link.
// dashboardLink is the full URL to /console/kol; minAmount is the minimum withdrawal threshold in USD;
// commissionRate is the KOL commission rate (e.g. 0.20 for 20%).
func BuildKolWelcomeEmail(lang, name, systemName, dashboardLink string, minAmount, commissionRate float64) (subject, content string) {
	lang = normalizeEmailLang(lang)

	type copy struct {
		subject    string
		badge      string
		heading    string
		subtitle   string
		imgLabel   string
		stat1Label string
		stat2Label string
		stat2Val   string
		stat3Label string
		stepTitle  string
		step1h     string
		step1b     string
		step2h     string
		step2b     string
		step3h     string
		step3b     string
		btnLabel   string
		footnote   string
		imgURL     string
		imgAlt     string
	}

	commPct := fmt.Sprintf("%.0f%%", commissionRate*100)
	minAmt := fmt.Sprintf("$%.2f", minAmount)

	var t copy
	switch lang {
	case "zh":
		t = copy{
			subject:    fmt.Sprintf("【%s】欢迎加入达人计划", systemName),
			badge:      "Welcome to KOL Program",
			heading:    fmt.Sprintf("欢迎加入，%s！", name),
			subtitle:   fmt.Sprintf("您已正式成为 %s 达人合作伙伴。每成功邀请一位用户充值，您将获得对应佣金，资金将通过 PayPal 定期打款。", systemName),
			imgLabel:   "控制台预览",
			stat1Label: "充值佣金比例", stat2Label: "佣金冻结期", stat2Val: "10天", stat3Label: "最低提现金额",
			stepTitle: "三步开始赚取佣金",
			step1h:    "登录达人控制台",
			step1b:    "登录后可查看邀请码、已邀请人数、佣金明细及累计收益，并复制您的专属推广链接。",
			step2h:    "分享推广链接",
			step2b:    "通过您的推广链接注册的用户，<strong style=\"color:#475569;\">前 3 次充值</strong>均可产生佣金，自动计入您的账户。",
			step3h:    "申请提现",
			step3b:    "佣金冻结期（10天）满后即可申请提现，资金打款至您的 PayPal 账户。",
			btnLabel:  "前往达人控制台",
			footnote:  fmt.Sprintf("打款日期为每月 1 日和 15 日 · 每次提现最低 %s", minAmt),
			imgURL:    "https://raw.githubusercontent.com/hahahahali/kol-assets/main/zh.jpg",
			imgAlt:    "达人中心指引",
		}
	case "ja":
		t = copy{
			subject:    fmt.Sprintf("【%s】アフィリエイトプログラムへようこそ", systemName),
			badge:      "Welcome to KOL Program",
			heading:    fmt.Sprintf("ようこそ、%s！", name),
			subtitle:   fmt.Sprintf("%s のアフィリエイトパートナーになりました。招待したユーザーがチャージするたびにコミッションが発生し、PayPal で定期的にお支払いします。", systemName),
			stat1Label: "コミッション率", stat2Label: "保留期間", stat2Val: "10日", stat3Label: "最低出金額",
			imgLabel:   "ダッシュボードプレビュー",
			stepTitle: "3ステップで稼ぎ始める",
			step1h:    "ダッシュボードにログイン",
			step1b:    "招待コード、招待人数、コミッション明細、累計収益を確認し、専属リンクをコピーできます。",
			step2h:    "リンクをシェア",
			step2b:    "あなたのリンクから登録したユーザーの<strong style=\"color:#475569;\">最初の3回のチャージ</strong>がコミッション対象となり、自動的に反映されます。",
			step3h:    "出金を申請",
			step3b:    "保留期間（10日）終了後に出金申請が可能です。資金は PayPal アカウントに送金されます。",
			btnLabel:  "ダッシュボードへ",
			footnote:  fmt.Sprintf("毎月 1 日・15 日に支払い · 最低出金額 %s", minAmt),
			imgURL:    "https://raw.githubusercontent.com/hahahahali/kol-assets/main/ja.jpg",
			imgAlt:    "アフィリエイトガイド",
		}
	case "fr":
		t = copy{
			subject:    fmt.Sprintf("【%s】Bienvenue dans le programme d'affiliation", systemName),
			badge:      "Welcome to KOL Program",
			heading:    fmt.Sprintf("Bienvenue, %s !", name),
			subtitle:   fmt.Sprintf("Vous êtes officiellement partenaire affilié de %s. Chaque utilisateur invité qui effectue un rechargement vous rapporte une commission, versée régulièrement via PayPal.", systemName),
			stat1Label: "Taux de commission", stat2Label: "Période de gel", stat2Val: "10j", stat3Label: "Retrait minimum",
			imgLabel:   "Aperçu du tableau de bord",
			stepTitle: "3 étapes pour commencer à gagner",
			step1h:    "Connectez-vous au tableau de bord",
			step1b:    "Consultez votre code d'invitation, le nombre d'invités, les détails des commissions et copiez votre lien de parrainage.",
			step2h:    "Partagez votre lien",
			step2b:    "Les <strong style=\"color:#475569;\">3 premiers rechargements</strong> de chaque utilisateur inscrit via votre lien génèrent une commission, créditée automatiquement.",
			step3h:    "Demandez un retrait",
			step3b:    "Après la période de gel (10 jours), vous pouvez demander un retrait vers votre compte PayPal.",
			btnLabel:  "Accéder au tableau de bord",
			footnote:  fmt.Sprintf("Paiements le 1er et le 15 de chaque mois · Retrait minimum %s", minAmt),
			imgURL:    "https://raw.githubusercontent.com/hahahahali/kol-assets/main/fr.jpg",
			imgAlt:    "Guide affilié",
		}
	case "ru":
		t = copy{
			subject:    fmt.Sprintf("【%s】Добро пожаловать в партнёрскую программу", systemName),
			badge:      "Welcome to KOL Program",
			heading:    fmt.Sprintf("Добро пожаловать, %s!", name),
			subtitle:   fmt.Sprintf("Вы официально стали партнёром %s. За каждое пополнение приглашённого пользователя вы получаете комиссию, которая выплачивается через PayPal.", systemName),
			stat1Label: "Ставка комиссии", stat2Label: "Период заморозки", stat2Val: "10д", stat3Label: "Мин. вывод",
			imgLabel:   "Предпросмотр кабинета",
			stepTitle: "3 шага для начала заработка",
			step1h:    "Войдите в кабинет",
			step1b:    "Просматривайте реферальный код, количество приглашённых, детали комиссий и копируйте вашу ссылку.",
			step2h:    "Поделитесь ссылкой",
			step2b:    "<strong style=\"color:#475569;\">Первые 3 пополнения</strong> каждого приглашённого засчитываются как комиссия и автоматически зачисляются на ваш счёт.",
			step3h:    "Запросите вывод",
			step3b:    "После периода заморозки (10 дней) подайте заявку на вывод средств на ваш счёт PayPal.",
			btnLabel:  "Перейти в кабинет",
			footnote:  fmt.Sprintf("Выплаты 1-го и 15-го каждого месяца · Мин. вывод %s", minAmt),
			imgURL:    "https://raw.githubusercontent.com/hahahahali/kol-assets/main/ru.jpg",
			imgAlt:    "Руководство партнёра",
		}
	case "vi":
		t = copy{
			subject:    fmt.Sprintf("【%s】Chào mừng bạn đến với chương trình cộng tác viên", systemName),
			badge:      "Welcome to KOL Program",
			heading:    fmt.Sprintf("Chào mừng, %s!", name),
			subtitle:   fmt.Sprintf("Bạn đã chính thức trở thành đối tác cộng tác viên của %s. Mỗi lần người được mời nạp tiền, bạn sẽ nhận hoa hồng và được thanh toán định kỳ qua PayPal.", systemName),
			stat1Label: "Tỷ lệ hoa hồng", stat2Label: "Thời gian đóng băng", stat2Val: "10ng", stat3Label: "Rút tối thiểu",
			imgLabel:   "Xem trước bảng điều khiển",
			stepTitle: "3 bước để bắt đầu kiếm tiền",
			step1h:    "Đăng nhập vào bảng điều khiển",
			step1b:    "Xem mã mời, số người được mời, chi tiết hoa hồng và sao chép liên kết giới thiệu của bạn.",
			step2h:    "Chia sẻ liên kết",
			step2b:    "<strong style=\"color:#475569;\">3 lần nạp tiền đầu tiên</strong> của mỗi người đăng ký qua liên kết của bạn đều tạo ra hoa hồng, tự động ghi vào tài khoản.",
			step3h:    "Yêu cầu rút tiền",
			step3b:    "Sau thời gian đóng băng (10 ngày), bạn có thể yêu cầu rút tiền về tài khoản PayPal.",
			btnLabel:  "Vào bảng điều khiển",
			footnote:  fmt.Sprintf("Thanh toán ngày 1 và 15 hàng tháng · Rút tối thiểu %s", minAmt),
			imgURL:    "https://raw.githubusercontent.com/hahahahali/kol-assets/main/vi.jpg",
			imgAlt:    "Hướng dẫn cộng tác viên",
		}
	default: // en
		t = copy{
			subject:    fmt.Sprintf("【%s】Welcome to the Affiliate Program", systemName),
			badge:      "Welcome to KOL Program",
			heading:    fmt.Sprintf("Welcome, %s!", name),
			subtitle:   fmt.Sprintf("You are now an official affiliate partner of %s. Earn a commission every time an invited user makes a deposit, paid out regularly via PayPal.", systemName),
			stat1Label: "Commission Rate", stat2Label: "Freeze Period", stat2Val: "10d", stat3Label: "Min. Withdrawal",
			imgLabel:   "Dashboard Preview",
			stepTitle: "3 steps to start earning",
			step1h:    "Log in to the dashboard",
			step1b:    "View your invite code, referral count, commission details, and copy your unique referral link.",
			step2h:    "Share your link",
			step2b:    "The <strong style=\"color:#475569;\">first 3 deposits</strong> of each user who signs up via your link earn commission, credited automatically.",
			step3h:    "Request a withdrawal",
			step3b:    "After the freeze period (10 days), submit a withdrawal request to your PayPal account.",
			btnLabel:  "Go to Dashboard",
			footnote:  fmt.Sprintf("Payments on the 1st and 15th of each month · Min. withdrawal %s", minAmt),
			imgURL:    "https://raw.githubusercontent.com/hahahahali/kol-assets/main/en.jpg",
			imgAlt:    "Affiliate guide",
		}
	}

	subject = t.subject

	stepRow := func(n, h, b string) string {
		return fmt.Sprintf(`<table width="100%%" cellpadding="0" cellspacing="0" border="0" style="margin-bottom:16px;"><tr><td width="28" style="vertical-align:top;padding-top:1px;"><div style="width:24px;height:24px;background:rgba(79,70,229,0.09);border:1px solid rgba(79,70,229,0.18);border-radius:50%%;text-align:center;line-height:22px;font-size:10px;font-weight:700;color:#4F46E5;">%s</div></td><td style="padding-left:13px;vertical-align:top;"><p style="margin:0 0 2px;font-size:13px;font-weight:600;color:#1D1D1F;line-height:1.5;hyphens:none;-webkit-hyphens:none;word-break:normal;">%s</p><p style="margin:0;font-size:13px;color:#6E6E73;line-height:1.6;">%s</p></td></tr></table>`, n, h, b)
	}

	main := fmt.Sprintf(`
    <tr><td style="padding:40px 52px 8px;">
      <p style="margin:0 0 14px;">%s</p>
      <h1 style="margin:0 0 10px;font-size:26px;font-weight:700;color:#1D1D1F;line-height:1.25;letter-spacing:-0.6px;">%s</h1>
      <p style="margin:0 0 32px;font-size:15px;color:#6E6E73;line-height:1.7;">%s</p>
      %s
      <div style="height:1px;background:rgba(195,186,242,0.28);margin-bottom:28px;"></div>
      <p style="margin:0 0 18px;font-size:13px;font-weight:700;color:#1D1D1F;letter-spacing:-0.1px;">%s</p>
      %s%s%s
      <p style="margin:0 0 10px;font-size:11px;font-weight:600;color:#AEAEB2;letter-spacing:0.06em;text-transform:uppercase;">%s</p>
      <img src="%s" style="max-width:100%%;border-radius:12px;margin:0 0 24px;display:block;" alt="%s">
      %s
    </td></tr>
    <tr><td style="padding:28px 52px 0;">
      <p style="margin:0;font-size:12px;color:#AEAEB2;text-align:center;line-height:1.7;">%s</p>
    </td></tr>
    <tr><td style="padding:16px 0 0;"></td></tr>`,
		emailBadgePurple(t.badge),
		t.heading,
		t.subtitle,
		emailStatCard3(t.stat1Label, commPct, t.stat2Label, t.stat2Val, t.stat3Label, minAmt),
		t.stepTitle,
		stepRow("1", t.step1h, t.step1b),
		stepRow("2", t.step2h, t.step2b),
		stepRow("3", t.step3h, t.step3b),
		t.imgLabel, t.imgURL, t.imgAlt,
		emailPrimaryBtn(appendLangQuery(dashboardLink, "en"), t.btnLabel),
		t.footnote,
	)
	content = emailWrapper(main)
	return
}


// BuildWithdrawalPaidEmail returns subject and HTML body for the withdrawal payment notification.
func BuildWithdrawalPaidEmail(lang, username, systemName string, amount float64, txId string) (subject, content string) {
	lang = normalizeEmailLang(lang)

	type copy struct {
		subject    string
		badge      string
		heading    string
		subtitle   string
		amtLabel   string
		txLabel    string
		bodyText   string
		security   string
		noShare    string
	}

	var t copy
	switch lang {
	case "zh":
		t = copy{
			subject:  fmt.Sprintf("【%s】您的提现申请已打款", systemName),
			badge:    "打款确认",
			heading:  "您的提现已打款",
			subtitle: fmt.Sprintf("Hi %s，您申请的提现已通过 PayPal 完成转账，请登录 PayPal 确认到账。", username),
			amtLabel: "打款金额",
			txLabel:  "PayPal 交易单号",
			bodyText: "如资金未在 1–3 个工作日内到账，请凭交易单号联系 PayPal 客服或通过帮助中心与我们联系。",
			security: "如非本人操作，请忽略此邮件，账户不会受到任何影响。",
			noShare:  "请勿将任何链接或验证信息分享给他人。",
		}
	case "ja":
		t = copy{
			subject:  fmt.Sprintf("【%s】出金申請の支払いが完了しました", systemName),
			badge:    "支払い確認",
			heading:  "出金が完了しました",
			subtitle: fmt.Sprintf("%s 様、申請された出金が PayPal 経由で送金されました。PayPal アカウントでご確認ください。", username),
			amtLabel: "支払い金額",
			txLabel:  "PayPal トランザクション ID",
			bodyText: "1〜3 営業日以内に着金しない場合は、取引 ID を記載の上 PayPal サポートまたはヘルプセンターよりお問い合わせください。",
			security: "お心当たりのない場合は、このメールを無視してください。",
			noShare:  "リンクや認証情報を第三者と共有しないでください。",
		}
	case "fr":
		t = copy{
			subject:  fmt.Sprintf("【%s】Votre retrait a été effectué", systemName),
			badge:    "Paiement confirmé",
			heading:  "Votre retrait a été effectué",
			subtitle: fmt.Sprintf("Bonjour %s, votre retrait a été envoyé via PayPal. Connectez-vous à PayPal pour confirmer la réception.", username),
			amtLabel: "Montant versé",
			txLabel:  "ID de transaction PayPal",
			bodyText: "Si les fonds n'arrivent pas sous 1 à 3 jours ouvrés, contactez le support PayPal avec l'ID de transaction ou contactez-nous via le centre d'aide.",
			security: "Si vous n'êtes pas à l'origine de cette action, ignorez cet e-mail.",
			noShare:  "Ne partagez aucun lien ni information de vérification avec qui que ce soit.",
		}
	case "ru":
		t = copy{
			subject:  fmt.Sprintf("【%s】Ваш вывод средств обработан", systemName),
			badge:    "Выплата подтверждена",
			heading:  "Ваш вывод выполнен",
			subtitle: fmt.Sprintf("Здравствуйте, %s! Ваш вывод средств был отправлен через PayPal. Войдите в PayPal для подтверждения получения.", username),
			amtLabel: "Сумма выплаты",
			txLabel:  "ID транзакции PayPal",
			bodyText: "Если средства не поступят в течение 1–3 рабочих дней, обратитесь в поддержку PayPal с указанием ID транзакции или свяжитесь с нами через центр помощи.",
			security: "Если вы не запрашивали это, просто проигнорируйте письмо.",
			noShare:  "Не передавайте ссылки или данные верификации третьим лицам.",
		}
	case "vi":
		t = copy{
			subject:  fmt.Sprintf("【%s】Yêu cầu rút tiền của bạn đã được xử lý", systemName),
			badge:    "Thanh toán xác nhận",
			heading:  "Tiền rút của bạn đã được chuyển",
			subtitle: fmt.Sprintf("Xin chào %s, khoản rút tiền của bạn đã được chuyển qua PayPal. Vui lòng đăng nhập PayPal để xác nhận nhận tiền.", username),
			amtLabel: "Số tiền chuyển",
			txLabel:  "Mã giao dịch PayPal",
			bodyText: "Nếu tiền chưa về trong 1–3 ngày làm việc, hãy liên hệ hỗ trợ PayPal kèm mã giao dịch hoặc liên hệ chúng tôi qua trung tâm hỗ trợ.",
			security: "Nếu bạn không thực hiện hành động này, hãy bỏ qua email này.",
			noShare:  "Không chia sẻ bất kỳ liên kết hay thông tin xác minh nào với người khác.",
		}
	default: // en
		t = copy{
			subject:  fmt.Sprintf("【%s】Your withdrawal has been processed", systemName),
			badge:    "Payment Confirmed",
			heading:  "Your withdrawal has been sent",
			subtitle: fmt.Sprintf("Hi %s, your withdrawal has been sent via PayPal. Please log in to PayPal to confirm receipt.", username),
			amtLabel: "Amount Sent",
			txLabel:  "PayPal Transaction ID",
			bodyText: "If funds don't arrive within 1–3 business days, contact PayPal support with the transaction ID or reach us via the help center.",
			security: "If you didn't request this, you can safely ignore this email.",
			noShare:  "Never share any links or verification info with others.",
		}
	}

	subject = t.subject
	main := fmt.Sprintf(`
    <tr><td style="padding:40px 52px 8px;">
      <p style="margin:0 0 14px;">%s</p>
      <h1 style="margin:0 0 10px;font-size:26px;font-weight:700;color:#1D1D1F;line-height:1.25;letter-spacing:-0.6px;">%s</h1>
      <p style="margin:0 0 32px;font-size:15px;color:#6E6E73;line-height:1.7;">%s</p>
      <div style="background:linear-gradient(160deg,rgba(236,253,245,0.9) 0%%,rgba(220,252,231,0.7) 100%%);border:1px solid rgba(52,211,153,0.28);border-radius:16px;padding:24px;box-shadow:inset 0 1px 0 rgba(255,255,255,0.95),0 2px 8px rgba(16,185,129,0.07);margin-bottom:22px;">
        <table width="100%%" cellpadding="0" cellspacing="0" border="0">
          <tr><td style="padding-bottom:16px;border-bottom:1px solid rgba(52,211,153,0.2);">
            <p style="margin:0;font-size:11px;font-weight:600;color:#6EE7B7;letter-spacing:0.08em;text-transform:uppercase;">%s</p>
            <p style="margin:5px 0 0;font-size:32px;font-weight:800;color:#1D1D1F;letter-spacing:-0.6px;">$%.2f</p>
          </td></tr>
          <tr><td style="padding-top:16px;">
            <p style="margin:0;font-size:11px;font-weight:600;color:#6EE7B7;letter-spacing:0.08em;text-transform:uppercase;">%s</p>
            <p style="margin:5px 0 0;font-size:14px;font-weight:600;color:#1D1D1F;letter-spacing:0.04em;font-family:ui-monospace,SFMono-Regular,monospace;">%s</p>
          </td></tr>
        </table>
      </div>
      <p style="margin:0;font-size:13px;color:#6E6E73;line-height:1.75;">%s</p>
    </td></tr>
    <tr><td style="padding:24px 0 0;"></td></tr>
    <tr><td style="padding:0 52px 28px;">%s</td></tr>`,
		emailBadgeGreen(t.badge),
		t.heading,
		t.subtitle,
		t.amtLabel,
		amount,
		t.txLabel,
		txId,
		t.bodyText,
		emailInfoBox(t.security+`<br><span style="color:#AEAEB2;">`+t.noShare+`</span>`),
	)
	content = emailWrapper(main)
	return
}
