---
title: 事件动作
description: 事件动作
type: 0
group: 💡 概念
menuName: 事件动作
icon:
order: 9
---

> 1.7.0 及以上版本

上一节我们介绍了如何实现联动，是比较基础和局限的，而事件动作是更简单、更灵活、更高级的用法，可以解决复杂的 UI 交互场景，支持渲染器事件监听和响应设计，无需关心组件层级关系。例如：

- http 请求：发送 http 请求
- 弹窗提示：执行弹窗、抽屉打开和 toast 提示
- 页面跳转：页面链接跳转
- 浏览器相关：回退、前进、后退、刷新
- 刷新组件：联动刷新表单数据，即数据重新加载
- 组件状态：控制指定组件的显示/隐藏、启用/禁用、展示态/编辑态
- 组件特性动作：执行指定组件的专有动作，例如执行表单的提交动作
- 组件数据：更新指定组件的数据域
- 广播：多个组件监听同一个事件做出不同响应
- JS 脚本：通过编写 JS 代码片段实现所需逻辑，同时支持 JS 代码内执行动作
- 逻辑编排：条件、循环、排他、并行

## 基本使用

### onEvent

通过`onEvent`属性实现渲染器事件与响应动作的绑定。`onEvent`内配置事件和动作映射关系，`actions`是事件对应的响应动作的集合。

```json
{
  "type": "button",
  "label": "尝试点击、鼠标移入/移出",
  "level": "primary",
  "onEvent": {
    "click": { // 监听点击事件
      "actions": [ // 执行的动作列表
        {
          "actionType": "toast", // 执行toast提示动作
          "args": { // 动作参数
            "msgType": "info",
            "msg": "派发点击事件"
          }
        }
      ]
    },
    "mouseenter": {{ // 监听鼠标移入事件
      "actions": [
        {
          "actionType": "toast",
          "args": {
            "msgType": "info",
            "msg": "派发鼠标移入事件"
          }
        }
      ]
    },
    "mouseleave": {{ // 监听鼠标移出事件
      "actions": [
        {
          "actionType": "toast",
          "args": {
            "msgType": "info",
            "msg": "派发鼠标移出事件"
          }
        }
      ]
    }
  }
}
```

```schema
{
  type: 'page',
  body: [
    {
      type: 'button',
      label: '尝试点击、鼠标移入/移出',
      level: 'primary',
      onEvent: {
        click: {
          actions: [
            {
              actionType: 'toast',
              args: {
                msgType: 'info',
                msg: '派发点击事件'
              }
            }
          ]
        },
        mouseenter: {
          actions: [
            {
              actionType: 'toast',
              args: {
                msgType: 'info',
                msg: '派发鼠标移入事件'
              }
            }
          ]
        },
        mouseleave: {
          actions: [
            {
              actionType: 'toast',
              args: {
                msgType: 'info',
                msg: '派发鼠标移出事件'
              }
            }
          ]
        }
      }
    }
  ]
}
```

### 上下文

执行动作时，可以通过`${event.data}`获取事件对象的数据、通过`${__rendererData}`获取组件当前数据域，例如：

```schema
{
  "type": "page",
  data: {
    p1: 'p1'
  },
  "body": {
    "type": "form",
    debug: true,
    "api": {
      url: "https://3xsw4ap8wah59.cfc-execute.bj.baidubce.com/api/awo-ui-mock/mock2/form/saveForm",
      method: "post",
      data: {
        "&": '$$',
        job: 'coder'
      }
    },
    data: {
      job: 'hr'
    },
    "body": [
      {
        type: 'alert',
        "body": "监听姓名值变化，执行动作时读取输入的内容；监听年龄值变化，执行动作时读取input-text组件当前数据域（表单数据）",
        "level": "info",
        "className": "mb-1"
      },
      {
        "type": "input-text",
        "name": "name",
        "label": "姓名：",
        onEvent: {
          change: {
            actions: [
              {
                actionType: 'toast',
                args: {
                  msg: '${name}'
                }
              }
            ]
          }
        }
      },
      {
        "type": "input-text",
        "name": "age",
        "label": "年龄：",
        onEvent: {
          change: {
            actions: [
              {
                actionType: 'toast',
                args: {
                  msg: '${__rendererData|json}'
                }
              }
            ]
          }
        }
      }
    ],
    onEvent: {
      submitSucc: {
        actions: [
          {
            actionType: 'toast',
            args: {
              msg: '${event.data|json}'
            }
          },
          {
            actionType: 'toast',
            args: {
              msg: '${__rendererData|json}'
            }
          }
        ]
      }
    }
  }
}
```

### 运行日志

可以在浏览器控制台查看运行日志，类似如下

```
run action ajax
  [ajax] action args, data {api: {…}, messages: {…}}
  [ajax] action end event {context: {…}, type: 'click', prevented: false, stoped: false, preventDefault: ƒ, …}
```

代表运行了 ajax 动作，第二行是传递的参数和数据，第三行是执行完动作之后的 `event` 值，可以用做后续动作的参数。

## 事件与动作分类

事件包含`渲染器事件`和`广播事件`。

