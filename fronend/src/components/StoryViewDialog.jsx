import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import './StoryViewDialog.css'

function StoryViewDialog({ isOpen, story, onClose }) {
  if (!isOpen || !story) return null

  return (
    <div className="story-view-dialog-overlay" onClick={onClose}>
      <div className="story-view-dialog" onClick={(e) => e.stopPropagation()}>
        <div className="story-view-dialog-header">
          <h3 className="story-view-dialog-title">{story.title || '未命名故事'}</h3>
          <div className="story-view-dialog-actions">
            <button
              className="story-view-dialog-close-btn"
              onClick={onClose}
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
        <div className="story-view-dialog-body">
          <div className="story-view-dialog-content">
            <ReactMarkdown remarkPlugins={[remarkGfm]}>
              {story.content || ''}
            </ReactMarkdown>
          </div>
        </div>
      </div>
    </div>
  )
}

export default StoryViewDialog

