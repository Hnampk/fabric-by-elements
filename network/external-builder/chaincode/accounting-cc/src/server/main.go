package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	pb "example.com/simple-chaincode/accounting_service"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

const (
	port = ":50052"
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
		return err
	}

	return nil
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

func deposit(txID string, accountID string, value float64) error {
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

	err = queryInsertDailyTransaction(tx, accountID, "", currentAccountTxCount, "+", balance, value, balance+value, 0, txID)
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

func withdraw(txID string, accountID string, value float64) error {
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

	err = queryInsertDailyTransaction(tx, accountID, "", currentAccountTxCount, "-", balance, value, balance-value, 0, txID)
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
