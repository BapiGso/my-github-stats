# GitHub Stats Decorator - Golang 版本（东方主题右侧角色）

一个基于 github-readme-stats 的增强版 GitHub 卡片生成器：左侧是原始的统计卡片或语言卡片，右侧拼接你在本仓库 assets 目录下自绘的东方角色 SVG。

- 左侧来源：
  - Stats: https://github-readme-stats.vercel.app/api
  - Top Langs: https://github-readme-stats.vercel.app/api/top-langs/
- 右侧来源：
  - 本项目的 assets 目录中的 SVG（通过 GET 参数 role 指定文件名）

## 特性

- 📊 统计卡片（Stats）与 🌐 语言卡片（Top Languages）
- 🖼️ 右侧角色插图来自本地 assets/*.svg，按需替换
- ⚙️ 通过查询参数自定义（可选）：svg_width、svg_height、role_x、role_y
- 🚀 服务端使用 Golang 实现，仍兼容 Vercel 部署

## 目录结构

```
assets/
  sakuya.svg          # 示例角色 SVG（你可以放自己的文件）
api/
  stats.go            # /api/stats 端点（Golang）
  top-langs.go        # /api/top-langs 端点（Golang）
  stats.js/top-langs.js（旧版 Node，仅保留兼容，后续可删除）
```

## 使用方式

- Stats 卡片：

```
![GitHub Stats](https://your-domain.vercel.app/api/stats?username=YOUR_USERNAME&role=sakuya)
```

- Top Languages 卡片：

```
![Top Languages](https://your-domain.vercel.app/api/top-langs?username=YOUR_USERNAME&role=sakuya)
```

参数说明：
- username: 必填，GitHub 用户名
- role: 可选，assets 下的文件名（不含 .svg 后缀），例如 role=sakuya 即使用 assets/sakuya.svg
- svg_width / svg_height: 可选，外层组合 SVG 的宽高（默认 stats: 500x200，top-langs: 500x170）
- role_x / role_y: 可选，右侧角色在组合图中的偏移位置，默认 x=360, y=0

注意：传入的其他查询参数会原样透传给 github-readme-stats（例如 theme, hide_border, card_width 等）。

## 本地开发

- 直接运行 Go 服务器（仅用于本地调试）：

```
GO_LOCAL_SERVER=1 go run ./api/*.go
# 打开 http://localhost:3000/api/stats?username=torvalds&role=sakuya
```

## 致谢

- github-readme-stats - 提供基础统计 SVG
- Vercel - Serverless 部署
- 東方Project - 上海爱丽丝幻乐团 - 角色灵感来源

## LICENSE

MIT
