# GuanLan

观澜 (GuanLan)：语出《孟子》“观水有术，必观其澜”。寓意看股票趋势不能只看短期水花，而要观察中长期的波澜壮阔。

```shell
# uv init -p 3.12
uv sync

# clone qlib source code for run script and learn code
# git clone https://github.com/microsoft/qlib.git
```

```shell
go mod tidy
go install tool
buf mod update
```

## Data

[yahoo finance数据下载](https://github.com/ranaroussi/yfinance)
[回测框架](https://github.com/mementum/backtrader)

添加新功能 记录我操作股票/基金的记录

- 建仓/加仓/减仓/清仓 记录
- 买入/卖出 记录包含的 税费、佣金、滑点、流动性成本
- 成本的计算
- 分红记录
- 总资本的计算 包含 本金、利息、分红、收益
- 总资本的成份分析
