import { useState, useEffect } from 'react'
import ChatContainer from './components/ChatContainer'
import Sidebar from './components/Sidebar'
import './App.css'

function App() {
  const [messages, setMessages] = useState([])
  const [selectedModel, setSelectedModel] = useState('openai')
  const [models, setModels] = useState([])
  const [isLoading, setIsLoading] = useState(false)

  useEffect(() => {
    fetchModels()
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

  const handleSendMessage = async (message) => {
    if (!message.trim() || isLoading) return

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

  const handleNewChat = () => {
    setMessages([])
  }

  return (
    <div className="app">
      <Sidebar
        models={models}
        selectedModel={selectedModel}
        onModelChange={setSelectedModel}
        onNewChat={handleNewChat}
      />
      <ChatContainer
        messages={messages}
        onSendMessage={handleSendMessage}
        isLoading={isLoading}
      />
    </div>
  )
}

export default App

