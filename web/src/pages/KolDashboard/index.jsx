import React, { useEffect, useState } from 'react';
import {
  Card,
  Table,
  Typography,
  Tag,
  Tabs,
  TabPane,
  Button,
  Modal,
  InputNumber,
  Space,
  Descriptions,
  Input,
  Form,
  Slider,
  Divider,
} from '@douyinfe/semi-ui';
import { API, showError, showSuccess } from '../../helpers';
import { useTranslation } from 'react-i18next';
import { useSecureVerification } from '../../hooks/common/useSecureVerification';
import SecureVerificationModal from '../../components/common/modals/SecureVerificationModal';

const { Text } = Typography;

function getNextPayoutDate() {
  const now = new Date();
  const day = now.getDate();
  const year = now.getFullYear();
  const month = now.getMonth();
  if (day < 15) {
    return new Date(year, month, 15);
  }
  return new Date(year, month + 1, 1);
}

function formatPayoutDate(date, t) {
  return t('{{month}}月{{day}}日', { month: date.getMonth() + 1, day: date.getDate() });
}

const KolDashboard = () => {
  const { t } = useTranslation();
  const [dashboard, setDashboard] = useState(null);
  const [invitees, setInvitees] = useState([]);
  const [inviteesTotal, setInviteesTotal] = useState(0);
  const [inviteesPage, setInviteesPage] = useState(1);
  const [commissions, setCommissions] = useState([]);
  const [commissionsTotal, setCommissionsTotal] = useState(0);
  const [commissionsPage, setCommissionsPage] = useState(1);
  const [withdrawals, setWithdrawals] = useState([]);
  const [withdrawalsTotal, setWithdrawalsTotal] = useState(0);
  const [withdrawalsPage, setWithdrawalsPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [withdrawOpen, setWithdrawOpen] = useState(false);
  const [withdrawLoading, setWithdrawLoading] = useState(false);
  const [withdrawForm, setWithdrawForm] = useState({ amount: 0, paypal_email: '', paypal_name: '' });
  const [rebateRate, setRebateRate] = useState(0);
  const [rebateLoading, setRebateLoading] = useState(false);
  const [affCodeEditing, setAffCodeEditing] = useState(false);
  const [affCodeInput, setAffCodeInput] = useState('');
  const [affCodeLoading, setAffCodeLoading] = useState(false);
  const pageSize = 10;

  const loadDashboard = async () => {
    try {
      const res = await API.get('/api/kol/dashboard');
      if (res.data.success) {
        setDashboard(res.data.data);
        setRebateRate(Math.round((res.data.data.kol_rebate_rate || 0) * 100));
      }
    } catch (e) {
      showError(t('加载失败'));
    }
  };

  const loadInvitees = async (page = 1) => {
    try {
      const res = await API.get(`/api/kol/invitees?p=${page}&page_size=${pageSize}`);
      if (res.data.success) {
        setInvitees(res.data.data.items || []);
        setInviteesTotal(res.data.data.total || 0);
      }
    } catch (e) { showError(t('加载失败')); }
  };

  const loadCommissions = async (page = 1) => {
    try {
      const res = await API.get(`/api/kol/commissions?p=${page}&page_size=${pageSize}`);
      if (res.data.success) {
        setCommissions(res.data.data.items || []);
        setCommissionsTotal(res.data.data.total || 0);
      }
    } catch (e) { showError(t('加载失败')); }
  };

  const loadWithdrawals = async (page = 1) => {
    try {
      const res = await API.get(`/api/kol/withdrawals?p=${page}&page_size=${pageSize}`);
      if (res.data.success) {
        setWithdrawals(res.data.data.items || []);
        setWithdrawalsTotal(res.data.data.total || 0);
      }
    } catch (e) { showError(t('加载失败')); }
  };

  useEffect(() => {
    setLoading(true);
    Promise.all([loadDashboard(), loadInvitees(), loadCommissions(), loadWithdrawals()])
      .finally(() => setLoading(false));
  }, []);

  const handleWithdrawSuccess = () => {
    const nextPayout = getNextPayoutDate();
    const dateStr = formatPayoutDate(nextPayout, t);
    showSuccess(t('提现申请已提交！款项将于 {{date}} 打款至您的 PayPal 账户，请注意查收。', { date: dateStr }));
    setWithdrawOpen(false);
    setWithdrawForm({ amount: 0, paypal_email: '', paypal_name: '' });
    loadDashboard();
    loadWithdrawals();
  };

  const {
    isModalVisible,
    verificationMethods,
    verificationState,
    withVerification,
    executeVerification,
    cancelVerification,
    setVerificationCode,
    switchVerificationMethod,
  } = useSecureVerification({
    onSuccess: (result) => {
      if (result?.success) {
        handleWithdrawSuccess();
      }
    },
  });

  const submitWithdrawApiCall = async () => {
    const res = await API.post('/api/kol/withdraw', {
      amount: withdrawForm.amount,
      paypal_email: withdrawForm.paypal_email,
      paypal_name: withdrawForm.paypal_name,
    });
    if (!res.data.success) {
      throw new Error(res.data.message || t('提现失败'));
    }
    return res.data;
  };

  const handleWithdraw = async () => {
    if (!withdrawForm.amount || withdrawForm.amount <= 0) {
      showError(t('请输入有效金额'));
      return;
    }
    if (!withdrawForm.paypal_email) {
      showError(t('请填写 PayPal 账户邮箱'));
      return;
    }
    if (!withdrawForm.paypal_name) {
      showError(t('请填写 PayPal 收款人姓名'));
      return;
    }
    setWithdrawLoading(true);
    try {
      const result = await withVerification(submitWithdrawApiCall, {
        title: t('申请提现'),
        description: t('为了保护账户安全，请验证您的身份。'),
        preferredMethod: 'passkey',
      });
      if (result?.success) {
        handleWithdrawSuccess();
      }
    } catch (e) {
      showError(e.message || t('提现失败'));
    } finally {
      setWithdrawLoading(false);
    }
  };

  const handleSaveRebate = async () => {    setRebateLoading(true);
    try {
      const res = await API.put('/api/kol/rebate-rate', { rebate_rate: rebateRate / 100 });
      if (res.data.success) {
        showSuccess(t('推荐折扣已更新'));
      } else {
        showError(res.data.message || t('保存失败'));
      }
    } catch (e) {
      showError(t('保存失败'));
    } finally {
      setRebateLoading(false);
    }
  };

  const startEditAffCode = () => {
    setAffCodeInput(dashboard?.aff_code || '');
    setAffCodeEditing(true);
  };

  const cancelEditAffCode = () => {
    setAffCodeEditing(false);
    setAffCodeInput('');
  };

  const handleSaveAffCode = async () => {
    const code = affCodeInput.trim();
    if (code.length < 4 || code.length > 20) {
      showError(t('邀请码长度必须在 4～20 个字符之间'));
      return;
    }
    if (!/^[a-zA-Z0-9_]+$/.test(code)) {
      showError(t('邀请码只能包含字母、数字和下划线'));
      return;
    }
    setAffCodeLoading(true);
    try {
      const res = await API.put('/api/kol/aff_code', { aff_code: code });
      if (res.data.success) {
        setDashboard((d) => ({ ...d, aff_code: res.data.data.aff_code }));
        setAffCodeEditing(false);
        setAffCodeInput('');
        showSuccess(t('邀请码已更新，请及时更新已分享的链接'));
      } else {
        showError(res.data.message || t('保存失败'));
      }
    } catch (e) {
      showError(t('保存失败'));
    } finally {
      setAffCodeLoading(false);
    }
  };

  const inviteeColumns = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: t('用户名'), dataIndex: 'username' },
    { title: t('邮箱'), dataIndex: 'email' },
    { title: t('累计充值'), dataIndex: 'total_recharge', render: (v) => `$${(v || 0).toFixed(2)}` },
    { title: t('累计佣金'), dataIndex: 'total_commission', render: (v) => `$${(v || 0).toFixed(2)}` },
  ];

  const commissionColumns = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: t('被邀请人 ID'), dataIndex: 'invitee_id', width: 120 },
    { title: t('订单号'), dataIndex: 'trade_no', width: 200, ellipsis: true },
    { title: t('充值金额'), dataIndex: 'recharge_amount', render: (v) => `$${(v || 0).toFixed(2)}` },
    { title: t('返佣比例'), dataIndex: 'commission_rate', render: (v) => `${((v || 0) * 100).toFixed(0)}%` },
    { title: t('佣金金额'), dataIndex: 'commission_amount', render: (v) => `$${(v || 0).toFixed(2)}` },
    {
      title: t('状态'),
      dataIndex: 'status',
      render: (v) => v === 'settled'
        ? <Tag color='green'>{t('已到账')}</Tag>
        : <Tag color='orange'>{t('冻结中')}</Tag>,
    },
    { title: t('时间'), dataIndex: 'created_at', render: (v) => v ? new Date(v * 1000).toLocaleString() : '-' },
  ];

  const withdrawalStatusMap = {
    pending:  { color: 'amber', text: '处理中' },
    approved: { color: 'blue',  text: '打款确认中' },
    rejected: { color: 'red',   text: '已拒绝' },
    paid:     { color: 'green', text: '已打款' },
  };

  const withdrawalColumns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: t('金额'), dataIndex: 'amount', render: (v) => `$${(v || 0).toFixed(2)}` },
    { title: t('PayPal 邮箱'), dataIndex: 'paypal_email', render: (v) => v || '-' },
    {
      title: t('状态'),
      dataIndex: 'status',
      render: (v) => {
        const s = withdrawalStatusMap[v] || { color: 'grey', text: v };
        return <Tag color={s.color}>{t(s.text)}</Tag>;
      },
    },
    { title: t('交易单号'), dataIndex: 'paypal_transaction_id', render: (v) => v || '-', ellipsis: true },
    { title: t('申请时间'), dataIndex: 'created_at', render: (v) => v ? new Date(v * 1000).toLocaleString() : '-' },
  ];

  const affLink = dashboard?.aff_code
    ? `${window.location.origin}/register?aff=${dashboard.aff_code}`
    : '';

  const minAmount = dashboard?.min_withdrawal_amount ?? 0;
  const canWithdraw = minAmount === 0 ? dashboard?.kol_balance > 0 : dashboard?.kol_balance >= minAmount;

  const nextPayout = getNextPayoutDate();
  const nextPayoutStr = formatPayoutDate(nextPayout, t);

  return (
    <div style={{ padding: '24px' }}>
      {/* Overview Cards */}
      <div className='grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4 mb-6'>
        <Card bodyStyle={{ padding: 16 }}>
          <Text type='tertiary' size='small'>{t('邀请人数')}</Text>
          <div className='text-2xl font-bold mt-1'>{dashboard?.invitee_count || 0}</div>
        </Card>
        <Card bodyStyle={{ padding: 16 }}>
          <Text type='tertiary' size='small'>{t('可提现余额')}</Text>
          <div className='text-2xl font-bold mt-1 text-green-600'>
            ${(dashboard?.kol_balance || 0).toFixed(2)}
          </div>
        </Card>
        <Card bodyStyle={{ padding: 16 }}>
          <Text type='tertiary' size='small'>{t('冻结中佣金')}</Text>
          <div className='text-2xl font-bold mt-1 text-orange-500'>
            ${(dashboard?.kol_pending_balance || 0).toFixed(2)}
          </div>
          <Text type='tertiary' size='small'>{t('10 天后自动解冻')}</Text>
        </Card>
        <Card bodyStyle={{ padding: 16 }}>
          <Text type='tertiary' size='small'>{t('历史总佣金')}</Text>
          <div className='text-2xl font-bold mt-1'>
            ${(dashboard?.kol_history_balance || 0).toFixed(2)}
          </div>
        </Card>
        <Card bodyStyle={{ padding: 16 }}>
          <Text type='tertiary' size='small'>{t('返佣比例')}</Text>
          <div className='text-2xl font-bold mt-1'>
            {Math.max(0, Math.round((dashboard?.commission_rate || 0) * 100) - rebateRate)}%
          </div>
        </Card>
      </div>

      {/* Invite Code & Actions */}
      <Card className='mb-6' bodyStyle={{ padding: 16 }}>
        <Descriptions row>
          <Descriptions.Item itemKey={t('邀请码')}>
            {affCodeEditing ? (
              <div>
                <Space align='center'>
                  <Input
                    value={affCodeInput}
                    onChange={setAffCodeInput}
                    placeholder='e.g. chainNo1'
                    maxLength={20}
                    style={{ width: 160 }}
                    onEnterPress={handleSaveAffCode}
                  />
                  <Button size='small' type='primary' loading={affCodeLoading} onClick={handleSaveAffCode}>{t('保存')}</Button>
                  <Button size='small' onClick={cancelEditAffCode}>{t('取消')}</Button>
                </Space>
                <div style={{ marginTop: 4 }}>
                  <Text type='tertiary' size='small'>{t('4~20位，字母 / 数字 / 下划线')}</Text>
                </div>
              </div>
            ) : (
              <Space align='center'>
                <Text copyable strong>{dashboard?.aff_code || '-'}</Text>
                <Button size='small' theme='borderless' onClick={startEditAffCode}>{t('修改')}</Button>
              </Space>
            )}
          </Descriptions.Item>
          <Descriptions.Item itemKey={t('邀请链接')}>
            {affCodeEditing ? (
              <Text type='warning' size='small'>⚠️ {t('保存后旧链接立即失效，请更新已分享的链接')}</Text>
            ) : (
              <Text copyable size='small'>{affLink || '-'}</Text>
            )}
          </Descriptions.Item>
        </Descriptions>
        <div className='mt-3'>
          <Space>
            <Button
              type='primary'
              onClick={() => setWithdrawOpen(true)}
              disabled={!canWithdraw}
            >
              {t('申请提现')}
            </Button>
            {!canWithdraw && (
              <Text type='tertiary' size='small'>
                {t('最低提现 ${{min}}，当前余额不足', { min: minAmount })}
              </Text>
            )}
            <Text type='tertiary' size='small'>
              {t('打款日为每月 1 日和 15 日，下次：{{date}}', { date: nextPayoutStr })}
            </Text>
          </Space>
        </div>

        <Divider margin='12px' />

        {/* Rebate Rate Setting */}
        <div>
          <Text strong>{t('推荐折扣')}</Text>
          <div style={{ marginTop: 4, marginBottom: 12 }}>
            <Text type='tertiary' size='small'>
              {t('您可以将一定比例的佣金转化为被邀请者的购买优惠，被邀请者通过您的邀请链接注册后，前 3 次充值可享受对应折扣')}
            </Text>
            <br />
            <Text type='tertiary' size='small'>
              {(() => {
                const commPct = Math.round((dashboard?.commission_rate || 0) * 100);
                const exampleRate = commPct > 0 ? Math.round(commPct / 2) : 10;
                const paid = (100 * (1 - exampleRate / 100)).toFixed(0);
                const commission = (100 * (commPct / 100 - exampleRate / 100)).toFixed(0);
                return t('（例：设置 {{rate}}% 后，好友充值 $100 实付 ${{paid}}，您获得佣金 ${{commission}}）', { rate: exampleRate, paid, commission });
              })()}
            </Text>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
            <div style={{ flex: 1 }}>
              {(() => {
                const maxVal = Math.round((dashboard?.commission_rate || 0) * 100);
                return (
                  <Slider
                    value={rebateRate}
                    onChange={setRebateRate}
                    min={0}
                    max={maxVal > 0 ? maxVal : 20}
                    step={1}
                    marks={{
                      0: '0%',
                      [maxVal > 0 ? maxVal : 20]: `${maxVal > 0 ? maxVal : 20}%`,
                    }}
                    disabled={!dashboard}
                  />
                );
              })()}
            </div>
            <Text strong style={{ minWidth: 40, textAlign: 'right' }}>{rebateRate}%</Text>
            <Button onClick={handleSaveRebate} loading={rebateLoading} size='small'>
              {t('保存')}
            </Button>
          </div>
        </div>
      </Card>

      {/* Data Tabs */}
      <Card>
        <Tabs>
          <TabPane tab={t('用户管理')} itemKey='invitees'>
            <Table columns={inviteeColumns} dataSource={invitees} loading={loading}
              pagination={{ total: inviteesTotal, pageSize, currentPage: inviteesPage,
                onPageChange: (p) => { setInviteesPage(p); loadInvitees(p); } }} size='small' />
          </TabPane>
          <TabPane tab={t('佣金流水')} itemKey='commissions'>
            <Table columns={commissionColumns} dataSource={commissions} loading={loading}
              pagination={{ total: commissionsTotal, pageSize, currentPage: commissionsPage,
                onPageChange: (p) => { setCommissionsPage(p); loadCommissions(p); } }} size='small' />
          </TabPane>
          <TabPane tab={t('提现记录')} itemKey='withdrawals'>
            <Table columns={withdrawalColumns} dataSource={withdrawals} loading={loading}
              pagination={{ total: withdrawalsTotal, pageSize, currentPage: withdrawalsPage,
                onPageChange: (p) => { setWithdrawalsPage(p); loadWithdrawals(p); } }} size='small' />
          </TabPane>
        </Tabs>
      </Card>

      {/* Withdraw Modal */}
      <Modal
        title={t('申请提现')}
        visible={withdrawOpen}
        onOk={handleWithdraw}
        onCancel={() => { setWithdrawOpen(false); setWithdrawForm({ amount: 0, paypal_email: '', paypal_name: '' }); }}
        confirmLoading={withdrawLoading}
        okText={t('提交申请')}
      >
        <div className='mb-4' style={{ background: '#f0f9ff', borderRadius: 8, padding: '10px 14px' }}>
          <Text type='tertiary' size='small'>
            {t('款项将于 {{date}} 打款至您的 PayPal 账户', { date: nextPayoutStr })}
          </Text>
        </div>
        <Form layout='vertical'>
          <Form.Section text={t('提现金额')}>
            <div className='mb-3'>
              <Text size='small' type='tertiary'>{t('可提现余额')}: <Text strong>${(dashboard?.kol_balance || 0).toFixed(2)}</Text> &nbsp;|&nbsp; {t('最低提现')}: <Text strong>${minAmount}</Text></Text>
            </div>
            <InputNumber
              value={withdrawForm.amount}
              onChange={(v) => setWithdrawForm((f) => ({ ...f, amount: v }))}
              min={minAmount}
              max={dashboard?.kol_balance || 0}
              step={1}
              prefix='$'
              style={{ width: '100%' }}
            />
          </Form.Section>
          <Form.Section text='PayPal'>
            <div className='mb-3'>
              <Text size='small' type='secondary' style={{ display: 'block', marginBottom: 6 }}>{t('PayPal 账户邮箱')} *</Text>
              <Input
                value={withdrawForm.paypal_email}
                onChange={(v) => setWithdrawForm((f) => ({ ...f, paypal_email: v }))}
                placeholder='yourname@example.com'
              />
            </div>
            <div>
              <Text size='small' type='secondary' style={{ display: 'block', marginBottom: 6 }}>{t('收款人姓名')} *</Text>
              <Input
                value={withdrawForm.paypal_name}
                onChange={(v) => setWithdrawForm((f) => ({ ...f, paypal_name: v }))}
                placeholder={t('与 PayPal 账户一致的真实姓名')}
              />
            </div>
          </Form.Section>
        </Form>
      </Modal>
      <SecureVerificationModal
        visible={isModalVisible}
        verificationMethods={verificationMethods}
        verificationState={verificationState}
        onVerify={executeVerification}
        onCancel={cancelVerification}
        onCodeChange={setVerificationCode}
        onMethodSwitch={switchVerificationMethod}
        title={verificationState.title}
        description={verificationState.description}
      />
    </div>
  );
};

export default KolDashboard;
