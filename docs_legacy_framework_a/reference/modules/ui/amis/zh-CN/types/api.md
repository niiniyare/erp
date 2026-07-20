> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: API
description:
type: 0
group: 🔧 类型
menuName: API
icon:
order: 20
---

API 类型用于配置请求接口的格式，涉及请求方式、请求地址、请求数据体等等相关配置

## 简单配置

如果你只需要配置简单的 ajax 接口，可以使用简单字符串格式，如下：

```
[<method>:]<url>
```

- **method**：get、post、put、delete，默认为 get
- **url**：接口地址，即模板字符串

示例：

```json
{
  "api": "get:/api/initData", // get 请求
  "api": "post:/api/initData", // post 请求
  "api": "put:/api/initData", // put 请求
  "api": "delete:/api/initData" // delete 请求
}
```

## 接口返回格式（重要）

所有配置在 amis 组件中的接口，都要符合下面的返回格式

```json
{
  "status": 0,
  "msg": "",
  "data": {
    ...其他字段
  }
}
```

- **status**: 返回 `0`，表示当前接口正确返回，否则按错误请求处理；
- **msg**: 返回接口处理信息，主要用于表单提交或请求失败时的 `toast` 显示；
- **data**: 必须返回一个具有 `key-value` 结构的对象。

**`status`**、**`msg`** 和 **`data`** 字段为接口返回的必要字段。

> 1.1.7

为了方便更多场景使用，还兼容了以下这些错误返回格式：

1. errorCode 作为 status、errorMessage 作为 msg
2. errno 作为 status、errmsg/errstr 作为 msg
3. error 作为 status、errmsg 作为 msg
4. error.code 作为 status、error.message 作为 msg
5. message 作为 msg

### 正确的格式

```json
{
  "status": 0,
  "msg": "",
  "data": {
    // 正确
    "text": "World!"
  }
}
```

### 错误的格式

直接返回字符串或者数组都是不推荐的

```json
{
  "status": 0,
  "msg": "",
  "data": "some string" // 错误，使用 key 包装
}
```

### 兼容格式

> 1.0.19 及以上版本。

为了支持多种后端，amis 支持直接返回数据的方式，无需返回 status 和将数据放在 data 字段中，比如下面的例子：

```json
{
  "username": "amis",
  "email": "amis@amis.com"
}
```

但这种方式无法显示错误信息，只能通过返回 http 状态码来标识错误。

### 配置弹框时间

可以通过 `msgTimeout` 控制弹框时间，它的时间是毫秒

```json
{
  "status": 2,
  "msg": "error",
  "msgTimeout": 10000,
  "data": {}
}
```

## 复杂配置

API 还支持配置对象类型

### 基本用法

```json
{
    "api": {
        "method": "get",
        "url": "xxxx",
        "data": {
            xxx
        },
        ... // 其他配置
    }
}
```

### 配置请求方式

可以配置`method`指定接口的请求方式，支持：`get`、`post`、`put`、`delete`、`patch`。

> `method`值留空时：
>
> - 在初始化接口中，默认为`get`请求
> - 在`form`提交接口，默认为`post`请求

### 配置请求地址

可以配置`url`指定接口请求地址，支持[模板字符串](../concepts/template)。

### 配置请求数据

可以配置`data`，配置自定义接口请求数据体。

```schema: scope="body"
{
    "type": "form",
    "api": {
        "method": "post",
        "url": "/api/mock2/form/saveForm",
        "data": {
            "myName": "${name}",
            "myEmail": "${email}"
        }

    },
    "body": [
      {
        "type": "input-text",
        "name": "name",
        "label": "姓名："
      },
      {
        "name": "email",
        "type": "input-email",
        "label": "邮箱："
      }
    ]
}
```

支持[数据映射](../concepts/data-mapping)

> 当`method`配置为`get`时，`data`中的值默认会添加到请求路径中

需要注意一下，配置了数据发送，默认如果值是 `undefined` 也会作为空字符发送，比如以上这个例子直接提交会发送, name 和 email 两个字段，值分别为空字符。由于历史原因这个已经不能再修改了。如果你想实现没有填写的值不发送，则需要配置成。

```
"data": {
  "myName": "${name|default:undefined}",
  "myEmail": "${email|default:undefined}"
}
```

这样 `undefined` 的值不会发送了。

