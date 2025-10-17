---
title: 表达式
description:
type: 0
group: 💡 概念
menuName: 表达式
icon:
order: 13
---

## 使用场景

awo-ui 中有很多场景会用到表达式。

- 模板中变量取值。 如：`my name is ${xxx}`
- api 地址参数取值 如 `http://mydomain.com/api/xxx?id=${id}`
- api 发送&接收数据映射

  ```
  {
    "type": "crud",
    "api": {
      method: "post"
      url: "http://mydomain.com/api/xxx",
      data: {
        skip: "${(page - 1) * perPage}",
        take: "${perPage}"
      }
    },
    ...
  }
  ```

- 组件显示与隐藏条件

  ```
  {
    "name": "xxxText",
    "type": "input-text",
    "visibleOn": "${ xxxFeature.on }"
  }
  ```

- 表单默认值

  ```
  {
    "name": "xxxText",
    "type": "input-text",
    "value": "${ TODAY() }"
  }
  ```

- 等等

awo-ui 中表达式有两种语法：

- 一种是纯 js 表达式，如 `data.xxx === 1`。
- 另一种是用 `${` 和 `}` 包裹的表达式。如：`${ xxx === 1}`。

```json
{
  "type": "tpl",
  "tpl": "当前作用域中变量 show 是 1 的时候才可以看得到我哦~",
  "visibleOn": "${show === 1}"
}
```

第一种是早期的版本，偷懒直接用的 [Javascript](https://developer.mozilla.org/zh-CN/docs/Web/JavaScript)，灵活性虽高，但是安全性欠佳。建议使用新版本规则，新规则跟 tpl 模板取值规则完全一样，不用来回切换语法。

## 表达式语法

表达式主要由三部分组成：开始字符、表达式内容和结束字符。其中开始字符固定是 `${` 结束字符固定是 `}`，中间内容才是表达式正文。这里说的语法主要还是表达式正文的语法。

规则主要包含：

- 变量名： `xxx变量`、`xxx变量.xxx属性`、`xxx变量[xxx属性]`
- 布尔值： `true` 或者 `false`
- null：`null`
- undefined: `undefined`
- 数字： `123` 或者 `123.23`
- 字符串： `"string"` 或者 `'string'`
- 字符模板

  ```
  `my name is ${name}`
  ```

- 数组： `[1, 2, 3]`
- 对象： `{a: 1, b: 2}`
- 组合使用如： `{a: 1: b: [1, 2, 3], [key]: yyy变量}`
- 三元表达式： `xx变量 == 1 ? 2 : 3`
- 二元表达式： `xx变量 && yy 变量` 、 `xx变量 || yy 变量`、 `xx变量 == 123`

  - `+` 相加
  - `-` 相减
  - `*` 乘
  - `/` 除
  - `**` pow 运算
  - `||` 或者
  - `&&` 并且
  - `|` 或运算
  - `^` 异或运算
  - `&` 与运算
  - `==` 等于比较
  - `!=` 不等于
  - `===` 恒等于
  - `!==` 不恒等于
  - `<` 小于
  - `<=` 小于或等于
  - `>` 大于
  - `>=` 大于或等于
  - `<<` 左移
  - `>>` 右移
  - `>>>` 带符号位的右移运算符

- 一元表达式： `!xx变量`、`~xx变量`

* `+` 一元加法
* `-` 一元减法
* `~` 否运算符，加 1 取反
* `!` 取反

- 函数调用：`SUM(1, 2, 3)`
- 箭头函数：`() => abc` 注意这个箭头函数只支持单表达式，不支持多条语句。主要配置其他函数使用如：`ARRAY_MAP(arr, item => item.abc)`
- 括号包裹，修改运算优先级：`(10 - 2) * 3`

示例：

```schema
{
  "type": "page",
  "data": {
    "a": 1,
    "key": "y",
    "obj": {
      "x": 2,
      "y": 3
    },
    "arr": [1, 2, 3]
  },
  "body": [
    "a is ${a} <br />",
    "a + 1 is ${a + 1} <br />",
    "obj.x is ${obj.x} <br />",
    "obj['x'] is ${obj['x']} <br />",
    "obj[key] is ${obj[key]} <br />",
    "arr[0] is ${arr[0]} <br />",
    "arr[a] is ${arr[a]} <br />",
    "arr[a + 1] is ${arr[a + 1]} <br />"
  ]
}
```

_特殊字符变量名_

> 1.6.1 及以上版本

默认变量名不支持特殊字符比如 `${ xxx.yyy }` 意思取 xxx 变量的 yyy 属性，如果变量名就是 `xxx.yyy` 怎么获取？这个时候需要用到转义语法，如：`${ xxx\.yyy }`

### 公式

除了支持简单表达式外，还集成了很多公式(函数)如 `${ AVG(1, 2, 3, 4)}`：

```schema
{
  "type": "page",
  "data": {
    "a": "",
    "语文成绩": 81
  },
  "body": [
    "1, 2, 3, 4 的平均数位 ${ AVG(1, 2, 3, 4)}",

    "当前成绩：${IF(语文成绩 > 80, '优秀', '继续努力')}"
  ]
}
```

## 函数调用示例

函数支持嵌套，参数支持常量及变量

```schema
{
  "type": "page",
  "body": [
    {
      "type": "form",
      "wrapWithPanel": false,
      "data": {
        "val": 3.5
      },
      "body": [
        {
          "type": "static",
          "label": "IF(true, 2, 3)",
          "tpl": "${IF(true, 2, 3)}"
        },
        {
          "type": "static",
          "label": "MAX(1, -1, 2, 3, 5, -9)",
          "tpl": "${MAX(1, -1, 2, 3, 5, -9)}"
        },
        {
          "type": "static",
          "label": "ROUND(3.5)",
          "tpl": "${ROUND(3.5)}"
        },
        {
          "type": "static",
          "label": "ROUND(val)",
          "tpl": "${ROUND(val)}"
        },
        {
          "type": "static",
          "label": "AVG(4, 6, 10, 10, 10)",
          "tpl": "${AVG(4, 6, 10, 10, 10)}"
        },
        {
          "type": "static",
          "label": "UPPERMONEY(7682.01)",
          "tpl": "${UPPERMONEY(7682.01)}"
        },
        {
          "type": "static",
          "label": "TIMESTAMP(DATE(2021, 11, 21, 0, 0, 0), 'x')",
          "tpl": "${TIMESTAMP(DATE(2021, 11, 21, 0, 0, 0), 'x')}"
        },
        {
          "type": "static",
          "label": "DATETOSTR(NOW(), 'YYYY-MM-DD')",
          "tpl": "${DATETOSTR(NOW(), 'YYYY-MM-DD')}"
        }
      ]
    }
  ]
}
```

下面是目前所支持函数的使用手册

!!!include(awo-ui-formula/lib/doc.md)!!!
