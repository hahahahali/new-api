/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import React, { useEffect, useState, useRef } from 'react';
import { Button, Col, Form, Row, Spin } from '@douyinfe/semi-ui';
import {
  API,
  removeTrailingSlash,
  showError,
  showSuccess,
  verifyJSON,
} from '../../../helpers';
import { useTranslation } from 'react-i18next';

export default function SettingsGeneralPayment(props) {
  const { t } = useTranslation();
  const sectionTitle = props.hideSectionTitle ? undefined : t('通用设置');
  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState({
    ServerAddress: '',
    CustomCallbackAddress: '',
    TopupGroupRatio: '',
    PayMethods: '',
    AmountOptions: '',
    AmountDiscount: '',
    AmountOptionNames: '',
    AmountOptionDescs: '',
  });
  const [originInputs, setOriginInputs] = useState({});
  const formApiRef = useRef(null);

  useEffect(() => {
    if (props.options && formApiRef.current) {
      const currentInputs = {
        ServerAddress: props.options.ServerAddress || '',
        CustomCallbackAddress: props.options.CustomCallbackAddress || '',
        TopupGroupRatio: props.options.TopupGroupRatio || '',
        PayMethods: props.options.PayMethods || '',
        AmountOptions: props.options.AmountOptions || '',
        AmountDiscount: props.options.AmountDiscount || '',
        AmountOptionNames: props.options.AmountOptionNames || '',
        AmountOptionDescs: props.options.AmountOptionDescs || '',
      };
      setInputs(currentInputs);
      setOriginInputs({ ...currentInputs });
      formApiRef.current.setValues(currentInputs);
    }
  }, [props.options]);

  const handleFormChange = (values) => {
    setInputs(values);
  };

  const submitGeneralSettings = async () => {
    if (
      originInputs.TopupGroupRatio !== inputs.TopupGroupRatio &&
      !verifyJSON(inputs.TopupGroupRatio)
    ) {
      showError(t('充值分组倍率不是合法的 JSON 字符串'));
      return;
    }

    if (
      originInputs.PayMethods !== inputs.PayMethods &&
      !verifyJSON(inputs.PayMethods)
    ) {
      showError(t('充值方式设置不是合法的 JSON 字符串'));
      return;
    }

    if (
      originInputs.AmountOptions !== inputs.AmountOptions &&
      inputs.AmountOptions.trim() !== '' &&
      !verifyJSON(inputs.AmountOptions)
    ) {
      showError(t('自定义充值数量选项不是合法的 JSON 数组'));
      return;
    }

    if (
      originInputs.AmountDiscount !== inputs.AmountDiscount &&
      inputs.AmountDiscount.trim() !== '' &&
      !verifyJSON(inputs.AmountDiscount)
    ) {
      showError(t('充值金额折扣配置不是合法的 JSON 对象'));
      return;
    }

    if (
      originInputs.AmountOptionNames !== inputs.AmountOptionNames &&
      inputs.AmountOptionNames.trim() !== '' &&
      !verifyJSON(inputs.AmountOptionNames)
    ) {
      showError(t('套餐名称配置不是合法的 JSON 格式'));
      return;
    }

    if (
      originInputs.AmountOptionDescs !== inputs.AmountOptionDescs &&
      inputs.AmountOptionDescs.trim() !== '' &&
      !verifyJSON(inputs.AmountOptionDescs)
    ) {
      showError(t('套餐副标题配置不是合法的 JSON 格式'));
      return;
    }

    setLoading(true);
    try {
      const options = [
        {
          key: 'ServerAddress',
          value: removeTrailingSlash(inputs.ServerAddress),
        },
      ];

      if (inputs.CustomCallbackAddress !== '') {
        options.push({
          key: 'CustomCallbackAddress',
          value: removeTrailingSlash(inputs.CustomCallbackAddress),
        });
      }
      if (originInputs.TopupGroupRatio !== inputs.TopupGroupRatio) {
        options.push({ key: 'TopupGroupRatio', value: inputs.TopupGroupRatio });
      }
      if (originInputs.PayMethods !== inputs.PayMethods) {
        options.push({ key: 'PayMethods', value: inputs.PayMethods });
      }
      if (originInputs.AmountOptions !== inputs.AmountOptions) {
        options.push({
          key: 'payment_setting.amount_options',
          value: inputs.AmountOptions,
        });
      }
      if (originInputs.AmountDiscount !== inputs.AmountDiscount) {
        options.push({
          key: 'payment_setting.amount_discount',
          value: inputs.AmountDiscount,
        });
      }
      if (originInputs.AmountOptionNames !== inputs.AmountOptionNames) {
        options.push({
          key: 'payment_setting.amount_option_names',
          value: inputs.AmountOptionNames,
        });
      }
      if (originInputs.AmountOptionDescs !== inputs.AmountOptionDescs) {
        options.push({
          key: 'payment_setting.amount_option_descs',
          value: inputs.AmountOptionDescs,
        });
      }

      const results = await Promise.all(
        options.map((option) =>
          API.put('/api/option/', {
            key: option.key,
            value: option.value,
          }),
        ),
      );

      const errorResults = results.filter((res) => !res.data.success);
      if (errorResults.length === 0) {
        showSuccess(t('更新成功'));
        setOriginInputs({ ...inputs });
        props.refresh && props.refresh();
      } else {
        errorResults.forEach((res) => {
          showError(res.data.message);
        });
      }
    } catch (error) {
      showError(t('更新失败'));
    }
    setLoading(false);
  };

  return (
    <Spin spinning={loading}>
      <Form
        initValues={inputs}
        onValueChange={handleFormChange}
        getFormApi={(api) => (formApiRef.current = api)}
      >
        <Form.Section text={sectionTitle}>
          <Form.Input
            field='ServerAddress'
            label={t('服务器地址')}
            placeholder={'https://yourdomain.com'}
            style={{ width: '100%' }}
            extraText={t(
              '该服务器地址将影响支付回调地址以及默认首页展示的地址，请确保正确配置',
            )}
          />
          <Row
            gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
            style={{ marginTop: 16 }}
          >
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.Input
                field='CustomCallbackAddress'
                label={t('回调地址')}
                placeholder={t('例如：https://yourdomain.com')}
                extraText={t(
                  '留空时默认使用服务器地址作为回调地址，填写后将覆盖默认值',
                )}
              />
            </Col>
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.TextArea
                field='TopupGroupRatio'
                label={t('充值分组倍率')}
                placeholder={t('为一个 JSON 文本，键为组名称，值为倍率')}
                autosize
              />
            </Col>
          </Row>
          <Row
            gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
            style={{ marginTop: 16 }}
          >
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.TextArea
                field='PayMethods'
                label={t('充值方式设置')}
                placeholder={t('为一个 JSON 文本')}
                autosize
              />
            </Col>
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.TextArea
                field='AmountOptions'
                label={t('自定义充值数量选项')}
                placeholder={t(
                  '为一个 JSON 数组，例如：[10, 20, 50, 100, 200, 500]',
                )}
                autosize
                extraText={t(
                  '设置用户可选择的充值数量选项，例如：[10, 20, 50, 100, 200, 500]',
                )}
              />
            </Col>
          </Row>
          <Row style={{ marginTop: 16 }}>
            <Col span={24}>
              <Form.TextArea
                field='AmountDiscount'
                label={t('充值金额折扣配置')}
                placeholder={t(
                  '为一个 JSON 对象，例如：{"100": 0.95, "200": 0.9, "500": 0.85}',
                )}
                autosize
                extraText={t(
                  '设置不同充值金额对应的折扣，键为充值金额，值为折扣率，例如：{"100": 0.95, "200": 0.9, "500": 0.85}',
                )}
              />
            </Col>
          </Row>
          <Row style={{ marginTop: 16 }}>
            <Col span={24}>
              <Form.TextArea
                field='AmountOptionNames'
                label={t('套餐名称配置')}
                placeholder={t(
                  '支持 JSON 数组或多语言对象，例如：["Starter", "Basic"] 或 {"zh": ["入门版"], "en": ["Starter"]}',
                )}
                autosize
                extraText={t(
                  '支持多语言格式，例如：{"zh": ["入门版", "基础版"], "en": ["Starter", "Basic"]}',
                )}
              />
            </Col>
          </Row>
          <Row style={{ marginTop: 16 }}>
            <Col span={24}>
              <Form.TextArea
                field='AmountOptionDescs'
                label={t('套餐副标题配置')}
                placeholder={t(
                  '支持 JSON 数组或多语言对象，例如：["适合初次体验", "适合个人爱好者"] 或 {"zh": ["适合初次体验"], "en": ["For beginners"]}',
                )}
                autosize
                extraText={t(
                  '支持多语言格式，例如：{"zh": ["适合初次体验", "适合个人爱好者"], "en": ["Perfect for beginners", "For hobbyists"]}',
                )}
              />
            </Col>
          </Row>
          <Button onClick={submitGeneralSettings} style={{ marginTop: 16 }}>
            {t('保存通用设置')}
          </Button>
        </Form.Section>
      </Form>
    </Spin>
  );
}
