#!/bin/bash
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

INSTALL_DIR="/opt/chat-system"
SERVICE_NAME="chat-system"
REPO_URL="https://github.com/5p2O5pav/Simplechat-go.git"  # 修改为实际地址

require_root() {
    if [[ $EUID -ne 0 ]]; then
        echo -e "${RED}请以 root 用户运行：sudo bash install.sh${NC}"
        exit 1
    fi
}

install_deps() {
    echo -e "${YELLOW}>> 安装依赖...${NC}"
    apt-get update -qq
    apt-get install -y git curl build-essential
    if ! command -v go &>/dev/null; then
        echo "安装 Go 1.21..."
        wget -q https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
        tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
        rm go1.21.5.linux-amd64.tar.gz
        echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
        export PATH=$PATH:/usr/local/go/bin
    fi
    echo -e "${GREEN}✓ 依赖就绪${NC}"
}

clone_or_update() {
    mkdir -p "$INSTALL_DIR"
    if [[ -d "$INSTALL_DIR/.git" ]]; then
        echo "更新已有代码..."
        cd "$INSTALL_DIR"
        git pull
    else
        git clone "$REPO_URL" "$INSTALL_DIR"
    fi
}

build_app() {
    cd "$INSTALL_DIR"
    go mod download
    go build -o chat-system ./cmd/server
    echo -e "${GREEN}✓ 编译完成${NC}"
}

setup_env() {
    if [[ ! -f "$INSTALL_DIR/.env" ]]; then
        cp "$INSTALL_DIR/.env.example" "$INSTALL_DIR/.env"
        echo -e "${YELLOW}请编辑 $INSTALL_DIR/.env 填写必要配置${NC}"
        read -p "按 Enter 继续编辑（或 Ctrl+C 退出）" 
        nano "$INSTALL_DIR/.env"
    fi
}

setup_systemd() {
    cat > /etc/systemd/system/${SERVICE_NAME}.service << EOF
[Unit]
Description=Chat System Go Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/chat-system
Restart=always
RestartSec=10
EnvironmentFile=$INSTALL_DIR/.env

[Install]
WantedBy=multi-user.target
EOF
    systemctl daemon-reload
    systemctl enable ${SERVICE_NAME}
    systemctl start ${SERVICE_NAME}
    sleep 2
    if systemctl is-active --quiet ${SERVICE_NAME}; then
        echo -e "${GREEN}✓ 服务已启动${NC}"
    else
        echo -e "${RED}✗ 服务启动失败，请检查日志：journalctl -u ${SERVICE_NAME} -n 20${NC}"
    fi
}

setup_webhook() {
    source "$INSTALL_DIR/.env"
    read -p "请输入你的域名（例如 chat.example.com）： " DOMAIN
    if [[ -z "$DOMAIN" ]]; then
        echo -e "${YELLOW}跳过 Webhook 设置${NC}"
        return
    fi
    WEBHOOK_URL="https://${DOMAIN}/telegram-webhook"
    RESPONSE=$(curl -s -X POST "https://api.telegram.org/bot${BOT_TOKEN}/setWebhook" \
        -d "url=${WEBHOOK_URL}" -d "secret_token=${WEBHOOK_SECRET}")
    if echo "$RESPONSE" | grep -q '"ok":true'; then
        echo -e "${GREEN}✓ Webhook 设置成功: ${WEBHOOK_URL}${NC}"
        sed -i "s|^DOMAIN=.*|DOMAIN=\"${DOMAIN}\"|" "$INSTALL_DIR/.env"
    else
        echo -e "${RED}✗ Webhook 设置失败：$RESPONSE${NC}"
    fi
    echo -e "${YELLOW}请到 Cloudflare 将域名回源到本机 ${PORT} 端口${NC}"
}

install_chat_cmd() {
    cat > /usr/local/bin/chat << 'EOF'
#!/bin/bash
[[ ! -f /opt/chat-system/chat-system ]] && echo "Chat System 未安装" && exit 1
[[ $EUID -ne 0 ]] && exec sudo systemctl $@ chat-system
exec systemctl $@ chat-system
EOF
    chmod +x /usr/local/bin/chat
    echo -e "${GREEN}✓ 快捷命令 'chat' 已就绪（用法：chat status|start|stop|restart）${NC}"
}

show_finish() {
    echo ""
    echo -e "${GREEN}================================${NC}"
    echo -e "${GREEN}  Chat System Go 安装成功！${NC}"
    echo -e "${GREEN}================================${NC}"
    echo -e "管理面板：${CYAN}chat status|start|stop|restart${NC}"
    echo -e "服务状态：${CYAN}systemctl status ${SERVICE_NAME}${NC}"
    echo -e "上传目录：${CYAN}${INSTALL_DIR}/public/uploads${NC}"
    echo -e "Webhook 密钥：${CYAN}${WEBHOOK_SECRET}${NC} （请勿泄露）"
    echo -e "${GREEN}================================${NC}"
}

# 主流程
require_root
install_deps
clone_or_update
build_app
setup_env
setup_systemd
setup_webhook
install_chat_cmd
show_finish
