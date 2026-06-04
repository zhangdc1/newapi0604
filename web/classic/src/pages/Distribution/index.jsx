import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Button,
  Card,
  Col,
  Form,
  Row,
  Space,
  Table,
  Typography,
} from '@douyinfe/semi-ui';
import { API } from '../../helpers/api';
import {
  copy,
  isAdmin,
  showError,
  showSuccess,
} from '../../helpers';
import { renderQuota } from '../../helpers/render';

const { Title, Text } = Typography;

const defaultSetting = {
  enabled: false,
  reward_mode: 1,
  reward_rate: 10,
  fixed_reward_quota: 0,
  referral_limit: 0,
  freeze_days: 0,
  min_settlement_quota: 0,
  admin_recharge_trigger: false,
  redeem_temporary_attribution: true,
};

const getItems = (payload) => payload?.data?.items || [];

export default function Distribution() {
  const { t } = useTranslation();
  const [overview, setOverview] = useState(null);
  const [records, setRecords] = useState([]);
  const [invites, setInvites] = useState([]);
  const [setting, setSetting] = useState(defaultSetting);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);

  const load = async () => {
    setLoading(true);
    try {
      const [overviewRes, recordsRes, invitesRes] = await Promise.all([
        API.get('/api/user/distribution/overview'),
        API.get('/api/user/distribution/commission-records'),
        API.get('/api/user/distribution/invites'),
      ]);
      if (overviewRes.data.success) setOverview(overviewRes.data.data);
      setRecords(getItems(recordsRes.data));
      setInvites(getItems(invitesRes.data));

      if (isAdmin()) {
        const settingRes = await API.get('/api/distribution/admin/settings');
        if (settingRes.data.success) {
          setSetting({ ...defaultSetting, ...settingRes.data.data });
        }
      }
    } catch (err) {
      showError(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const settle = async () => {
    try {
      const res = await API.post('/api/user/distribution/settle');
      if (res.data.success) {
        showSuccess(t('奖励已转入账户余额'));
        await load();
      } else {
        showError(res.data.message || t('转入失败'));
      }
    } catch (err) {
      showError(err);
    }
  };

  const saveSetting = async (values) => {
    setSaving(true);
    try {
      const payload = {
        ...setting,
        ...values,
        reward_mode: Number(values.reward_mode || setting.reward_mode || 1),
        reward_rate: Number(values.reward_rate || 0),
        fixed_reward_quota: Number(values.fixed_reward_quota || 0),
        referral_limit: Number(values.referral_limit || 0),
        freeze_days: Number(values.freeze_days || 0),
        min_settlement_quota: Number(values.min_settlement_quota || 0),
      };
      const res = await API.put('/api/distribution/admin/settings', payload);
      if (res.data.success) {
        showSuccess(t('保存成功'));
        await load();
      } else {
        showError(res.data.message || t('保存失败'));
      }
    } catch (err) {
      showError(err);
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className='dashboard-container'>
      <Space vertical align='start' style={{ width: '100%' }}>
        <Title heading={3}>{t('分销中心')}</Title>

        <Row gutter={[16, 16]} style={{ width: '100%' }}>
          <Col xs={24} md={12} xl={6}>
            <Card title={t('邀请人数')}>
              <Title heading={4}>{overview?.invite_count || 0}</Title>
            </Card>
          </Col>
          <Col xs={24} md={12} xl={6}>
            <Card title={t('有效邀请')}>
              <Title heading={4}>{overview?.effective_invites || 0}</Title>
            </Card>
          </Col>
          <Col xs={24} md={12} xl={6}>
            <Card title={t('可转奖励')}>
              <Title heading={4}>{renderQuota(overview?.available_quota || 0)}</Title>
            </Card>
          </Col>
          <Col xs={24} md={12} xl={6}>
            <Card title={t('累计奖励')}>
              <Title heading={4}>{renderQuota(overview?.total_earned_quota || 0)}</Title>
            </Card>
          </Col>
        </Row>

        <Card title={t('邀请链接')} style={{ width: '100%' }}>
          <Space wrap>
            <Text copyable={false}>{overview?.invite_link || '-'}</Text>
            <Button
              onClick={() => {
                copy(overview?.invite_link || '');
                showSuccess(t('复制成功'));
              }}
              disabled={!overview?.invite_link}
            >
              {t('复制')}
            </Button>
            <Button
              theme='solid'
              onClick={settle}
              disabled={(overview?.available_quota || 0) <= 0}
            >
              {t('转入账户余额')}
            </Button>
          </Space>
        </Card>

        {isAdmin() && (
          <Card title={t('分销设置')} style={{ width: '100%' }}>
            <Form
              initValues={setting}
              onSubmit={saveSetting}
              key={JSON.stringify(setting)}
            >
              <Row gutter={[16, 8]}>
                <Col xs={24} md={8}>
                  <Form.Switch field='enabled' label={t('开启分销系统')} />
                </Col>
                <Col xs={24} md={8}>
                  <Form.Select field='reward_mode' label={t('奖励模式')}>
                    <Form.Select.Option value={1}>{t('按比例')}</Form.Select.Option>
                    <Form.Select.Option value={2}>{t('固定额度')}</Form.Select.Option>
                  </Form.Select>
                </Col>
                <Col xs={24} md={8}>
                  <Form.InputNumber field='reward_rate' label={t('奖励比例')} suffix='%' />
                </Col>
                <Col xs={24} md={8}>
                  <Form.InputNumber field='fixed_reward_quota' label={t('固定奖励额度')} />
                </Col>
                <Col xs={24} md={8}>
                  <Form.InputNumber field='referral_limit' label={t('奖励充值次数限制')} />
                </Col>
                <Col xs={24} md={8}>
                  <Form.InputNumber field='freeze_days' label={t('冻结天数')} />
                </Col>
                <Col xs={24} md={8}>
                  <Form.InputNumber field='min_settlement_quota' label={t('最低转入额度')} />
                </Col>
                <Col xs={24} md={8}>
                  <Form.Switch field='admin_recharge_trigger' label={t('后台加额触发奖励')} />
                </Col>
                <Col xs={24} md={8}>
                  <Form.Switch field='redeem_temporary_attribution' label={t('兑换码临时归因')} />
                </Col>
              </Row>
              <Button htmlType='submit' theme='solid' loading={saving}>
                {t('保存设置')}
              </Button>
            </Form>
          </Card>
        )}

        <Card title={t('奖励记录')} style={{ width: '100%' }}>
          <Table
            loading={loading}
            dataSource={records}
            pagination={false}
            rowKey='id'
            columns={[
              { title: 'ID', dataIndex: 'id' },
              { title: t('被邀请用户'), dataIndex: 'referred_user_id' },
              { title: t('来源'), dataIndex: 'source' },
              {
                title: t('充值额度'),
                render: (_, row) => renderQuota(row.increased_quota),
              },
              {
                title: t('奖励额度'),
                render: (_, row) => renderQuota(row.commission_quota),
              },
              { title: t('状态'), dataIndex: 'status' },
            ]}
          />
        </Card>

        <Card title={t('邀请用户')} style={{ width: '100%' }}>
          <Table
            loading={loading}
            dataSource={invites}
            pagination={false}
            rowKey='id'
            columns={[
              { title: 'ID', dataIndex: 'id' },
              { title: t('用户名'), dataIndex: 'username' },
              { title: t('奖励次数'), dataIndex: 'reward_count' },
              {
                title: t('贡献充值'),
                render: (_, row) => renderQuota(row.total_increased_quota),
              },
              {
                title: t('贡献奖励'),
                render: (_, row) => renderQuota(row.total_commission_quota),
              },
            ]}
          />
        </Card>
      </Space>
    </div>
  );
}
