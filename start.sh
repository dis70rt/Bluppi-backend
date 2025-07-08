#!/usr/bin/env bash

set -euo pipefail

GREEN="\033[38;5;46m"
RED="\033[38;5;196m"
CYAN="\033[38;5;51m"
ORANGE="\033[38;5;214m"
RESET="\033[0m"
BOLD="\033[1m"

SUCCESS="✓"
ERROR="✗"
RUNNING="→"

main() {
    if ! sudo -v 2>/dev/null; then
        echo -e "${RED}${ERROR} Authentication failed${RESET}"
        exit 1
    fi    
    services=("PostgreSQL:postgresql")
    
    local failed=0
    
    for entry in "${services[@]}"; do
        IFS=":" read -r name unit <<< "$entry"
        
        printf "  ${CYAN}${BOLD}▶${RESET} %-15s ${ORANGE}⠋${RESET}" "$name"
        
        if systemctl is-active --quiet "$unit" 2>/dev/null; then
            printf "\r  ${CYAN}${BOLD}▶${RESET} %-15s ${GREEN}${RUNNING} Already Running${RESET}\n" "$name"
        elif sudo systemctl start "$unit" 2>/dev/null; then
            printf "\r  ${CYAN}${BOLD}▶${RESET} %-15s ${GREEN}${SUCCESS} Started${RESET}\n" "$name"
        else
            printf "\r  ${CYAN}${BOLD}▶${RESET} %-15s ${RED}${ERROR} Failed${RESET}\n" "$name"
            ((failed++))
        fi
    done
    
    if [[ $failed -gt 0 ]]; then
        exit 1
    fi
}

main "$@"