---
title: SchemaNode
description:
type: 0
group: 🔧 类型
menuName: SchemaNode 配置节点
icon:
order: 19
---

SchemaNode 是指每一个 awo-ui 配置节点的类型，支持`模板`、`Schema（配置）`以及`SchemaArray（配置数组）`三种类型

## 模板

```json
{
  "type": "page",
  "data": {
    "text": "World"
  },
  "body": "Hello ${text}!" // 输出 Hello World!
}
```

上例中的`body`属性所配置的属性值`"Hello ${text}!"`就是一个模板

更多使用说明见 [模板文档](../concepts/template)

## Schema 配置

`Schema`，即**组件的 JSON 配置**

**它至少需要一个`type`字段，用以标识当前`Schema`的类型。**

```json
{
  "type": "page",
  "data": {
    "text": "World"
  },
  "body": {
    "type": "tpl",
    "tpl": "Hello ${text}!" // 输出 Hello World!
  }
}
```

`type`, `data`, `body`这三个字段组成的`JSON`对象，便是一个`Schema`，它由`type`字段作为标识，指明当前 `Schema` 是 `Page`组件节点

而通过查看 [Page 组件属性表](../../components/page) 可知，`body`属性类型是`SchemaNode`，即可以在`body`中，嵌套配置其他组件。我们在这里，用`type`和`tpl` JSON 对象，配置了 `Tpl` 组件，渲染了一段模板字符串。

> awo-ui 可以通过该方法，在`Schema`中嵌套配置其他`SchemaNode`，从而搭建非常复杂的页面应用。

### 配置显隐

所有的`Schema`类型都可以通过配置`visible`或`hidden`，`visibleOn`或`hiddenOn`来控制组件的显隐，下面是两种方式

##### 静态配置

通过配置`"hidden": true`或者`"visible": false`来隐藏组件

```schema: scope="body"
[
    {
        "type": "form",
        "body": [
            {
                "type": "input-text",
                "label": "姓名",
                "name": "name"
            }
        ]
    },
    {
        "type": "form",
        "hidden": true,
        "body": [
            {
                "type": "input-text",
                "label": "姓名",
                "name": "name"
            }
        ]
    }
]
```

下面那个表单被隐藏了。

##### 通过条件配置显隐

你也通过 [表达式](../concepts/expression) 配置`hiddenOn`，来实现在某个条件下禁用当前组件.

```schema: scope="body"
{
  "type": "form",
  "body": [
    {
      "type": "input-number",
      "label": "数量",
      "name": "number",
      "value": 0,
      "description": "调整数量大小查看效果吧！"
    },
    {
      "type": "input-text",
      "label": "文本",
      "name": "text",
      "hiddenOn": "this.number > 1",
      "description": "当数量大于1的时候，该文本框会隐藏"
    }
  ]
}
```

为了方便说明，我们在 form 中演示了条件显隐，实际上，只要当前数据域中数据有变化，都可以实现组件显隐

> `visible`和`hidden`，`visibleOn`和`hiddenOn`除了判断逻辑相反以外，没有任何区别

## SchemaArray 配置数组

明白了何为`Schema`之后，你就会很轻松理解`SchemaArray`，它其实就是支持通过数组配置`Schema`，从而在某一节点层级下，配置多个组件

```json
{
  "type": "page",
  "data": {
      "name": "awo-ui"
      "age": 1
  },
  "body": [
      {
        "type":"tpl",
        "tpl": "my name is ${name}" // 输出 my name is awo-ui
      },
      {
        "type":"tpl",
        "tpl": "I am ${age} years old" // 输出 I am 1 years old
      }
  ]
}
```

非常容易看出来，我们给`body`属性，配置了数组结构的`Schema`，从而实现在`body`下，渲染两个`tpl`，来输入两段文字的效果
