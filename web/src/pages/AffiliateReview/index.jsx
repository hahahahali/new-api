import React, { useEffect, useState } from 'react';
import {
  Button,
  Card,
  Popconfirm,
  Popover,
  Table,
  Tag,
  Typography,
  Space,
} from '@douyinfe/semi-ui';
import { IconCopy } from '@douyinfe/semi-icons';
import { API, showError, showSuccess, copy } from '../../helpers';
import { useTranslation } from 'react-i18next';

const { Text, Title } = Typography;

const STATUS_COLORS = {
  pending: 'orange',
  approved: 'green',
  rejected: 'red',
};

const STATUS_LABELS = {
  pending: '待审核',
  approved: '已通过',
  rejected: '已拒绝',
};

const SocialPopover = ({ record }) => {
  const socials = [
    record.instagram && { label: 'Instagram', value: `@${record.instagram}` },
    record.tiktok    && { label: 'TikTok',    value: `@${record.tiktok}` },
    record.youtube   && { label: 'YouTube',   value: record.youtube },
  ].filter(Boolean);

  if (!socials.length) return <Text type='tertiary'>-</Text>;

  const content = (
    <div style={{ padding: '8px 4px', minWidth: 240 }}>
      {socials.map(({ label, value }) => (
        <div
          key={label}
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '6px 0',
            borderBottom: '1px solid var(--semi-color-border)',
          }}
        >
          <div>
            <Text type='tertiary' size='small'>{label}</Text>
            <div>
              <Text
                style={{ userSelect: 'text', cursor: 'text', fontSize: 13 }}
              >
                {value}
              </Text>
            </div>
          </div>
          <Button
            size='small'
            theme='borderless'
            type='tertiary'
            icon={<IconCopy />}
            onClick={() => {
              copy(value);
              showSuccess('已复制');
            }}
          />
        </div>
      ))}
    </div>
  );

  return (
    <Popover content={content} trigger='click' position='bottomLeft' showArrow>
      <Text link style={{ cursor: 'pointer' }}>
        查看 ({socials.length})
      </Text>
    </Popover>
  );
};

const AffiliateReview = () => {
  const { t } = useTranslation();
  const [applications, setApplications] = useState([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize] = useState(20);
  const [loading, setLoading] = useState(false);
  const [filterStatus, setFilterStatus] = useState('');
  const [actionLoading, setActionLoading] = useState({});

  const loadApplications = async (p = page, s = filterStatus) => {
    setLoading(true);
    try {
      const params = new URLSearchParams({ page: p, page_size: pageSize });
      if (s) params.append('status', s);
      const res = await API.get(`/api/affiliate/applications?${params.toString()}`);
      if (res.data.success) {
        setApplications(res.data.data || []);
        setTotal(res.data.total || 0);
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
    loadApplications(1, filterStatus);
    setPage(1);
  }, [filterStatus]);

  const handleApprove = async (id) => {
    setActionLoading((prev) => ({ ...prev, [`approve_${id}`]: true }));
    try {
      const res = await API.post(`/api/affiliate/applications/${id}/approve`);
      if (res.data.success) {
        showSuccess(t('已通过申请，邮件已发送'));
        loadApplications();
      } else {
        showError(res.data.message);
      }
    } catch (e) {
      showError(t('操作失败'));
    } finally {
      setActionLoading((prev) => ({ ...prev, [`approve_${id}`]: false }));
    }
  };

  const handleReject = async (id) => {
    setActionLoading((prev) => ({ ...prev, [`reject_${id}`]: true }));
    try {
      const res = await API.post(`/api/affiliate/applications/${id}/reject`);
      if (res.data.success) {
        showSuccess(t('已拒绝申请'));
        loadApplications();
      } else {
        showError(res.data.message);
      }
    } catch (e) {
      showError(t('操作失败'));
    } finally {
      setActionLoading((prev) => ({ ...prev, [`reject_${id}`]: false }));
    }
  };

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: t('姓名'), dataIndex: 'name', width: 120 },
    { title: t('邮箱'), dataIndex: 'email', width: 220 },
    { title: t('国家'), dataIndex: 'country', width: 130, render: (v) => v || '-' },
    {
      title: t('社交账号'),
      width: 120,
      render: (_, record) => <SocialPopover record={record} />,
    },
    {
      title: t('状态'),
      dataIndex: 'status',
      width: 100,
      render: (v) => (
        <Tag color={STATUS_COLORS[v] || 'grey'}>
          {t(STATUS_LABELS[v] || v)}
        </Tag>
      ),
    },
    {
      title: t('申请时间'),
      dataIndex: 'created_at',
      width: 180,
      render: (v) => v ? new Date(v * 1000).toLocaleString() : '-',
    },
    {
      title: t('操作'),
      width: 160,
      render: (_, record) => {
        if (record.status !== 'pending') return null;
        return (
          <Space>
            <Popconfirm
              title={t('确认通过该申请？通过后将自动发送邀请邮件。')}
              onConfirm={() => handleApprove(record.id)}
              okText={t('通过')}
              cancelText={t('取消')}
            >
              <Button
                theme='solid'
                type='primary'
                size='small'
                loading={actionLoading[`approve_${record.id}`]}
              >
                {t('通过')}
              </Button>
            </Popconfirm>
            <Popconfirm
              title={t('确认拒绝该申请？')}
              onConfirm={() => handleReject(record.id)}
              okText={t('拒绝')}
              cancelText={t('取消')}
            >
              <Button
                theme='light'
                type='danger'
                size='small'
                loading={actionLoading[`reject_${record.id}`]}
              >
                {t('拒绝')}
              </Button>
            </Popconfirm>
          </Space>
        );
      },
    },
  ];

  return (
    <div style={{ paddingTop: 40 }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title heading={4} style={{ margin: 0 }}>{t('达人申请')}</Title>
        <Space>
          {['', 'pending', 'approved', 'rejected'].map((s) => (
            <Button
              key={s}
              theme={filterStatus === s ? 'solid' : 'light'}
              type={filterStatus === s ? 'primary' : 'tertiary'}
              size='small'
              onClick={() => setFilterStatus(s)}
            >
              {s === '' ? t('全部') : t(STATUS_LABELS[s])}
            </Button>
          ))}
        </Space>
      </div>
      <Card>
        <Table
          columns={columns}
          dataSource={applications}
          rowKey='id'
          loading={loading}
          pagination={{
            currentPage: page,
            pageSize,
            total,
            onPageChange: (p) => {
              setPage(p);
              loadApplications(p);
            },
          }}
          scroll={{ x: 'max-content' }}
        />
      </Card>
    </div>
  );
};

export default AffiliateReview;
