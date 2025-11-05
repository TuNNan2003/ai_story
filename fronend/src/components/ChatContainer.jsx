import { useRef, useEffect } from 'react'
import MessageList from './MessageList'
import EnhancedInputArea from './EnhancedInputArea'
import './ChatContainer.css'

function ChatContainer({ messages, onSendMessage, isLoading, models, selectedModel, onModelChange, enableTypewriter = true, onLoadMore, canLoadMore, isLoadingMore }) {
  const containerRef = useRef(null)

  useEffect(() => {
    // 自动滚动到底部（仅在消息数量增加时，且不是加载更早的消息时）
    if (containerRef.current) {
      // 如果是在加载新消息（消息数量增加且最后一条是助手消息），滚动到底部
      // 如果是加载更早的消息，不滚动（保持当前位置）
      if (!isLoadingMore) {
        containerRef.current.scrollTop = containerRef.current.scrollHeight
      }
    }
  }, [messages, isLoadingMore])

  return (
    <div className="chat-container">
      <div className="chat-messages" ref={containerRef}>
        <MessageList
          messages={messages}
          isLoading={isLoading}
          enableTypewriter={enableTypewriter}
          onLoadMore={onLoadMore}
          canLoadMore={canLoadMore}
          isLoadingMore={isLoadingMore}
        />
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

