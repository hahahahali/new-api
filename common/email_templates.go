package common

import "fmt"

// normalizeEmailLang returns a supported email language code, falling back to "en".
func normalizeEmailLang(lang string) string {
	switch lang {
	case "zh", "en", "ja", "fr", "es":
		return lang
	default:
		return "en"
	}
}

// BuildAffiliateVerificationEmail returns subject and HTML body for the email verification code.
func BuildAffiliateVerificationEmail(lang, systemName, code string) (subject, content string) {
	lang = normalizeEmailLang(lang)
	switch lang {
	case "zh":
		subject = fmt.Sprintf("【%s】邮箱验证码", systemName)
		content = fmt.Sprintf(`<p>您好，</p>
<p>您正在申请加入 <strong>%s</strong> 达人合作计划，邮箱验证码为：</p>
<p style="font-size:32px;font-weight:bold;letter-spacing:10px;color:#d97706;text-align:center;padding:20px 0;">%s</p>
<p>验证码有效期为 <strong>10 分钟</strong>，请尽快完成提交。</p>
<p>如果您并未发起此申请，请忽略此邮件。</p>`,
			systemName, code)
	case "ja":
		subject = fmt.Sprintf("【%s】メール認証コード", systemName)
		content = fmt.Sprintf(`<p>こんにちは、</p>
<p><strong>%s</strong> のアフィリエイトプログラムにご申請いただきありがとうございます。認証コードは以下の通りです：</p>
<p style="font-size:32px;font-weight:bold;letter-spacing:10px;color:#d97706;text-align:center;padding:20px 0;">%s</p>
<p>このコードの有効期限は <strong>10分間</strong> です。お早めにご入力ください。</p>
<p>お心当たりのない場合は、このメールを無視してください。</p>`,
			systemName, code)
	case "fr":
		subject = fmt.Sprintf("【%s】Code de vérification", systemName)
		content = fmt.Sprintf(`<p>Bonjour,</p>
<p>Vous postulez au programme d'affiliation <strong>%s</strong>. Votre code de vérification est :</p>
<p style="font-size:32px;font-weight:bold;letter-spacing:10px;color:#d97706;text-align:center;padding:20px 0;">%s</p>
<p>Ce code est valable <strong>10 minutes</strong>. Veuillez le saisir rapidement.</p>
<p>Si vous n'êtes pas à l'origine de cette demande, ignorez cet e-mail.</p>`,
			systemName, code)
	case "es":
		subject = fmt.Sprintf("【%s】Código de verificación", systemName)
		content = fmt.Sprintf(`<p>Hola,</p>
<p>Está solicitando unirse al programa de afiliados de <strong>%s</strong>. Su código de verificación es:</p>
<p style="font-size:32px;font-weight:bold;letter-spacing:10px;color:#d97706;text-align:center;padding:20px 0;">%s</p>
<p>Este código es válido durante <strong>10 minutos</strong>. Por favor, ingréselo a la brevedad.</p>
<p>Si no realizó esta solicitud, ignore este correo.</p>`,
			systemName, code)
	default: // en
		subject = fmt.Sprintf("【%s】Email verification code", systemName)
		content = fmt.Sprintf(`<p>Hello,</p>
<p>You are applying to join the <strong>%s</strong> affiliate program. Your verification code is:</p>
<p style="font-size:32px;font-weight:bold;letter-spacing:10px;color:#d97706;text-align:center;padding:20px 0;">%s</p>
<p>This code is valid for <strong>10 minutes</strong>. Please submit your application promptly.</p>
<p>If you did not initiate this request, please ignore this email.</p>`,
			systemName, code)
	}
	return
}

