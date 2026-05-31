---
title: Action Button
description:
type: 0
group: ⚙ Components
menuName: Action Button
icon:
order: 26
---

Action button is one of the primary methods for triggering page behaviors.

## Basic Usage

Here we simply implement an interaction where clicking a button opens a dialog.

```schema: scope="body"
{
  "label": "Dialog",
  "type": "button",
  "actionType": "dialog",
  "dialog": {
    "title": "Dialog",
    "body": "This is a simple dialog."
  }
}
```

## Styling

### Size

Configure `size` to display different sizes

```schema: scope="body"
{
  "type": "button-toolbar",
  "buttons": [
    {
      "type": "button",
      "label": "Default Size"
    },
    {
      "type": "button",
      "label": "Extra Small",
      "size": "xs"
    },
    {
      "type": "button",
      "label": "Small",
      "size": "sm"
    },
    {
      "type": "button",
      "label": "Medium",
      "size": "md"
    },
    {
      "type": "button",
      "label": "Large",
      "size": "lg"
    }
  ]
}
```

### Theme

You can configure `level` or `primary` to display different styles.

```schema: scope="body"
{
  "type": "button-toolbar",
  "buttons": [
    {
      "type": "button",
      "label": "Default"
    },
    {
      "type": "button",
      "label": "Primary",
      "level": "primary"
    },
    {
      "type": "button",
      "label": "Secondary",
      "level": "secondary"
    },
    {
      "type": "button",
      "label": "Info",
      "level": "info"
    },
    {
      "type": "button",
      "label": "Success",
      "level": "success"
    },
    {
      "type": "button",
      "label": "Warning",
      "level": "warning"
    },
    {
      "type": "button",
      "label": "Danger",
      "level": "danger"
    },
    {
      "type": "button",
      "label": "Light",
      "level": "light"
    },
    {
      "type": "button",
      "label": "Dark",
      "level": "dark"
    },
    {
      "type": "button",
      "label": "Link",
      "level": "link"
    }
  ]
}
```

### Icon

You can configure the `icon` property to display an icon on the button

```schema: scope="body"
{
  "label": "Dialog",
  "type": "button",
  "actionType": "dialog",
  "icon": "fa fa-plus",
  "dialog": {
    "title": "Dialog",
    "body": "This is a simple dialog."
  }
}
```

The icon can also be a URL address, for example

```schema: scope="body"
{
  "label": "Dialog",
  "type": "button",
  "actionType": "dialog",
  "icon": "https://suda.cdn.bcebos.com/images%2F2021-01%2Fdiamond.svg",
  "dialog": {
    "title": "Dialog",
    "body": "This is a simple dialog."
  }
}
```

If `label` is configured as an empty string, only the `icon` will be displayed

```schema: scope="body"
{
  "label": "",
  "type": "button",
  "actionType": "dialog",
  "icon": "fa fa-plus",
  "dialog": {
    "title": "Dialog",
    "body": "This is a simple dialog."
  }
}
```

## Pre-action Confirmation

You can configure `confirmText` to display a confirmation dialog before any action is performed.

```schema: scope="body"
{
    "label": "Ajax Request",
    "type": "button",
    "actionType": "ajax",
    "confirmText": "Are you sure you want to send this request?",
    "api": "/api/mock2/form/saveForm"
}
```

## Ajax Request

You can implement ajax requests by configuring `"actionType":"ajax"` and `api`.

```schema: scope="body"
{
    "label": "Ajax Request",
    "type": "button",
    "actionType": "ajax",
    "api": "/api/mock2/form/saveForm"
}
```

### Redirect to a Page After Successful Request

##### Configure Relative Path for Single Page Navigation

```schema: scope="body"
{
    "label": "Ajax Request",
    "type": "button",
    "actionType": "ajax",
    "api": "/api/mock2/form/saveForm",
    "redirect": "../docs/start/getting-started"
}
```

##### Configure Full Path for Direct Navigation

```schema: scope="body"
{
    "label": "Ajax Request",
    "type": "button",
    "actionType": "ajax",
    "api": "/api/mock2/form/saveForm",
    "redirect": "https://www.baidu.com/"
}
```