### 不处理 key 中的路径

> since 1.5.0

默认请求数据体中配置 key，如果带路径会自动转成对象如：

```
"api": {
    "method": "post",
    "url": "/api/mock2/form/saveForm",
    "data": {
        "a.b": "${name}",
        "c[d]": "${name}"
    }

}
```

最终发送出去的数据格式为

```
{
  a: {
    b: "xxx"
  },
  c: {
    d: "xxx"
  }
}
```

如果数据映射中的 key 不想被处理路径则需要配置 `convertKeyToPath` 为 false 如：

```
"api": {
    "method": "post",
    "url": "/api/mock2/form/saveForm",
    "convertKeyToPath": false,
    "data": {
        "a.b": "${name}",
        "c[d]": "${name}"
    }

}
```

这样发送的数据格式为

```
{
  "a.b": "xxx",
  "c[d]": "xxx"
}
```

### 配置请求数据格式

可以配置`dataType`，来指定请求的数据体格式，默认为`json`

> 下面例子中 api 没有配置`data`属性，因为`form`会默认将所有表单项的值进行提交。

#### application/json

默认是`application/json`，不需要额外配置。

> 注意：当数据域里的 key 为 `&` 且值为 `$$` 时, 将所有原始数据打平设置到 `data` 中.

```schema: scope="body"
{
    "type": "form",
    "title": "默认JSON格式",
    "api": {
        "method": "post",
        "url": "/api/mock2/form/saveForm",
        "data":{
            "&": "$$$$" // 获取表单数据域中的所有值
        }
    },
    "body": [
      {
        "type": "input-text",
        "name": "name",
        "label": "姓名："
      },
      {
        "name": "email",
        "type": "input-email",
        "label": "邮箱："
      }
    ]
  }
```

#### application/x-www-form-urlencoded

配置`"dataType": "form"`，可配置发送体格式为`application/x-www-form-urlencoded`

```schema: scope="body"
{
    "type": "form",
    "api": {
        "method": "post",
        "url": "/api/mock2/form/saveForm",
        "dataType": "form"
    },
    "body": [
      {
        "type": "input-text",
        "name": "name",
        "label": "姓名："
      },
      {
        "name": "email",
        "type": "input-email",
        "label": "邮箱："
      }
    ]
  }
```

#### multipart/form-data

配置`"dataType": "form-data"`，可配置发送体格式为`multipart/form-data`

```schema: scope="body"
{
    "type": "form",
    "api": {
        "method": "post",
        "url": "/api/mock2/form/saveForm",
        "dataType": "form-data"
    },
    "body": [
      {
        "type": "input-text",
        "name": "name",
        "label": "姓名："
      },
      {
        "name": "email",
        "type": "input-email",
        "label": "邮箱："
      }
    ]
  }
```

当表单项中文件类型数据，则自动使用`multipart/form-data`数据体

```schema: scope="body"
{
    "type": "form",
    "api": {
        "method": "post",
        "url": "/api/mock2/form/saveForm"
    },
    "body": [
      {
        "type": "input-text",
        "name": "name",
        "label": "姓名："
      },
      {
        "name": "file",
        "type": "input-file",
        "label": "附件：",
        "asBlob": true
      }
    ]
  }
```

> `asBlob`配置项会指定当前 File 控件不再自己上传了，而是直接把文件数据作为表单项的值，文件内容会在 Form 表单提交的接口里面一起带上。

### 配置自定义请求头

可以配置`headers`对象，添加自定义请求头

```schema: scope="body"
{
    "type": "form",
    "api": {
        "method": "post",
        "url": "/api/mock2/form/saveForm",
        "headers": {
            "my-header": "aaa"
        }
    },
    "body": [
      {
        "type": "input-text",
        "name": "name",
        "label": "姓名："
      },
      {
        "name": "email",
        "type": "input-email",
        "label": "邮箱："
      }
    ]
  }
```

### 配置请求条件

可以配置表达式`sendOn`来实现：当符合某个条件的情况下，接口才触发请求

