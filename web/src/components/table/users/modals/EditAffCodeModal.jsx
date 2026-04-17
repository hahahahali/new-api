import React, { useEffect, useState } from 'react';
import { Input, Modal, Typography } from '@douyinfe/semi-ui';
import { API, showError, showSuccess } from '../../../../helpers';

const { Text } = Typography;

const EditAffCodeModal = ({ visible, onCancel, onSuccess, user, t }) => {
  const [affCode, setAffCode] = useState('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (visible && user) {
      setAffCode(user.aff_code || '');
    }
  }, [visible, user]);

  const handleConfirm = async () => {
    if (!affCode.trim()) {
      showError(t('邀请码不能为空'));
      return;
    }
    setLoading(true);
    try {
      const res = await API.put(`/api/kol/admin/users/${user.id}/aff_code`, {
        aff_code: affCode.trim(),
      });
      if (res.data.success) {
        showSuccess(t('邀请码修改成功'));
        onSuccess?.();
        onCancel();
      } else {
        showError(res.data.message || t('操作失败'));
      }
    } catch (e) {
      showError(t('操作失败'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal
      title={t('修改邀请码')}
      visible={visible}
      onCancel={onCancel}
      onOk={handleConfirm}
      confirmLoading={loading}
      okText={t('确认')}
      cancelText={t('取消')}
    >
      <div className='mb-3'>
        <Text type='tertiary'>
          {t('用户')}: <Text strong>{user?.username}</Text>
        </Text>
      </div>
      <Input
        value={affCode}
        onChange={setAffCode}
        placeholder={t('请输入新邀请码（最多 32 位）')}
        maxLength={32}
        showClear
      />
    </Modal>
  );
};

export default EditAffCodeModal;
