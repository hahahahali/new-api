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
  Badge,
} from '@douyinfe/semi-ui';
import { API, showError, showSuccess, renderQuota } from '../../helpers';
import { useTranslation } from 'react-i18next';

const { Text } = Typography;

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
  const [withdrawAmount, setWithdrawAmount] = useState(0);
  const [withdrawLoading, setWithdrawLoading] = useState(false);
  const [connectLoading, setConnectLoading] = useState(false);
  const [connectStatus, setConnectStatus] = useState(null);
  const pageSize = 10;

  const loadDashboard = async () => {
    try {
      const res = await API.get('/api/kol/dashboard');
      if (res.data.success) {
        setDashboard(res.data.data);
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
    } catch (e) {
      showError(t('加载失败'));
    }
  };

  const loadCommissions = async (page = 1) => {
    try {
      const res = await API.get(`/api/kol/commissions?p=${page}&page_size=${pageSize}`);
      if (res.data.success) {
        setCommissions(res.data.data.items || []);
        setCommissionsTotal(res.data.data.total || 0);
      }
    } catch (e) {
      showError(t('加载失败'));
    }
  };

  const loadWithdrawals = async (page = 1) => {
    try {
      const res = await API.get(`/api/kol/withdrawals?p=${page}&page_size=${pageSize}`);
      if (res.data.success) {
        setWithdrawals(res.data.data.items || []);
        setWithdrawalsTotal(res.data.data.total || 0);
      }
    } catch (e) {
      showError(t('加载失败'));
    }
  };

  useEffect(() => {
    setLoading(true);
    Promise.all([loadDashboard(), loadInvitees(), loadCommissions(), loadWithdrawals(), checkConnectStatus()])
      .finally(() => setLoading(false));
  }, []);

  const checkConnectStatus = async () => {
    try {
      const res = await API.get('/api/kol/stripe-connect/status');
      if (res.data.success) {
        setConnectStatus(res.data.data);
      }
    } catch (e) {
      // Stripe Connect may not be enabled, ignore
    }
  };

  const handleOnboard = async () => {
    setConnectLoading(true);
    try {
      const res = await API.post('/api/kol/stripe-connect/onboard');
      if (res.data.success && res.data.data?.onboarding_url) {
        window.open(res.data.data.onboarding_url, '_blank');
        showSuccess(t('已打开 Stripe Connect 认证页面'));
      } else {
        showError(res.data.message || t('操作失败'));
      }
    } catch (e) {
      showError(t('操作失败'));
    } finally {
      setConnectLoading(false);
    }
  };

  const handleWithdraw = async () => {
    if (!withdrawAmount || withdrawAmount <= 0) {
      showError(t('请输入有效金额'));
      return;
    }
    setWithdrawLoading(true);
    try {
      const res = await API.post('/api/kol/withdraw', { amount: withdrawAmount });
      if (res.data.success) {
        showSuccess(res.data.message || t('提现成功'));
        setWithdrawOpen(false);
        setWithdrawAmount(0);
        loadDashboard();
        loadWithdrawals();
      } else {
        showError(res.data.message || t('提现失败'));
      }
    } catch (e) {
      showError(t('提现失败'));
    } finally {
      setWithdrawLoading(false);
    }
  };

  const inviteeColumns = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: t('用户名'), dataIndex: 'username' },
    { title: t('邮箱'), dataIndex: 'email' },
    {
      title: t('累计充值'),
      dataIndex: 'total_recharge',
      render: (v) => `$${(v || 0).toFixed(2)}`,
    },
    {
      title: t('累计佣金'),
      dataIndex: 'total_commission',
      render: (v) => `$${(v || 0).toFixed(2)}`,
    },
  ];

  const commissionColumns = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: t('被邀请人 ID'), dataIndex: 'invitee_id', width: 120 },
    { title: t('订单号'), dataIndex: 'trade_no', width: 200, ellipsis: true },
    {
      title: t('充值金额'),
      dataIndex: 'recharge_amount',
      render: (v) => `$${(v || 0).toFixed(2)}`,
    },
    {
      title: t('返佣比例'),
      dataIndex: 'commission_rate',
      render: (v) => `${((v || 0) * 100).toFixed(0)}%`,
    },
    {
      title: t('佣金金额'),
      dataIndex: 'commission_amount',
      render: (v) => `$${(v || 0).toFixed(2)}`,
    },
    {
      title: t('状态'),
      dataIndex: 'status',
      render: (v) => v === 'settled'
        ? <Tag color='green'>{t('已到账')}</Tag>
        : <Tag color='orange'>{t('冻结中')}</Tag>,
    },
    {
      title: t('时间'),
      dataIndex: 'created_at',
      render: (v) => v ? new Date(v * 1000).toLocaleString() : '-',
    },
  ];

  const withdrawalStatusMap = {
    pending: { color: 'amber', text: '处理中' },
    approved: { color: 'blue', text: '打款确认中' },
    rejected: { color: 'red', text: '已拒绝' },
    paid: { color: 'green', text: '已打款' },
  };

  const withdrawalColumns = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    {
      title: t('金额'),
      dataIndex: 'amount',
      render: (v) => `$${(v || 0).toFixed(2)}`,
    },
    {
      title: t('状态'),
      dataIndex: 'status',
      render: (v) => {
        const s = withdrawalStatusMap[v] || { color: 'grey', text: v };
        return <Tag color={s.color}>{t(s.text)}</Tag>;
      },
    },
    {
      title: t('拒绝原因'),
      dataIndex: 'reject_reason',
      render: (v) => v || '-',
    },
    {
      title: t('申请时间'),
      dataIndex: 'created_at',
      render: (v) => v ? new Date(v * 1000).toLocaleString() : '-',
    },
  ];

  const affLink = dashboard?.aff_code
    ? `${window.location.origin}/register?aff=${dashboard.aff_code}`
    : '';

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
          <Text type='tertiary' size='small'>{t('3 天后自动解冻')}</Text>
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
            {((dashboard?.commission_rate || 0) * 100).toFixed(0)}%
          </div>
        </Card>
      </div>

      {/* Invite Code & Link */}
      <Card className='mb-6' bodyStyle={{ padding: 16 }}>
        <Descriptions row>
          <Descriptions.Item itemKey={t('邀请码')}>
            <Text copyable strong>{dashboard?.aff_code || '-'}</Text>
          </Descriptions.Item>
          <Descriptions.Item itemKey={t('邀请链接')}>
            <Text copyable size='small'>{affLink || '-'}</Text>
          </Descriptions.Item>
        </Descriptions>
        <div className='mt-3'>
          <Space>
            <Button
              type='primary'
              onClick={() => setWithdrawOpen(true)}
              disabled={
                !dashboard?.kol_balance ||
                dashboard.kol_balance < (dashboard?.min_withdrawal_amount || 50) ||
                !connectStatus?.onboarded
              }
            >
              {t('申请提现')}
            </Button>
            {!connectStatus?.onboarded && (
              <Button
                theme='outline'
                loading={connectLoading}
                onClick={handleOnboard}
              >
                {connectStatus?.account_id ? t('继续 Stripe 认证') : t('开通 Stripe Connect')}
              </Button>
            )}
            {connectStatus?.onboarded && (
              <Tag color='green'>{t('Stripe Connect 已认证')}</Tag>
            )}
            {connectStatus && !connectStatus.onboarded && connectStatus.account_id && (
              <Text type='warning' size='small'>
                {t('Stripe Connect 认证未完成，请点击按钮继续')}
              </Text>
            )}
          </Space>
        </div>
      </Card>

      {/* Data Tabs */}
      <Card>
        <Tabs>
          <TabPane tab={t('用户管理')} itemKey='invitees'>
            <Table
              columns={inviteeColumns}
              dataSource={invitees}
              loading={loading}
              pagination={{
                total: inviteesTotal,
                pageSize,
                currentPage: inviteesPage,
                onPageChange: (p) => { setInviteesPage(p); loadInvitees(p); },
              }}
              size='small'
            />
          </TabPane>
          <TabPane tab={t('佣金流水')} itemKey='commissions'>
            <Table
              columns={commissionColumns}
              dataSource={commissions}
              loading={loading}
              pagination={{
                total: commissionsTotal,
                pageSize,
                currentPage: commissionsPage,
                onPageChange: (p) => { setCommissionsPage(p); loadCommissions(p); },
              }}
              size='small'
            />
          </TabPane>
          <TabPane tab={t('提现记录')} itemKey='withdrawals'>
            <Table
              columns={withdrawalColumns}
              dataSource={withdrawals}
              loading={loading}
              pagination={{
                total: withdrawalsTotal,
                pageSize,
                currentPage: withdrawalsPage,
                onPageChange: (p) => { setWithdrawalsPage(p); loadWithdrawals(p); },
              }}
              size='small'
            />
          </TabPane>
        </Tabs>
      </Card>

      {/* Withdraw Modal */}
      <Modal
        title={t('申请提现')}
        visible={withdrawOpen}
        onOk={handleWithdraw}
        onCancel={() => setWithdrawOpen(false)}
        confirmLoading={withdrawLoading}
      >
        <div className='mb-3'>
          <Text>{t('当前可提现余额')}: <Text strong>${(dashboard?.kol_balance || 0).toFixed(2)}</Text></Text>
        </div>
        <div className='mb-3'>
          <Text>{t('最低提现金额')}: <Text strong>${(dashboard?.min_withdrawal_amount || 50).toFixed(2)}</Text></Text>
        </div>
        <InputNumber
          value={withdrawAmount}
          onChange={setWithdrawAmount}
          min={dashboard?.min_withdrawal_amount || 50}
          max={dashboard?.kol_balance || 0}
          step={1}
          prefix='$'
          style={{ width: '100%' }}
        />
      </Modal>
    </div>
  );
};

export default KolDashboard;