```schema: scope="body"
{
    "title": "",
    "type": "form",
    "mode": "horizontal",
    "body": [
      {
        "label": "选项1",
        "type": "radios",
        "name": "a",
        "inline": true,
        "options": [
          {
            "label": "选项A",
            "value": 1
          },
          {
            "label": "选项B",
            "value": 2
          },
          {
            "label": "选项C",
            "value": 3
          }
        ]
      },
      {
        "label": "选项2",
        "type": "select",
        "size": "sm",
        "name": "b",
        "source": {
            "method": "get",
            "url": "/api/mock2/options/level2?a=${a}",
            "sendOn": "this.a === 2"
        },
        "description": "只有<code>选项1</code>选择<code>B</code>的时候，才触发<code>选项2</code>的<code>source</code>接口重新拉取"
      }
    ],
    "actions": []
}
```

查看 **选项 2** 的`source`属性，他是 API 类型值，支持配置`sendOn` [表达式](../concepts/expression)，实现根据条件请求接口。

### 配置接口缓存

当你在某种情况下，需要非常频繁的请求同一个接口，例如列表中，每一行中都有一个 Service 进行数据拉取操作，

```schema: scope="body"
{
    "type": "crud",
    "api": "/api/mock2/sample?waitSeconds=1",
    "columns": [
        {
            "name": "id",
            "label": "ID"
        },
        {
            "type": "service",
            "label": "数据",
            "api": "/api/mock2/page/initData",
            "body": {
                "type": "tpl",
                "tpl": "当前日期是：${date}"
            }
        }
    ]
}
```

如上，如果你打开浏览器网络面板，会发现`/api/mock2/page/initData`接口将重复请求多次，次数为你当前列表的数据条数。

这往往并不理想，你可以设置`cache`来设置缓存时间，单位是毫秒，在设置的缓存时间内，同样的请求将不会重复发起，而是会获取缓存好的请求响应数据。

```schema: scope="body"
{
    "type": "crud",
    "api": "/api/mock2/sample?waitSeconds=1",
    "columns": [
        {
            "name": "id",
            "label": "ID"
        },
        {
            "type": "service",
            "label": "数据",
            "api": {
                "method": "get",
                "url": "/api/mock2/page/initData",
                "cache": 2000
            },
            "body": {
                "type": "tpl",
                "tpl": "当前日期是：${date}"
            }
        }
    ]
}
```

这下我们再打开网络面板，发现只有一条请求了

### 配置返回数据

如果接口返回的数据结构不符合预期，可以通过配置 `responseData`来修改，同样支持[数据映射](../concepts/data-mapping)，可用来映射的数据为接口的实际数据（接口返回的 `data` 部分），额外加 `api` 变量。其中 `api.query` 为接口发送的 query 参数，`api.body` 为接口发送的内容体原始数据。

> 注意：当数据域里的 key 为 `&` 且值为 `$$` 时, 表示将所有原始数据打平设置到 `data` 中.

```json
{
  "type": "page",
  "initApi": {
    "method": "get",
    "url": "/api/xxx",
    "responseData": {
      "&": "$$",
      "first": "${items|first}"
    }
  }
}
```

假如接口实际返回为：

```json
{
  "status": 0,
  "msg": "",
  "data": {
    "items": [{"a": 1}, {"a": 2}]
  }
}
```

经过映射，给组件的数据为：

```json
{
  "items": [{"a": 1}, {"a": 2}],
  "first": {"a": 1}
}
```

另一个常用示例是 `"type": "select"` 的 `source` 数据源，如果接口返回的是：

```json
{
  "data": [
    {
      "myLabel": "lab",
      "myValue": 1
    }
  ]
}
```

而 select 所需的数据格式是 `[{"label": "lab", "value": 1}]`，如何进行映射？

方法是

```json
{
  "type": "select",
  "source": {
    "method": "get",
    "url": "http://xxx",
    "responseData": {
      "options": "${items|pick:label~myLabel,value~myValue}"
    }
  }
}
```

需要注意上面例子中 `items` 是因为数据直接放在了 `data` 中，如果是放在其他字段中就换成对应的字段名。

### 配置请求适配器

amis 的 API 配置，如果无法配置出你想要的请求结构，那么可以配置`requestAdaptor`发送适配器

**发送适配器** 是指在接口请求前，对请求进行一些自定义处理，例如修改发送数据体、添加请求头、等等，基本用法是，获取暴露的`api`参数，并且对该参数进行一些修改，并`return`出去：

##### 暴露的参数

发送适配器暴露以下参数以供用户进行操作:

