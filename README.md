# grandma 讲故事的地方

grandma是一个前后端分离的AI故事创作APP，用户可以与AI合作，提供思路，和AI商讨故事大纲，让AI为他写下完整的故事

项目整体采用前后端分离开发
后端使用golang作为主要开发语言
前端使用React + Vite开发

## v0.1 Features

### 前端：
- ✅ 类似GPT的页面UI风格，以打字机样式展示后端发送的大模型响应
- ✅ 可以自由选择多种模型，并将用户输入发送给后端
- ✅ 流式接收AI响应，实时显示
- ✅ 用户说的话在右侧，大模型返回的内容在左侧，并用色块区分
- ✅ 前端在请求时仅携带本次用户请求的内容
- ✅ 简化的颜色方案：仅三种颜色（背景、用户消息框、AI消息框）
- ✅ AI响应时可输入但不能发送
- ✅ Markdown渲染：支持大模型返回的Markdown内容渲染（代码块、表格、列表等）

### 后端：
- ✅ 接入多种大模型API（OpenAI、Anthropic）
- ✅ 接收用户发送，调用大模型API
- ✅ web框架使用gin进行开发
- ✅ 将大模型的返回以流式接口返回给前端
- 针对文本场景，做出如下处理：
    - 用户的每个对话有独立的**对话ID**
    - 对话中用户的每一段提问，用户的每一段发言，都有*独立的***文档ID**
-   采用微服务架构，后端拆分为如下模块：
    - 对话模块：接收用户请求，将用户请求传递给大模型，并用流式输出返回大模型响应
    - 对话列表模块：接收用户请求，返回用户的历史对话列表
    - 文档管理模块
    - 对话管理模块

## v0.2 Features

### 后端：
- 接入数据库，支持对话信息的存储和文档的存储

## 快速开始

### 后端

1. 进入后端目录
```bash
cd backend
```

2. 安装依赖
```bash
go mod download
```

3. 配置环境变量
创建 `.env` 文件，配置API密钥：
```env
OPENAI_API_KEY=your_openai_api_key_here
ANTHROPIC_API_KEY=your_anthropic_api_key_here
PORT=8080
```

4. 运行服务器
```bash
go run main.go
```

后端将在 `http://localhost:8080` 启动

### 前端

1. 进入前端目录
```bash
cd fronend
```

2. 安装依赖
```bash
npm install
```

3. 启动开发服务器
```bash
npm run dev
```

前端将在 `http://localhost:3000` 启动

## 项目结构

```
grandma/
├── backend/          # 后端代码（Golang + Gin）
│   ├── config/       # 配置管理
│   ├── handlers/     # 请求处理
│   ├── models/       # 数据模型
│   ├── services/     # 业务逻辑（大模型API接入）
│   └── main.go       # 入口文件
├── fronend/          # 前端代码（React + Vite）
│   ├── src/
│   │   ├── components/   # React组件
│   │   ├── App.jsx       # 主应用
│   │   └── main.jsx      # 入口文件
│   └── package.json
└── README.md
```

## 使用说明

1. 确保后端和前端都已启动
2. 在浏览器访问 `http://localhost:3000`
3. 在侧边栏选择要使用的大模型
4. 在输入框中输入你的想法或问题
5. 按 Enter 或点击发送按钮
6. AI的响应将以打字机效果逐字显示

## 支持的模型

- **OpenAI**: GPT-3.5 Turbo
- **Anthropic**: Claude 3.5 Sonnet

## 开发计划

- [ ] 历史对话记录
- [ ] 故事大纲生成
- [ ] 故事保存和导出
- [ ] 更多大模型支持
