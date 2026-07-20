> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: PDF Viewer
description:
type: 0
group: ⚙ 组件
menuName: PDFViewer 渲染
icon:
order: 24
---

## 基本用法

```schema: scope="body"
{
  "type": "pdf-viewer",
  "id": "pdf-viewer",
  "src": "/examples/static/simple.pdf",
  "width": 500
}
```

## 配合文件上传实现预览功能

配置和 `input-file` 相同的 `name` 即可

```schema: scope="body"
{
  "type": "form",
  "title": "",
  "wrapWithPanel": false,
  "body": [
    {
      "type": "input-file",
      "name": "file",
      "label": "File",
      "asBlob": true,
      "accept": ".pdf"
    },
    {
      "type": "pdf-viewer",
      "id": "pdf-viewer",
      "name": "file",
      "width": 500
    }
  ]
}
```

## 属性表

| 属性名     | 类型   | 默认值 | 说明       |
| ---------- | ------ | ------ | ---------- |
| src        | Api    |        | 文档地址   |
| width      | number |        | 宽度       |
| height     | number | -      | 高度       |
| background | string | #fff   | PDF 背景色 |