- **api**：当前请求的 api 对象，一般包含下面几个属性：
  - url：当前接口 Api 地址
  - method：当前请求的方式
  - data：请求的数据体, 注意当请求方式为 `get` 时，`data` 在传入适配器时会被删除，请通过 query 读取，或者修改
  - query：请求的查询参数，所有请求参数都会被合并到 query 中，包含 data 参数和 url 参数
  - headers：请求的头部信息
  - context: 发送请求时的上下文数据
- **context** 发送请求时的上下文数据

> 6.5.0 版本开始支持在发送适配中修改 query 参数，之前的版本只能修改 url 参数。

##### 字符串形式

如果在 JSON 文件中配置的话，`requestAdaptor`只支持字符串形式。

用法示例：

```schema: scope="body"
{
    "type": "form",
    "api": {
        "method": "post",
        "url": "/api/mock2/form/saveForm",
        "requestAdaptor": "console.log(context); // 打印上下文数据\nreturn {\n    ...api,\n    data: {\n        ...api.data,    // 获取暴露的 api 中的 data 变量\n        foo: 'bar'      // 新添加数据\n    }\n}"
    },
    "body": [
      {
        "type": "input-text",
        "name": "name",
        "label": "姓名："
      },
      {
        "name": "email",
        "type": "input-email",
        "label": "邮箱："
      }
    ]
  }
```

上例中的适配器实际上是如下代码的字符串形式：

```js
// 进行一些操作

console.log(context); // 打印上下文数据

// 一定要将调整后的 api 对象 return 出去
return {
  ...api,
  data: {
    ...api.data, // 获取暴露的 api 中的 data 变量
    foo: 'bar' // 新添加数据
  }
};
```

字符串形式的适配器代码最后会自动包裹成函数，你只需要补充内部的函数实现，并将修改好的 `api` 对象 `return` 出去：

```js
function (api, context) {
  // 你的适配器代码在这里
}
```

##### 函数形式

如果你的使用环境为 js 文件，则可以直接传入函数，如下：

```js
const schema = {
  type: 'form',
  api: {
    method: 'post',
    url: '/api/mock2/form/saveForm',
    requestAdaptor: function (api, context) {
      console.log(context); // 打印上下文数据

      return {
        ...api,
        data: {
          ...api.data, // 获取暴露的 api 中的 data 变量
          foo: 'bar' // 新添加数据
        }
      };
    }
  },
  body: [
    {
      type: 'input-text',
      name: 'name',
      label: '姓名：'
    },
    {
      name: 'text',
      type: 'input-email',
      label: '邮箱：'
    }
  ]
};
```

上面例子中，我们获取暴露的`api`对象中的`data`变量，并且为其添加了一个新的字段`foo`，并且一起返回出去就可以了，这样我们的请求数据体中就会加上我们这个新的字段。

你也可以使用`debugger`自行进行调试。

#### 拦截请求

如果 api 发送适配器中，修改 api 对象，在 api 对象里面放入 `mockResponse` 属性，则会拦截请求发送，amis 内部会直接使用 `mockResponse` 的结果返回。

```js
const schema = {
  type: 'form',
  api: {
    method: 'post',
    url: '/api/mock2/form/saveForm',
    requestAdaptor: function (api, context) {
      return {
        // 模拟 http 请求返回
        mockResponse: {
          status: 200, // http 返回状态
          data: {
            // http 返回结果
            status: 0, // amis 返回数据的状态
            data: {
              name: '模拟返回的值'
            }
          }
        }
      };
    }
  },
  body: [
    {
      type: 'input-text',
      name: 'name',
      label: '姓名：'
    },
    {
      name: 'text',
      type: 'input-email',
      label: '邮箱：'
    }
  ]
};
```

### 配置接收适配器

