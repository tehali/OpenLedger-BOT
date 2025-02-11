package bot

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// loadAccounts 从文件加载账号列表
func (o *OpenLedger) loadAccounts() ([]string, error) {
	file, err := os.Open("accounts.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to open accounts.txt: %w", err)
	}
	defer file.Close()

	var accounts []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if account := strings.TrimSpace(scanner.Text()); account != "" {
			accounts = append(accounts, account)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading accounts.txt: %w", err)
	}

	return accounts, nil
} 