import { useState, useEffect, useRef } from 'react'
import DeleteConfirmDialog from './DeleteConfirmDialog'
import './WorkHistory.css'

function WorkHistory({ works, currentWorkId, onSelectWork, isCollapsed, onToggleCollapse, onRenameWork, onDeleteWork, width, onWidthChange, onNewWork }) {
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

  const handleRenameClick = (e, work) => {
    e.stopPropagation()
    setEditingId(work.id)
    setEditTitle(work.title)
  }

  const handleRenameSubmit = async (e, workId) => {
    e.stopPropagation()
    if (editTitle.trim() && editTitle.trim() !== works.find(w => w.id === workId)?.title) {
      await onRenameWork(workId, editTitle.trim())
    }
    setEditingId(null)
    setEditTitle('')
  }

  const handleRenameCancel = (e) => {
    e.stopPropagation()
    setEditingId(null)
    setEditTitle('')
  }

  const handleKeyDown = (e, workId) => {
    if (e.key === 'Enter') {
      handleRenameSubmit(e, workId)
    } else if (e.key === 'Escape') {
      handleRenameCancel(e)
    }
  }

  return (
    <>
      <div
        className={`work-history ${isCollapsed ? 'collapsed' : ''}`}
        style={!isCollapsed && width ? { width: `${width}px` } : {}}
      >
        <div className="work-history-header" onClick={onToggleCollapse}>
          <span className="work-history-title">我的创作</span>
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
            <div className="work-history-list">
              <button onClick={onNewWork} className="new-work-button">
                <svg width="14" height="14" viewBox="0 0 16 16" fill="none">
                  <path
                    d="M8 1v14M1 8h14"
                    stroke="currentColor"
                    strokeWidth="2"
                    strokeLinecap="round"
                  />
                </svg>
                新建创作
              </button>
              {works.length === 0 ? (
                <div className="work-history-empty">暂无创作</div>
              ) : (
                works.map((work) => (
                  <div
                    key={work.id}
                    className={`work-item ${
                      currentWorkId === work.id ? 'active' : ''
                    }`}
                    onClick={() => onSelectWork(work.id)}
                  >
                    {editingId === work.id ? (
                      <div className="work-item-edit" onClick={(e) => e.stopPropagation()}>
                        <input
                          type="text"
                          value={editTitle}
                          onChange={(e) => setEditTitle(e.target.value)}
                          onKeyDown={(e) => handleKeyDown(e, work.id)}
                          onBlur={(e) => handleRenameSubmit(e, work.id)}
                          className="work-item-edit-input"
                          autoFocus
                        />
                      </div>
                    ) : (
                      <>
                        <div className="work-item-content">
                          <div className="work-item-title" title={work.title}>{work.title || '未命名创作'}</div>
                          <div className="work-item-time">
                            {new Date(work.updated_at).toLocaleDateString('zh-CN')}
                          </div>
                        </div>
                        <div className="work-item-actions">
                          <button
                            className="work-item-rename-btn"
                            onClick={(e) => handleRenameClick(e, work)}
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
                            className="work-item-delete-btn"
                            onClick={(e) => {
                              e.stopPropagation()
                              const dontShowAgain = localStorage.getItem('dontShowDeleteConfirm') === 'true'
                              if (dontShowAgain) {
                                onDeleteWork(work.id)
                              } else {
                                setDeleteTargetId(work.id)
                                setDeleteTargetTitle(work.title)
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
          </>
        )}
        {!isCollapsed && (
          <div
            ref={resizeRef}
            className="work-history-resizer"
            onMouseDown={handleResizeStart}
          />
        )}
      </div>
      {isCollapsed && (
        <div className="work-history-sidebar" onClick={onToggleCollapse}>
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
            onDeleteWork(deleteTargetId)
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

export default WorkHistory

