# 领域模型与通用字段

本文档定义 GuanLan 当前阶段的核心领域对象、任务状态和通用字段。产品边界见 [01-product-scope.md](./01-product-scope.md)。

## 1. 统一任务状态

数据任务、分析任务、训练任务和回测任务共用以下状态：

| 状态 | 说明 |
|------|------|
| `pending` | 等待执行 |
| `running` | 正在执行 |
| `success` | 执行成功 |
| `failed` | 执行失败 |
| `partial_success` | 部分成功 |
| `cancelled` | 已取消 |

## 2. 任务公共字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `task_id` | UUID | 任务唯一标识 |
| `task_type` | enum | `data_sync` / `analysis` / `training` / `backtest` |
| `target_object` | string | 目标对象标识，如指数代码、股票代码或任务范围 |
| `trigger_method` | enum | `scheduled` / `manual` / `retry` |
| `status` | enum | 统一任务状态 |
| `created_at` | datetime | 创建时间 |
| `started_at` | datetime | 开始时间 |
| `ended_at` | datetime | 结束时间 |
| `retry_count` | int | 重试次数 |
| `failure_reason` | string | 失败原因，可空 |
| `log_ref` | string | 关联日志入口 |

## 3. 结果追溯字段

分析结果、训练产物和回测报告必须记录以下追溯字段：

| 字段 | 说明 |
|------|------|
| `data_version` | 使用的数据版本 |
| `param_version` | 使用的参数版本 |
| `model_version` | 使用的模型版本，可空 |
| `run_timestamp` | 运行时间 |
| `result_status` | `success` / `partial` / `failed` |

## 4. 核心对象

### 4.1 指数数据集

| 字段 | 说明 |
|------|------|
| `index_code` | 指数代码，如 `000905.SH` |
| `market` | 市场，`A` / `US` |
| `index_name` | 指数名称 |
| `data_completeness` | 数据完整率 |
| `last_sync_time` | 最近同步时间 |
| `sync_status` | 同步状态 |

### 4.2 指数成分股快照

| 字段 | 说明 |
|------|------|
| `index_code` | 所属指数 |
| `stock_code` | 股票代码 |
| `snap_date` | 快照日期 |
| `weight` | 权重，可空 |
| `is_active` | 是否仍在指数中 |

### 4.3 日频行情

| 字段 | 说明 |
|------|------|
| `stock_code` | 股票代码 |
| `market` | 市场 |
| `trade_date` | 交易日期 |
| `open` | 开盘价 |
| `high` | 最高价 |
| `low` | 最低价 |
| `close` | 收盘价 |
| `volume` | 成交量 |
| `amount` | 成交额，可空 |
| `adj_factor` | 复权因子，可空 |
| `source` | 数据源 |
| `data_version` | 数据版本 |

### 4.4 个股数据状态

| 字段 | 说明 |
|------|------|
| `stock_code` | 股票代码 |
| `market` | 市场 |
| `training_index_code` | 命中的预置训练数据指数代码，可空 |
| `data_start_date` | 数据起始日期 |
| `data_end_date` | 数据截止日期 |
| `completeness` | 完整率 |
| `missing_ranges` | 缺失区间列表 |
| `last_update` | 最近更新时间 |

### 4.5 股票池条目

| 字段 | 说明 |
|------|------|
| `stock_code` | 股票代码 |
| `market` | 市场 |
| `tags` | 标签列表 |
| `priority` | 优先级 |
| `notes` | 备注 |
| `is_active` | 是否参与每日分析 |
| `added_at` | 加入时间 |
| `removed_at` | 移除时间，可空 |
| `source` | `manual` / `history_recovery` |
| `last_action` | 最近操作，`add` / `remove` / `disable` / `enable` |
| `last_action_at` | 最近操作时间 |

### 4.6 交易记录

| 字段 | 说明 |
|------|------|
| `trade_id` | 交易记录 ID |
| `trade_date` | 交易日期 |
| `stock_code` | 股票代码 |
| `stock_name` | 股票名称 |
| `side` | `buy` / `sell` |
| `price` | 成交价格 |
| `quantity` | 成交数量 |
| `total_fee` | 总费用，含佣金、印花税等 |
| `note` | 备注，可空 |
| `created_at` | 创建时间 |

### 4.7 现金分红

| 字段 | 说明 |
|------|------|
| `dividend_id` | 分红记录 ID |
| `dividend_date` | 分红日期 |
| `stock_code` | 股票代码 |
| `dividend_per_share` | 每股分红金额，可空 |
| `total_dividend` | 总分红金额 |
| `bonus_share_ratio` | 送股比例，预留字段，当前不参与计算 |
| `transfer_share_ratio` | 转增比例，预留字段，当前不参与计算 |
| `note` | 备注，可空 |

### 4.8 现金流记录

| 字段 | 说明 |
|------|------|
| `cash_flow_id` | 现金流记录 ID |
| `flow_date` | 发生日期 |
| `amount` | 金额，入金为正，出金为负 |
| `flow_type` | `deposit` / `withdrawal` / `trade` / `dividend` |
| `source_ref` | 来源记录 ID，可空 |
| `note` | 备注，可空 |

### 4.9 持仓状态

