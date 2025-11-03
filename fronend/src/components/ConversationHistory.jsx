import { useState, useEffect } from 'react'
import './ConversationHistory.css'

function ConversationHistory({ conversations, currentConversationId, onSelectConversation, isCollapsed, onToggleCollapse }) {
  return (
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
                <div className="conversation-item-title">{conv.title}</div>
                <div className="conversation-item-time">
                  {new Date(conv.updated_at).toLocaleDateString('zh-CN')}
                </div>
              </div>
            ))
          )}
        </div>
      )}
    </div>
  )
}

export default ConversationHistory

