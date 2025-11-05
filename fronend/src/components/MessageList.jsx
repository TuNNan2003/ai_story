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
          {isLoadingMore ? (
            <div className="load-more-indicator">
              <div className="load-more-spinner">
                <span></span>
                <span></span>
                <span></span>
              </div>
              <span className="load-more-text">加载中...</span>
            </div>
          ) : (
            <div className="load-more-hint">向上滚动加载更多</div>
          )}
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

