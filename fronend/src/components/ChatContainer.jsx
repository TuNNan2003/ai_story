import { useRef, useEffect } from 'react'
import MessageList from './MessageList'
import EnhancedInputArea from './EnhancedInputArea'
import './ChatContainer.css'

function ChatContainer({ messages, onSendMessage, isLoading, models, selectedModel, onModelChange, enableTypewriter = true }) {
  const containerRef = useRef(null)

  useEffect(() => {
    // 自动滚动到底部
    if (containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight
    }
  }, [messages])

  return (
    <div className="chat-container">
      <div className="chat-messages" ref={containerRef}>
        <MessageList messages={messages} isLoading={isLoading} enableTypewriter={enableTypewriter} />
      </div>
      <EnhancedInputArea
        onSendMessage={onSendMessage}
        isLoading={isLoading}
        models={models}
        selectedModel={selectedModel}
        onModelChange={onModelChange}
        hasMessages={messages.length > 0}
      />
    </div>
  )
}

export default ChatContainer