### Show Feedback Dialog After Successful Request

```schema: scope="body"
{
    "type": "button",
    "label": "Ajax Feedback Dialog",
    "actionType": "ajax",
    "api": "/api/mock2/form/saveForm",
    "feedback": {
        "title": "Operation Successful",
        "body": "${id} has been processed successfully"
    }
}
```

For more information, see the [Dialog Documentation](./dialog#feedback-%E5%8F%8D%E9%A6%88%E5%BC%B9%E6%A1%86)

### Refresh Target Component After Successful Request

1. The target component needs to have a `name` attribute configured
2. Add `"reload": "xxx"` to the Action, where `xxx` is the `name` attribute value of the target component. If configuring multiple components, separate `name` values with commas. Additionally, if you want to carry data during reload, you can configure it like `{"reload": "xxx?a=${a}&b=${b}"}`, which not only refreshes the target component but also passes the current environment data a and b to xxx.

```schema
{
  "type": "page",
  "body": [
    {
      "type": "button",
      "label": "Ajax Request",
      "actionType": "ajax",
      "api": "/api/mock2/form/saveForm",
      "reload": "crud"
    },
    {
      "type": "divider"
    },
    {
      "type": "crud",
      "name": "crud",
      "api": "/api/mock2/sample?waitSeconds=1",
      "columns": [
        {
            "name": "id",
            "label": "ID"
        },
        {
            "name": "engine",
            "label": "Rendering engine"
        },
        {
            "name": "browser",
            "label": "Browser"
        },
        {
            "name": "platform",
            "label": "Platform(s)"
        },
        {
            "name": "version",
            "label": "Engine version"
        },
        {
            "name": "grade",
            "label": "CSS grade"
        }
      ]
    }
  ]
}
```

> Configure `"reload": "window"` to refresh the current page

### Custom Toast Text

You can customize the toast messages returned by the interface by configuring `messages`

```schema: scope="body"
{
    "type": "button",
    "label": "Ajax Request",
    "actionType": "ajax",
    "api": "/api/mapping",
    "messages": {
        "success": "Success! Hooray",
        "failed": "Failed..."
    }
}
```

Note that if the API result returns a `msg` field, the API return will be used with priority.

**Property Table**

| Property | Type                                                                                     | Default | Description                                                                                                                                   |
| -------- | ---------------------------------------------------------------------------------------- | ------ | -------------------------------------------------------------------------------------------------------------------------------------- |
| api      | [Api](../../docs/types/api)                                                              | -      | Request URL, refer to [api](../../docs/types/api) format specification.                                                                                  |
| redirect | [Template String](../../docs/concepts/template#%E6%A8%A1%E6%9D%BF%E5%AD%97%E7%AC%A6%E4%B8%B2) | -      | Specifies the path to redirect to after the current request ends, can use `${xxx}` for value extraction.                                                                                     |
| feedback | `DialogObject`                                                                           | -      | For ajax type actions, when ajax returns normally, it can also pop up a dialog for other interactions. The returned data can be used in this dialog. Format can refer to [Dialog](./Dialog) |
| messages | `object`                                                                                 | -      | `success`: Prompt after successful ajax operation, can be unspecified, defaults to API return when not specified. `failed`: Ajax operation failure prompt.                                     |

## Download Request

> Version 1.4.0 and above

You can implement download requests by configuring `"actionType":"download"` and `api`. It's actually a special case of `ajax` that automatically adds `"responseType": "blob"` to the api.

```schema: scope="body"
{
    "label": "Download",
    "type": "action",
    "actionType": "download",
    "api": "/api/download"
}
```

The above example cannot be tested due to environment constraints. You need to configure `content-type` and `Content-Disposition` in the returned header, for example

```
Content-Type: application/pdf
Content-Disposition: attachment; filename="download.pdf"
```

If the interface has cross-domain issues, in addition to common cors headers, you also need to add the following header

```
Access-Control-Expose-Headers:  Content-Disposition
```

## Save to Local

> Version 1.10.0 and above

Similar to the download interface functionality above, but does not require returning a `Content-Disposition` header, only needs to solve cross-domain issues. Mainly used for simple scenarios, such as downloading text

```schema: scope="body"
{
    "label": "Save",
    "type": "action",
    "actionType": "saveAs",
    "api": "/api/download"
}
```

> This feature currently doesn't use the fetcher method in env and doesn't support POST

By default, it will automatically extract the filename from the url. If there isn't one, you need to specify it, for example

```schema: scope="body"
{
    "label": "Save",
    "type": "action",
    "actionType": "saveAs",
    "fileName": "Downloaded filename",
    "api": "/api/download"
}
```

## Countdown

Mainly used for verification code sending scenarios. By setting countdown `countDown` (in seconds), the button is disabled for a period of time after clicking:

```schema: scope="body"
{
  "type": "form",
  "body": [
    {
      "name": "phone",
      "type": "input-text",
      "required": true,
      "label": "Phone Number",
      "addOn": {
        "label": "Send Verification Code",
        "type": "button",
        "countDown": 60,
        "countDownTpl": "Resend in ${timeLeft} seconds",
        "actionType": "ajax",
        "api": "/api/mock2/form/saveForm?phone=${phone}"
      }
    }
  ]
}
```

You can also control the displayed text through `countDownTpl`, where the `${timeLeft}` variable is the remaining time.

## Navigation Links

### Single Page Navigation

```schema: scope="body"
{
    "label": "Go to Introduction Page",
    "type": "button",
    "level": "info",
    "actionType": "link",
    "link": "../docs/index"
}

```

**Property Table**

| Property     | Type     | Default | Description                                                                                                                |
| ---------- | -------- | ------ | ------------------------------------------------------------------------------------------------------------------- |
| actionType | `string` | `link` | Single page navigation                                                                                                            |
| link       | `string` | `link` | Specifies the navigation address. Unlike url, this is a single page navigation method that doesn't render the browser. Please specify pages within the awo-ui platform. Can use `${xxx}` for value extraction. |

### Direct Navigation

```schema: scope="body"
{
    "label": "Open Baidu",
    "type": "button",
    "level": "success",
    "actionType": "url",
    "url": "http://www.baidu.com"
}
```

**Property Table**

| Property     | Type      | Default  | Description                                             |
| ---------- | --------- | ------- | ------------------------------------------------ |
| actionType | `string`  | `url`   | Page navigation                                         |
| url        | `string`  | -       | After button click, will open the specified page. Can use `${xxx}` for value extraction. |
| blank      | `boolean` | `false` | If `true`, will open in a new tab page.              |

## Send Email

```schema: scope="body"
{
  "label": "Send Email",
  "type": "button",
  "actionType": "email",
  "to": "awo-ui@baidu.com",
  "cc": "baidu@baidu.com",
  "subject": "This is the email subject",
  "body": "This is the email content"
}
```

### Asynchronous Data Retrieval

```schema: scope="body"
{
  "type": "page",
  "initApi": "/api/mock2/mail/mailInfo",
  "body": {
    "label": "Send Email",
    "type": "button",
    "actionType": "email",
    "to": "${to}",
    "cc": "${cc}",
    "subject": "${subject}",
    "body": "${body}"
  }
}
```

**Property Table**

| Property     | Type     | Default  | Description                             |
| ---------- | -------- | ------- | -------------------------------- |
| actionType | `string` | `email` | Shows a popup after clicking             |
| to         | `string` | -       | Recipient email, can use ${xxx} for value extraction.   |
| cc         | `string` | -       | CC email, can use ${xxx} for value extraction.     |
| bcc        | `string` | -       | BCC email, can use ${xxx} for value extraction. |
| subject    | `string` | -       | Email subject, can use ${xxx} for value extraction.     |
| body       | `string` | -       | Email content, can use ${xxx} for value extraction.     |

## Dialog

```schema: scope="body"
{
  "label": "Dialog Form",
  "type": "button",
  "level": "primary",
  "actionType": "dialog",
  "dialog": {
    "title": "Form Settings",
    "body": {
      "type": "form",
      "api": "/api/mock2/form/saveForm",
      "body": [
        {
          "type": "input-text",
          "name": "text",
          "label": "Text"
        }
      ]
    }
  }
}
```

### Example of Dialog Combined with Reload to Refresh Dropdown

Below is a typical scenario where there's a dropdown, and a button that can open a dialog to add new data. After adding, the dropdown needs to re-fetch the latest list (this example doesn't show updates because the add functionality isn't implemented, but if you look at network requests, you'll see it makes a new request).

```schema: scope="body"
{
    "type": "form",
    "api": "/api/mock2/form/saveForm",
    "name": "myForm",
    "body": [
        {
          "type": "select",
          "name": "group",
          "label": "Group",
          "source": "/api/mock2/form/getOptions"
        },
        {
          "label": "Add Group",
          "type": "button",
          "level": "primary",
          "actionType": "dialog",
          "reload": "myForm.group",
          "dialog": {
            "title": "Add Group",
            "body": {
              "type": "form",
              "api": "/api/mock2/form/saveForm",
              "body": [
                {
                  "type": "input-text",
                  "name": "groupName",
                  "label": "Group Name"
                }
              ]
            }
          }
        }
    ]
}
```

You can see that `reload` is `myForm.group`, where the first part is the form's name, and the second part is the dropdown's name.

**Property Table**

| Property        | Type                       | Default   | Description                                          |
| ------------- | -------------------------- | -------- | --------------------------------------------- |
| actionType    | `string`                   | `dialog` | Shows a popup after clicking                          |
| dialog        | `string` or `DialogObject` | -        | Specifies dialog content, format can refer to [Dialog](./dialog)    |
| nextCondition | `boolean`                  | -        | Can be used to set conditions for the next data item, defaults to `true`. |

## Drawer

```schema: scope="body"
{
  "label": "Drawer Form",
  "type": "button",
  "actionType": "drawer",
  "drawer": {
    "title": "Form Settings",
    "body": {
      "type": "form",
      "api": "/api/mock2/form/saveForm?waitSeconds=1",
      "body": [
        {
          "type": "input-text",
          "name": "text",
          "label": "Text"
        }
      ]
    }
  }
}
```

**Property Table**

| Property     | Type                       | Default   | Description                                       |
| ---------- | -------------------------- | -------- | ------------------------------------------ |
| actionType | `string`                   | `drawer` | Shows a sidebar after clicking                       |
| drawer     | `string` or `DrawerObject` | -        | Specifies drawer content, format can refer to [Drawer](./drawer) |

## Copy Text

```schema: scope="body"
{
    "label": "Copy Text",
    "type": "button",
    "actionType": "copy",
    "content": "http://www.baidu.com"
}
```

You can set the copy format through `copyFormat`, default is text

```schema: scope="body"
{
    "label": "Copy Rich Text",
    "type": "button",
    "actionType": "copy",
    "copyFormat": "text/html",
    "content": "<a href='http://www.baidu.com'>link</a> <b>bold</b>"
}
```

**Property Table**

| Property     | Type                                 | Default | Description                                 |
| ---------- | ------------------------------------ | ------ | ------------------------------------ |
| actionType | `string`                             | `copy` | Copy content to clipboard                 |
| content    | [Template](../../docs/concepts/template) | -      | Specifies the content to copy. Can use `${xxx}` for value extraction. |

## Refresh Other Components

**Property Table**

| Property     | Type     | Default   | Description                                                                        |
| ---------- | -------- | -------- | --------------------------------------------------------------------------- |
| actionType | `string` | `reload` | Refresh target component                                                                |
| target     | `string` | -        | Target component name to refresh (component's `name` value, configured by yourself), separate multiple values with `,`. |

## Component-Specific Action Types

### Add a Row to Table in Form

This actionType is a specialized action for [FormItem-Table](./form/input-table#按钮触发新增行)

### Form Validation

In the form below, the form items included in the button's `required` attribute will be validated first. Only after all fields are validated will the form's inherent items be validated. Note that when the `required` in the button conflicts with the corresponding form item's own `required` attribute, the final validation method is `"required": true`.

```schema: scope="body"
{
    "type":"button",
    "label":"Open Dialog Form",
    "level": "primary",
    "actionType":"dialog",
    "dialog":{
        "type":"dialog",
        "title":"System Notification",
        "closeOnEsc": true,
        "body": [
            {
                "type":"form",
                "title":"Form",
                "api":"/api/mock2/form/saveForm",
                "body":[
                    {
                        "label":"字段a",
                        "type":"input-text",
                        "name":"a",
                        "required":true
                    },
                    {
                        "name":"b",
                        "label":"字段b",
                        "type":"input-text",
                        "validations": {
                          "minimum": 1,
                          "isNumeric": true,
                          "isInt": true
                        },
                        "required": false
                    },
                    {
                        "name":"c",
                        "label":"字段c",
                        "type":"input-text"
                    },
                    {
                        "name":"d",
                        "label":"字段d",
                        "type":"input-text",
                        "required":true
                    }
                ]
            }
        ],
        "actions":[
            {
                "type":"submit",
                "label":"提交-校验字段b",
                "actionType":"submit",
                "required":["b"],
                "level": "info"
            },
            {
                "type":"submit",
                "label":"提交-校验字段b, c",
                "actionType":"submit",
                "required":["b", "c"],
                "level": "info"
            }
        ]
    }
}
```

### 重置表单

在 form 中，配置`"type": "reset"`的按钮，可以实现重置表单数据的功能

```schema: scope="body"
{
    "type": "form",
    "api": "/api/mock2/form/saveForm",
    "body": [
        {
            "type": "input-text",
            "name": "username",
            "placeholder": "请输入用户名",
            "label": "用户名"
        },
        {
            "type": "input-password",
            "name": "password",
            "label": "密码",
            "placeholder": "请输入密码"
        },
        {
            "type": "checkbox",
            "name": "rememberMe",
            "option": "记住登录"
        }
    ],
    "actions": [
        {
            "type": "reset",
            "label": "重置"
        },
        {
            "type": "submit",
            "label": "提交",
            "level": "primary"
        }
    ]
}
```

### 清空表单

在 form 中，配置`"actionType": "clear"`的按钮，可以实现清空表单数据的功能，跟重置不同的是，重置其实是还原到初始值，并不一定是清空。

```schema: scope="body"
{
    "type": "form",
    "api": "/api/mock2/form/saveForm",
    "body": [
        {
            "type": "input-text",
            "name": "username",
            "placeholder": "请输入用户名",
            "label": "用户名",
            "value": "rick"
        },
        {
            "type": "input-password",
            "name": "password",
            "label": "密码",
            "placeholder": "请输入密码"
        },
        {
            "type": "checkbox",
            "name": "rememberMe",
            "option": "记住登录"
        }
    ],
    "actions": [
        {
            "type": "button",
            "actionType": "clear",
            "label": "清空"
        },
        {
            "type": "reset",
            "label": "重置"
        },
        {
            "type": "submit",
            "label": "提交",
            "level": "primary"
        }
    ]
}
```

### 重置表单并提交

`actionType` 配置成 `"reset-and-submit"`

### 清空表单并提交

`actionType` 配置成 `"clear-and-submit"`

## 自定义点击事件

> 1.3.0 版本新增功能

如果上面的的行为不满足需求，还可以通过字符串形式的 `onClick` 来定义点击事件，这个字符串会转成 JavaScript 函数，并支持异步（如果是用 sdk 需要自己编译一个 es2017 版本）。

```schema: scope="body"
{
    "label": "点击",
    "type": "button",
    "onClick": "alert('点击了按钮'); console.log(props);"
}
```

awo-ui 会传入两个参数 `event` 和 `props`，`event` 就是 React 的事件，而 `props` 可以拿到这个组件的其他属性，同时还能调用 amis 中的内部方法。

```schema: scope="body"
{
    "label": "点击",
    "type": "button",
    "onClick": "props.onAction(event, {actionType:'dialog', dialog: {title: '弹框', body: '这是代码调用的弹框'}})"
}
```

我们将前面的代码拿出来方便分析：

```javascript
// event 和 props 前面提到过，而 onAction 就是 awo-ui 内部的方法，可以用来调用其他 action，需要传递两个参数，一个是 event，另一个就是 action 类型及所需的参数。
props.onAction(event, {
  actionType: 'dialog',
  dialog: {title: '弹框', body: '这是代码调用的弹框'}
});
```

这个函数如果返回 `false` 就会阻止 awo-ui 其他 action 的执行，比如这个例子

```schema: scope="body"
{
  "label": "弹框",
  "type": "button",
  "actionType": "dialog",
  "onClick": "alert('点击按钮');",
  "dialog": {
    "title": "Dialog",
    "body": "This is a simple dialog."
  }
}
```

它的行为是先执行 alert，再执行弹框，但如果我们加上一个 `return false`，就会发现后面的 awo-ui 弹框不执行了。

```schema: scope="body"
{
  "label": "弹框",
  "type": "button",
  "actionType": "dialog",
  "onClick": "alert('点击按钮'); return false;",
  "dialog": {
    "title": "Dialog",
    "body": "This is a simple dialog."
  }
}
```

如果是在表单项中，还能通过 `props.formStore.setValues();` 来修改其它表单项值

```schema: scope="body"
{
  "type": "form",
  "api": "/api/mock2/form/saveForm",
  "body": [
    {
      "type": "input-text",
      "name": "name",
      "label": "姓名："
    },
    {
      "type": "input-text",
      "name": "email",
      "label": "邮箱："
    },
    {
      "label": "修改姓名",
      "name": "name",
      "type": "button",
      "onClick": "props.formStore.setValues({name: 'awo-ui', email: 'amis@baidu.com'});"
    }
  ]
}
```

## 全局键盘快捷键触发

> 1.3.0 版本新增功能

可以通过 `hotKey` 属性来配置键盘快捷键触发，比如下面的例子

```schema: scope="body"
{
  "label": "使用 ⌘+o 或 ctrl+o 来弹框",
  "type": "button",
  "hotKey": "command+o,ctrl+o",
  "actionType": "dialog",
  "dialog": {
    "title": "Dialog",
    "body": "This is a simple dialog."
  }
}
```

除了 ctrl 和 command 还支持 shift、alt。

其它键盘特殊按键的命名列表：backspace, tab, clear, enter, return, esc, escape, space, up, down, left, right, home, end, pageup, pagedown, del, delete, f1 - f19, num_0 - num_9, num_multiply, num_add, num_enter, num_subtract, num_decimal, num_divide。

> 注意这个主要用于实现页面级别快捷键，如果要实现回车提交功能，请将 `input-text` 放在 `form` 里，而不是给 button 配一个 `enter` 的快捷键。

## Action 作为容器组件

> 1.5.0 及以上版本

action 还可以使用 `body` 来渲染其他组件，让那些不支持行为的组件支持点击事件，比如下面的例子

```schema: scope="body"
[{
  "type": "action",
  "body": [{
    "type": "color",
    "value": "#108cee"
  }],
  "actionType": "dialog",
  "dialog": {
    "title": "Dialog",
    "body": "This is a simple dialog."
  }
},{
  "type": "action",
  "body": {
    "type": "image",
    "src": "https://internal-awo-ui-res.cdn.bcebos.com/images/2020-1/1578395692722/4f3cb4202335.jpeg@s_0,w_216,l_1,f_jpg,q_80"
  },
  "tooltip": "Click to show dialog",
  "actionType": "dialog",
  "dialog": {
    "title": "Dialog",
    "body": "This is a simple dialog."
  }
}]
```

In this mode, button configuration options like `label`, `size`, `icon` etc. are not supported, because it only serves as a container component without presentation.

## Button Tooltip

Set tooltip through `tooltip`

```schema: scope="body"
{
  "label": "弹框",
  "type": "button",
  "actionType": "link",
  "link": "../index",
  "tooltip": "Click link to navigate"
}
```

If the button is disabled, you need to use `disabledTip`

```schema: scope="body"
{
  "label": "弹框",
  "type": "button",
  "actionType": "link",
  "disabled": true,
  "link": "../index",
  "disabledTip": "Disabled"
}
```

You can also set the popup position through `tooltipPlacement`

```schema: scope="body"
{
  "label": "弹框",
  "type": "button",
  "actionType": "link",
  "link": "../index",
  "tooltipPlacement": "right",
  "tooltip": "Click link to navigate"
}
```

## 通用属性表

所有`actionType`都支持的通用配置项

| 属性名             | 类型                                 | 默认值      | 说明                                                                                                                                                                        |
| ------------------ | ------------------------------------ | ----------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| type               | `string`                             | `action`    | 指定为 Page 渲染器。                                                                                                                                                        |
| actionType         | `string`                             | -           | 【必填】这是 action 最核心的配置，来指定该 action 的作用类型，支持：`ajax`、`link`、`url`、`drawer`、`dialog`、`confirm`、`cancel`、`prev`、`next`、`copy`、`close`。       |
| label              | `string`                             | -           | 按钮文本。可用 `${xxx}` 取值。                                                                                                                                              |
| level              | `string`                             | `default`   | 按钮样式，支持：`link`、`primary`、`secondary`、`info`、`success`、`warning`、`danger`、`light`、`dark`、`default`。                                                        |
| size               | `string`                             | -           | 按钮大小，支持：`xs`、`sm`、`md`、`lg`。                                                                                                                                    |
| icon               | `string`                             | -           | 设置图标，例如`fa fa-plus`。                                                                                                                                                |
| iconClassName      | `string`                             | -           | 给图标上添加类名。                                                                                                                                                          |
| rightIcon          | `string`                             | -           | 在按钮文本右侧设置图标，例如`fa fa-plus`。                                                                                                                                  |
| rightIconClassName | `string`                             | -           | 给右侧图标上添加类名。                                                                                                                                                      |
| active             | `boolean`                            | -           | 按钮是否高亮。                                                                                                                                                              |
| activeLevel        | `string`                             | -           | 按钮高亮时的样式，配置支持同`level`。                                                                                                                                       |
| activeClassName    | `string`                             | `is-active` | 给按钮高亮添加类名。                                                                                                                                                        |
| block              | `boolean`                            | -           | 用`display:"block"`来显示按钮。                                                                                                                                             |
| confirmText        | [模板](../../docs/concepts/template) | -           | 当设置后，操作在开始前会询问用户。可用 `${xxx}` 取值。                                                                                                                      |
| reload             | `string`                             | -           | 指定此次操作完后，需要刷新的目标组件名字（组件的`name`值，自己配置的），多个请用 `,` 号隔开。                                                                               |
| tooltip            | `string`                             | -           | 鼠标停留时弹出该段文字，也可以配置对象类型：字段为`title`和`content`。可用 `${xxx}` 取值。                                                                                  |
| disabledTip        | `string`                             | -           | 被禁用后鼠标停留时弹出该段文字，也可以配置对象类型：字段为`title`和`content`。可用 `${xxx}` 取值。                                                                          |
| tooltipPlacement   | `string`                             | `top`       | 如果配置了`tooltip`或者`disabledTip`，指定提示信息位置，可配置`top`、`bottom`、`left`、`right`。                                                                            |
| close              | `boolean` or `string`                | -           | 当`action`配置在`dialog`或`drawer`的`actions`中时，配置为`true`指定此次操作完后关闭当前`dialog`或`drawer`。当值为字符串，并且是祖先层弹框的名字的时候，会把祖先弹框关闭掉。 |
| required           | `Array<string>`                      | -           | 配置字符串数组，指定在`form`中进行操作之前，需要指定的字段名的表单项通过验证                                                                                                |

## 事件表

当前组件会对外派发以下事件，可以通过`onEvent`来监听这些事件，并通过`actions`来配置执行的动作，详细查看[事件动作](../../docs/concepts/event-action)。

| Event Name   | Event Parameters                               | Description           |
| ---------- | -------------------------------------- | -------------- |
| click      | `nativeEvent: MouseEvent` Mouse event object | Triggered on click     |
| mouseenter | `nativeEvent: MouseEvent` Mouse event object | Triggered on mouse enter |
| mouseleave | `nativeEvent: MouseEvent` Mouse event object | Triggered on mouse leave |

## Action Table

None