// BuildAffiliateApprovalEmail returns subject and HTML body for the approval notification.
func BuildAffiliateApprovalEmail(lang, name, systemName, link string) (subject, content string) {
	lang = normalizeEmailLang(lang)
	switch lang {
	case "zh":
		subject = fmt.Sprintf("【%s】您的达人合作申请已通过", systemName)
		content = fmt.Sprintf(`<p>您好 %s，</p>
<p>恭喜！您申请加入 <strong>%s</strong> 达人计划已通过审核。</p>
<p>请点击以下链接完成注册，注册后您将自动加入达人分组并获得专属佣金功能：</p>
<p><a href="%s">%s</a></p>
<p>如果链接无法点击，请将以下地址复制到浏览器打开：<br>%s</p>
<p><strong>注意：该链接为一次性链接，仅可使用一次。</strong></p>
<p>📅 <strong>打款周期说明：</strong>每位被邀请用户的<strong>前 3 次充值订单</strong>可参与返佣（佣金比例 20%%），冻结期 10 天后自动到账。每月 <strong>1 日</strong> 和 <strong>15 日</strong> 通过 PayPal 打款，达到最低提现金额后可在达人中心申请提现。</p>
<p>如有疑问请联系我们。</p>`,
			name, systemName, link, link, link)
	case "ja":
		subject = fmt.Sprintf("【%s】アフィリエイト申請が承認されました", systemName)
		content = fmt.Sprintf(`<p>%s 様、</p>
<p>おめでとうございます！<strong>%s</strong> のアフィリエイトプログラムへのご申請が承認されました。</p>
<p>以下のリンクから登録を完了してください。登録後、自動的にアフィリエイトグループに追加され、専用のコミッション機能がご利用いただけます：</p>
<p><a href="%s">%s</a></p>
<p>リンクが開けない場合は、以下のURLをブラウザにコピーしてください：<br>%s</p>
<p><strong>注意：このリンクは一度のみ使用可能です。</strong></p>
<p>📅 <strong>支払いスケジュール：</strong>招待した各ユーザーの<strong>最初の3回の購入</strong>がコミッション対象（コミッション率20%%）となり、10日間の保留期間後に自動的に反映されます。毎月 <strong>1日</strong> と <strong>15日</strong> にPayPalで支払いが行われます。</p>
<p>ご不明な点がございましたら、お気軽にお問い合わせください。</p>`,
			name, systemName, link, link, link)
	case "fr":
		subject = fmt.Sprintf("【%s】Votre demande d'affiliation a été approuvée", systemName)
		content = fmt.Sprintf(`<p>Bonjour %s,</p>
<p>Félicitations ! Votre demande d'adhésion au programme d'affiliation <strong>%s</strong> a été approuvée.</p>
<p>Cliquez sur le lien ci-dessous pour finaliser votre inscription. Une fois inscrit(e), vous serez automatiquement ajouté(e) au groupe affilié avec accès aux fonctions de commission :</p>
<p><a href="%s">%s</a></p>
<p>Si le lien ne s'ouvre pas, copiez l'adresse suivante dans votre navigateur :<br>%s</p>
<p><strong>Attention : ce lien est à usage unique.</strong></p>
<p>📅 <strong>Calendrier de paiement :</strong> les <strong>3 premiers achats</strong> de chaque utilisateur invité sont éligibles à la commission (taux 20%%), versée après une période de gel de 10 jours. Les paiements sont effectués via PayPal le <strong>1er</strong> et le <strong>15</strong> de chaque mois.</p>
<p>N'hésitez pas à nous contacter pour toute question.</p>`,
			name, systemName, link, link, link)
	case "es":
		subject = fmt.Sprintf("【%s】Su solicitud de afiliado ha sido aprobada", systemName)
		content = fmt.Sprintf(`<p>Hola %s,</p>
<p>¡Felicitaciones! Su solicitud para unirse al programa de afiliados de <strong>%s</strong> ha sido aprobada.</p>
<p>Haga clic en el siguiente enlace para completar su registro. Después de registrarse, será agregado automáticamente al grupo de afiliados con acceso a las funciones de comisión:</p>
<p><a href="%s">%s</a></p>
<p>Si el enlace no funciona, copie la siguiente dirección en su navegador:<br>%s</p>
<p><strong>Nota: este enlace es de un solo uso.</strong></p>
<p>📅 <strong>Calendario de pagos:</strong> las <strong>primeras 3 compras</strong> de cada usuario invitado son elegibles para comisión (tasa 20%%), liberada después de un período de retención de 10 días. Los pagos se realizan vía PayPal el <strong>1</strong> y el <strong>15</strong> de cada mes.</p>
<p>Contáctenos si tiene alguna pregunta.</p>`,
			name, systemName, link, link, link)
	default: // en
		subject = fmt.Sprintf("【%s】Your affiliate application has been approved", systemName)
		content = fmt.Sprintf(`<p>Hello %s,</p>
<p>Congratulations! Your application to join the <strong>%s</strong> affiliate program has been approved.</p>
<p>Please click the link below to complete your registration. Once registered, you will be automatically added to the affiliate group with access to commission features:</p>
<p><a href="%s">%s</a></p>
<p>If the link doesn't open, copy the following address into your browser:<br>%s</p>
<p><strong>Note: this link can only be used once.</strong></p>
<p>📅 <strong>Payout schedule:</strong> the <strong>first 3 purchases</strong> of each invited user are eligible for commission (20%% rate), released after a 10-day holding period. Payments are made via PayPal on the <strong>1st</strong> and <strong>15th</strong> of each month.</p>
<p>Feel free to contact us if you have any questions.</p>`,
			name, systemName, link, link, link)
	}
	return
}

