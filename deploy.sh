#!/bin/bash
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

INSTALL_DIR="/opt/chat-system"
SERVICE_NAME="chat-system"

print_banner() {
    clear
    echo -e "${CYAN}"
    echo "╔══════════════════════════════════════╗"
    echo "║   Chat System Go 管理面板           ║"
    echo "╚══════════════════════════════════════╝"
    echo -e "${NC}"
}

press_enter() { echo ""; read -p "按 Enter 返回..."; }

main_menu() {
    while true; do
        print_banner
        if systemctl is-active --quiet $SERVICE_NAME; then
            echo -e "服务: ${GREEN}运行中${NC}"
        else
            echo -e "服务: ${RED}已停止${NC}"
        fi
        echo ""
        echo "1) 启动服务"
        echo "2) 停止服务"
        echo "3) 重启服务"
        echo "4) 服务状态"
        echo "5) 查看日志"
        echo "6) 更新代码并重新编译"
        echo "7) 设置 Webhook"
        echo "0) 退出"
        read -p "请选择: " choice
        case $choice in
            1) systemctl start $SERVICE_NAME && echo -e "${GREEN}已启动${NC}" || echo -e "${RED}启动失败${NC}"; press_enter ;;
            2) systemctl stop $SERVICE_NAME && echo -e "${GREEN}已停止${NC}"; press_enter ;;
            3) systemctl restart $SERVICE_NAME && echo -e "${GREEN}已重启${NC}"; press_enter ;;
            4) systemctl status $SERVICE_NAME --no-pager -l | head -20; press_enter ;;
            5) journalctl -u $SERVICE_NAME -n 50 --no-pager; press_enter ;;
            6)
                cd $INSTALL_DIR
                git pull
                go build -o chat-system ./cmd/server
                systemctl restart $SERVICE_NAME
                echo -e "${GREEN}更新完成${NC}"
                press_enter
                ;;
            7)
                source $INSTALL_DIR/.env
                read -p "域名: " DOMAIN
                if [[ -n "$DOMAIN" ]]; then
                    WEBHOOK_URL="https://${DOMAIN}/telegram-webhook"
                    RESPONSE=$(curl -s -X POST "https://api.telegram.org/bot${BOT_TOKEN}/setWebhook" -d "url=${WEBHOOK_URL}" -d "secret_token=${WEBHOOK_SECRET}")
                    if echo "$RESPONSE" | grep -q '"ok":true'; then
                        echo -e "${GREEN}Webhook 设置成功${NC}"
                        sed -i "s|^DOMAIN=.*|DOMAIN=\"${DOMAIN}\"|" $INSTALL_DIR/.env
                    else
                        echo -e "${RED}设置失败: $RESPONSE${NC}"
                    fi
                fi
                press_enter
                ;;
            0) exit 0 ;;
            *) echo -e "${RED}无效选项${NC}"; press_enter ;;
        esac
    done
}

main_menu
