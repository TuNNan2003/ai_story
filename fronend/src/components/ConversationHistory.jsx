import { useState } from 'react'
import './ConversationHistory.css'

function ConversationHistory({ conversations, currentConversationId, onSelectConversation, isCollapsed, onToggleCollapse, onRenameConversation }) {
  const [editingId, setEditingId] = useState(null)
  const [editTitle, setEditTitle] = useState('')

  const handleRenameClick = (e, conversation) => {
    e.stopPropagation()
    setEditingId(conversation.id)
    setEditTitle(conversation.title)
  }

  const handleRenameSubmit = async (e, conversationId) => {
    e.stopPropagation()
    if (editTitle.trim() && editTitle.trim() !== conversations.find(c => c.id === conversationId)?.title) {
      await onRenameConversation(conversationId, editTitle.trim())
    }
    setEditingId(null)
    setEditTitle('')
  }

  const handleRenameCancel = (e) => {
    e.stopPropagation()
    setEditingId(null)
    setEditTitle('')
  }

  const handleKeyDown = (e, conversationId) => {
    if (e.key === 'Enter') {
      handleRenameSubmit(e, conversationId)
    } else if (e.key === 'Escape') {
      handleRenameCancel(e)
    }
  }

  return (
    <>
      <div className={`conversation-history ${isCollapsed ? 'collapsed' : ''}`}>
        <div className="conversation-history-header" onClick={onToggleCollapse}>
          <span className="conversation-history-title">历史对话</span>
          <svg
            className={`collapse-icon ${isCollapsed ? 'collapsed' : ''}`}
            width="16"
            height="16"
            viewBox="0 0 16 16"
            fill="none"
          >
            <path
              d="M4 6l4 4 4-4"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
        </div>
        {!isCollapsed && (
          <div className="conversation-history-list">
            {conversations.length === 0 ? (
              <div className="conversation-history-empty">暂无历史对话</div>
            ) : (
              conversations.map((conv) => (
                <div
                  key={conv.id}
                  className={`conversation-item ${
                    currentConversationId === conv.id ? 'active' : ''
                  }`}
                  onClick={() => onSelectConversation(conv.id)}
                >
                  {editingId === conv.id ? (
                    <div className="conversation-item-edit" onClick={(e) => e.stopPropagation()}>
                      <input
                        type="text"
                        value={editTitle}
                        onChange={(e) => setEditTitle(e.target.value)}
                        onKeyDown={(e) => handleKeyDown(e, conv.id)}
                        onBlur={(e) => handleRenameSubmit(e, conv.id)}
                        className="conversation-item-edit-input"
                        autoFocus
                      />
                    </div>
                  ) : (
                    <>
                      <div className="conversation-item-content">
                        <div className="conversation-item-title">{conv.title}</div>
                        <div className="conversation-item-time">
                          {new Date(conv.updated_at).toLocaleDateString('zh-CN')}
                        </div>
                      </div>
                      <button
                        className="conversation-item-rename-btn"
                        onClick={(e) => handleRenameClick(e, conv)}
                        title="重命名"
                      >
                        <svg width="14" height="14" viewBox="0 0 16 16" fill="none">
                          <path
                            d="M11.333 2.667a1.414 1.414 0 0 1 2 2L5.333 12l-2.666.667L3.333 10l8-8z"
                            stroke="currentColor"
                            strokeWidth="1.5"
                            strokeLinecap="round"
                            strokeLinejoin="round"
                          />
                        </svg>
                      </button>
                    </>
                  )}
                </div>
              ))
            )}
          </div>
        )}
      </div>
      {isCollapsed && (
        <div className="conversation-history-sidebar" onClick={onToggleCollapse}>
          <svg
            width="20"
            height="20"
            viewBox="0 0 16 16"
            fill="none"
            className="sidebar-expand-icon"
          >
            <path
              d="M6 4l4 4-4 4"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
        </div>
      )}
    </>
  )
}

export default ConversationHistory

