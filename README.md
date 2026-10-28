# MicroMD

超级轻量的 Markdown 阅读器。Windows 上是一个约 3 MB 的单文件 exe（原生窗口 + WebView2），
界面是一个零依赖的 HTML 文件，所有渲染都在本地完成，不联网。

## 下载使用

到 [Releases](../../releases/latest) 下载 `MicroMD-<版本>-win-x64.exe`，双击即用，无需安装。

- 系统要求：Windows 10 / 11（自带 WebView2 运行时）
- 启动即出窗口，内容随后载入
- 打开方式：
  - 双击 exe → 显示上次的草稿或欢迎页
  - 把 `.md` 拖到窗口里
  - 把 `.md` 拖到 exe 图标上，或命令行 `MicroMD.exe notes.md`
- 保存：`Ctrl+S`。用「命令行 / 拖到 exe」打开的文档会**直接写回原文件**，
  其他方式（拖进窗口、点「打开」）会导出下载

没有 WebView2 运行时（或想免安装免编译）时，直接用浏览器打开 `web/index.html`，
功能一致，只是保存走下载。

## 功能

- 实时分栏预览（左编辑 / 右预览），可拖动分隔条，双击复位
- 三种布局：分栏、纯预览、纯编辑
- 亮色 / 暗色主题，跟随系统，可手动切换
- 自动保存草稿（localStorage），关掉再开会接着上次
- 字数统计、预计阅读时长、双向滚动同步
- 大文件保护：超过 200 KB 时降低渲染频率，超长草稿不写入本地存储

## 快捷键

| 快捷键 | 功能 |
| --- | --- |
| `Ctrl+O` | 打开文件 |
| `Ctrl+S` | 保存 / 导出 |
| `Ctrl+B` | 粗体 |
| `Ctrl+I` | 斜体 |
| `Ctrl+K` | 插入链接 |
| `Tab` | 插入两个空格 |

## 支持的语法

标题、粗体 / 斜体 / 删除线 / 高亮、行内代码、围栏代码块、引用、有序与无序列表（可嵌套）、
任务列表、表格（含对齐）、分隔线、链接与图片、自动链接、反斜杠转义。

出于安全考虑，内嵌原始 HTML 会被转义显示；`javascript:` 链接会被丢弃。

## 项目结构

```
main.go                    原生外壳：Win32 窗口 + WebView2 + 原生打开/保存
web/index.html             全部界面与 Markdown 解析器（单文件、零依赖）
assets/icon.ico, icon.png  应用图标
tools/icon/main.go         图标生成器（矢量绘制，可重新生成）
micromd.bat                浏览器版启动器（无 WebView2 时使用）
.github/workflows/         自动打包发布
```

页面资源内嵌在 exe 里，首次运行解压到 `%LOCALAPPDATA%\MicroMD\`，
WebView2 的用户数据也在该目录下。

## 从源码构建

需要 Go 1.21+、Windows。

```powershell
go install github.com/tc-hib/go-winres@latest
go-winres simply --arch amd64 --manifest gui --icon assets/icon.ico `
  --product-name MicroMD --file-description "MicroMD Markdown Viewer" `
  --file-version 0.1.0 --product-version 0.1.0

go build -trimpath -buildvcs=false -ldflags "-H windowsgui -s -w" -o MicroMD.exe .
```

`go-winres` 生成的 `rsrc_windows_amd64.syso` 负责写入图标和版本信息（已加入 `.gitignore`）。
重新生成图标：`go run tools/icon/main.go`。

## 发布

`.github/workflows/release.yml` 在两种情况下运行：

- 推送 `v*` 标签
- 在 Actions 页面手动运行 `Release`（可填版本号，留空则自动生成 `v0.0.<run_number>`）

流程：构建 → 重写该 tag 指向当前提交 → 创建或更新 GitHub Release，上传
`MicroMD-<版本>-win-x64.exe` 及其 SHA256 校验文件。
