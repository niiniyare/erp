---
title: Button
description:
type: 0
group: ⚙ Components
menuName: Button
icon:
order: 29
---

## Basic Usage

```schema: scope="body"
{
  "label": "Open Dialog",
  "type": "button",
  "actionType": "dialog",
  "dialog": {
    "title": "Dialog",
    "body": "This is a simple dialog."
  }
}
```

`button` is actually an alias for `action`. For more usage, see [action](./action)

## Property Table

| Property           | Type                                                                                                                             | Default  | Description                                                  |
| ---------------- | -------------------------------------------------------------------------------------------------------------------------------- | ------- | ----------------------------------------------------- |
| className        | `string`                                                                                                                         |         | Specifies the class name to add to the button                                  |
| url              | `string`                                                                                                                         |         | Click to jump to address, specifying this property makes button behavior consistent with a link |
| size             | `'xs' \| 'sm' \| 'md' \| 'lg' `                                                                                                  |         | Set button size                                          |
| actionType       | `'button' \| 'reset' \| 'submit'\| 'clear'\| 'url'`                                                                              | button  | Set button type                                          |
| level            | `'link' \| 'primary' \| 'enhance' \| 'secondary' \| 'info'\|'success' \| 'warning' \| 'danger' \| 'light'\| 'dark' \| 'default'` | default | Set button style                                          |
| tooltip          | `'string' \| 'TooltipObject'`                                                                                                    |         | Tooltip content                                          |
| tooltipPlacement | `'top' \| 'right' \| 'bottom' \| 'left' `                                                                                        | top     | Tooltip position                                          |
| tooltipTrigger   | `'hover' \| 'focus'`                                                                                                             |         | Trigger tooltip                                           |
| disabled         | `'boolean'`                                                                                                                      | false   | Button disabled state                                          |
| block            | `'boolean'`                                                                                                                      | false   | Option to adjust button width to its parent width                        |
| loading          | `'boolean'`                                                                                                                      | false   | Show button loading effect                                 |
| loadingOn        | `'string'`                                                                                                                       |         | Show button loading expression                               |
