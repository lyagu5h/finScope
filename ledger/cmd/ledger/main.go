package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/lyagu5h/finScope/ledger/internal/ledger"
)



func main() {
	store := ledger.NewStore()

	transactions := []ledger.Transaction{
		{
			Amount: 1499.90, 
			Category: "Food", 
			Description: "Покупка продуктов в Магнит", 
			Date: time.Date(2025, 1, 12, 14, 30, 0, 0, time.UTC),
		},
		{
			Amount: 56000.00, 
			Category: "Salary", 
			Description: "Зарплата", 
			Date: time.Date(2025, 1, 10, 8, 0, 0, 0, time.UTC),
		},
	}

	f, err := os.Open("budgets.json")

	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	defer f.Close()

	reader := bufio.NewReader(f)
	if err := store.LoadBudgets(reader); err != nil {
		fmt.Println("Error: cannot load budgets.json: ", err)
	}

	for _, t := range transactions {
		err := store.AddTransaction(t)
		if err != nil {
			fmt.Println(err)
		}
	}

	fmt.Println(store.ListTransactions())
}