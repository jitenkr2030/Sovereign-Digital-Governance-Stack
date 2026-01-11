# 审计日志规范

## 概述

本文档定义CSIC平台的审计日志规范，包括日志格式、存储要求和验证流程。

## 审计日志类型

### 1. 操作审计日志

记录所有用户操作和系统变更。

```json
{
  "audit_id": "audit_001",
  "timestamp": "2024-01-15T10:30:00Z",
  "user_id": "usr_001",
  "user_name": "admin",
  "user_role": "ADMIN",
  "department": "Financial Regulation",
  "action": "CREATE",
  "resource_type": "EXCHANGE",
  "resource_id": "exc_001",
  "resource_name": "Test Exchange",
  "ip_address": "10.0.0.100",
  "user_agent": "Mozilla/5.0",
  "request_body": {
    "name": "Test Exchange",
    "license_number": "LIC-2024-001"
  },
  "response_code": 201,
  "previous_hash": "a1b2c3...",
  "current_hash": "d4e5f6...",
  "nonce": 1000,
  "result": "SUCCESS",
  "error_message": null
}
```

### 2. 系统审计日志

记录系统级事件和状态变更。

```json
{
  "system_audit_id": "sys_001",
  "timestamp": "2024-01-15T10:30:00Z",
  "event_type": "SERVICE_START",
  "service_name": "api-gateway",
  "host": "server-001",
  "message": "Service started successfully",
  "metrics": {
    "memory_usage": 512,
    "cpu_usage": 15
  }
}
```

### 3. 安全审计日志

记录安全相关事件。

```json
{
  "security_audit_id": "sec_001",
  "timestamp": "2024-01-15T10:30:00Z",
  "event_type": "LOGIN_SUCCESS",
  "user_id": "usr_001",
  "ip_address": "10.0.0.100",
  "geo_location": {
    "country": "CN",
    "region": "Beijing"
  },
  "device_fingerprint": "abc123...",
  "risk_score": 10,
  "mfa_used": true,
  "failure_reason": null
}
```

### 4. 数据变更审计日志

记录数据修改操作。

```json
{
  "data_audit_id": "data_001",
  "timestamp": "2024-01-15T10:30:00Z",
  "table_name": "exchanges",
  "record_id": "exc_001",
  "operation": "UPDATE",
  "before": {
    "status": "ACTIVE",
    "version": 5
  },
  "after": {
    "status": "SUSPENDED",
    "version": 6
  },
  "changed_fields": ["status", "version"],
  "user_id": "usr_002"
}
```

## 必录字段

### 所有日志通用字段

| 字段名 | 类型 | 必录 | 说明 |
|--------|------|------|------|
| log_id | string | 是 | 唯一标识符 |
| timestamp | ISO8601 | 是 | 事件时间戳 |
| log_type | enum | 是 | 日志类型 |

### 操作审计必录字段

| 字段名 | 类型 | 必录 | 说明 |
|--------|------|------|------|
| user_id | string | 是 | 执行用户ID |
| user_role | string | 是 | 用户角色 |
| action | string | 是 | 操作类型 |
| resource_type | string | 是 | 资源类型 |
| resource_id | string | 是 | 资源ID |
| ip_address | string | 是 | 客户端IP |
| request_body | json | 是 | 请求内容 |
| response_code | int | 是 | 响应状态码 |

## 审计链结构

### 哈希链

每个审计日志包含前一个日志的哈希值，形成不可变的链式结构。

```
GENESIS ──> audit_001 ──> audit_002 ──> audit_003 ──> ...
           (hash1)      (hash2)      (hash3)
```

### 哈希计算

```go
func calculateHash(auditLog AuditLog, previousHash string) string {
    data := fmt.Sprintf("%s|%s|%s|%s|%s|%d",
        auditLog.UserID,
        auditLog.Action,
        auditLog.ResourceType,
        auditLog.ResourceID,
        previousHash,
        auditLog.Timestamp.Unix(),
    )
    
    hash := sha256.Sum256([]byte(data))
    return hex.EncodeToString(hash[:])
}
```

## 存储要求

### 存储介质

| 介质 | 用途 | 写入方式 |
|------|------|----------|
| WORM存储 | 主存储 | 一次写入，多次读取 |
| OpenSearch | 索引和搜索 | 实时写入 |
| 备份存储 | 异地备份 | 定期同步 |

### 保留期限

| 日志类型 | 保留期限 | 归档策略 |
|----------|----------|----------|
| 操作审计 | 7年 | 归档到冷存储 |
| 安全审计 | 7年 | 永久保留 |
| 系统审计 | 1年 | 汇总后保留7年 |
| 数据变更 | 7年 | 归档到冷存储 |

### 存储加密

- 传输加密: TLS 1.3
- 存储加密: AES-256-GCM
- 密钥管理: HSM

## 验证流程

### 完整性验证

```bash
# 验证审计链完整性
./scripts/audit/verify-chain.sh --start-date 2024-01-01 --end-date 2024-01-31

# 验证单个日志
./scripts/audit/verify-log.sh --log-id audit_001

# 验证哈希链
./scripts/audit/verify-hash-chain.sh --chain-id main
```

### 一致性检查

```bash
# 检查日志一致性
./scripts/audit/check-consistency.sh --type all

# 检查用户操作
./scripts/audit/check-user-actions.sh --user-id usr_001

# 检查资源变更
./scripts/audit/check-resource-changes.sh --resource-type EXCHANGE
```

## 访问控制

### 角色权限

| 角色 | 读取 | 导出 | 删除 |
|------|------|------|------|
| ADMIN | 全部 | 全部 | 不可 |
| AUDITOR | 全部 | 全部 | 不可 |
| OPERATOR | 本部门 | 不可 | 不可 |
| VIEWER | 本部门 | 不可 | 不可 |

### 查询示例

```sql
-- 查询用户操作记录
SELECT * FROM audit_logs
WHERE user_id = 'usr_001'
ORDER BY timestamp DESC
LIMIT 100;

-- 查询资源变更历史
SELECT * FROM audit_logs
WHERE resource_type = 'EXCHANGE'
  AND resource_id = 'exc_001'
ORDER BY timestamp DESC;

-- 查询敏感操作
SELECT * FROM audit_logs
WHERE action IN ('DELETE', 'REVOKE', 'FREEZE')
  AND timestamp >= NOW() - INTERVAL '24 HOURS';
```

## 合规要求

### 证据标准

- 时间戳可信
- 内容不可篡改
- 完整性可验证
- 访问有记录

### 导出格式

```json
{
  "export": {
    "export_id": "exp_001",
    "export_time": "2024-01-15T10:30:00Z",
    "exported_by": "usr_001",
    "exported_for": "Audit Investigation",
    "date_range": {
      "start": "2024-01-01T00:00:00Z",
      "end": "2024-01-15T10:30:00Z"
    },
    "total_records": 10000,
    "checksum": "sha256:abc123...",
    "signature": {
      "signer": "HSM Key 1",
      "value": "sig456...",
      "timestamp": "2024-01-15T10:30:00Z"
    }
  },
  "records": [...]
}
```
