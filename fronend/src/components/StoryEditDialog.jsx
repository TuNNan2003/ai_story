import { useState, useEffect } from 'react'
import './StoryEditDialog.css'

function StoryEditDialog({ isOpen, content, title: initialTitle, onClose, onSave, storyId }) {
  const [editedContent, setEditedContent] = useState(content || '')
  const [editedTitle, setEditedTitle] = useState(initialTitle || '')
  const [isSaving, setIsSaving] = useState(false)
  const [error, setError] = useState(null)

  // 当对话框打开时，更新内容
  useEffect(() => {
    if (isOpen) {
      setEditedContent(content || '')
      setEditedTitle(initialTitle || '')
      setError(null)
    }
  }, [isOpen, content, initialTitle])

  const handleSave = async () => {
    if (!editedContent.trim()) {
      setError('内容不能为空')
      return
    }

    setIsSaving(true)
    setError(null)

    try {
      await onSave(editedContent, editedTitle)
      // 保存成功后，关闭对话框
      onClose()
    } catch (err) {
      setError(err.message || '保存失败')
    } finally {
      setIsSaving(false)
    }
  }

  const handleClose = () => {
    setError(null)
    onClose()
  }

  if (!isOpen) return null

  return (
    <div className="story-edit-dialog-overlay" onClick={handleClose}>
      <div className="story-edit-dialog" onClick={(e) => e.stopPropagation()}>
        <div className="story-edit-dialog-header">
          <h3 className="story-edit-dialog-title">{storyId ? '编辑故事' : '添加为故事'}</h3>
          <div className="story-edit-dialog-actions">
            <button
              className="story-edit-dialog-close-btn"
              onClick={handleClose}
              title="关闭"
            >
              <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
                <path
                  d="M12 4l-8 8M4 4l8 8"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                />
              </svg>
            </button>
          </div>
        </div>
        <div className="story-edit-dialog-body">
          <div className="story-edit-dialog-field">
            <label className="story-edit-dialog-label">标题</label>
            <input
              type="text"
              className="story-edit-dialog-input"
              value={editedTitle}
              onChange={(e) => setEditedTitle(e.target.value)}
              placeholder="请输入故事标题"
            />
          </div>
          <div className="story-edit-dialog-field">
            <label className="story-edit-dialog-label">内容</label>
            <textarea
              className="story-edit-dialog-textarea"
              value={editedContent}
              onChange={(e) => setEditedContent(e.target.value)}
              placeholder="请输入故事内容"
              rows={25}
            />
          </div>
          {error && (
            <div className="story-edit-dialog-error">
              {error}
            </div>
          )}
        </div>
        <div className="story-edit-dialog-footer">
          <button
            className="story-edit-dialog-cancel-btn"
            onClick={handleClose}
            disabled={isSaving}
          >
            取消
          </button>
          <button
            className="story-edit-dialog-save-btn"
            onClick={handleSave}
            disabled={isSaving || !editedContent.trim()}
          >
            {isSaving ? '保存中...' : '保存'}
          </button>
        </div>
      </div>
    </div>
  )
}

export default StoryEditDialog

