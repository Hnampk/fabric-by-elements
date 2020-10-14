package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	pb "example.com/simple-chaincode/accounting_service"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

const (
	port                            = ":50052"
	ERROR_DUPLICATE_ENTRY           = "1062"
	VALID_DUPLICATE_DURATION        = 10 // 30s
	VALID_DUPLICATE_ENTRY           = "200"
	DAILY_TRANSACTION_ON_PROCESS    = 0
	DAILY_TRANSACTION_SUCCEED       = 1
	DAILY_TRANSACTION_REVERSED      = 2
	DAILY_TRANSACTION_FOR_REVERSION = 3
)

var (
	db        *sql.DB
	workerNum uint32 = 1
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedAccountingServer
}

func (s *server) CreateAccount(ctx context.Context, in *pb.CreateAccountRequest) (*pb.CreateAccountReply, error) {
	txID := in.GetTxID()
	accountID := in.GetAccountID()

	log.Printf("[Received request] %v:CreateAccount %v", txID, accountID)

	err := createAccount(txID, accountID)

	if err != nil {
		return &pb.CreateAccountReply{Status: false, Message: err.Error()}, nil
	}

	return &pb.CreateAccountReply{Status: true, Message: "Successfully createAccount " + accountID + "!"}, nil
}

func (s *server) Deposit(ctx context.Context, in *pb.DepositRequest) (*pb.DepositReply, error) {
	txID := in.GetTxID()
	accountID := in.GetAccountID()
	value := in.GetValue()

	log.Printf("[Received request] %v: Deposit %v to %v", txID, value, accountID)

	err := deposit(txID, accountID, float64(value))
	if err != nil {
		if strings.Contains(err.Error(), VALID_DUPLICATE_ENTRY) {
			return &pb.DepositReply{Status: true, Message: "Tx on processing!" + txID}, nil
		}

		return &pb.DepositReply{Status: false, Message: err.Error()}, nil
	}

	return &pb.DepositReply{Status: true, Message: "Successfully deposit!"}, nil
}

func (s *server) Withdraw(ctx context.Context, in *pb.WithdrawRequest) (*pb.WithdrawReply, error) {
	txID := in.GetTxID()
	accountID := in.GetAccountID()
	value := in.GetValue()

	log.Printf("[Received request] %v: Withdraw %v to %v", txID, value, accountID)

	err := withdraw(txID, accountID, float64(value))
	if err != nil {
		return &pb.WithdrawReply{Status: false, Message: err.Error()}, nil
	}

	return &pb.WithdrawReply{Status: true, Message: "Successfully deposit!"}, nil
}

func (s *server) Reverse(ctx context.Context, in *pb.ReverseRequest) (*pb.ReverseReply, error) {
	txID := in.GetTxID()

	log.Printf("[Received request] Reverse txID: %v ", txID)

	err := reverse(txID)
	if err != nil {
		return &pb.ReverseReply{Status: true, Message: err.Error()}, nil
	}

	return &pb.ReverseReply{Status: true, Message: "Reverse Successfully!"}, nil
}

