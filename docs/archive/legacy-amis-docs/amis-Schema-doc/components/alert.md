---
title: Alert
description:
type: 0
group: ⚙ Components
menuName: Alert
icon:
order: 27
---

Used for special text prompts, divided into four categories: info, success, warning, and danger.

## Basic Usage

The `level` attribute supports 4 preset styles: `info`, `success`, `warning`, `danger`.

```schema: scope="body"
[
  {
    "type": "alert",
    "body": "Info message",
    "level": "info",
    "className": "mb-1"
  },
  {
    "type": "alert",
    "title": "Info Title",
    "body": "Info message",
    "level": "info",
    "className": "mb-3"
  },
  {
    "type": "alert",
    "body": "Success message",
    "level": "success",
    "className": "mb-1"
  },
  {
    "type": "alert",
    "title": "Success Title",
    "body": "Success message",
    "level": "success",
    "className": "mb-3"
  },
  {
    "type": "alert",
    "body": "Warning message",
    "level": "warning",
    "className": "mb-1"
  },
  {
    "type": "alert",
    "title": "Warning Title",
    "body": "Warning message",
    "level": "warning",
    "className": "mb-3"
  },
  {
    "type": "alert",
    "body": "Danger message",
    "level": "danger",
    "className": "mb-1"
  },
  {
    "type": "alert",
    "title": "Danger Title",
    "body": "Danger message",
    "level": "danger",
  },
]
```

## Icon

After configuring `"showIcon": true`, icons are displayed to make information more prominent. You can customize icon content through the `icon` attribute. If the `icon` attribute is empty, a default icon will be added based on the `level` value.

```schema: scope="body"
[
  {
    "type": "alert",
    "body": "Info message",
    "level": "info",
    "showIcon": true,
    "className": "mb-1"
  },
  {
    "type": "alert",
    "title": "Info Title",
    "body": "Info message",
    "level": "info",
    "showIcon": true,
    "className": "mb-3"
  },
  {
    "type": "alert",
    "body": "Success message",
    "level": "success",
    "showIcon": true,
    "className": "mb-1"
  },
  {
    "type": "alert",
    "title": "Success Title",
    "body": "Success message",
    "level": "success",
    "showIcon": true,
    "className": "mb-3"
  },
  {
    "type": "alert",
    "body": "Warning message",
    "level": "warning",
    "showIcon": true,
    "className": "mb-1"
  },
  {
    "type": "alert",
    "title": "Warning Title",
    "body": "Warning message",
    "level": "warning",
    "showIcon": true,
    "className": "mb-3"
  },
  {
    "type": "alert",
    "body": "Danger message",
    "level": "danger",
    "showIcon": true,
    "className": "mb-1"
  },
  {
    "type": "alert",
    "title": "Danger Title",
    "body": "Danger message",
    "level": "danger",
    "showIcon": true,
    "className": "mb-3"
  },
  {
    "type": "alert",
    "body": "自定义ICON文案",
    "showIcon": true,
    "icon": "info-circle",
    "className": "mb-1"
  },
  {
    "type": "alert",
    "title": "自定义ICON标题",
    "body": "自定义ICON文案",
    "showIcon": true,
    "icon": "fa fa-list"
  }
]
```

## level 支持表达式

> 1.6.4 及以上版本

修改下面例子的 status 值为 2 就能看到变化

```schema:
{
  "type": "page",
  "data": {
    "status": 1
  },
  "body": [
    {
      "type": "alert",
      "level": "${IFS(status===1, 'danger', status===2, 'warning')}",
      "body": "这是内容区"
    }
  ]
}
```

同时 icon 和 showIcon 也都支持表达式

## 显示关闭按钮

配置`"showCloseButton": true`实现显示关闭按钮。

```schema: scope="body"
[
  {
    "type": "alert",
    "body": "显示关闭按钮的提示",
    "level": "info",
    "showCloseButton": true,
    "showIcon": true,
    "className": "mb-2"
  },
  {
    "type": "alert",
    "title": "可关闭提示",
    "body": "显示关闭按钮的提示",
    "level": "success",
    "showCloseButton": true,
    "showIcon": true
  }
]
```

## 属性表

| 属性名               | 类型                                      | 默认值    | 说明                                                     |
| -------------------- | ----------------------------------------- | --------- | -------------------------------------------------------- |
| type                 | `string`                                  | `"alert"` | 指定为 alert 渲染器                                      |
| className            | `string`                                  |           | 外层 Dom 的类名                                          |
| level                | `string`                                  | `info`    | 级别，可以是：`info`、`success`、`warning` 或者 `danger` |
| body                 | [SchemaNode](../../docs/types/schemanode) |           | 显示内容                                                 |
| showCloseButton      | `boolean`                                 | `false`   | 是否显示关闭按钮                                         |
| closeButtonClassName | `string`                                  |           | 关闭按钮的 CSS 类名                                      |
| showIcon             | `boolean`                                 | `false`   | 是否显示 icon                                            |
| icon                 | `string`                                  |           | 自定义 icon                                              |
| iconClassName        | `string`                                  |           | icon 的 CSS 类名                                         |