| 字段 | 说明 |
|------|------|
| `stock_code` | 股票代码 |
| `stock_name` | 股票名称 |
| `quantity` | 当前持仓数量 |
| `total_cost` | 当前持仓总成本 |
| `average_cost` | 移动加权平均成本 |
| `realized_pnl` | 累计已实现盈亏 |
| `dividend_income` | 累计分红收入 |
| `latest_price` | 最新估值价格，可空 |
| `market_value` | 当前持仓市值，可空 |
| `unrealized_pnl` | 持仓未实现盈亏，可空 |
| `updated_at` | 最近更新时间 |

### 4.10 估值快照

| 字段 | 说明 |
|------|------|
| `valuation_id` | 估值快照 ID |
| `valuation_date` | 估值日期 |
| `stock_code` | 股票代码，可空 |
| `price` | 单只股票估值价格，可空 |
| `total_asset_override` | 用户录入的合计总资产，可空 |
| `source` | `manual` / `csv_import` |
| `note` | 备注，可空 |

### 4.11 资产快照

| 字段 | 说明 |
|------|------|
| `snapshot_date` | 快照日期 |
| `cash_balance` | 现金余额 |
| `holding_market_value` | 持仓总市值 |
| `total_asset` | 总资产 |
| `source` | `valuation` / `daily_price` / `manual_total` |

### 4.12 年度复盘汇总

| 字段 | 说明 |
|------|------|
| `year` | 自然年 |
| `realized_pnl` | 年度已实现盈亏 |
| `dividend_income` | 年度分红收入 |
| `net_cash_flow` | 年度净出入金 |
| `begin_total_asset` | 期初总资产，可空 |
| `end_total_asset` | 期末总资产，可空 |
| `return_rate` | 简易收益率，可空 |
| `by_stock_breakdown` | 按股票分组的盈亏和分红贡献 |

### 4.13 分析结果

| 字段 | 说明 |
|------|------|
| `stock_code` | 股票代码 |
| `analysis_date` | 分析日期 |
| `signal` | `buy` / `sell` / `hold` |
| `confidence` | 置信度，0-1 |
| `risk_score` | 风险分数 |
| `suggested_stop_loss` | 建议止损价 |
| `suggested_target` | 建议目标价 |
| `position_advice` | 仓位建议比例 |
| `reason_summary` | 理由摘要 |
| `data_version` | 数据版本 |
| `model_version` | 模型版本 |

### 4.14 回测报告

| 字段 | 说明 |
|------|------|
| `backtest_id` | 回测任务 ID |
| `stock_range` | 股票范围 |
| `date_range` | 时间区间 |
| `benchmark` | 基准指数 |
| `total_return` | 总收益率 |
| `annual_return` | 年化收益率 |
| `max_drawdown` | 最大回撤 |
| `sharpe_ratio` | 夏普比率 |
| `win_rate` | 胜率 |
| `profit_loss_ratio` | 盈亏比 |

### 4.15 告警事件

| 字段 | 说明 |
|------|------|
| `alert_id` | 告警 ID |
| `alert_type` | `data` / `analysis` / `risk` / `system` |
| `severity` | `info` / `warning` / `critical` |
| `source` | 告警来源 |
| `message` | 告警消息 |
| `created_at` | 创建时间 |
| `resolved_at` | 解决时间，可空 |
| `status` | `active` / `resolved` |

### 4.16 服务健康状态

| 字段 | 说明 |
|------|------|
| `service_name` | 服务名称 |
| `status` | `healthy` / `degraded` / `down` |
| `last_check` | 最近检查时间 |
| `uptime` | 运行时间 |
| `metrics` | 关键指标快照 |

## 5. 投资组合计算规则

### 5.1 交易与持仓成本

1. 买入：现金余额减少 `price * quantity + total_fee`，持仓数量增加，持仓总成本增加同等金额。
2. 加仓：采用移动加权平均成本法，`average_cost = total_cost / quantity`。
3. 卖出：现金余额增加 `price * quantity - total_fee`，持仓数量减少，并按卖出数量乘以当前平均成本结转成本。
4. 已实现盈亏：卖出收入减去卖出部分成本和费用。
5. 清仓：持仓数量归零，该股票剩余总成本归零，已实现盈亏结清到累计值。

### 5.2 分红与成本下降

1. 现金分红增加现金余额。
2. 分红日持仓数量不变，持仓总成本减少总分红金额，平均成本随之下降。
3. 每股分红和总分红可二选一录入；若录入每股分红，系统按当日持仓数量计算总分红。
4. 送股和转增字段仅预留，当前阶段不改变持仓数量和成本。

### 5.3 总资产与年度复盘

1. 总资产 = 持仓总市值 + 现金余额。
2. 持仓市值优先使用手动估值快照或导入的每日收盘价；没有估值时不强制计算未实现盈亏。
3. 没有日频行情时，系统仍必须基于交易、分红、现金流和稀疏估值快照计算现金余额、持仓、成本、已实现盈亏和资产曲线。
4. 年度复盘通过日期范围查询交易、分红、现金流和估值快照，不建立额外年度流水表。

## 6. 页面通用能力

所有任务和状态页面统一支持：

1. 状态标签。
2. 最近更新时间。
3. 失败原因查看。
4. 关联日志入口。
5. 手动重试入口。
6. 日频图表的数据区间和缺失状态提示。
7. 交易、分红、现金流和估值记录按自然年过滤。