// BuildAffiliateRejectionEmail returns subject and HTML body for the rejection notification.
func BuildAffiliateRejectionEmail(lang, name, systemName, reason string) (subject, content string) {
	lang = normalizeEmailLang(lang)

	reasonBlock := ""
	if reason != "" {
		var label string
		switch lang {
		case "zh":
			label = "审核意见"
		case "ja":
			label = "審査コメント"
		case "fr":
			label = "Commentaire de l'examinateur"
		case "es":
			label = "Comentario del revisor"
		default:
			label = "Reviewer comment"
		}
		reasonBlock = fmt.Sprintf(`<p><strong>%s:</strong></p><p style="padding:12px;background:#f5f5f5;border-left:4px solid #e53e3e;margin:0;">%s</p>`, label, reason)
	}

	switch lang {
	case "zh":
		subject = fmt.Sprintf("【%s】关于您的达人合作申请", systemName)
		content = fmt.Sprintf(`<p>您好 %s，</p>
<p>感谢您申请加入 <strong>%s</strong> 达人计划。</p>
<p>经过审核，我们遗憾地通知您，此次申请暂未通过。</p>
%s
<p>如果您认为这是一个错误，或希望了解更多信息，欢迎随时联系我们。待条件满足后，您也可以重新提交申请。</p>
<p>感谢您对 <strong>%s</strong> 的关注与支持。</p>`,
			name, systemName, reasonBlock, systemName)
	case "ja":
		subject = fmt.Sprintf("【%s】アフィリエイト申請について", systemName)
		content = fmt.Sprintf(`<p>%s 様、</p>
<p><strong>%s</strong> のアフィリエイトプログラムへのご申請ありがとうございます。</p>
<p>審査の結果、誠に残念ながら今回のご申請は承認されませんでした。</p>
%s
<p>ご不明な点がございましたら、いつでもお問い合わせください。条件が整い次第、再度ご申請いただけます。</p>
<p><strong>%s</strong> チームより</p>`,
			name, systemName, reasonBlock, systemName)
	case "fr":
		subject = fmt.Sprintf("【%s】Concernant votre demande d'affiliation", systemName)
		content = fmt.Sprintf(`<p>Bonjour %s,</p>
<p>Merci d'avoir postulé au programme d'affiliation <strong>%s</strong>.</p>
<p>Après examen, nous avons le regret de vous informer que votre demande n'a pas été retenue cette fois-ci.</p>
%s
<p>Si vous pensez qu'il s'agit d'une erreur ou souhaitez plus d'informations, n'hésitez pas à nous contacter. Vous pouvez également soumettre une nouvelle demande lorsque vous remplissez les conditions requises.</p>
<p>Merci de votre intérêt pour <strong>%s</strong>.</p>`,
			name, systemName, reasonBlock, systemName)
	case "es":
		subject = fmt.Sprintf("【%s】Sobre su solicitud de afiliado", systemName)
		content = fmt.Sprintf(`<p>Hola %s,</p>
<p>Gracias por solicitar unirse al programa de afiliados de <strong>%s</strong>.</p>
<p>Tras la revisión, lamentamos informarle que su solicitud no fue aprobada en esta ocasión.</p>
%s
<p>Si cree que esto es un error o desea más información, no dude en contactarnos. También puede volver a presentar una solicitud cuando cumpla los requisitos.</p>
<p>Gracias por su interés en <strong>%s</strong>.</p>`,
			name, systemName, reasonBlock, systemName)
	default: // en
		subject = fmt.Sprintf("【%s】Regarding your affiliate application", systemName)
		content = fmt.Sprintf(`<p>Hello %s,</p>
<p>Thank you for applying to the <strong>%s</strong> affiliate program.</p>
<p>After review, we regret to inform you that your application was not approved at this time.</p>
%s
<p>If you believe this is an error or would like more information, please feel free to contact us. You are also welcome to reapply once you meet the requirements.</p>
<p>Thank you for your interest in <strong>%s</strong>.</p>`,
			name, systemName, reasonBlock, systemName)
	}
	return
}

