# Release Guide

## 发布前检查

```bash
gofmt -w .
go test -race ./...
go vet ./...
git diff --check
git status --short
```

确认工作区干净，并确认 `需求.md`、`.idea/` 等本地文件没有被跟踪。

## 发布步骤

1. 在 `CHANGELOG.md` 的 `Unreleased` 下整理本次变更。
2. 将内容移动到新版本标题，例如 `## [0.1.2] - YYYY-MM-DD`。
3. 提交版本变更：

   ```bash
   git add CHANGELOG.md README.md
   git commit -m "release v0.1.2"
   git push origin main
   ```

4. 创建并推送带说明的 annotated tag：

   ```bash
   git tag -a v0.1.2 -m "release v0.1.2"
   git push origin v0.1.2
   ```

5. 在 GitHub 创建同名 Release，粘贴对应版本的 Changelog 内容。
6. 验证模块索引：

   ```bash
   GOPROXY=https://proxy.golang.org \
     go list -m -json github.com/EmptyZeroRain/json2struct@v0.1.2
   ```

遵循 Semantic Versioning：破坏 API 用 major，新增兼容功能用 minor，兼容性修复用 patch。
