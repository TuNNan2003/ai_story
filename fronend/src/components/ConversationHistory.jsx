import { useState, useEffect, useRef } from 'react'
import DeleteConfirmDialog from './DeleteConfirmDialog'
import './ConversationHistory.css'

function ConversationHistory({ conversations, currentConversationId, onSelectConversation, isCollapsed, onToggleCollapse, onRenameConversation, onDeleteConversation, width, onWidthChange, stories, onSelectStory, onEditStory, onDeleteStory }) {
  const [editingId, setEditingId] = useState(null)
  const [editTitle, setEditTitle] = useState('')
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [deleteTargetId, setDeleteTargetId] = useState(null)
  const [deleteTargetTitle, setDeleteTargetTitle] = useState('')
  const [isResizing, setIsResizing] = useState(false)
  const resizeRef = useRef(null)

  useEffect(() => {
    const handleMouseMove = (e) => {
      if (isResizing && resizeRef.current) {
        const newWidth = e.clientX
        if (newWidth >= 200 && newWidth <= 500) {
          onWidthChange(newWidth)
        }
      }
    }

    const handleMouseUp = () => {
      setIsResizing(false)
    }

    if (isResizing) {
      document.addEventListener('mousemove', handleMouseMove)
      document.addEventListener('mouseup', handleMouseUp)
      document.body.style.cursor = 'col-resize'
      document.body.style.userSelect = 'none'
    }

    return () => {
      document.removeEventListener('mousemove', handleMouseMove)
      document.removeEventListener('mouseup', handleMouseUp)
      document.body.style.cursor = ''
      document.body.style.userSelect = ''
    }
  }, [isResizing, onWidthChange])

  const handleResizeStart = (e) => {
    e.preventDefault()
    setIsResizing(true)
  }

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
      <div
        className={`conversation-history ${isCollapsed ? 'collapsed' : ''}`}
        style={!isCollapsed && width ? { width: `${width}px` } : {}}
      >
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
          <>
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
                          <div className="conversation-item-title" title={conv.title}>{conv.title}</div>
                          <div className="conversation-item-time">
                            {new Date(conv.updated_at).toLocaleDateString('zh-CN')}
                          </div>
                        </div>
                        <div className="conversation-item-actions">
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
                          <button
                            className="conversation-item-delete-btn"
                            onClick={(e) => {
                              e.stopPropagation()
                              // 检查localStorage中是否设置了不再提醒
                              const dontShowAgain = localStorage.getItem('dontShowDeleteConfirm') === 'true'
                              if (dontShowAgain) {
                                onDeleteConversation(conv.id)
                              } else {
                                setDeleteTargetId(conv.id)
                                setDeleteTargetTitle(conv.title)
                                setDeleteDialogOpen(true)
                              }
                            }}
                            title="删除"
                          >
                            <svg width="14" height="14" viewBox="0 0 16 16" fill="none">
                              <path
                                d="M4 4l8 8M12 4l-8 8"
                                stroke="currentColor"
                                strokeWidth="1.5"
                                strokeLinecap="round"
                                strokeLinejoin="round"
                              />
                            </svg>
                          </button>
                        </div>
                      </>
                    )}
                  </div>
                ))
              )}
            </div>
            <div className="story-history-section">
              <div className="story-history-header">
                <span className="story-history-title">我的故事</span>
              </div>
              <div className="story-history-list">
                {!stories || stories.length === 0 ? (
                  <div className="story-history-empty">暂无故事</div>
                ) : (
                  stories.map((story) => (
                    <div
                      key={story.id}
                      className="story-item"
                    >
                      <div 
                        className="story-item-content"
                        onClick={() => onSelectStory && onSelectStory(story)}
                      >
                        <div className="story-item-title" title={story.title}>{story.title || '未命名故事'}</div>
                        <div className="story-item-time">
                          {new Date(story.created_at).toLocaleDateString('zh-CN')}
                        </div>
                      </div>
                      <div className="story-item-actions">
                        <button
                          className="story-item-edit-btn"
                          onClick={(e) => {
                            e.stopPropagation()
                            onEditStory && onEditStory(story)
                          }}
                          title="修改"
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
                        <button
                          className="story-item-delete-btn"
                          onClick={(e) => {
                            e.stopPropagation()
                            onDeleteStory && onDeleteStory(story)
                          }}
                          title="删除"
                        >
                          <svg width="14" height="14" viewBox="0 0 16 16" fill="none">
                            <path
                              d="M4 4l8 8M12 4l-8 8"
                              stroke="currentColor"
                              strokeWidth="1.5"
                              strokeLinecap="round"
                              strokeLinejoin="round"
                            />
                          </svg>
                        </button>
                      </div>
                    </div>
                  ))
                )}
              </div>
            </div>
          </>
        )}
        {!isCollapsed && (
          <div
            ref={resizeRef}
            className="conversation-history-resizer"
            onMouseDown={handleResizeStart}
          />
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
      <DeleteConfirmDialog
        isOpen={deleteDialogOpen}
        onConfirm={() => {
          if (deleteTargetId) {
            onDeleteConversation(deleteTargetId)
          }
          setDeleteDialogOpen(false)
          setDeleteTargetId(null)
          setDeleteTargetTitle('')
        }}
        onCancel={() => {
          setDeleteDialogOpen(false)
          setDeleteTargetId(null)
          setDeleteTargetTitle('')
        }}
        conversationTitle={deleteTargetTitle}
      />
    </>
  )
}

export default ConversationHistory

