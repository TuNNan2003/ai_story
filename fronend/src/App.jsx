import { useState, useEffect } from 'react'
import ChatContainer from './components/ChatContainer'
import ConversationHistory from './components/ConversationHistory'
import './App.css'

function App() {
  const [messages, setMessages] = useState([])
  const [selectedModel, setSelectedModel] = useState('openai')
  const [models, setModels] = useState([])
  const [isLoading, setIsLoading] = useState(false)
  const [currentConversationId, setCurrentConversationId] = useState(null)
  const [conversations, setConversations] = useState([])
  const [isHistoryCollapsed, setIsHistoryCollapsed] = useState(false)
  const [isHistoryView, setIsHistoryView] = useState(false)

  useEffect(() => {
    fetchModels()
    fetchConversations()
  }, [])

  const fetchModels = async () => {
    try {
      const response = await fetch('/api/models')
      const data = await response.json()
      setModels(data.models)
      if (data.models.length > 0) {
        setSelectedModel(data.models[0].id)
      }
    } catch (error) {
      console.error('Failed to fetch models:', error)
    }
  }

  const fetchConversations = async () => {
    try {
      const response = await fetch('/api/conversations?page=1&page_size=50')
      const data = await response.json()
      setConversations(data.conversations || [])
    } catch (error) {
      console.error('Failed to fetch conversations:', error)
    }
  }

  const handleNewChat = async () => {
    try {
      const response = await fetch('/api/conversations/new', {
        method: 'POST',
      })
      if (response.ok) {
        const data = await response.json()
        setCurrentConversationId(data.id)
        setMessages([])
        setIsHistoryView(false) // 新对话，不是历史视图
        fetchConversations()
      }
    } catch (error) {
      console.error('Failed to create new conversation:', error)
    }
  }

  const handleSelectConversation = async (conversationId) => {
    setCurrentConversationId(conversationId)
    try {
      // 获取对话的文档列表
      const response = await fetch(`/api/documents?conversation_id=${conversationId}`)
      if (response.ok) {
        const data = await response.json()
        // 按照 created_at 排序（后端已经按 created_at ASC 排序）
        const conversationMessages = data.documents.map((doc) => ({
          id: doc.id,
          role: doc.role,
          content: doc.content,
        }))
        setMessages(conversationMessages)
        setIsHistoryView(true) // 加载历史对话后，设置为历史视图
      }
    } catch (error) {
      console.error('Failed to load conversation:', error)
    }
  }

  const getOrCreateConversationId = async () => {
    // 如果已有对话ID，直接返回
    if (currentConversationId) {
      return currentConversationId
    }
    // 如果没有对话ID，创建新对话
    // 注意：这里不刷新列表，因为新对话还没有标题，等到第一次发送消息后再刷新
    try {
      const response = await fetch('/api/conversations/new', {
        method: 'POST',
      })
      if (response.ok) {
        const data = await response.json()
        setCurrentConversationId(data.id)
        return data.id
      }
    } catch (error) {
      console.error('Failed to create conversation:', error)
    }
    return null
  }

  const handleSendMessage = async (message) => {
    if (!message.trim() || isLoading) return

    // 检查是否是新对话（第一次发送消息）
    const isNewConversation = !currentConversationId

    // 发送新消息时，切换到非历史视图（使用打字机效果）
    setIsHistoryView(false)

    // 首先获取或创建对话ID
    const conversationId = await getOrCreateConversationId()
    if (!conversationId) {
      console.error('Failed to get conversation ID')
      return
    }

    const userMessage = {
      id: Date.now(),
      role: 'user',
      content: message,
    }

    setMessages((prev) => [...prev, userMessage])
    setIsLoading(true)

    const assistantMessage = {
      id: Date.now() + 1,
      role: 'assistant',
      content: '',
    }
    setMessages((prev) => [...prev, assistantMessage])

    try {
      // 前端仅携带本次用户请求的内容
      const response = await fetch('/api/chat', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          conversation_id: conversationId,
          model: selectedModel,
          messages: [
            {
              role: 'user',
              content: message,
            },
          ],
        }),
      })

      if (!response.ok) {
        throw new Error('Failed to get response')
      }

      const reader = response.body.getReader()
      const decoder = new TextDecoder()
      let fullContent = ''

      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        const chunk = decoder.decode(value, { stream: true })
        fullContent += chunk

        setMessages((prev) => {
          const newMessages = [...prev]
          const lastMessage = newMessages[newMessages.length - 1]
          if (lastMessage.role === 'assistant') {
            lastMessage.content = fullContent
          }
          return newMessages
        })
      }
      // 只有在新对话第一次发送消息时才刷新列表（更新标题）
      // 已存在的对话不需要刷新，因为用户已经在当前对话中
      if (isNewConversation) {
        await fetchConversations()
      }
    } catch (error) {
      console.error('Error:', error)
      setMessages((prev) => {
        const newMessages = [...prev]
        const lastMessage = newMessages[newMessages.length - 1]
        if (lastMessage.role === 'assistant') {
          lastMessage.content = '抱歉，发生了错误：' + error.message
        }
        return newMessages
      })
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="app">
      <ConversationHistory
        conversations={conversations}
        currentConversationId={currentConversationId}
        onSelectConversation={handleSelectConversation}
        isCollapsed={isHistoryCollapsed}
        onToggleCollapse={() => setIsHistoryCollapsed(!isHistoryCollapsed)}
      />
      <div className="main-content">
        <div className="new-chat-button-container">
          <button onClick={handleNewChat} className="new-chat-button-top">
            <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
              <path
                d="M8 1v14M1 8h14"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
              />
            </svg>
            新对话
          </button>
        </div>
        <ChatContainer
          messages={messages}
          onSendMessage={handleSendMessage}
          isLoading={isLoading}
          models={models}
          selectedModel={selectedModel}
          onModelChange={setSelectedModel}
          enableTypewriter={!isHistoryView}
        />
      </div>
    </div>
  )
}

export default App

