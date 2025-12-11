package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("使用方法: go run gen_password.go <密码>")
		fmt.Println("示例: go run gen_password.go admin123")
		fmt.Println()
		fmt.Println("生成默认管理员密码...")
		generateDefault()
		os.Exit(0)
	}

	password := os.Args[1]

	// 生成bcrypt hash
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("========================================")
	fmt.Println("密码生成成功!")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("原始密码:", password)
	fmt.Println("加密密码:", string(hashedPassword))
	fmt.Println()
	fmt.Println("SQL插入语句:")
	fmt.Println("----------------------------------------")
	fmt.Printf("INSERT INTO `admins` (`username`, `password`, `nickname`, `email`, `status`)\n")
	fmt.Printf("VALUES ('your_username', '%s', '管理员', 'admin@example.com', 1);\n", string(hashedPassword))
	fmt.Println("----------------------------------------")
	fmt.Println()
	fmt.Println("提示: 请将 'your_username' 替换为实际的用户名")
}

func generateDefault() {
	passwords := map[string]string{
		"admin123": "默认管理员密码",
		"password": "简单密码示例",
	}

	fmt.Println("========================================")
	fmt.Println("常用密码示例")
	fmt.Println("========================================")
	fmt.Println()

	for pwd, desc := range passwords {
		hash, _ := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
		fmt.Printf("%s (%s)\n", desc, pwd)
		fmt.Printf("Hash: %s\n", string(hash))
		fmt.Println()
	}
}
