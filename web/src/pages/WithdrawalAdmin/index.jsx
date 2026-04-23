import React, { useEffect, useState } from 'react';
import {
  Button,
  Card,
  Input,
  Modal,
  Space,
  Table,
  Tag,
  Typography,
} from '@douyinfe/semi-ui';
import { API, showError, showSuccess } from '../../helpers';
import { useTranslation } from 'react-i18next';
import { useSecureVerification } from '../../hooks/common/useSecureVerification';
import SecureVerificationModal from '../../components/common/modals/SecureVerificationModal';

const { Text, Title } = Typography;

const STATUS_COLORS = { pending: 'orange', approved: 'blue', rejected: 'red', paid: 'green' };
const STATUS_LABELS = { pending: '处理中', approved: '打款确认中', rejected: '已拒绝', paid: '已打款' };

const WithdrawalAdmin = () => {
  const { t } = useTranslation();
  const [items, setItems] = useState([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize] = useState(20);
  const [loading, setLoading] = useState(false);
  const [filterStatus, setFilterStatus] = useState('pending');

  // Pay modal state
  const [payModal, setPayModal] = useState({ visible: false, id: null, username: '', amount: 0 });
  const [txId, setTxId] = useState('');
  const [payLoading, setPayLoading] = useState(false);

  const load = async (p = page, s = filterStatus) => {
    setLoading(true);
    try {
      const params = new URLSearchParams({ page: p, page_size: pageSize });
      if (s) params.append('status', s);
      const res = await API.get(`/api/kol/admin/withdrawals?${params.toString()}`);
      if (res.data.success) {
        setItems(res.data.data.items || []);
        setTotal(res.data.data.total || 0);
      } else {
        showError(res.data.message);
      }
    } catch (e) {
      showError(t('加载失败'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load(1, filterStatus);
    setPage(1);
  }, [filterStatus]);

  const handlePaySuccess = () => {
    showSuccess(t('已确认打款，通知邮件已发送'));
    setPayModal({ visible: false, id: null });
    load(page, filterStatus);
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
        handlePaySuccess();
      }
    },
  });

  const openPayModal = (record) => {
    setTxId('');
    setPayModal({ visible: true, id: record.id, username: record.username || record.user_email, amount: record.amount });
  };

  const submitPayApiCall = async () => {
    const res = await API.post(`/api/kol/admin/withdrawals/${payModal.id}/pay`, {
      paypal_transaction_id: txId.trim(),
    });
    if (!res.data.success) {
      throw new Error(res.data.message || t('操作失败'));
    }
    return res.data;
  };

  const handlePay = async () => {
    if (!txId.trim()) {
      showError(t('请填写 PayPal 交易单号'));
      return;
    }
    setPayLoading(true);
    try {
      const result = await withVerification(submitPayApiCall, {
        title: t('确认打款'),
        description: t('为了保护账户安全，请验证您的身份。'),
        preferredMethod: 'passkey',
      });
      if (result?.success) {
        handlePaySuccess();
      }
    } catch (e) {
      showError(e.message || t('操作失败'));
    } finally {
      setPayLoading(false);
    }
  };

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: t('达人'), width: 150, render: (_, r) => (
      <div>
        <Text strong>{r.username || '-'}</Text>
        <Text type='tertiary' size='small' style={{ display: 'block' }}>{r.user_email || '-'}</Text>
      </div>
    )},
    { title: t('金额'), dataIndex: 'amount', width: 100, render: (v) => <Text strong>${(v || 0).toFixed(2)}</Text> },
    { title: t('PayPal 邮箱'), dataIndex: 'paypal_email', width: 200, render: (v) => v || '-' },
    { title: t('收款人姓名'), dataIndex: 'paypal_name', width: 140, render: (v) => v || '-' },
    {
      title: t('状态'), dataIndex: 'status', width: 100,
      render: (v) => <Tag color={STATUS_COLORS[v] || 'grey'}>{t(STATUS_LABELS[v] || v)}</Tag>,
    },
    { title: t('交易单号'), dataIndex: 'paypal_transaction_id', width: 160, render: (v) => v || '-', ellipsis: true },
    { title: t('申请时间'), dataIndex: 'created_at', width: 180, render: (v) => v ? new Date(v * 1000).toLocaleString() : '-' },
    {
      title: t('操作'), width: 100,
      render: (_, record) => {
        if (record.status !== 'pending' && record.status !== 'approved') return null;
        return (
          <Button theme='solid' type='primary' size='small' onClick={() => openPayModal(record)}>
            {t('确认打款')}
          </Button>
        );
      },
    },
  ];

  return (
    <div style={{ paddingTop: 40 }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title heading={4} style={{ margin: 0 }}>{t('提现管理')}</Title>
        <Space>
          {['', 'pending', 'paid', 'rejected'].map((s) => (
            <Button
              key={s}
              theme={filterStatus === s ? 'solid' : 'light'}
              type={filterStatus === s ? 'primary' : 'tertiary'}
              size='small'
              onClick={() => setFilterStatus(s)}
            >
              {s === '' ? t('全部') : t(STATUS_LABELS[s] || s)}
            </Button>
          ))}
        </Space>
      </div>
      <Card>
        <Table
          columns={columns}
          dataSource={items}
          rowKey='id'
          loading={loading}
          pagination={{
            currentPage: page, pageSize, total,
            onPageChange: (p) => { setPage(p); load(p); },
          }}
          scroll={{ x: 'max-content' }}
        />
      </Card>

      <Modal
        title={t('确认 PayPal 打款')}
        visible={payModal.visible}
        onCancel={() => setPayModal({ visible: false, id: null })}
        onOk={handlePay}
        okText={t('确认打款')}
        cancelText={t('取消')}
        confirmLoading={payLoading}
      >
        <div style={{ marginBottom: 16 }}>
          <Text type='secondary'>
            {t('达人：{{name}}，金额：${{amount}}', { name: payModal.username, amount: (payModal.amount || 0).toFixed(2) })}
          </Text>
        </div>
        <Text style={{ display: 'block', marginBottom: 8 }}>{t('PayPal 交易单号')} *</Text>
        <Input
          value={txId}
          onChange={setTxId}
          placeholder={t('请输入 PayPal 交易完成后的交易单号')}
        />
        <Text type='tertiary' size='small' style={{ display: 'block', marginTop: 8 }}>
          {t('确认后将自动向达人发送打款通知邮件。')}
        </Text>
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

export default WithdrawalAdmin;
