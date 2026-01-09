import { useState, useEffect, useRef } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Card, Button, Space, Typography, Breadcrumb, Spin, Modal, Input, message, Tabs, Tooltip } from 'antd'
import { SaveOutlined, HistoryOutlined, RocketOutlined, ReloadOutlined } from '@ant-design/icons'
import Editor from '@monaco-editor/react'
import { configApi } from '../api/configs'
import type { Config } from '../api/client'

const { Title, Text } = Typography

export default function ConfigEditorPage() {
  const { configId } = useParams<{ configId: string }>()
  const navigate = useNavigate()
  const [config, setConfig] = useState<Config | null>(null)
  const [content, setContent] = useState('')
  const [originalContent, setOriginalContent] = useState('')
  const [version, setVersion] = useState(0)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [commitModalOpen, setCommitModalOpen] = useState(false)
  const [commitMessage, setCommitMessage] = useState('')
  const editorRef = useRef<unknown>(null)

  const fetchConfig = async () => {
    if (!configId) return
    try {
      const { data } = await configApi.get(Number(configId))
      setConfig(data.config)
      setContent(data.content)
      setOriginalContent(data.content)
      setVersion(data.version)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchConfig()
  }, [configId])

  const hasChanges = content !== originalContent

  const handleSave = async () => {
    if (!configId || !hasChanges) return
    setCommitModalOpen(true)
  }

  const handleConfirmSave = async () => {
    if (!configId) return
    setSaving(true)
    try {
      const { data } = await configApi.update(Number(configId), {
        content,
        message: commitMessage || '更新配置',
      })
      setOriginalContent(content)
      setVersion(data.version.version)
      setCommitModalOpen(false)
      setCommitMessage('')
      message.success('保存成功')
    } finally {
      setSaving(false)
    }
  }

  const handleFormat = () => {
    if (!editorRef.current) return
    try {
      const formatted = JSON.stringify(JSON.parse(content), null, 2)
      setContent(formatted)
    } catch {
      message.error('JSON 格式化失败，请检查语法')
    }
  }

  const getLanguage = () => {
    if (config?.file_type === 'yaml') return 'yaml'
    return 'json'
  }

  if (loading) {
    return <div style={{ textAlign: 'center', padding: 100 }}><Spin size="large" /></div>
  }

  return (
    <div style={{ height: 'calc(100vh - 160px)', display: 'flex', flexDirection: 'column' }}>
      <Breadcrumb
        items={[
          { title: <a onClick={() => navigate('/projects')}>项目列表</a> },
          { title: <a onClick={() => navigate(`/projects/${config?.project_id}`)}>项目详情</a> },
          { title: config?.name || '配置编辑' },
        ]}
        style={{ marginBottom: 16 }}
      />

      <Card
        style={{ flex: 1, display: 'flex', flexDirection: 'column' }}
        bodyStyle={{ flex: 1, display: 'flex', flexDirection: 'column', padding: 0 }}
        title={
          <Space>
            <Title level={5} style={{ margin: 0 }}>{config?.name}</Title>
            <Text type="secondary">v{version}</Text>
            {hasChanges && <Text type="warning">（未保存）</Text>}
          </Space>
        }
        extra={
          <Space>
            <Tooltip title="格式化">
              <Button icon={<ReloadOutlined />} onClick={handleFormat} disabled={getLanguage() !== 'json'} />
            </Tooltip>
            <Button icon={<HistoryOutlined />} onClick={() => navigate(`/configs/${configId}/versions`)}>
              版本历史
            </Button>
            <Button icon={<RocketOutlined />} onClick={() => navigate(`/configs/${configId}/release`)}>
              发布
            </Button>
            <Button
              type="primary"
              icon={<SaveOutlined />}
              onClick={handleSave}
              disabled={!hasChanges}
            >
              保存
            </Button>
          </Space>
        }
      >
        <Tabs
          defaultActiveKey="code"
          style={{ flex: 1 }}
          tabBarStyle={{ margin: 0, paddingLeft: 16 }}
          items={[
            {
              key: 'code',
              label: '代码编辑',
              children: (
                <div style={{ height: 'calc(100vh - 320px)' }}>
                  <Editor
                    height="100%"
                    language={getLanguage()}
                    value={content}
                    onChange={(value) => setContent(value || '')}
                    onMount={(editor) => { editorRef.current = editor }}
                    options={{
                      minimap: { enabled: false },
                      fontSize: 14,
                      lineNumbers: 'on',
                      scrollBeyondLastLine: false,
                      automaticLayout: true,
                      tabSize: 2,
                    }}
                    theme="vs-light"
                  />
                </div>
              ),
            },
          ]}
        />
      </Card>

      <Modal
        title="提交更改"
        open={commitModalOpen}
        onCancel={() => setCommitModalOpen(false)}
        onOk={handleConfirmSave}
        confirmLoading={saving}
        okText="保存"
      >
        <Input.TextArea
          value={commitMessage}
          onChange={(e) => setCommitMessage(e.target.value)}
          placeholder="请输入提交说明（可选）"
          rows={3}
        />
      </Modal>
    </div>
  )
}
