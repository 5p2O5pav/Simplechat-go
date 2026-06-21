#!/bin/bash
# 监控脚本：服务自恢复、资源告警、每日统计
CONFIG_FILE="/opt/chat-system/.env"
if [[ -f "$CONFIG_FILE" ]]; then
    source "$CONFIG_FILE"
else
    echo "配置文件不存在"
    exit 1
fi

HOST=$(hostname)
SERVICE="chat-system"

# 服务重启
if ! systemctl is-active --quiet "$SERVICE"; then
    systemctl start "$SERVICE"
    sleep 5
    if systemctl is-active --quiet "$SERVICE"; then
        MSG="⚠️ $HOST Chat System 已自动恢复"
    else
        MSG="🚨 $HOST Chat System 异常，自动重启失败"
    fi
    curl -s -X POST "https://api.telegram.org/bot${BOT_TOKEN}/sendMessage" \
        -d "chat_id=${CHAT_ID}" -d "message_thread_id=${MONITOR_TOPIC_ID}" -d "text=${MSG}" > /dev/null
fi

# 磁盘/内存告警（可选）
DISK=$(df / | tail -1 | awk '{print $5}' | tr -d '%')
if [[ $DISK -gt 90 ]]; then
    curl -s -X POST "https://api.telegram.org/bot${BOT_TOKEN}/sendMessage" \
        -d "chat_id=${CHAT_ID}" -d "message_thread_id=${MONITOR_TOPIC_ID}" \
        -d "text=⚠️ $HOST 磁盘使用率 ${DISK}%" > /dev/null
fi

# 每日统计（如果参数为 --daily 或当前小时为0）
if [[ "$1" == "--daily" ]] || [[ $(date +%H) == "00" ]]; then
    STATS=$(mktemp)
    TODAY=$(date +%Y-%m-%d)
    echo "📊 每日运行统计 [${TODAY}]" > "$STATS"
    echo "=========================" >> "$STATS"
    # 从 SQLite 统计活跃会话数（使用 sqlite3 命令）
    if command -v sqlite3 &>/dev/null; then
        DB_PATH="${DB_PATH:-./data/chat.db}"
        if [[ -f "$DB_PATH" ]]; then
            COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM sessions WHERE expire_at > strftime('%s','now')*1000;")
            echo "• 活跃会话数: ${COUNT:-0}" >> "$STATS"
        fi
    fi
    # 磁盘、内存、CPU
    echo "• 磁盘: $(df -h / | tail -1 | awk '{print $3"/"$2" ("$5")"}')" >> "$STATS"
    echo "• 内存: $(free -h | grep Mem | awk '{print $3"/"$2}')" >> "$STATS"
    echo "• CPU负载: $(uptime | awk -F'load average:' '{print $2}')" >> "$STATS"
    # 上传文件统计
    UPLOAD_PATH="/opt/chat-system/public/uploads"
    if [[ -d "$UPLOAD_PATH" ]]; then
        UPLOAD_SIZE=$(du -sh "$UPLOAD_PATH" 2>/dev/null | cut -f1)
        UPLOAD_COUNT=$(find "$UPLOAD_PATH" -type f | wc -l)
        echo "• 上传文件: ${UPLOAD_COUNT} 个，共 ${UPLOAD_SIZE}" >> "$STATS"
    fi
    curl -s -X POST "https://api.telegram.org/bot${BOT_TOKEN}/sendMessage" \
        -d "chat_id=${CHAT_ID}" -d "message_thread_id=${MONITOR_TOPIC_ID}" \
        -d "text=$(cat "$STATS")" > /dev/null
    rm "$STATS"
fi
