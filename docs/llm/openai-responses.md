# OpenAI Responses 后端

当前 OpenAI 后端使用 `POST /v1/responses`，并开启 `stream: true`。

## 当前已经实现的能力

- 将运行时输入项映射到 Responses API 的 `input`
- 将统一工具定义映射到 OpenAI `function` tools
- 通过 SSE 读取响应流
- 增量输出 assistant 文本
- 解析函数调用事件
- 保留 `response_id`，用于工具执行后继续同一轮对话

## 当前处理的事件

后端目前重点处理以下事件：

- `response.created`
- `response.output_text.delta`
- `response.output_item.added`
- `response.function_call_arguments.done`
- `response.completed`
- `response.failed`

## 请求结构

首轮请求发送：

- `model`
- `instructions`
- 用户输入
- 所有可用工具
- `stream: true`

如果模型请求工具，则后续请求会发送：

- `previous_response_id`
- 一个或多个 `function_call_output`
- 同一组工具定义

## 当前限制

- 目前只覆盖主路径事件，未覆盖 Responses 的全部事件类型
- 错误处理已可用，但还不够细
- SSE 解析采用标准 `data:` 行模式，尚未加入更强健的多段聚合处理
