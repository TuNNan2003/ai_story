import { useState, useEffect, useRef } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import './MarkdownRenderer.css'

function MarkdownRenderer({ text, speed = 20, enableTypewriter = true }) {
  const [displayedText, setDisplayedText] = useState('')
  const prevTextRef = useRef('')
  const currentIndexRef = useRef(0)
  const timeoutIdRef = useRef(null)

  useEffect(() => {
    if (!text) {
      setDisplayedText('')
      prevTextRef.current = ''
      currentIndexRef.current = 0
      if (timeoutIdRef.current) {
        clearTimeout(timeoutIdRef.current)
        timeoutIdRef.current = null
      }
      return
    }

    // 如果禁用打字机效果，直接显示完整文本
    if (!enableTypewriter) {
      setDisplayedText(text)
      prevTextRef.current = text
      currentIndexRef.current = text.length
      if (timeoutIdRef.current) {
        clearTimeout(timeoutIdRef.current)
        timeoutIdRef.current = null
      }
      return
    }

    // 如果文本是之前文本的延续（向后追加），继续打字新部分
    if (text.startsWith(prevTextRef.current)) {
      // 文本是追加的，保持已显示的部分，继续打字新部分
      const newTextStart = prevTextRef.current.length
      if (currentIndexRef.current < newTextStart) {
        // 确保当前索引至少等于已有文本长度
        currentIndexRef.current = newTextStart
        setDisplayedText(text.slice(0, currentIndexRef.current))
      }
    } else {
      // 新文本，重新开始
      setDisplayedText('')
      currentIndexRef.current = 0
      prevTextRef.current = ''
      if (timeoutIdRef.current) {
        clearTimeout(timeoutIdRef.current)
        timeoutIdRef.current = null
      }
    }

    // 继续打字
    const type = () => {
      if (currentIndexRef.current < text.length) {
        setDisplayedText(text.slice(0, currentIndexRef.current + 1))
        currentIndexRef.current++
        timeoutIdRef.current = setTimeout(type, speed)
      } else {
        prevTextRef.current = text
        timeoutIdRef.current = null
      }
    }

    // 如果还有未显示的文本，开始打字
    if (currentIndexRef.current < text.length && !timeoutIdRef.current) {
      type()
    } else if (currentIndexRef.current >= text.length) {
      // 文本已完整显示
      setDisplayedText(text)
      prevTextRef.current = text
    }

    return () => {
      if (timeoutIdRef.current) {
        clearTimeout(timeoutIdRef.current)
        timeoutIdRef.current = null
      }
    }
  }, [text, speed, enableTypewriter])

  return (
    <div className="markdown-content">
      <ReactMarkdown remarkPlugins={[remarkGfm]}>
        {displayedText}
      </ReactMarkdown>
    </div>
  )
}

export default MarkdownRenderer

