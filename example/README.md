# line-balancer 示例

计算达成目标产量所需的「节拍时间（takt time）」，用最长工序优先 + 最佳适配
的贪心法把工序分配到工作站，并给出瓶颈工位与线平衡率。

本目录 `tasks.csv` 是 8 道工序（秒）。

## 运行

```bash
go run . -in example/tasks.csv -demand 400 -time 28800
```

含义：8 小时（28800 秒）要产出 400 件 → 节拍 = 72 s/件。工具会把 8 道工序
贪心装入若干工位，并标出负载最高的瓶颈工位与线平衡率。

## 调产量看瓶颈变化

```bash
go run . -in example/tasks.csv -demand 800        # 节拍变 36s，assemble(60s) 超标 → 明显瓶颈
go run . -in example/tasks.csv -demand 200        # 节拍变 144s，工位更少
```

> 缺 `-in` / `-demand`（或 ≤0）、文件不存在、或秒数为负 → 受控报错（exit 1），不 panic。
> CSV 支持表头 `task,seconds`（含 UTF-8 BOM 也能解析），列之间用逗号分隔。