同样的，如果后端返回的响应结构不符合 amis 的[接口格式要求](#%E6%8E%A5%E5%8F%A3%E8%BF%94%E5%9B%9E%E6%A0%BC%E5%BC%8F-%E9%87%8D%E8%A6%81-)，而后端不方便调整时，可以配置`adaptor`实现接收适配器

**接收适配器** 是指在接口请求后，对响应进行一些自定义处理，例如修改响应的数据结构、修改响应的数据等等。

例如：接口正确返回的格式中，会返回`"code": 200`，而 amis 中，接口返回格式需要`"status": 0`，这时候就需要接收适配器进行调整结构。

##### 暴露的参数

接收适配器器暴露以下参数以供用户进行操作:

- **payload**：当前请求的响应 payload，即 response.data
- **response**：当前请求的原始响应
- **api**：api 上的配置项，还可以通过 `api.data` 获得数据域里的内容
- **context** 发送请求时的上下文数据

##### 字符串形式

如果在 JSON 文件中配置的话，`adaptor`只支持字符串形式。

用法示例：

```schema: scope="body"
{
  "type": "form",
  "api": {
    "method": "post",
    "url": "/api/mock2/form/saveForm",
    "adaptor": "console.log(context); // 打印上下文数据 \nreturn {\n    ...payload,\n    status: payload.code === 200 ? 0 : payload.code\n}"
  },
  "body": [
    {
      "type": "input-text",
      "name": "name",
      "label": "姓名："
    },
    {
      "name": "file",
      "type": "input-file",
      "label": "附件：",
      "asBlob": true
    }
  ]
}
```

上例中的适配器实际上是如下代码的字符串形式：

```js
// 进行一些操作

// 一定要将调整后的 payload 对象 return 出去
return {
  ...payload,
  status: payload.code === 200 ? 0 : payload.code
};
```

字符串形式的适配器代码最后会自动包裹成函数，你只需要补充内部的函数实现，并将修改好的 `payload` 对象 `return` 出去：

```js
function (payload, response, api, context) {
  // 你的适配器代码在这里
}
```

##### 函数形式

如果你的使用环境为 js 文件，则可以直接传入函数，如下：

```js
const schema = {
  type: 'form',
  api: {
    method: 'post',
    url: '/api/mock2/form/saveForm',
    adaptor: function (payload, response, api, context) {
      console.log(context); // 打印上下文数据
      return {
        ...payload,
        status: payload.code === 200 ? 0 : payload.code
      };
    }
  },
  body: [
    {
      type: 'input-text',
      name: 'name',
      label: '姓名：'
    },
    {
      name: 'email',
      type: 'input-email',
      label: '邮箱：'
    }
  ]
};
```

### 配置文件下载

如果 API 返回的是文件下载，需要加上这个配置：

```json
{
    "api": {
        ... // 其他配置
        "responseType": "blob"
    }
}
```

比如点一个按钮下载的完整示例是：

```json
{
  "type": "button",
  "actionType": "ajax",
  "api": {
    "method": "post",
    "url": "/api",
    "responseType": "blob"
  }
}
```

还需要在这个 `/api` 返回的 header 中配置 `content-type` 和 `Content-Disposition`，比如

```
Content-Type: application/pdf
Content-Disposition: attachment; filename="download.pdf"
```

如果只有 `Content-Type`，比如 Excel 的 `application/vnd.openxmlformats-officedocument.spreadsheetml.sheet`，则应该使用[页面跳转](../../components/action#直接跳转)的方式来实现下载。

如果是跨域请求，还需要配置

```
Access-Control-Expose-Headers: Content-Disposition
```

如果自己覆盖了 `fetcher` 函数，需要有类似如下代码，具体可以参考 `embed.tsx` 里的实现

```javascript
let response = await axios(config);
response = await attachmentAdpator(response, __);
```

### 配置提示信息

可以通过`messages`自定义接口请求提示信息。

```schema: scope="body"
{
    "type": "form",
    "api": {
        "method": "post",
        "url": "/api/mock2/form/saveForm",
        "data": {
            "myName": "${name}",
            "myEmail": "${email}"
        },
        "messages": {
          "success": '好耶，成功了！',
          "failed": '糟糕，失败了！'
        }

    },
    "body": [
      {
        "type": "input-text",
        "name": "name",
        "label": "姓名："
      },
      {
        "name": "email",
        "type": "input-email",
        "label": "邮箱："
      }
    ]
}
```

### replaceData

返回的数据是否替换掉当前的数据，默认为 `false`（即追加），设置为`true`就是完全替换当前数据。

## 自动刷新

凡是拉取类接口，默认都会开启自动刷新，比如 form 配置 initApi: `/api/initForm?tpl=${tpl}`。这个接口会在 form 初始化的请求。当接口中有参数时，amis 会监控这个接口的实际结果是否有变化，假如这个时候 form 里面有个字段名为 tpl 的表单项，它的值发生变化，这个 form 的 initApi 又会请求一次。

```schema:scope="body"
{
  "type": "form",
  "mode": "horizontal",
  "title": "监听表单内部的修改",
  "initApi": "/api/mock2/form/initData?tpl=${tpl}",
  "actions": [],
  "body": [
    {
      "label": "数据模板",
      "type": "select",
      "labelClassName": "text-muted",
      "name": "tpl",
      "value": "tpl1",
      "inline": true,
      "options": [
        {
          "label": "模板1",
          "value": "tpl1"
        },
        {
          "label": "模板2",
          "value": "tpl2"
        },
        {
          "label": "模板3",
          "value": "tpl3"
        }
      ],
      "description": "<span class=\"text-danger\">请修改这里看效果</span>"
    },
    {
      "label": "名称",
      "type": "static",
      "labelClassName": "text-muted",
      "name": "name"
    },
    {
      "label": "作者",
      "type": "static",
      "labelClassName": "text-muted",
      "name": "author"
    },
    {
      "label": "请求时间",
      "type": "static-datetime",
      "labelClassName": "text-muted",
      "format": "YYYY-MM-DD HH:mm:ss",
      "name": "date"
    }
  ]
}
```

这个功能是自动开启的，直接配置 api 地址（`/api/xxx?xx=${xxx}`），或者对象形式配置 `{method: 'get', url: '/api/xxx?xx=${xxx}'}` 都会自动启动这个功能。如果想要关闭这个功能，有两种方式。

1. 把依赖的数据写在对象形式的 data 里面比如 `{method: 'get', url: '/api/xxx', data: {'xx': "${xxx}"}}`。
2. 对象形式添加 `autoRefresh: false` 属性。

【重点】利用这个 feature 结合 `sendOn` 接口发送条件，可以做到接口的串行加载。比如，接口 2 的地址上写上接口 1 会返回的某个字段，然后配置接口 2 的发送条件为这个字段必须存在时。这样接口 2 就会等接口 1 完了才会加载。

## 跟踪数据自动刷新

> since 1.1.6

之前的版本，配置的 api 默认就会具备自动刷新功能，除非显式的配置 `autoRefresh: false` 来关闭。自动刷新主要通过跟踪 api 的 url 属性来完成的，如果 url 中了使用了某个变量，而这个变量发生变化则会触发自动刷新。
也就说这个 url 地址，既能控制 api 请求的 query 参数，同时又起到跟踪变量重新刷新接口的作用。这个设定大部分情况下都是合理的，但是有时候想要跟踪 url 参数以外的变量就做不到了。所以新增了此属性 `trackExpression`，显式的配置需要跟踪的变量如：

> 如果`trackExpression` 追踪的数据是**对象数据**，可以使用[数据映射](../../docs/concepts/data-mapping)的`json`方法将数据序列化之后再比较，例如`"trackExpression": "${fieldToTrack|json}"`

```schema: scope="body"
{
    "title": "",
    "type": "form",
    "mode": "horizontal",
    "body": [
      {
        "label": "选项1",
        "type": "radios",
        "name": "a",
        "inline": true,
        "options": [
          {
            "label": "选项A",
            "value": 1
          },
          {
            "label": "选项B",
            "value": 2
          },
          {
            "label": "选项C",
            "value": 3
          }
        ]
      },
      {
        "label": "选项2",
        "type": "select",
        "size": "sm",
        "name": "b",
        "source": {
          "method": "get",
          "url": "/api/mock2/options/level2",
          "trackExpression": "${a}"
        },
        "description": "切换<code>选项1</code>的值，会触发<code>选项2</code>的<code>source</code> 接口重新拉取"
      }
    ],
    "actions": []
}
```

## GraphQL

1.7.0 及之前的版本需要通过配置 `data` 里的 `query` 和 `variables` 字段可以实现 GraphQL 查询

```schema: scope="body"
{
  "type": "form",
  "api": {
    "method": "post",
    "url": "/api/mock2/form/saveForm",
    "data": {
      "query": "mutation AddUser($name: String!, $email: String!) { \
        insert_user(object: { title: $title, email: $email }) { \
          title \
          email \
        } \
      }",
      "variables": {
         "name": "${name}",
         "email": "${email}"
      }
    }
  },
  "body": [
    {
      "type": "input-text",
      "name": "name",
      "label": "姓名："
    },
    {
      "name": "email",
      "type": "input-email",
      "label": "邮箱："
    }
  ]
}
```

1.8.0 及以上版本简化了 GraphQL 的支持，增加了 `graphql` 属性，如果配置了就会自动并自动将 data 当成 `variables`，上面的例子可以简化为下面的写法，除了简化之外还方便了可视化编辑器编辑

```schema: scope="body"
{
  "type": "form",
  "api": {
    "method": "post",
    "url": "/api/mock2/form/saveForm",
    "graphql": "mutation AddUser($name: String!, $email: String!) { \
        insert_user(object: { name: $name, email: $email }) { \
          name \
          email \
        } \
    }"
  },
  "body": [
    {
      "type": "input-text",
      "name": "name",
      "label": "姓名："
    },
    {
      "name": "email",
      "type": "input-email",
      "label": "邮箱："
    }
  ]
}
```

如果设置了 `data` 会被当成 `variables`，比如在 CRUD 里设置分页参数，比如下面的例子

```json
{
  "type": "crud",
  "api": {
    "url": "/api/mock2/sample",
    "method": "post",
    "graphql": "{ pages(page: $page, perPage: $perPage) { id, engine } }",
    "data": {
      "page": "${page}",
      "perPage": "${perPage}"
    }
  },
  "columns": [
    {
      "name": "id",
      "label": "ID"
    },
    {
      "name": "engine",
      "label": "Rendering engine"
    }
  ]
}
```

## 属性表

| 字段名          | 说明         | 类型                                                                                                 | 备注                                                                                                                                                                                          |
| --------------- | ------------ | ---------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| method          | 请求方式     | 字符串                                                                                               | 支持：get、post、put、delete                                                                                                                                                                  |
| url             | 请求地址     | [模板字符串](https://suda.bce.baidu.com/docs/template#%E6%A8%A1%E6%9D%BF%E5%AD%97%E7%AC%A6%E4%B8%B2) | -                                                                                                                                                                                             |
| data            | 请求数据     | 对象或字符串                                                                                         | 支持数据映射                                                                                                                                                                                  |
| dataType        | 数据体格式   | 字符串                                                                                               | 默认为 `json` 可以配置成 `form` 或者 `form-data`。当 `data` 中包含文件时，自动会采用 `form-data（multipart/form-data）` 格式。当配置为 `form` 时为 `application/x-www-form-urlencoded` 格式。 |
| qsOptions       | --           | 对象或字符串                                                                                         | 当 dataType 为 form 或者 form-data 的时候有用。具体参数请参考这里，默认设置为: `{ arrayFormat: 'indices', encodeValuesOnly: true }`                                                           |
| headers         | 请求头       | 对象                                                                                                 | -                                                                                                                                                                                             |
| sendOn          | 请求条件     | [表达式](../concepts/expression)                                                                     | -                                                                                                                                                                                             |
| cache           | 接口缓存时间 | 整型数字                                                                                             | -                                                                                                                                                                                             |
| requestAdaptor  | 发送适配器   | 字符串                                                                                               | ，支持字符串串格式，或者直接就是函数如：                                                                                                                                                      |
| adaptor         | 接收适配器   | 字符串                                                                                               | 如果接口返回不符合要求，可以通过配置一个适配器来处理成 amis 需要的。同样支持 Function 或者 字符串函数体格式                                                                                   |
| replaceData     | 替换当前数据 | 布尔                                                                                                 | 返回的数据是否替换掉当前的数据，默认为 `false`，即：`追加`，设置成 `true` 就是完全替换。                                                                                                      |
| responseType    | 返回类型     | 字符串                                                                                               | 如果是下载需要设置为 'blob'                                                                                                                                                                   |
| autoRefresh     | 是否自动刷新 | 布尔                                                                                                 | 配置是否需要自动刷新接口。                                                                                                                                                                    |
| responseData    | 配置返回数据 | 对象                                                                                                 | 对返回结果做个映射                                                                                                                                                                            |
| trackExpression | 跟踪变量     | 字符串                                                                                               | 配置跟踪变量表达式                                                                                                                                                                            |
| messages        | 提示信息     | 对象                                                                                                 | 配置接口请求的提示信息，messages.success 表示请求成功提示信息、messages.failed 表示请求失败提示信息，2.4.1 及以上版本                                                                         |
