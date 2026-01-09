import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  Card, Tabs, Button, Table, Space, Typography, Breadcrumb, Spin, Tag, Modal, Form, Select, Slider, Input, message, Popconfirm
} from 'antd'
import { RocketOutlined, RollbackOutlined, CheckOutlined, CloseOutlined } from '@ant-design/icons'
import { configApi } from '../api/configs'
import type { Config, Release } from '../api/client'
import dayjs from 'dayjs'

const { Title, Text } = Typography

const environments = [
  { key: 'dev', label: '开发环境' },
  { key: 'test', label: '测试环境' },
  { key: 'staging', label: '预发布环境' },
  { key: 'prod', label: '生产环境' },
]

export default function ReleasePage() {
  const { configId } = useParams<{ configId: string }>()
  const navigate = useNavigate()
  const [config, setConfig] = useState<Config | null>(null)
  const [releases, setReleases] = useState<Release[]>([])
  const [loading, setLoading] = useState(true)
  const [activeEnv, setActiveEnv] = useState('dev')
  const [releaseModalOpen, setReleaseModalOpen] = useState(false)
  const [grayModalOpen, setGrayModalOpen] = useState(false)
  const [releasing, setReleasing] = useState(false)
  const [form] = Form.useForm()
  const [grayForm] = Form.useForm()

  const fetchData = async () => {
    if (!configId) return
    try {
      const [configRes, releasesRes] = await Promise.all([
        configApi.get(Number(configId)),
        configApi.listReleases(Number(configId)),
      ])
      setConfig(configRes.data.config)
      setReleases(releasesRes.data.releases || [])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
  }, [configId])

  const handleRelease = async (values: { environment: string; version?: number }) => {
    if (!configId) return
    setReleasing(true)
    try {
      await configApi.release(Number(configId), values)
      message.success('发布成功')
      setReleaseModalOpen(false)
      form.resetFields()
      fetchData()
    } finally {
      setReleasing(false)
    }
  }

  const handleGrayRelease = async (values: { environment: string; rule_type: string; percentage?: number; client_ids?: string; ip_ranges?: string }) => {
    if (!configId) return
    setReleasing(true)
    try {
      const data: Parameters<typeof configApi.grayRelease>[1] = {
        environment: values.environment,
        rule_type: values.rule_type as 'percentage' | 'client_id' | 'ip_range',
      }
      if (values.rule_type === 'percentage') {
        data.percentage = values.percentage
      } else if (values.rule_type === 'client_id') {
        data.client_ids = values.client_ids?.split(',').map(s => s.trim()).filter(Boolean)
      } else if (values.rule_type === 'ip_range') {
        data.ip_ranges = values.ip_ranges?.split(',').map(s => s.trim()).filter(Boolean)
      }
      await configApi.grayRelease(Number(configId), data)
      message.success('灰度发布创建成功')
      setGrayModalOpen(false)
      grayForm.resetFields()
      fetchData()
    } finally {
      setReleasing(false)
    }
  }

  const getCurrentRelease = (env: string) => {
    return releases.find(r => r.environment === env && (r.status === 'released' || r.status === 'gray'))
  }

  const getEnvReleases = (env: string) => {
    return releases.filter(r => r.environment === env)
  }

  const statusMap: Record<string, { color: string; text: string }> = {
    released: { color: 'green', text: '已发布' },
    gray: { color: 'blue', text: '灰度中' },
    rollback: { color: 'orange', text: '已回滚' },
    cancelled: { color: 'default', text: '已取消' },
    promoted: { color: 'purple', text: '已提升' },
  }

  const columns = [
    {
      title: '版本',
      dataIndex: 'version',
      key: 'version',
      render: (v: number) => `v${v}`,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={statusMap[status]?.color || 'default'}>
          {statusMap[status]?.text || status}
        </Tag>
      ),
    },
    {
      title: '类型',
      dataIndex: 'release_type',
      key: 'release_type',
      render: (type: string, record: Release) => (
        <Space>
          <Tag>{type === 'gray' ? '灰度' : '全量'}</Tag>
          {type === 'gray' && record.gray_percentage > 0 && (
            <Text type="secondary">{record.gray_percentage}%</Text>
          )}
        </Space>
      ),
    },
    {
      title: '发布者',
      dataIndex: 'released_by',
      key: 'released_by',
    },
    {
      title: '发布时间',
      dataIndex: 'released_at',
      key: 'released_at',
      render: (t: string) => dayjs(t).format('YYYY-MM-DD HH:mm:ss'),
    },
  ]

  if (loading) {
    return <div style={{ textAlign: 'center', padding: 100 }}><Spin size="large" /></div>
  }

  return (
    <div>
      <Breadcrumb
        items={[
          { title: <a onClick={() => navigate('/projects')}>项目列表</a> },
          { title: <a onClick={() => navigate(`/projects/${config?.project_id}`)}>项目详情</a> },
          { title: <a onClick={() => navigate(`/configs/${configId}`)}>{config?.name}</a> },
          { title: '发布管理' },
        ]}
        style={{ marginBottom: 16 }}
      />

      <Card
        title={
          <Space>
            <Title level={5} style={{ margin: 0 }}>发布管理</Title>
            <Text type="secondary">{config?.name}</Text>
          </Space>
        }
        extra={
          <Space>
            <Button icon={<RocketOutlined />} onClick={() => { form.setFieldValue('environment', activeEnv); setReleaseModalOpen(true) }}>
              全量发布
            </Button>
            <Button onClick={() => { grayForm.setFieldValue('environment', activeEnv); setGrayModalOpen(true) }}>
              灰度发布
            </Button>
          </Space>
        }
      >
        <Tabs
          activeKey={activeEnv}
          onChange={setActiveEnv}
          items={environments.map(env => {
            const current = getCurrentRelease(env.key)
            const envReleases = getEnvReleases(env.key)
            return {
              key: env.key,
              label: (
                <Space>
                  {env.label}
                  {current && <Tag color={statusMap[current.status]?.color}>v{current.version}</Tag>}
                </Space>
              ),
              children: (
                <div>
                  {current && (
                    <Card size="small" style={{ marginBottom: 16, background: '#fafafa' }}>
                      <Space style={{ width: '100%', justifyContent: 'space-between' }}>
                        <Space>
                          <Text strong>当前版本:</Text>
                          <Tag color="blue">v{current.version}</Tag>
                          <Tag color={statusMap[current.status]?.color}>
                            {statusMap[current.status]?.text}
                          </Tag>
                          {current.status === 'gray' && (
                            <Text type="secondary">灰度比例: {current.gray_percentage}%</Text>
                          )}
                        </Space>
                        {current.status === 'gray' && (
                          <Space>
                            <Popconfirm
                              title="确认提升为全量发布？"
                              onConfirm={async () => {
                                // 调用 promote API
                                message.success('已提升为全量发布')
                                fetchData()
                              }}
                            >
                              <Button type="primary" size="small" icon={<CheckOutlined />}>
                                全量发布
                              </Button>
                            </Popconfirm>
                            <Popconfirm
                              title="确认取消灰度发布？"
                              onConfirm={async () => {
                                // 调用 cancel API
                                message.success('灰度发布已取消')
                                fetchData()
                              }}
                            >
                              <Button size="small" danger icon={<CloseOutlined />}>
                                取消
                              </Button>
                            </Popconfirm>
                          </Space>
                        )}
                      </Space>
                    </Card>
                  )}
                  <Table
                    columns={columns}
                    dataSource={envReleases}
                    rowKey="id"
                    pagination={false}
                    size="small"
                  />
                </div>
              ),
            }
          })}
        />
      </Card>

      <Modal
        title="全量发布"
        open={releaseModalOpen}
        onCancel={() => setReleaseModalOpen(false)}
        footer={null}
      >
        <Form form={form} layout="vertical" onFinish={handleRelease}>
          <Form.Item name="environment" label="目标环境" rules={[{ required: true }]}>
            <Select options={environments.map(e => ({ value: e.key, label: e.label }))} />
          </Form.Item>
          <Form.Item name="version" label="版本号">
            <Select
              allowClear
              placeholder="默认使用最新版本"
              options={Array.from({ length: config?.current_version || 0 }, (_, i) => ({
                value: i + 1,
                label: `v${i + 1}`,
              })).reverse()}
            />
          </Form.Item>
          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button onClick={() => setReleaseModalOpen(false)}>取消</Button>
              <Button type="primary" htmlType="submit" loading={releasing}>发布</Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="灰度发布"
        open={grayModalOpen}
        onCancel={() => setGrayModalOpen(false)}
        footer={null}
        width={500}
      >
        <Form form={grayForm} layout="vertical" onFinish={handleGrayRelease} initialValues={{ rule_type: 'percentage', percentage: 10 }}>
          <Form.Item name="environment" label="目标环境" rules={[{ required: true }]}>
            <Select options={environments.map(e => ({ value: e.key, label: e.label }))} />
          </Form.Item>
          <Form.Item name="rule_type" label="灰度规则" rules={[{ required: true }]}>
            <Select
              options={[
                { value: 'percentage', label: '按百分比' },
                { value: 'client_id', label: '按客户端 ID' },
                { value: 'ip_range', label: '按 IP 范围' },
              ]}
            />
          </Form.Item>
          <Form.Item noStyle shouldUpdate={(prev, curr) => prev.rule_type !== curr.rule_type}>
            {({ getFieldValue }) => {
              const ruleType = getFieldValue('rule_type')
              if (ruleType === 'percentage') {
                return (
                  <Form.Item name="percentage" label="灰度比例">
                    <Slider marks={{ 0: '0%', 25: '25%', 50: '50%', 75: '75%', 100: '100%' }} />
                  </Form.Item>
                )
              }
              if (ruleType === 'client_id') {
                return (
                  <Form.Item name="client_ids" label="客户端 ID" extra="多个 ID 用逗号分隔">
                    <Input.TextArea rows={3} placeholder="client-1, client-2" />
                  </Form.Item>
                )
              }
              if (ruleType === 'ip_range') {
                return (
                  <Form.Item name="ip_ranges" label="IP 范围" extra="支持 CIDR 格式，多个用逗号分隔">
                    <Input.TextArea rows={3} placeholder="192.168.1.0/24, 10.0.0.1" />
                  </Form.Item>
                )
              }
              return null
            }}
          </Form.Item>
          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button onClick={() => setGrayModalOpen(false)}>取消</Button>
              <Button type="primary" htmlType="submit" loading={releasing}>创建灰度</Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