func main() {
	var err error

	db, err = sql.Open("mysql", "root:vnpay123@/accounting_service")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	if len(os.Args) > 1 {
		workerNumInt, err := strconv.Atoi(os.Args[1])
		workerNum = uint32(workerNumInt)

		if err != nil {
			fmt.Println("supplied workerNum is not valid!")
			return
		}
	}

	fmt.Println("Number of stream workers: ", workerNum)

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{grpc.NumStreamWorkers(workerNum), grpc.MaxConcurrentStreams(1000)}
	s := grpc.NewServer(opts...)

	pb.RegisterAccountingServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// ============================= START query SQL functions =============================
func queryInsertAccount(tx *sql.Tx, accountID string, txID string) error {
	_, err := tx.Exec(`
	INSERT INTO account_balance (account_id, hash) VALUES (?, ?);
	`, accountID, txID)

	if err != nil {
		log.Println("[ERROR]", err)
		return err
	}

	return nil
}

func querySelectAccountBalance(accountID string) (*sql.Row, error) {
	balanceStatement, err := db.Prepare(`
		SELECT * FROM account_balance
		WHERE account_id = ?;
		`) // ? = placeholder

	if err != nil {
		log.Panicln(err)
		return nil, err
	}

	defer balanceStatement.Close() // Close the statement when we leave main() / the program terminates

	return balanceStatement.QueryRow(accountID), nil
}

func queryInsertDailyTransaction(tx *sql.Tx, accountID string, traceID string, currentAccountTxCount int64, code string, preBalance float64, value float64, balance float64, status int, hash string) error {
	uniqueID := accountID + "-" + strconv.FormatInt(currentAccountTxCount+1, 10)
	preID := accountID + "-" + strconv.FormatInt(currentAccountTxCount, 10)

	_, err := tx.Exec(`
	INSERT INTO daily_transaction(unique_id, account_id, trace_id, pre_id, code, pre_balance, value, balance, status, hash)
	VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`, uniqueID, accountID, traceID, preID, code, preBalance, value, balance, status, hash)

	if err != nil {
		log.Println("[ERROR]", err)

		if strings.Contains(err.Error(), ERROR_DUPLICATE_ENTRY) {
			dailyTxStatement, err2 := db.Prepare(`
			SELECT unix_timestamp(date) FROM daily_transaction
			WHERE unique_id = ?;
			`) //
			defer dailyTxStatement.Close()
			var date int64
			err2 = dailyTxStatement.QueryRow(uniqueID).Scan(&date)

			if err2 != nil {
				log.Println("[ERROR]", err2)
				return errors.Errorf("[queryInsertDailyTransaction] account not exist:", accountID)
			}

			if time.Now().Unix()-date > VALID_DUPLICATE_DURATION {
				// this unique id has been used before!
				return err
			}

			return errors.Errorf(VALID_DUPLICATE_ENTRY + ", created at: " + strconv.Itoa(int(date*1000)))
		}

		return err
	}

	return nil
}

func querySelectDailyTransactionByTxID(txID string) (*sql.Row, error) {
	dtStatement, err := db.Prepare(`
		SELECT unique_id, id, account_id, trace_id, pre_id, code, pre_balance, value, balance, status, hash, unix_timestamp(date) 
		FROM daily_transaction
		WHERE hash = ?
		`) // ? = placeholder

	if err != nil {
		log.Panicln(err)
		return nil, err
	}

	defer dtStatement.Close() // Close the statement when we leave main() / the program terminates

	return dtStatement.QueryRow(txID), nil
}

func querySelectDailyTransactionCreateAccount(accountID string) (*sql.Row, error) {
	dtStatement, err := db.Prepare(`
	SELECT (account_id, pre_id, date) FROM daily_transaction
	WHERE account_id = ?
	LIMIT 1
	`)

	defer dtStatement.Close()

	if err != nil {
		log.Println("[ERROR]", err)
		return nil, err
	}

	return dtStatement.QueryRow(accountID), nil
}

func queryUpdateDailyTransactionByUniqueID(tx *sql.Tx, uniqueID string, status int) error {
	_, err := tx.Exec(`
	UPDATE daily_transaction
	SET status=?
	WHERE unique_id=?;
	`, status, uniqueID)

	if err != nil {
		log.Println("[ERROR]", err)
		return err
	}

	return nil
}

// ============================= END query SQL functions =============================

func updateAccountBalance(tx *sql.Tx, accountID string, balance float64, accountTxCount int64, lastBalance float64) error {
	_, err := tx.Exec(`
	UPDATE account_balance
	SET balance = ?, account_tx_count = ?, last_balance = ?
	WHERE account_id = ?;
	`, balance, accountTxCount, lastBalance, accountID)

	if err != nil {
		log.Println("[ERROR]", err)
		return err
	}

	return nil
}

func deposit(txID string, accountID string, value float64, status ...int) error {
	var uniqueID int64
	var balance float64
	var currentAccountTxCount int64
	var lastBalance float64
	var hash string

	// verify account & balance
	queryABrow, err := querySelectAccountBalance(accountID)
	if err != nil {
		log.Println("[ERROR]", err)
		return err
	}

	err = queryABrow.Scan(&uniqueID, &accountID, &balance, &currentAccountTxCount, &lastBalance, &hash)
	if err != nil {
		log.Println("[ERROR]", err)
		return errors.Errorf("[deposit] account not exist:", accountID)
	}

	log.Println(accountID, "'s balance:", balance, ", processed", currentAccountTxCount, "transactions")

	err = depositPreCheck(accountID, balance, value)
	if err != nil {
		log.Println("[ERROR]", err)
		return err
	}

	// try to deposit
	tx, err := db.Begin()
	if err != nil {
		log.Println("[ERROR]", err)
		return err
	}

	defer tx.Rollback()

	if len(status) > 0 {
		// for reverse query
		err = queryInsertDailyTransaction(tx, accountID, "", currentAccountTxCount, "+", balance, value, balance+value, status[0], txID)
	} else {
		err = queryInsertDailyTransaction(tx, accountID, "", currentAccountTxCount, "+", balance, value, balance+value, DAILY_TRANSACTION_ON_PROCESS, txID)
	}

	if err != nil {
		return err
	}

	err = updateAccountBalance(tx, accountID, balance+value, currentAccountTxCount+1, balance)
	if err != nil {
		return err
	}

	err = tx.Commit()

	if err != nil {
		log.Println("[ERROR]", err)
		return err
	}

	log.Println("Successfully deposit", value, "into account", accountID, ", new balace:", balance+value)
	return nil
}

func withdraw(txID string, accountID string, value float64, status ...int) error {
	var uniqueID int64
	var balance float64
	var currentAccountTxCount int64
	var lastBalance float64
	var hash string

	// verify account & balance
	queryABrow, err := querySelectAccountBalance(accountID)
	if err != nil {
		log.Println("[ERROR]", err)
		return err
	}

	err = queryABrow.Scan(&uniqueID, &accountID, &balance, &currentAccountTxCount, &lastBalance, &hash)
	if err != nil {
		log.Println("[ERROR]", err)
		return errors.Errorf("[withdraw] account not exist", accountID)
	}

	log.Println(accountID, "'s balance:", balance, ", processed", currentAccountTxCount, "transactions")

	err = withdrawPreCheck(accountID, balance, value)
	if err != nil {
		log.Println("[ERROR]", err)
		return err
	}

	// try to withdraw
	tx, err := db.Begin()
	if err != nil {
		log.Println("[ERROR]", err)
		return err
	}

	defer tx.Rollback()

	if len(status) > 0 {
		// for reverse query
		err = queryInsertDailyTransaction(tx, accountID, "", currentAccountTxCount, "-", balance, value, balance-value, status[0], txID)
	} else {
		err = queryInsertDailyTransaction(tx, accountID, "", currentAccountTxCount, "-", balance, value, balance-value, DAILY_TRANSACTION_ON_PROCESS, txID)
	}

	if err != nil {
		return err
	}

	err = updateAccountBalance(tx, accountID, balance-value, currentAccountTxCount+1, balance)
	if err != nil {
		return err
	}

	err = tx.Commit()

	if err != nil {
		log.Println("[ERROR]", err)
		return err
	}

	log.Println("Successfully withdraw", value, "from account", accountID, ", new balace:", balance-value)
	return nil
}

func reverse(txID string) error {
	var uniqueID string
	var id int64
	var accountID string
	var traceID string
	var preID string
	var code string
	var prebalance float64
	var value float64
	var balance float64
	var status int
	var hash string
	var date int64

	queryDTrow, err := querySelectDailyTransactionByTxID(txID)
	if err != nil {
		log.Println("[DailyTransactionByTxID] tx not exist:", txID)
		return err
	}

	err = queryDTrow.Scan(&uniqueID, &id, &accountID, &traceID, &preID, &code, &prebalance, &value, &balance, &status, &hash, &date)
	if err != nil {
		log.Println("[ERROR]", err)
		return nil
	}

	log.Println("tx detail: ")

	if status == DAILY_TRANSACTION_ON_PROCESS {
		// start reverse process
		if code == "+" {
			err := withdraw(txID, accountID, value, DAILY_TRANSACTION_FOR_REVERSION)

			if err != nil {
				log.Println("[ERROR]", err)
				return errors.Errorf("[ERROR] failed to reverse transaction: ", txID)
			}
		} else {
			err := deposit(txID, accountID, value, DAILY_TRANSACTION_FOR_REVERSION)

			if err != nil {
				log.Println("[ERROR]", err)
				return errors.Errorf("[ERROR] failed to reverse transaction: ", txID)
			}
		}

		tx, err := db.Begin()
		if err != nil {
			log.Println("[ERROR]", err)
			return err
		}

		defer tx.Rollback()

		// update daily transaction status
		err = queryUpdateDailyTransactionByUniqueID(tx, uniqueID, DAILY_TRANSACTION_REVERSED)
		if err != nil {
			log.Println("[ERROR]", err)
			return err
		}

		err = tx.Commit()

		if err != nil {
			log.Println("[ERROR]", err)
			return err
		}

	} else {
		// ???
	}

	return nil
}

func reverseAccountBalance() {

}

func depositPreCheck(accountID string, balance float64, value float64) error {
	if balance < 0 {
		return errors.Errorf("Negative account balance: Account:", accountID, ", balance:", balance)
	}

	if value <= 0 {
		return errors.Errorf("Invalid value", value)
	}

	return nil
}

func withdrawPreCheck(accountID string, balance float64, value float64) error {
	err := depositPreCheck(accountID, balance, value)
	if err != nil {
		return err
	}

	if balance-value < 0 {
		return errors.Errorf("Account balance does not enough:", balance)
	}

	return nil
}

func createAccount(txID string, accountID string) error {
	tx, err := db.Begin()
	if err != nil {
		log.Panicln(err)
		return err
	}

	defer tx.Rollback()

	err = queryInsertAccount(tx, accountID, txID)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		log.Panicln(err)
		return err
	}

	// queryCreateAccountRow, err := querySelectDailyTransactionCreateAccount(accountID)
	// err = queryCreateAccountRow.Scan()
	// if err != nil {
	// 	log.Println("[ERROR]", err)
	// 	return err
	// }

	log.Println("Successfully create account", accountID)
	return nil
}
