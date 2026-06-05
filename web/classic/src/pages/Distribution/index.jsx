import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Button,
  Card,
  Col,
  Form,
  Input,
  Row,
  Space,
  Table,
  Tag,
  Typography,
} from '@douyinfe/semi-ui';
import { API } from '../../helpers/api';
import { copy, isAdmin, showError, showSuccess } from '../../helpers';
import { renderQuota } from '../../helpers/render';

const { Title, Text } = Typography;

const defaultSetting = {
  enabled: false,
  normal_reward_mode: 1,
  normal_reward_rate: 10,
  normal_fixed_reward_quota: 0,
  normal_referral_limit: 5,
  agent_reward_mode: 1,
  agent_reward_rate: 10,
  agent_fixed_reward_quota: 0,
  agent_referral_limit: 0,
  freeze_days: 0,
  min_settlement_quota: 0,
  admin_recharge_trigger: false,
  redeem_temporary_attribution: true,
};

const getItems = (payload) => payload?.data?.items || [];

function numberValue(value, fallback = 0) {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : fallback;
}

export default function Distribution() {
  const { t } = useTranslation();
  const [overview, setOverview] = useState(null);
  const [records, setRecords] = useState([]);
  const [invites, setInvites] = useState([]);
  const [agents, setAgents] = useState([]);
  const [agentKeyword, setAgentKeyword] = useState('');
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
        const [settingRes, agentsRes] = await Promise.all([
          API.get('/api/distribution/admin/settings'),
          API.get('/api/distribution/admin/agents', { params: { page_size: 10 } }),
        ]);
        if (settingRes.data.success) {
          setSetting({ ...defaultSetting, ...settingRes.data.data });
        }
        setAgents(getItems(agentsRes.data));
      }
    } catch (err) {
      showError(err);
    } finally {
      setLoading(false);
    }
  };

  const searchAgents = async () => {
    try {
      const res = await API.get('/api/distribution/admin/agents', {
        params: { keyword: agentKeyword, page_size: 10 },
      });
      if (res.data.success) {
        setAgents(getItems(res.data));
      } else {
        showError(res.data.message || t('查询失败'));
      }
    } catch (err) {
      showError(err);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const copyText = async (text, successMessage) => {
    if (!text) return;
    if (await copy(text)) {
      showSuccess(successMessage);
    } else {
      showError(t('复制失败'));
    }
  };

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
        normal_reward_mode: numberValue(values.normal_reward_mode, 1),
        normal_reward_rate: numberValue(values.normal_reward_rate),
        normal_fixed_reward_quota: numberValue(values.normal_fixed_reward_quota),
        normal_referral_limit: numberValue(values.normal_referral_limit),
        agent_reward_mode: numberValue(values.agent_reward_mode, 1),
        agent_reward_rate: numberValue(values.agent_reward_rate),
        agent_fixed_reward_quota: numberValue(values.agent_fixed_reward_quota),
        agent_referral_limit: numberValue(values.agent_referral_limit),
        freeze_days: numberValue(values.freeze_days),
        min_settlement_quota: numberValue(values.min_settlement_quota),
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

  const updateAgent = async (userId, isAgentValue) => {
    try {
      const res = await API.put(`/api/distribution/admin/agents/${userId}`, {
        is_agent: isAgentValue,
      });
      if (res.data.success) {
        showSuccess(t('保存成功'));
        await searchAgents();
      } else {
        showError(res.data.message || t('保存失败'));
      }
    } catch (err) {
      showError(err);
    }
  };

  const statusText = (status) => {
    if (status === 1) return t('待结算');
    if (status === 2) return t('已结算');
    if (status === 3) return t('已失效');
    return t('未知');
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

        <Card title={t('邀请信息')} style={{ width: '100%' }}>
          <Space vertical align='start' style={{ width: '100%' }}>
            <Space wrap>
              <Text>{t('邀请码')}:</Text>
              <Text strong>{overview?.aff_code || '-'}</Text>
              <Button
                onClick={() => copyText(overview?.aff_code, t('邀请码已复制'))}
                disabled={!overview?.aff_code}
              >
                {t('复制邀请码')}
              </Button>
            </Space>
            <Space wrap>
              <Text>{t('邀请链接')}:</Text>
              <Text copyable={false}>{overview?.invite_link || '-'}</Text>
              <Button
                onClick={() => copyText(overview?.invite_link, t('邀请链接已复制'))}
                disabled={!overview?.invite_link}
              >
                {t('复制邀请链接')}
              </Button>
              <Button
                theme='solid'
                onClick={settle}
                disabled={(overview?.available_quota || 0) <= 0}
              >
                {t('转入账户余额')}
              </Button>
            </Space>
          </Space>
        </Card>

        {isAdmin() && (
          <Card title={t('分销设置')} style={{ width: '100%' }}>
            <Form initValues={setting} onSubmit={saveSetting} key={JSON.stringify(setting)}>
              <Row gutter={[16, 8]}>
                <Col xs={24} md={8}>
                  <Form.Switch field='enabled' label={t('开启分销系统')} />
                </Col>
                <Col xs={24} md={8}>
                  <Form.Switch field='admin_recharge_trigger' label={t('后台加额触发奖励')} />
                </Col>
                <Col xs={24} md={8}>
                  <Form.Switch field='redeem_temporary_attribution' label={t('兑换码临时归因')} />
                </Col>

                <Col xs={24}>
                  <Title heading={5}>{t('普通用户奖励设置')}</Title>
                </Col>
                <Col xs={24} md={6}>
                  <Form.Select field='normal_reward_mode' label={t('奖励模式')}>
                    <Form.Select.Option value={1}>{t('按比例')}</Form.Select.Option>
                    <Form.Select.Option value={2}>{t('固定额度')}</Form.Select.Option>
                  </Form.Select>
                </Col>
                <Col xs={24} md={6}>
                  <Form.InputNumber field='normal_reward_rate' label={t('奖励比例')} suffix='%' />
                </Col>
                <Col xs={24} md={6}>
                  <Form.InputNumber field='normal_fixed_reward_quota' label={t('固定奖励额度')} />
                </Col>
                <Col xs={24} md={6}>
                  <Form.InputNumber field='normal_referral_limit' label={t('奖励充值次数限制')} />
                </Col>

                <Col xs={24}>
                  <Title heading={5}>{t('代理奖励设置')}</Title>
                </Col>
                <Col xs={24} md={6}>
                  <Form.Select field='agent_reward_mode' label={t('奖励模式')}>
                    <Form.Select.Option value={1}>{t('按比例')}</Form.Select.Option>
                    <Form.Select.Option value={2}>{t('固定额度')}</Form.Select.Option>
                  </Form.Select>
                </Col>
                <Col xs={24} md={6}>
                  <Form.InputNumber field='agent_reward_rate' label={t('奖励比例')} suffix='%' />
                </Col>
                <Col xs={24} md={6}>
                  <Form.InputNumber field='agent_fixed_reward_quota' label={t('固定奖励额度')} />
                </Col>
                <Col xs={24} md={6}>
                  <Form.InputNumber field='agent_referral_limit' label={t('奖励充值次数限制')} />
                </Col>

                <Col xs={24} md={8}>
                  <Form.InputNumber field='freeze_days' label={t('冻结天数')} />
                </Col>
                <Col xs={24} md={8}>
                  <Form.InputNumber field='min_settlement_quota' label={t('最低转入额度')} />
                </Col>
              </Row>
              <Button htmlType='submit' theme='solid' loading={saving}>
                {t('保存设置')}
              </Button>
            </Form>
          </Card>
        )}

        {isAdmin() && (
          <Card title={t('代理管理')} style={{ width: '100%' }}>
            <Space style={{ marginBottom: 12 }} wrap>
              <Input
                value={agentKeyword}
                onChange={setAgentKeyword}
                placeholder={t('输入用户ID或用户名')}
                style={{ width: 240 }}
              />
              <Button onClick={searchAgents}>{t('查询')}</Button>
            </Space>
            <Table
              loading={loading}
              dataSource={agents}
              pagination={false}
              rowKey='id'
              columns={[
                { title: 'ID', dataIndex: 'id' },
                { title: t('用户名'), dataIndex: 'username' },
                { title: t('显示名称'), dataIndex: 'display_name' },
                {
                  title: t('分销身份'),
                  render: (_, row) =>
                    row.is_agent ? <Tag color='green'>{t('代理')}</Tag> : <Tag>{t('普通用户')}</Tag>,
                },
                {
                  title: t('操作'),
                  render: (_, row) => (
                    <Button onClick={() => updateAgent(row.id, !row.is_agent)}>
                      {row.is_agent ? t('取消代理') : t('设为代理')}
                    </Button>
                  ),
                },
              ]}
            />
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
              {
                title: t('分销身份'),
                render: (_, row) =>
                  row.referrer_is_agent ? <Tag color='green'>{t('代理')}</Tag> : <Tag>{t('普通用户')}</Tag>,
              },
              {
                title: t('状态'),
                render: (_, row) => statusText(row.status),
              },
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
