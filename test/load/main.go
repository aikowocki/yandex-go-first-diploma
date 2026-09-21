package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"sync"
	"time"
)

const (
	baseURL       = "http://localhost:8081"
	accrualURL    = "http://localhost:8080"
	numUsers      = 10
	ordersPerUser = 5
)

func main() {
	start := time.Now()

	// Регистрация механики вознаграждения в accrual
	goodsBody := []byte(`{"match":"Bork","reward":10,"reward_type":"%"}`)
	resp, err := http.Post(accrualURL+"/api/goods", "application/json", bytes.NewReader(goodsBody))
	if err != nil {
		fmt.Printf("failed to register goods: %v\n", err)
	} else {
		fmt.Printf("goods registration: %d\n", resp.StatusCode)
		_ = resp.Body.Close()
	}

	var wg sync.WaitGroup
	for i := 0; i < numUsers; i++ {
		wg.Add(1)
		go func(userIdx int) {
			defer wg.Done()
			runUser(userIdx)
		}(i)
	}
	wg.Wait()

	fmt.Printf("\nDone: %d users, %d orders each, took %s\n", numUsers, ordersPerUser, time.Since(start))
}

func runUser(idx int) {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar, Timeout: 10 * time.Second}

	user := fmt.Sprintf("loaduser_%d_%d", idx, time.Now().UnixNano())

	// Register
	body, _ := json.Marshal(map[string]string{"login": user, "password": "test123"})
	resp, err := client.Post(baseURL+"/api/user/register", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("[user %d] register error: %v\n", idx, err)
		return
	}
	_ = resp.Body.Close()
	if resp.StatusCode != 200 {
		fmt.Printf("[user %d] register: %d\n", idx, resp.StatusCode)
		return
	}

	// Submit orders
	for j := 0; j < ordersPerUser; j++ {
		orderNum := generateLuhn(idx, j)

		// Register in accrual
		orderBody, _ := json.Marshal(map[string]any{
			"order": orderNum,
			"goods": []map[string]any{
				{"description": "Товар Bork", "price": 1000 + j*100},
			},
		})
		resp, err := client.Post(accrualURL+"/api/orders", "application/json", bytes.NewReader(orderBody))
		if err == nil {
			_ = resp.Body.Close()
		}

		// Submit to gophermart
		resp, err = client.Post(baseURL+"/api/user/orders", "text/plain", bytes.NewReader([]byte(orderNum)))
		if err != nil {
			fmt.Printf("[user %d] order %s error: %v\n", idx, orderNum, err)
			continue
		}
		_ = resp.Body.Close()
		fmt.Printf("[user %d] order %s: %d\n", idx, orderNum, resp.StatusCode)
	}

	// Wait for accrual processing
	time.Sleep(60 * time.Second)

	// Check balance
	resp, err = client.Get(baseURL + "/api/user/balance")
	if err != nil {
		fmt.Printf("[user %d] balance error: %v\n", idx, err)
		return
	}
	data, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	fmt.Printf("[user %d] balance: %s\n", idx, string(data))

	// Withdraw
	withdrawOrder := generateLuhn(idx+1000, 0)
	wBody, _ := json.Marshal(map[string]any{"order": withdrawOrder, "sum": 1.5})
	resp, err = client.Post(baseURL+"/api/user/balance/withdraw", "application/json", bytes.NewReader(wBody))
	if err != nil {
		fmt.Printf("[user %d] withdraw error: %v\n", idx, err)
		return
	}
	_ = resp.Body.Close()
	fmt.Printf("[user %d] withdraw: %d\n", idx, resp.StatusCode)

	// Balance after withdraw
	resp, err = client.Get(baseURL + "/api/user/balance")
	if err != nil {
		fmt.Printf("[user %d] balance after withdraw error: %v\n", idx, err)
		return
	}
	data, _ = io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	fmt.Printf("[user %d] balance after withdraw: %s\n", idx, string(data))

	// Withdrawals list
	resp, err = client.Get(baseURL + "/api/user/withdrawals")
	if err != nil {
		fmt.Printf("[user %d] withdrawals error: %v\n", idx, err)
		return
	}
	data, _ = io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	fmt.Printf("[user %d] withdrawals: %s\n", idx, string(data))
}

// generateLuhn generates a valid Luhn number based on user and order index
func generateLuhn(userIdx, orderIdx int) string {
	base := fmt.Sprintf("%04d%04d%03d", userIdx, orderIdx, time.Now().UnixNano()%1000)
	checkDigit := luhnCheckDigit(base)
	return base + strconv.Itoa(checkDigit)
}

func luhnCheckDigit(number string) int {
	sum := 0
	for i := len(number) - 1; i >= 0; i-- {
		d := int(number[i] - '0')
		if (len(number)-i)%2 == 1 {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
	}
	return (10 - sum%10) % 10
}