- 渲染器事件，由具体的渲染器组件提供，每个渲染器组件暴露的事件可以查看具体的[组件文档的事件表](../../components/page#事件表)；
- 广播事件，即自定义事件，可以自定义派发的事件名称`eventName`，其他渲染器可以监听该自定义事件并配置响应动作。

动作包含`通用动作`、`组件动作`、`广播动作`、`自定义动作`，可以通过配置`actionType`来指定具体执行什么动作。

## 触发通用动作

通用动作包含发送 http 请求、跳转链接、浏览器回退、浏览器刷新、打开/关闭弹窗、打开/关闭抽屉、打开对话框、弹出 Toast 提示、复制、发送邮件、刷新、控制显示隐藏、控制启用禁用状态、更新数据。

### 发送 http 请求

通过配置`actionType: 'ajax'`实现 http 请求发送，该动作需实现 `env.fetcher` 请求器。

```schema
{
  type: 'page',
  data: {
    name: 'lll'
  },
  body: [
    {
      type: 'button',
      id: 'b_001',
      label: '发送 Ajax 请求',
      level: 'primary',
      "confirmText": "确认要发出这个请求？",
      onEvent: {
        click: {
          actions: [
            {
              actionType: 'ajax',
              args: {
                api: {
                  url: 'https://3xsw4ap8wah59.cfc-execute.bj.baidubce.com/api/awo-ui-mock/mock2/form/saveForm?name=${name}',
                  method: 'get',
                  "responseData": {
                    "resId": "${id}"
                  }
                },
                messages: {
                  success: '成功了！欧耶',
                  failed: '失败了呢。。'
                }
              },
              data: {
                age: 18
              }
            },
            {
              actionType: 'toast',
              expression: '${event.data.responseResult.responseStatus === 0}',
              args: {
                msg: '${event.data.responseResult|json}'
              }
            }
          ]
        }
      }
    },
    {
      type: 'button',
      id: 'b_001',
      label: '发送 Ajax 请求（静默模式）',
      level: 'primary',
      "confirmText": "确认要发出这个请求？",
      className: 'm',
      onEvent: {
        click: {
          actions: [
            {
              actionType: 'ajax',
              args: {
                api: {
                  url: 'https://3xsw4ap8wah59.cfc-execute.bj.baidubce.com/api/awo-ui-mock/initData',
                  method: 'post'
                },
                messages: {
                  success: '成功了！欧耶',
                  failed: '失败了呢。。'
                },
                options: {
                  silent: true
                }
              },
              data: {
                age: 18
              }
            },
            {
              actionType: 'toast',
              expression: '${event.data.responseResult.responseStatus === 0}',
              args: {
                msg: '${event.data.responseResult|json}'
              }
            }
          ]
        }
      }
    }
  ]
}
```

**动作属性（args）**

> `< 1.8.0 及以下版本`，以下属性与 args 同级。

| 属性名   | 类型                                | 默认值 | 说明                      |
| -------- | ----------------------------------- | ------ | ------------------------- |
| api      | [API](../../../docs/types/api)      | -      | 接口配置                  |
| options  | `object`                            | -      | 其他配置                  |
| messages | `{success: string, failed: string}` | -      | 请求成功/失败后的提示信息 |

**其他属性**

| 属性名    | 类型     | 默认值 | 说明                                                                          |
| --------- | -------- | ------ | ----------------------------------------------------------------------------- |
| outputVar | `string` | -      | 请求响应结果缓存在`${event.data.responseResult}`或`${event.data.{outputVar}}` |

请求响应结果的结构如下：

```json
{
  // 状态码
  "responseStatus": 0,
  // 响应数据
  "responseData": {
    "xxx": "xxx"
  },
  // 响应消息
  "responseMsg": "ok"
}
```

### 打开弹窗（模态）

通过配置`actionType: 'dialog'`实现 Dialog 弹窗。

```schema
{
  type: 'page',
  body: [
    {
      type: 'button',
      className: 'ml-2',
      label: '打开弹窗（模态）',
      level: 'primary',
      onEvent: {
        click: {
          actions: [
            {
              actionType: 'dialog',
              args: {
                dialog: {
                  type: 'dialog',
                  title: '模态弹窗',
                  id: 'dialog_001',
                  data: {
                    myage: '22'
                  },
                  body: [
                    {
                      type: 'tpl',
                      tpl: '<p>对，你打开了模态弹窗</p>',
                      inline: false
                    },
                    {
                      type: 'input-text',
                      name: 'myname',
                      mode: 'horizontal',
                      onEvent: {
                        change: {
                          actions: [
                            {
                              actionType: 'confirm',
                              componentId: 'dialog_001'
                            }
                          ]
                        }
                      }
                    }
                  ],
                  onEvent: {
                    confirm: {
                      actions: [
                        {
                          actionType: 'toast',
                          args: {
                            msg: 'confirm'
                          }
                        }
                      ]
                    },
                    cancel: {
                      actions: [
                        {
                          actionType: 'toast',
                          args: {
                            msg: 'cancel'
                          }
                        }
                      ]
                    }
                  }
                }
              }
            }
          ]
        }
      }
    }
  ]
}
```

**动作属性（args）**

> `< 2.3.2 及以下版本`，以下属性与 args 同级。

| 属性名 | 类型                    | 默认值 | 说明                                                      |
| ------ | ----------------------- | ------ | --------------------------------------------------------- |
| dialog | `string`/`DialogObject` | -      | 指定弹框内容，格式可参考[Dialog](../../components/dialog) |

### 关闭弹窗（模态）

通过配置`actionType: 'closeDialog'`实现关闭当前弹窗；附加配置`componentId`可以实现关闭指定弹窗。

```schema
{
  type: 'page',
  body: [
    {
      type: 'button',
      id: 'b_003',
      label: '打开弹窗（模态）',
      level: 'primary',
      onEvent: {
        click: {
          actions: [
            {
              actionType: 'dialog',
              args: {
                dialog: {
                  type: 'dialog',
                  id: 'dialog_002',
                  title: '模态弹窗',
                  body: [
                    {
                      type: 'button',
                      label: '打开子弹窗，然后关闭它的父亲',
                      onEvent: {
                        click: {
                          actions: [
                            {
                              actionType: 'dialog',
                              args: {
                                dialog: {
                                  type: 'dialog',
                                  title: '模态子弹窗',
                                  body: [
                                    {
                                      type: 'button',
                                      label: '关闭指定弹窗（关闭父弹窗）',
                                      onEvent: {
                                        click: {
                                          actions: [
                                            {
                                              actionType: 'closeDialog',
                                              componentId: 'dialog_002'
                                            }
                                          ]
                                        }
                                      }
                                    }
                                  ]
                                }
                              }
                            }
                          ]
                        }
                      }
                    },
                    {
                      type: 'button',
                      label: '关闭当前弹窗',
                      className: 'ml-2',
                      onEvent: {
                        click: {
                          actions: [
                            {
                              actionType: 'closeDialog'
                            }
                          ]
                        }
                      }
                    }
                  ]
                }
              }
            }
          ]
        }
      }
    }
  ]
}
```

**其他属性**

| 属性名      | 类型     | 默认值 | 说明            |
| ----------- | -------- | ------ | --------------- |
| componentId | `string` | -      | 指定弹框组件 id |

### 打开抽屉（模态）

通过配置`actionType: 'drawer'`实现 Drawer 抽屉打开。

```schema
{
  type: 'page',
  body: [
    {
      type: 'button',
      id: 'b_004',
      level: 'primary',
      label: '打开抽屉（模态）',
      onEvent: {
        click: {
          actions: [
            {
              actionType: 'drawer',
              args: {
                drawer: {
                  type: 'drawer',
                  title: '模态抽屉',
                  body: [
                    {
                      type: 'tpl',
                      tpl: '<p>对，你打开了模态抽屉</p>',
                      inline: false
                    }
                  ],
                  onEvent: {
                    confirm: {
                      actions: [
                        {
                          actionType: 'toast',
                          args: {
                            msg: 'confirm'
                          }
                        }
                      ]
                    },
                    cancel: {
                      actions: [
                        {
                          actionType: 'toast',
                          args: {
                            msg: 'cancel'
                          }
                        }
                      ]
                    }
                  }
                }
              }
            }
          ]
        }
      }
    }
  ]
}
```

**动作属性（args）**

> `< 2.3.2 及以下版本`，以下属性与 args 同级。

| 属性名 | 类型                    | 默认值 | 说明                                                      |
| ------ | ----------------------- | ------ | --------------------------------------------------------- |
| drawer | `string`/`DrawerObject` | -      | 指定弹框内容，格式可参考[Drawer](../../components/drawer) |

### 关闭抽屉（模态）

通过配置`actionType: 'closeDrawer'`实现关闭当前抽屉；附加配置`componentId`可以实现关闭指定抽屉。

```schema
{
  type: 'page',
  body: [
    {
      type: 'button',
      label: '打开抽屉（模态）',
      level: 'primary',
      onEvent: {
        click: {
          actions: [
            {
              actionType: 'drawer',
              args: {
                drawer: {
                  type: 'drawer',
                  id: 'drawer_1',
                  title: '模态抽屉',
                  body: [
                    {
                      type: 'button',
                      label: '打开子抽屉，然后关闭它的父亲',
                      onEvent: {
                        click: {
                          actions: [
                            {
                              actionType: 'drawer',
                              args: {
                                drawer: {
                                  type: 'drawer',
                                  title: '模态子抽屉',
                                  body: [
                                    {
                                      type: 'button',
                                      label: '关闭指定抽屉(关闭父抽屉)',
                                      onEvent: {
                                        click: {
                                          actions: [
                                            {
                                              actionType: 'closeDrawer',
                                              componentId: 'drawer_1'
                                            }
                                          ]
                                        }
                                      }
                                    }
                                  ]
                                }
                              }
                            }
                          ]
                        }
                      }
                    },
                    {
                      type: 'button',
                      label: '关闭当前抽屉',
                      className: 'ml-2',
                      onEvent: {
                        click: {
                          actions: [
                            {
                              actionType: 'closeDrawer'
                            }
                          ]
                        }
                      }
                    }
                  ]
                }
              }
            }
          ]
        }
      }
    }
  ]
}
```

**其他属性**

| 属性名      | 类型     | 默认值 | 说明            |
| ----------- | -------- | ------ | --------------- |
| componentId | `string` | -      | 指定抽屉组件 id |

### 打开对话框

通过配置`actionType: 'alert'`或`actionType: 'confirm'`打开不同对话框，该动作分别需实现 env.alert: (msg: string) => void 和 env.confirm: (msg: string, title?: string) => boolean | Promise&lt;boolean&gt;。

#### 提示对话框

```schema
{
  type: 'page',
  data: {
    msg: '去吃饭了'
  },
  body: [
    {
      type: 'button',
      label: '提示对话框（模态）',
      level: 'primary',
      onEvent: {
        click: {
          actions: [
            {
              actionType: 'alert',
              args: {
                title: '提示',
                msg: '<a href="http://www.baidu.com" target="_blank">${msg}~</a>'
              }
            }
          ]
        }
      }
    }
  ]
}
```

**动作属性（args）**

> `< 1.8.0 及以下版本`，以下属性与 args 同级。

| 属性名 | 类型     | 默认值   | 说明           |
| ------ | -------- | -------- | -------------- |
| title  | `string` | 系统提示 | 对话框标题     |
| msg    | `string` | -        | 对话框提示内容 |

#### 确认对话框

```schema
{
  type: 'page',
  data: {
    title: '操作确认',
    msg: '要删除它吗？'
  },
  body: [
    {
      type: 'button',
      label: '确认对话框（模态）',
      level: 'primary',
      onEvent: {
        click: {
          actions: [
            {
              actionType: 'confirmDialog',
              args: {
                title: '${title}',
                msg: '<span style="color:red">${msg}</span>'
              }
            }
          ]
        }
      }
    }
  ]
}
```

**动作属性（args）**

> `< 1.8.0 及以下版本`，以下属性与 args 同级。

| 属性名 | 类型     | 默认值 | 说明           |
| ------ | -------- | ------ | -------------- |
| title  | `string` | -      | 对话框标题     |
| msg    | `string` | -      | 对话框提示内容 |

### 跳转链接

通过配置`actionType: 'url'`或`actionType: 'link'`实现链接跳转，该动作需实现 env.jumpTo(to: string, action?: any) => void 方法。

#### 打开页面

```schema
{
  type: 'page',
  data: {
    myname: 'lvxj',
    myjon: 'player'
  },
  body: [
    {
      type: 'button',
      label: '百度一下',
      level: 'primary',
      className: 'ml-2',
      onEvent: {
        click: {
          actions: [
            {
              actionType: 'url',
              args: {
                url: 'http://www.baidu.com',
                blank: true,
                params: {
                  name: 'jack',
                  jon: '${myjon}'
                },
                name: '${myname}',
                age: 18
              }
            }
          ]
        }
      }
    }
  ]
}
```

**动作属性（args）**

> `< 1.8.0 及以下版本`，以下属性与 args 同级。

| 属性名 | 类型      | 默认值  | 说明                                                       |
| ------ | --------- | ------- | ---------------------------------------------------------- |
| url    | `string`  | -       | 按钮点击后，会打开指定页面。可用 `${xxx}` 取值             |
| blank  | `boolean` | `false` | 如果为 `true` 将在新 tab 页面打开                          |
| params | `object`  | -       | 页面参数`{key:value}`，支持数据映射，`> 1.10.0 及以上版本` |

#### 打开单页

```schema
{
  type: 'page',
  body: [
    {
      type: 'button',
      label: '跳转至「表达式」',
      level: 'primary',
      className: 'ml-2',
      onEvent: {
        click: {
          actions: [
            {
              actionType: 'link',
              args: {
                link: './expression'
              }
            }
          ]
        }
      }
    }
  ]
}
```

**动作属性（args）**

> `< 1.8.0 及以下版本`，以下属性与 args 同级。

| 属性名 | 类型     | 默认值 | 说明                                                                                                              |
| ------ | -------- | ------ | ----------------------------------------------------------------------------------------------------------------- |
| link   | `string` | `link` | 用来指定跳转地址，跟 url 不同的是，这是单页跳转方式，不会渲染浏览器，请指定 awo-ui 平台内的页面。可用 `${xxx}` 取值 |
| params | `object` | -      | 页面参数`{key:value}`，支持数据映射，`> 1.9.0 及以上版本`                                                         |

### 浏览器回退

> 1.8.0 及以上版本

通过配置`actionType: 'goBack'`实现页面回退。

```schema
{
  type: 'page',
  body: [
    {
      type: 'button',
      label: '回退',
      level: 'primary',
      className: 'ml-2',
      onEvent: {
        click: {
          actions: [
            {
              actionType: 'goBack'
            }
          ]
        }
      }
    }
  ]
}
```

### 前进/后退到指定页面

> 1.8.0 及以上版本

通过配置`actionType: 'goPage'`实现浏览器页面的前进/后退。只有当历史记录中存在目标页面时才会生效。

```schema
{
  type: 'page',
  body: [
    {
      type: 'button',
      label: '后退到上上个页面',
      level: 'primary',
      className: 'ml-2',
      onEvent: {
        click: {
          actions: [
            {
              actionType: 'goPage',
              args: {
                delta: -2
              }
            }
          ]
        }
      }
    }
  ]
}
```

**动作属性（args）**

> `< 1.8.0 及以下版本`，以下属性与 args 同级。

| 属性名 | 类型     | 默认值 | 说明 |
| ------ | -------- | ------ | ---- |
| delta  | `string` | `0`    | 位置 |

### 浏览器刷新

> 1.8.0 及以上版本

通过配置`actionType: 'refresh'`实现浏览器刷新。

```schema
{
  type: 'page',
  body: [
    {
      type: 'button',
      label: '刷新页面',
      level: 'primary',
      className: 'ml-2',
      onEvent: {
        click: {
          actions: [
            {
              actionType: 'refresh'
            }
          ]
        }
      }
    }
  ]
}
```

### toast 提示

通过配置`actionType: 'toast'`实现弹出 toast 提示，该动作需实现 env.notify(type: ToastLevel, msg: string, conf?: ToastConf) => void 方法。

```schema
{
  type: 'page',
  body: [
    {
      type: 'button',
      label: '警告',
      level: 'warning',
      className: 'mr-2 mb-2',
      onEvent: {
        click: {
          actions: [
            {
              actionType: 'toast',
              args: {
                msgType: 'warning',
                msg: '我是全局警告消息，可以配置不同类型和弹出位置~',
                position: 'top-right'
              }
            }
          ]
        }
      }
    },
    {
      type: 'button',
      label: '成功',
      level: 'success',
      className: 'mr-2 mb-2',
      onEvent: {
        click: {
          actions: [
            {
              actionType: 'toast',
              args: {
                msgType: 'success',
                msg: '我是全局警告消息，可以配置不同类型和弹出位置~',
                position: 'top-right'
              }
            }
          ]
        }
      }
    },
    {
      type: 'button',
      label: '错误',
      level: 'danger',
      className: 'mr-2 mb-2',
      onEvent: {
        click: {
          actions: [
            {
              actionType: 'toast',
              args: {
                msgType: 'error',
                msg: '我是全局警告消息，可以配置不同类型和弹出位置~',
                position: 'top-right'
              }
            }
          ]
        }
      }
    },
    {
      type: 'button',
      label: '提示',
      level: 'info',
      className: 'mb-2',
      onEvent: {
        click: {
          actions: [
            {
              actionType: 'toast',
              args: {
                msgType: 'info',
                msg: '我是全局警告消息，可以配置不同类型和弹出位置~',
                position: 'top-right'
              }
            }
          ]
        }
      }
    },
    {
      type: 'tpl',
      tpl: '<br>',
    },
    {
      type: 'button',
      label: '左上',
      className: 'mr-2 mt-2',
      onEvent: {
        click: {
          actions: [
            {
              actionType: 'toast',
              args: {
                msgType: 'info',
                msg: '我是全局警告消息，可以配置不同类型和弹出位置~',
                position: 'top-left'
              }
            }
          ]
        }
      }
    },
    {
      type: 'button',
      label: '中上',
      className: 'mr-2 mt-2',
      onEvent: {
        click: {
          actions: [
            {
              actionType: 'toast',
              args: {
                msgType: 'info',
                msg: '我是全局警告消息，可以配置不同类型和弹出位置~',
                position: 'top-center'
              }
            }
          ]
        }
      }
    },
    {
      type: 'button',
      label: '右上',
      className: 'mt-2',
      onEvent: {
        click: {
          actions: [
            {
              actionType: 'toast',
              args: {
                msgType: 'info',
                msg: '我是全局警告消息，可以配置不同类型和弹出位置~',
                position: 'top-right'
              }
            }
          ]
        }
      }
    },
    {
      type: 'tpl',
      tpl: '<br>',
    },
    {
      type: 'button',
      label: '屏幕的中央',
      className: 'ml-10 mt-2 mb-2',
      onEvent: {
        click: {
          actions: [
            {
              actionType: 'toast',
              args: {
                msgType: 'info',
                msg: '我是全局警告消息，可以配置不同类型和弹出位置~',
                position: 'center'
              }
            }
          ]
        }
      }
    },
    {
      type: 'tpl',
      tpl: '<br>',
    },
    {
      type: 'button',
      label: '左下',
      className: 'mr-2',
      onEvent: {
        click: {
          actions: [
            {
              actionType: 'toast',
              args: {
                msgType: 'info',
                msg: '我是全局警告消息，可以配置不同类型和弹出位置~',
                position: 'bottom-left'
              }
            }
          ]
        }
      }
    },
    {
      type: 'button',
      label: '中下',
      className: 'mr-2',
      onEvent: {
        click: {
          actions: [
            {
              actionType: 'toast',
              args: {
                msgType: 'info',
                msg: '我是全局警告消息，可以配置不同类型和弹出位置~',
                position: 'bottom-center'
              }
            }
          ]
        }
      }
    },
    {
      type: 'button',
      label: '右下',
      onEvent: {
        click: {
          actions: [
            {
              actionType: 'toast',
              args: {
                msgType: 'info',
                msg: '我是全局警告消息，可以配置不同类型和弹出位置~',
                position: 'bottom-right'
              }
            }
          ]
        }
      }
    }
  ]
}
```

**动作属性（args）**

> `< 1.8.0 及以下版本`，以下属性与 args 同级。

| 属性名      | 类型      | 默认值                                  | 说明                                                                                             |
| ----------- | --------- | --------------------------------------- | ------------------------------------------------------------------------------------------------ |
| msgType     | `string`  | `"info"`                                | 消息类型 `info\|success\|error\|warning`                                                         |
| msg         | `string`  | -                                       | 消息内容                                                                                         |
| position    | `string`  | `top-center（移动端为center）`          | 提示显示位置 `top-right\|top-center\|top-left\|bottom-center\|bottom-left\|bottom-right\|center` |
| closeButton | `boolean` | `false`                                 | 是否展示关闭按钮                                                                                 |
| showIcon    | `boolean` | `true`                                  | 是否展示图标                                                                                     |
| timeout     | `number`  | `5000（error类型为6000，移动端为3000）` | 持续时间                                                                                         |

### 复制

通过配置`actionType: 'copy'`和复制属性实现文本的复制操作，该动作需实现 env.copy(contents: string, options?: any) => void 方法。

```schema
{
  type: 'page',
  body: [
    {
      type: 'button',
      label: '复制一段文本',
      level: 'primary',
      className: 'mr-2',
      onEvent: {
        click: {
          actions: [
            {
              actionType: 'copy',
              args: {
                content: 'http://www.baidu.com'
              }
            }
          ]
        }
      }
    },
    {
      type: 'button',
      label: '复制一段富文本',
      level: 'primary',
      onEvent: {
        click: {
          actions: [
            {
              actionType: 'copy',
              args: {
                copyFormat: 'text/html',
                content: "<a href='http://www.baidu.com'>link</a> <b>bold</b>"
              }
            }
          ]
        }
      }
    }
  ]
}
```

**动作属性（args）**

> `< 1.8.0 及以下版本`，以下属性与 args 同级。

| 属性名     | 类型                                 | 默认值      | 说明                               |
| ---------- | ------------------------------------ | ----------- | ---------------------------------- |
| copyFormat | `string`                             | `text/html` | 复制格式                           |
| content    | [模板](../../docs/concepts/template) | -           | 指定复制的内容。可用 `${xxx}` 取值 |

### 发送邮件

通过配置`actionType: 'email'`和邮件属性实现发送邮件操作。

```schema
{
  type: 'page',
  body: [
    {
      type: 'button',
      label: '发送邮件',
      level: 'primary',
      onEvent: {
        click: {
          actions: [
            {
              actionType: 'email',
              args: {
                to: 'awo-ui@baidu.com',
                cc: 'baidu@baidu.com',
                subject: '这是邮件主题',
                body: '这是邮件正文'
              }
            }
          ]
        }
      }
    }
  ]
}
```

**动作属性（args）**

> `< 1.8.0 及以下版本`，以下属性与 args 同级。

| 属性名  | 类型     | 默认值 | 说明                           |
| ------- | -------- | ------ | ------------------------------ |
| to      | `string` | -      | 收件人邮箱，可用 ${xxx} 取值   |
| cc      | `string` | -      | 抄送邮箱，可用 ${xxx} 取值     |
| bcc     | `string` | -      | 匿名抄送邮箱，可用 ${xxx} 取值 |
| subject | `string` | -      | 邮件主题，可用 ${xxx} 取值     |
| body    | `string` | -      | 邮件正文，可用 ${xxx} 取值     |

### 刷新

通过配置`actionType: 'reload'`实现对指定组件的刷新（重新加载）操作，仅支持`form`、`wizard`、`service`、`page`、`app`、`chart`、`crud`，以及支持动态数据的`输入类`组件，详见组件的`动作表`。更多示例请查看[刷新示例](../../../examples/action/reload/form)。

```schema
{
  type: 'page',
  body: [
    {
      type: 'button',
      label: '刷新',
      className: 'mb-2',
      level: 'primary',
      onEvent: {
        click: {
          actions: [
            {
              actionType: 'reload',
              componentId: 'form-reload'
            }
          ]
        }
      }
    },
    {
      type: 'form',
      id: 'form-reload',
      name: 'form-reload',
      initApi:
        'https://3xsw4ap8wah59.cfc-execute.bj.baidubce.com/api/awo-ui-mock/mock2/page/initData',
      title: '表单',
      body: [
        {
          type: 'input-text',
          id: 'date-input-01',
          name: 'date',
          label: '时间戳'
        }
      ]
    }
  ]
}
```

**动作属性**

| 属性名    | 类型      | 默认值 | 说明                                                               |
| --------- | --------- | ------ | ------------------------------------------------------------------ |
| resetPage | `boolean` | true   | 当目标组件为 `crud` 时，可以控制是否重置页码，`> 2.3.2 及以上版本` |

**其他属性**

| 属性名      | 类型     | 默认值 | 说明                  |
| ----------- | -------- | ------ | --------------------- |
| componentId | `string` | -      | 指定刷新的目标组件 id |

### 显示与隐藏

> 1.8.0 及以上版本

通过配置`actionType: 'show'`或`'hidden'`实现对指定组件的显示和隐藏，且显隐动作的控制高于组件显隐属性的设置。

```schema
{
  type: 'page',
  body: [
    {
      type: 'button',
      label: '显示',
      level: 'primary',
      className: 'mr-2',
      onEvent: {
        click: {
          actions: [
            {
              actionType: 'show',
              componentId: 'input-text_001'
            }
          ]
        }
      }
    },
    {
      type: 'button',
      label: '隐藏',
      level: 'primary',
      onEvent: {
        click: {
          actions: [
            {
              actionType: 'hidden',
              componentId: 'input-text_001'
            }
          ]
        }
      }
    },
    {
      type: 'input-text',
      label: '愿望',
      className: 'mt-2',
      id: 'input-text_001',
      mode: 'horizontal',
      hidden: true
    }
  ]
}
```

**其他属性**

| 属性名      | 类型     | 默认值 | 说明                        |
| ----------- | -------- | ------ | --------------------------- |
| componentId | `string` | -      | 指定显示或隐藏的目标组件 id |

### 控制状态

> 1.8.0 及以上版本

通过配置`actionType: 'enabled'`或`actionType: 'disabled'`实现对指定组件的启用和禁用，仅支持实现了状态控制功能的数据`输入类`组件。

```schema
{
  type: 'page',
  body: [
    {
      type: 'button',
      id: 'b_dis',
      label: '禁用',
      level: 'primary',
      className: 'mr-2 mb-2',
      disabled: true,
      onEvent: {
        click: {
          actions: [
            {
              actionType: 'disabled',
              componentId: 'form_disable'
            }
          ]
        }
      }
    },
    {
      type: 'button',
      label: '启用',
      level: 'primary',
      className: 'mb-2',
      onEvent: {
        click: {
          actions: [
            {
              actionType: 'enabled',
              componentId: 'form_disable'
            }
          ]
        }
      }
    },
    {
      type: 'form',
      id: 'form_disable',
      title: '表单',
      body: [
        {
          type: 'group',
          body: [
            {
              type: 'button',
              className: 'ml-2',
              label: '我的状态变了'
            },
            {
              type: 'button',
              className: 'ml-2',
              label: '禁用上面的按钮',
              level: 'primary',
              onEvent: {
                click: {
                  actions: [
                    {
                      actionType: 'disabled',
                      componentId: 'b_dis'
                    }
                  ]
                }
              }
            },
            {
              type: 'button',
              className: 'ml-2',
              label: '启用用上面的按钮',
              level: 'primary',
              onEvent: {
                click: {
                  actions: [
                    {
                      actionType: 'enabled',
                      componentId: 'b_dis'
                    }
                  ]
                }
              }
            }
          ]
        }
      ]
    }
  ]
}
```

**其他属性**

| 属性名      | 类型     | 默认值 | 说明                        |
| ----------- | -------- | ------ | --------------------------- |
| componentId | `string` | -      | 指定启用或禁用的目标组件 id |

### 更新数据

> 1.8.0 及以上版本

更新数据即更新指定组件数据域中的数据（data），通过配置`actionType: 'setValue'`实现组件`数据域变量更新`，通过它可以实现`组件间联动更新`、`数据回填`，更多示例请查看[更新数据示例](../../../examples/action/setdata/form)。

**注意事项**

- 数据类型支持范围：`基础类型`、`对象类型`、`数组类型`，数据类型取决于目标组件所需数据值类型
- 目标组件支持范围：`form`、`dialog`、`drawer`、`wizard`、`service`、`page`、`app`、`chart`，以及数据`输入类`组件
- < 2.3.2 及以下版本，虽然更新数据可以实现对组件数据域的更新，但如果更新数据动作的数据值来自前面的异步动作（例如 发送 http 请求、自定义 JS（异步）），则后面的动作只能通过事件变量`${event.data.xxx}`来获取异步动作产生的数据，无法通过当前数据域`${xxx}`直接获取更新后的数据。
- 它的值通常都是对象形式，比如 form 传递的值应该是类似 `{"user": "awo-ui"}`，这时就会更新表单里的 `user` 字段值为 `amis`

```schema
{
  type: 'page',
  title: '更新表单数据',
  data: {
    globalData: {
      myrole: '法官',
      mymsg: '该吃饭了!'
    }
  },
  body: [
    {
      type: 'button',
      label: '更新',
      level: 'primary',
      className: 'mb-2',
      onEvent: {
        click: {
          actions: [
            {
              actionType: 'setValue',
              componentId: 'form_data',
              args: {
                value: '${globalData}'
              }
            }
          ]
        }
      }
    },
    {
      type: 'form',
      id: 'form_data',
      title: '表单',
      debug: true,
      data: {
        myrole: '预言家',
        age: '18'
      },
      "initApi": "/api/mock2/form/initData",
      body: [
        {
          type: 'input-text',
          label: '角色',
          name: 'myrole',
          disabled: false,
          mode: 'horizontal'
        },
        {
          type: 'input-text',
          label: '年龄',
          name: 'age',
          disabled: false,
          mode: 'horizontal'
        }
      ]
    }
  ]
}
```

**动作属性（args）**

> `< 1.8.0 及以下版本`，以下属性与 args 同级。

| 属性名 | 类型     | 默认值 | 说明                                                                |
| ------ | -------- | ------ | ------------------------------------------------------------------- |
| value  | `any`    | -      | 值                                                                  |
| index  | `number` | -      | 当目标组件是`combo`时，可以指定更新的数据索引， `1.10.1 及以上版本` |

**其他属性**

| 属性名      | 类型     | 默认值 | 说明                  |
| ----------- | -------- | ------ | --------------------- |
| componentId | `string` | -      | 指定赋值的目标组件 id |

### 自定义 JS

通过配置`actionType: 'custom'`实现自定义 JS。JS 中可以访问以下对象和方法：

- context，渲染器上下文
- doAction() 动作执行方法，用于调用任何 actionType 指定的动作
- event，事件对象，可以调用 setData()、stopPropagation()、preventDefault()分别实现事件上下文设置、动作干预、事件干预，可以通过 event.data 获取事件上下文

```schema
{
  type: 'page',
  body: [
    {
      type: 'button',
      label: '发送一个 http 请求',
      level: 'primary',
      onEvent: {
        click: {
          actions: [
            {
              actionType: 'custom',
              script:
                "doAction({actionType: 'ajax', args: {api: '/api/mock2/form/saveForm'}});\n //event.stopPropagation();"
            }
          ]
        }
      }
    }
  ]
}
```

**动作属性（args）**

> `< 2.3.2 及以下版本`，以下属性与 args 同级。

| 属性名 | 类型                | 默认值 | 说明                                                                                                                                            |
| ------ | ------------------- | ------ | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| script | `string`/`function` | -      | 自定义 JS 脚本代码，代码内可以通过调用`doAction`执行任何[动作](../../docs/concepts/event-action#动作) ，通过事件对象`event`可以实现事件动作干预 |

#### 支持异步

> 2.0.3 及以上版本

- 方式一：通过返回 Promise 实例的方式

```schema
{
  type: 'page',
  body: [
    {
      "type": "form",
      "title": "表单",
      "body": [
        {
          "label": "编号",
          "type": "input-text",
          "name": "pId",
          "id": "u:e47e2c8e6be8",
          "mode": "horizontal",
          "addOn": {
            "label": "自动获取",
            "type": "button",
            "onEvent": {
              "click": {
                "actions": [
                  {
                    "componentId": "u:52cd013e120f",
                    "actionType": "disabled"
                  },
                  {
                    "script": "return new Promise((resolve, reject) => {setTimeout(() => {event.setData({...event.data, pId: '01027359'});resolve();}, 3000)})",
                    "actionType": "custom"
                  },
                  {
                    "componentId": "u:e47e2c8e6be8",
                    "args": {
                      "value": "${pId}"
                    },
                    "actionType": "setValue"
                  },
                  {
                    "componentId": "u:52cd013e120f",
                    "actionType": "enabled"
                  }
                ],
                "weight": 0
              }
            },
            "id": "u:52cd013e120f"
          }
        }
      ],
      "apiFromAPICenter": false,
      "id": "u:76203156676b"
    }
  ]
}
```

- 方式二：通过返回 Thunk 的方式

```schema
{
  type: 'page',
  body: [
    {
      "type": "form",
      "title": "表单",
      "body": [
        {
          "label": "编号",
          "type": "input-text",
          "name": "pId",
          "id": "u:e47e2c8e6be7",
          "mode": "horizontal",
          "addOn": {
            "label": "自动获取",
            "type": "button",
            "onEvent": {
              "click": {
                "actions": [
                  {
                    "componentId": "u:52cd013e120e",
                    "actionType": "disabled"
                  },
                  {
                    "script": "return (callback) => { setTimeout(() => {event.setData({...event.data, pId: '01027359' });callback();}, 3000) };",
                    "actionType": "custom"
                  },
                  {
                    "componentId": "u:e47e2c8e6be7",
                    "args": {
                      "value": "${pId}"
                    },
                    "actionType": "setValue"
                  },
                  {
                    "componentId": "u:52cd013e120e",
                    "actionType": "enabled"
                  }
                ],
                "weight": 0
              }
            },
            "id": "u:52cd013e120e"
          }
        }
      ],
      "apiFromAPICenter": false,
      "id": "u:76203156676a"
    }
  ]
}
```

#### 存储数据

有时在执行自定义 JS 的时候，希望该过程中产生的数据可以分享给后面的动作使用，此时可以通过`event.setData()`来实现事件上下文的设置，这样后面动作都可以通过事件上下文来获取共享的数据。

> 注意：直接调用`event.setData()`将修改事件的原有上下文，如果不希望覆盖可以通过`event.setData({...event.data, {xxx: xxx}})`来进行数据的合并。

## 触发其他组件的动作

通过配置`componentId`来触发指定组件的动作，组件动作配置通过`args`传入`(> 1.9.0 及以上版本)`，动作参数请查看对应的组件的[动作表](../../components/form/index#动作表)，更多示例请查看[组件事件动作示例](../../../examples/event/form)。

```schema
{
  type: 'page',
  body: [
    {
      type: 'button',
      label: '选中选项卡2',
      level: 'primary',
      className: 'mr-3 mb-3',
      onEvent: {
        click: {
          actions: [
            {
              actionType: 'changeActiveKey',
              componentId: 'tabs-change-receiver',
              args: {
                activeKey: 2
              }
            }
          ]
        }
      }
    },
    {
      id: 'tabs-change-receiver',
      type: 'tabs',
      mode: 'line',
      tabs: [
        {
          title: '选项卡1',
          body: '选项卡内容1'
        },
        {
          title: '选项卡2',
          body: '选项卡内容2'
        },
        {
          title: '选项卡3',
          body: '选项卡内容3'
        }
      ]
    }
  ]
}
```

## 触发广播动作

通过配置`actionType: 'broadcast'`实现触发一个广播。

```schema
{
  type: 'page',
  body: [
    {
      "name": "role",
      "type": "select",
      "label": "广播一下",
      "mode": "row",
      "options": [
        {
          "label": "海贼王的男人",
          "value": "路飞"
        },
        {
          "label": "海上华佗",
          "value": "乔巴"
        },
        {
          "label": "海上食神",
          "value": "山治"
        }
      ],
      "onEvent": {
        "change": {
          "actions": [
            {
              actionType: 'broadcast',
              args: {
                eventName: 'broadcast_1',
              },
              data: {
                myrole: '${role}',
                age: 18
              }
            }
          ]
        }
      }
    },
    {
      type: 'form',
      id: 'form_001_form_01',
      title: '表单1（优先级低）',
      name: 'sub-form1',
      body: [
        {
          type: 'input-text',
          label: '昵称',
          name: 'myname',
          disabled: false,
          mode: 'horizontal'
        }
      ],
      onEvent: {
        broadcast_1: {
          actions: [
            {
              actionType: 'reload',
              data: {
                myname: '${myrole}', // 从事件数据中取
              }
            },
            {
              "actionType": "toast",
              args: {
                "msgType": "info",
                "msg": "表单1刷新完成"
              }
            }
          ]
        }
      }
    },
    {
      type: 'form',
      name: 'form2',
      id: 'form_002',
      title: '表单2（优先级中）',
      body: [
        {
          type: 'input-text',
          label: '角色',
          name: 'myrole',
          disabled: false,
          mode: 'horizontal'
        },
        {
          type: 'input-text',
          label: '年龄',
          name: 'age',
          disabled: false,
          mode: 'horizontal'
        }
      ],
      onEvent: {
        broadcast_1: {
          weight: 2,
          actions: [
            {
              actionType: 'reload',
              data: {
                myrole: '${myrole}',
                age: '${age}'
              }
            },
            {
              "actionType": "toast",
              args: {
                "msgType": "info",
                "msg": "表单2刷新完成"
              }
            }
          ]
        }
      }
    },
    {
      type: 'form',
      name: 'form3',
      id: 'form_003',
      title: '表单3（优先级高）',
      body: [
        {
          type: 'input-text',
          id: 'form_003_text_01',
          label: '职业',
          name: 'job',
          disabled: false,
          mode: 'horizontal'
        }
      ],
      api: 'https://api/form/form3',
      onEvent: {
        broadcast_1: {
          weight: 3,
          actions: [
            {
              actionType: 'reload',
              data: {
                job: '${myrole}'
              }
            },
            {
              "actionType": "toast",
              args: {
                "msgType": "info",
                "msg": "表单3刷新完成"
              }
            }
          ]
        }
      }
    }
  ]
}
```

**动作属性（args）**

> `< 2.3.2 及以下版本`，以下属性与 args 同级。

| 属性名    | 类型     | 默认值 | 说明                                             |
| --------- | -------- | ------ | ------------------------------------------------ |
| eventName | `string` | -      | 广播动作对应的自定义事件名称，用于广播事件的监听 |

**其他属性**

| 属性名 | 类型     | 默认值 | 说明                                                     |
| ------ | -------- | ------ | -------------------------------------------------------- |
| weight | `number` | 0      | 可以通过配置动作执行优先级来控制所有监听者的动作执行顺序 |

## 自定义动作

除了以上内置动作，你还可以注册自己的动作。通过对`RendererAction`的`run`方法的实现可以定制自己的动作逻辑，最后通过`registerAction`注册到 awo-ui 事件动作中。

```javascript
import {
  ListenerAction,
  ListenerContext,
  registerAction,
  RendererAction
} from 'awo-ui-core';
import {RendererEvent} from 'awo-ui-core';

// 动作定义
interface IMyAction extends ListenerAction {
  actionType: 'my-action';
  args: {
    param1: string, // 动作参数1
    param2: string // 动作参数2
  };
}

/**
 * 我的动作实现
 */
export class MyAction implements RendererAction {
  run(action: IMyAction, renderer: ListenerContext, event: RendererEvent<any>) {
    const props = renderer.props;
    const {param1, param2} = action.args;

    // 你的动作逻辑
    // ...
  }
}

// 注册自定义动作
registerAction('my-action', new MyAction());
```

# 编排动作

通过配置不同的逻辑动作实现动作编排，支持嵌套。

## 条件

通过配置`expression: 表达式`来实现条件逻辑。

```schema
{
  type: 'page',
  body: {
    type: 'form',
    wrapWithPanel: false,
    data: {
      expression: 'okk'
    },
    body: [
      {
        type: 'button',
        label: '符合条件的执行',
        level: 'primary',
        onEvent: {
          click: {
            actions: [
              {
                actionType: 'toast',
                args: {
                  msgType: 'success',
                  msg: '我okk~'
                },
                expression: 'expression === "okk"'
              },
              {
                actionType: 'toast',
                args: {
                  msg: '1'
                },
                expression: 'expression === "nono"'
              },
              {
                actionType: 'toast',
                args: {
                  msgType: 'success',
                  msg: '我也okk~'
                },
                expression: 'expression === "okk"'
              }
            ]
          }
        }
      }
    ]
  }
}
```

**其他属性**

| 属性名     | 类型                                        | 默认值 | 说明                         |
| ---------- | ------------------------------------------- | ------ | ---------------------------- |
| expression | `boolean`\|[表达式](../concepts/expression) | -      | 执行条件，不设置表示默认执行 |

## 循环

通过配置`actionType: 'for'`实现循环逻辑。

### 单层循环

```schema
{
  type: 'page',
  body: {
    type: 'form',
    wrapWithPanel: false,
    data: {
      loopName: 'loopData',
      loopData: [
        {
          name: 'lv',
          age: '19'
        },
        {
          name: 'xj',
          age: '21'
        }
      ]
    },
    body: [
      {
        type: 'button',
        label: '循环发送两次请求，且每次携带了循环的数据',
        level: 'primary',
        onEvent: {
          click: {
            actions: [
              {
                actionType: 'loop',
                args: {
                  loopName: '${loopName}'
                },
                children: [
                  {
                    actionType: 'ajax',
                    args: {
                      api: 'https://3xsw4ap8wah59.cfc-execute.bj.baidubce.com/api/awo-ui-mock/mock2/form/saveForm?name=${name}&age=${age}'
                    }
                  }
                ]
              }
            ]
          }
        }
      }
    ]
  }
}
```

### 嵌套循环

```schema
{
  type: 'page',
  body: {
    type: 'form',
    wrapWithPanel: false,
    data: {
      loopName: 'loopData',
      loopData: [
        {
          name: 'lv',
          age: '19'
        },
        {
          name: 'xj',
          age: '21'
        }
      ]
    },
    body: [
      {
        type: 'button',
        label: '循环',
        className: 'm-2',
        onEvent: {
          click: {
            actions: [
              {
                actionType: 'loop',
                preventDefault: false,
                stopPropagation: false,
                args: {
                  loopName: '${loopName}'
                },
                children: [
                  {
                    actionType: 'toast',
                    args: {
                      msg: '第1层循环动作1${name}'
                    },
                    preventDefault: false,
                    stopPropagation: false
                  },
                  {
                    actionType: 'toast',
                    args: {
                      msg: '第1层循环动作2${name}'
                    }
                  },
                  {
                    actionType: 'loop',
                    args: {
                      loopName: '${loopName}'
                    },
                    children: [
                      {
                        actionType: 'toast',
                        args: {
                          msg: '第2层循环动作1${name}'
                        }
                      },
                      {
                        actionType: 'toast',
                        args: {
                          msg: '第2层循环动作2${name}'
                        },
                        preventDefault: false,
                        stopPropagation: false
                      },
                      {
                        actionType: 'continue'
                      },
                      {
                        actionType: 'toast',
                        args: {
                          msg: '第2层循环动作3${name}'
                        }
                      }
                    ]
                  }
                ]
              }
            ]
          }
        }
      }
    ]
  }
}
```

**动作属性（args）**

> `< 1.8.0 及以下版本`，以下属性与 args 同级。

| 属性名   | 类型     | 默认值 | 说明         |
| -------- | -------- | ------ | ------------ |
| loopName | `string` | -      | 循环变量名称 |

**其他属性**

> `< 2.3.2 及以下版本`，以下属性与 args 同级。

| 属性名   | 类型                                                 | 默认值 | 说明                                  |
| -------- | ---------------------------------------------------- | ------ | ------------------------------------- |
| children | Array<[动作](../../docs/concepts/event-action#动作)> | -      | 子动作，可以通过`break动作`来跳出循环 |

### Break 动作

通过配置`actionType: 'for'`和`actionType: 'break'`实现循环跳出。

```schema
{
  type: 'page',
  body: {
    type: 'form',
    wrapWithPanel: false,
    data: {
      loopName: 'loopData',
      loopData: [
        {
          name: 'lv',
          age: '19'
        },
        {
          name: 'xj',
          age: '21'
        }
      ]
    },
    body: [
      {
        type: 'button',
        label: '只执行了第一个动作就跳出了循环',
        level: 'primary',
        onEvent: {
          click: {
            actions: [
              {
                actionType: 'loop',
                args: {
                  loopName: '${loopName}'
                },
                children: [
                  {
                    actionType: "toast",
                    args: {
                      msgType: 'success',
                      msg: '第 1 个动作',
                      position: 'top-right'
                    }
                  },
                  {
                    actionType: 'break'
                  },
                  {
                    actionType: "toast",
                    args: {
                      msgType: 'success',
                      msg: '第 2 个动作',
                      position: 'top-right'
                    }
                  }
                ]
              }
            ]
          }
        }
      }
    ]
  }
}
```

### Continue 动作

通过配置`actionType: 'for'`和`actionType: 'continue'`实现循环跳过。

```schema
{
  type: 'page',
  body: {
    type: 'form',
    wrapWithPanel: false,
    data: {
      loopName: 'loopData',
      loopData: [
        {
          name: 'lv',
          age: '19'
        },
        {
          name: 'xj',
          age: '21'
        }
      ]
    },
    body: [
      {
        type: 'button',
        label: '只循环执行第一个动作',
        level: 'primary',
        onEvent: {
          click: {
            actions: [
              {
                actionType: 'loop',
                args: {
                  loopName: '${loopName}',
                  level: 3
                },
                children: [
                  {
                    actionType: "toast",
                    args: {
                      msgType: 'success',
                      msg: '第 1 个动作',
                      position: 'top-right'
                    }
                  },
                  {
                    actionType: 'continue'
                  },
                  {
                    actionType: "toast",
                    args: {
                      msgType: 'success',
                      msg: '最后的动作',
                      position: 'top-right'
                    }
                  }
                ]
              }
            ]
          }
        }
      }
    ]
  }
}
```

## 排他（switch）

通过配置`actionType: 'switch'`实现排他逻辑。

```schema
{
  type: 'page',
  body: {
    type: 'form',
    wrapWithPanel: false,
    data: {
      branchCont: 18
    },
    body: [
      {
        type: 'button',
        label: '只执行动作2',
        level: 'primary',
        onEvent: {
          click: {
            actions: [
              {
                actionType: 'switch',
                children: [
                  {
                    actionType: "toast",
                    args: {
                      msgType: 'info',
                      msg: '动作1',
                      position: 'top-right'
                    },
                    expression: 'this.branchCont > 19',
                    stopPropagation: true // 这里无效，因为条件不成立
                  },
                  {
                    actionType: "toast",
                    args: {
                      msgType: 'info',
                      msg: '动作2',
                      position: 'top-right'
                    },
                    expression: 'this.branchCont > 17'
                  },
                  {
                    actionType: "toast",
                    args: {
                      msgType: 'info',
                      msg: '动作3',
                      position: 'top-right'
                    },
                    expression: 'this.branchCont > 16'
                  }
                ]
              }
            ]
          }
        }
      }
    ]
  }
}
```

**其他属性**

| 属性名   | 类型                                                 | 默认值 | 说明                                                   |
| -------- | ---------------------------------------------------- | ------ | ------------------------------------------------------ |
| children | Array<[动作](../../docs/concepts/event-action#动作)> | -      | 子动作，每个子动作可以通过配置`expression`来匹配的条件 |

## 并行

通过配置`actionType: 'parallel'`实现并行执逻辑。

```schema
{
  type: 'page',
  body: {
    type: 'form',
    wrapWithPanel: false,
    data: {
      name: 'lvxj',
      age: 'kkkk'
    },
    body: [
      {
        type: 'button',
        label: '同时发送两个ajax请求，并显示请求返回',
        level: 'primary',
        onEvent: {
          click: {
            actions: [
              {
                actionType: 'parallel',
                children: [
                  {
                    actionType: 'ajax',
                    args: {
                      api: {
                        url: 'https://3xsw4ap8wah59.cfc-execute.bj.baidubce.com/api/awo-ui-mock/mock2/form/initData?name=${name}',
                        method: 'get'
                      },
                      messages: {
                        success: '请求1成功了！欧耶',
                        failed: '失败了呢。。'
                      }
                    },
                    outputVar: 'var1'
                  },
                  {
                    actionType: 'ajax',
                    args: {
                      api: {
                        url: 'https://3xsw4ap8wah59.cfc-execute.bj.baidubce.com/api/awo-ui-mock/mock2/form/saveForm?name=${name}',
                        method: 'get'
                      },
                      messages: {
                        success: '请求2成功了！欧耶',
                        failed: '失败了呢。。'
                      }
                    },
                    outputVar: 'var2'
                  },
                  {
                    actionType: 'ajax',
                    args: {
                      api: {
                        url: 'https://3xsw4ap8wah59.cfc-execute.bj.baidubce.com/api/awo-ui-mock/mock2/form/saveForm?name=${name}',
                        method: 'get'
                      },
                      messages: {
                        success: '请求3成功了！欧耶',
                        failed: '失败了呢。。'
                      }
                    },
                    outputVar: 'var3'
                  }
                ]
              },
              {
                actionType: 'toast',
                args: {
                  msg: 'var1:${event.data.var1|json}, var2:${event.data.var2|json}, var3:${event.data.var3|json}'
                }
              }
            ]
          }
        }
      }
    ]
  }
}
```

**其他属性**

| 属性名   | 类型                                                 | 默认值 | 说明   |
| -------- | ---------------------------------------------------- | ------ | ------ |
| children | Array<[动作](../../docs/concepts/event-action#动作)> | -      | 子动作 |

# 动作间数据传递

从事件触发开始，整个数据流包含事件本身产生的事件数据和动作产生的动作数据，事件源头产生的数据在 AMIS 事件动作机制底层已经自动加入渲染器数据域，可以通过`xxx`直接获取（`< 2.3.2 及以下版本 为 event.data.xxx`），而部分动作产生的数据如何流动需要交互设计者进行介入，对于数据流动可以通过数据映射，将上一个动作产生的数据作为动作参数写入下一个动作。

#### 传递数据

通过 `data` 指定输入的参数数据（`< 2.3.2 及以下版本`通过`args`传递数据），它是一个键值对。

```schema
{
  type: 'page',
  body: [
    {
      type: 'button',
      id: 'b_001',
      label: '发一个广播，携带动作参数',
      className: 'mb-2',
      level: 'primary',
      onEvent: {
        click: {
          actions: [
            {
              actionType: 'broadcast',
              args: {
                eventName: 'broadcast_1',
              },
              data: {
                name: 'lvxj',
                age: 18
              },
              description: '一个按钮的点击事件'
            }
          ]
        }
      }
    },
    {
      type: 'form',
      name: 'form1',
      id: 'form_001',
      title: '接收广播事件的参数',
      debug: true,
      body: [
        {
          type: 'input-text',
          id: 'form_001_text_01',
          label: '年龄',
          name: 'age',
          disabled: false,
          mode: 'horizontal'
        }
      ],
      data: {
        name: 'awo-ui'
      },
      onEvent: {
        broadcast_1: {
          actions: [
            {
              actionType: 'reload',
              data: {
                age: '${age}',
                name: '${name}'
              }
            }
          ]
        }
      }
    }
  ]
}
```

#### 引用 http 请求返回的数据

http 请求动作执行结束后，后面的动作可以通过 `${responseResult}`或`${{outputVar}}`来获取请求响应结果，响应结果的结构定义参考[发送 http 请求](../../docs/concepts/event-action#发送-http-请求)。

> `< 2.3.2 及以下版本 需要通过 ${event.data.{xxx}}`来获取以上信息，例如：${event.data.responseResult}

```schema
{
  type: 'page',
  body: [
    {
      type: 'button',
      label: '发送Ajax请求，并把返回数据传给弹窗',
      level: 'primary',
      onEvent: {
        click: {
          actions: [
            {
              actionType: 'ajax',
              args: {
                api: 'https://3xsw4ap8wah59.cfc-execute.bj.baidubce.com/api/awo-ui-mock/mock2/form/saveForm'
              }
            },
            {
              actionType: 'dialog',
              data: {
                id: '${event.data.responseResult.responseData.id}'
              },
              dialog: {
                type: 'dialog',
                id: 'dialog_005',
                title: '弹框标题1',
                data: {
                  id: '${id}'
                },
                body: [
                  {
                    type: 'form',
                    body: [
                      {
                        type: 'tpl',
                        tpl: '<p>请求返回的数据：id=${id}</p>',
                        inline: false
                      }
                    ]
                  }
                ]
              }
            }
          ]
        }
      }
    }
  ]
}
```

# 事件动作干预

事件动作干预是指执行完当前动作后，干预所监听事件默认处理逻辑和后续其他动作的执行。通过`preventDefault`、`stopPropagation`分别阻止监听事件默认行为和停止下一个动作执行。

## 阻止事件默认行为

有些组件内置了一些逻辑来帮助用户降低配置成本，但可能这些逻辑并不符合设计者的业务需求，这时可以通过`onEvent`来监听对应的事件，并通过`preventDefault`来阻止那些默认处理逻辑来达到想要的最终效果。更多示例请查看[阻止组件默认行为示例](../../../examples/action/prevent/form)。

```schema
{
  type: 'page',
  title: '弹窗确认后执行其他动作并阻止默认关闭',
  body: [
    {
      type: 'button',
      className: 'ml-2',
      label: '打开弹窗',
      level: 'primary',
      onEvent: {
        click: {
          actions: [
            {
              actionType: 'dialog',
              dialog: {
                type: 'dialog',
                title: '提示',
                id: 'dialog_001',
                data: {
                   myage: '22'
                },
                body: [
                  {
                    type: 'alert',
                    body: '输入Do not close，确认后将不关闭弹窗',
                    level: 'warning'
                  },
                  {
                    type: 'input-text',
                    name: 'command'
                  }
                ],
                onEvent: {
                  confirm: {
                    actions: [
                      {
                        actionType: 'toast',
                        args: {
                          msg: '不关闭'
                        },
                        preventDefault: 'command === "Do not close"'
                      }
                    ]
                  }
                }
              }
            }
          ]
        }
      }
    }
  ]
}
```

## 停止后续动作执行

通过`onEvent`可以对监听的事件配置一组动作，这些动作是顺序执行的，有时间设计者希望执行某个/些动作后就停止继续执行后面的动作，这时候可以通过`stopPropagation`来停止执行后面配置的所有动作。

```schema
{
  "type": "page",
  "title": "只执行3个动作中的前两个动作",
  "body": [
    {
      "type": "button",
      "label": "弹出2个提示",
      level: 'primary',
      "onEvent": {
        "click": {
          "actions": [
            {
              "actionType": "toast",
              args: {
                "msgType": 'info',
                "msg": '动作1'
              }
            },
            {
              "actionType": "toast",
              args: {
                "msgType": 'info',
                "msg": '动作2'
              },
              "stopPropagation": true
            },
            {
              "actionType": "toast",
              args: {
                "msgType": 'info',
                "msg": '动作3',
                "position": 'top-right'
              }
            }
          ]
        }
      }
    }
  ]
}
```

# 自定义组件接入事件动作

需求场景主要是想要自定义组件的内部事件暴露出去，能够通过对事件的监听来执行所需动作，并希望自定义组件自身的动作能够被其他组件调用。接入方法是通过`props.dispatchEvent`派发自身的各种事件，使其具备更灵活的交互设计能力；通过实现`doAction`方法实现其他组件对其专属动作的调用，需要注意的是，此处依赖内部的 `Scoped Context`来实现自身的注册，可以参考 [组件间通信](../../docs/extend/custom-react#组件间通信)。

# 属性表

| 属性名          | 类型                                        | 默认值  | 说明                                                                                                          |
| --------------- | ------------------------------------------- | ------- | ------------------------------------------------------------------------------------------------------------- |
| actionType      | `string`                                    | -       | 动作名称                                                                                                      |
| args            | `object`                                    | -       | 动作属性`{key:value}`，支持数据映射                                                                           |
| data            | `object`                                    | -       | 追加数据`{key:value}`，支持数据映射，如果是触发其他组件的动作，则该数据会传递给目标组件，`> 2.3.2 及以上版本` |
| dataMergeMode   | `string`                                    | 'merge' | 当配置了 data 的时候，可以控制数据追加方式，支持合并(`merge`)和覆盖(`override`)两种模式，`> 2.3.2 及以上版本` |
| preventDefault  | `boolean`\|[表达式](../concepts/expression) | false   | 阻止事件默认行为，`> 1.10.0 及以上版本支持表达式`                                                             |
| stopPropagation | `boolean`\|[表达式](../concepts/expression) | false   | 停止后续动作执行，`> 1.10.0 及以上版本支持表达式`                                                             |
| expression      | `boolean`\|[表达式](../concepts/expression) | -       | 执行条件，不设置表示默认执行                                                                                  |
| outputVar       | `string`                                    | -       | 输出数据变量名                                                                                                |
