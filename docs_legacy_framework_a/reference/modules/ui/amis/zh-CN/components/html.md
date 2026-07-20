> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Html
description:
type: 0
group: ⚙ 组件
menuName: Html
icon:
order: 49
---

## 基本用法

渲染一段 HTML

```schema
{
  "body": {
    "type": "html",
    "html": "支持 <code>Html</code>"
  }
}
```

> 当需要获取数据域中变量时，使用 [Tpl](./tpl) 。