// BuildWithdrawalPaidEmail returns subject and HTML body for the withdrawal payment notification.
func BuildWithdrawalPaidEmail(lang, username, systemName string, amount float64, txId string) (subject, content string) {
	lang = normalizeEmailLang(lang)
	switch lang {
	case "zh":
		subject = fmt.Sprintf("【%s】您的提现申请已打款", systemName)
		content = fmt.Sprintf(`<p>您好 %s，</p>
<p>您申请提现 <strong>$%.2f</strong> 的款项已通过 PayPal 打款至您的账户。</p>
<p><strong>PayPal 交易单号：%s</strong></p>
<p>请登录您的 PayPal 账户查收，如有疑问请联系我们。</p>
<p><strong>%s</strong> 团队</p>`,
			username, amount, txId, systemName)
	case "ja":
		subject = fmt.Sprintf("【%s】出金申請の支払いが完了しました", systemName)
		content = fmt.Sprintf(`<p>%s 様、</p>
<p>申請された <strong>$%.2f</strong> の出金がPayPalアカウントに送金されました。</p>
<p><strong>PayPalトランザクションID：%s</strong></p>
<p>PayPalアカウントにログインしてご確認ください。ご不明な点がございましたらお問い合わせください。</p>
<p><strong>%s</strong> チーム</p>`,
			username, amount, txId, systemName)
	case "fr":
		subject = fmt.Sprintf("【%s】Votre retrait a été effectué", systemName)
		content = fmt.Sprintf(`<p>Bonjour %s,</p>
<p>Votre retrait de <strong>$%.2f</strong> a été envoyé sur votre compte PayPal.</p>
<p><strong>Identifiant de transaction PayPal : %s</strong></p>
<p>Connectez-vous à votre compte PayPal pour vérifier la réception. Contactez-nous en cas de question.</p>
<p>L'équipe <strong>%s</strong></p>`,
			username, amount, txId, systemName)
	case "es":
		subject = fmt.Sprintf("【%s】Su retiro ha sido procesado", systemName)
		content = fmt.Sprintf(`<p>Hola %s,</p>
<p>Su retiro de <strong>$%.2f</strong> ha sido enviado a su cuenta de PayPal.</p>
<p><strong>ID de transacción de PayPal: %s</strong></p>
<p>Inicie sesión en su cuenta de PayPal para confirmar la recepción. Contáctenos si tiene alguna pregunta.</p>
<p>El equipo de <strong>%s</strong></p>`,
			username, amount, txId, systemName)
	default: // en
		subject = fmt.Sprintf("【%s】Your withdrawal has been processed", systemName)
		content = fmt.Sprintf(`<p>Hello %s,</p>
<p>Your withdrawal of <strong>$%.2f</strong> has been sent to your PayPal account.</p>
<p><strong>PayPal Transaction ID: %s</strong></p>
<p>Please log in to your PayPal account to confirm receipt. Contact us if you have any questions.</p>
<p>The <strong>%s</strong> team</p>`,
			username, amount, txId, systemName)
	}
	return
}
