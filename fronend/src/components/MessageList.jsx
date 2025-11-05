import Message from './Message'
import './MessageList.css'

function MessageList({ messages, isLoading, enableTypewriter = true, onLoadMore, canLoadMore = false, isLoadingMore = false }) {
  if (messages.length === 0) {
    return null
  }

  return (
    <div className="message-list">
      {canLoadMore && (
        <div className="load-more-container">
          <button
            className="load-more-button"
            onClick={onLoadMore}
            disabled={isLoadingMore}
          >
            {isLoadingMore ? '加载中...' : '加载更早的消息'}
          </button>
        </div>
      )}
      {messages.map((message) => (
        <Message key={message.id} message={message} enableTypewriter={enableTypewriter} />
      ))}
      {isLoading && messages[messages.length - 1]?.role === 'assistant' && (
        <div className="typing-indicator">
          <span></span>
          <span></span>
          <span></span>
        </div>
      )}
    </div>
  )
}

export default MessageList

